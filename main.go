package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Config struct {
	MQTTBroker   string
	MQTTPort     int
	MQTTUsername string
	MQTTPassword string
	MQTTTopic    string
	SensorID     string
	ReadInterval time.Duration
}

type TemperatureSensor struct {
	devicePath string
}

var PrevTemp = 0

func NewTemperatureSensor() (*TemperatureSensor, error) {
	// Find DS18B20 sensor
	devices, err := filepath.Glob("/sys/bus/w1/devices/28-*")
	if err != nil {
		return nil, fmt.Errorf("error searching for DS18B20 devices: %v", err)
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("no DS18B20 sensors found")
	}

	// Use the first device found
	devicePath := filepath.Join(devices[0], "w1_slave")

	log.Printf("Found DS18B20 sensor: %s", devices[0])

	return &TemperatureSensor{
		devicePath: devicePath,
	}, nil
}

func (ts *TemperatureSensor) ReadTemperature() (float64, error) {
	data, err := os.ReadFile(ts.devicePath)
	if err != nil {
		return 0, fmt.Errorf("error reading sensor data: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("invalid sensor data format")
	}

	// Check if the reading is valid
	if !strings.Contains(lines[0], "YES") {
		return 0, fmt.Errorf("sensor reading not ready")
	}

	// Extract temperature value
	tempIndex := strings.Index(lines[1], "t=")
	if tempIndex == -1 {
		return 0, fmt.Errorf("temperature value not found")
	}

	tempStr := lines[1][tempIndex+2:]
	tempRaw, err := strconv.Atoi(strings.TrimSpace(tempStr))
	if err != nil {
		return 0, fmt.Errorf("error parsing temperature: %v", err)
	}

	// Convert from millidegrees to degrees Celsius
	temperature := float64(tempRaw) / 1000.0

	return temperature, nil
}

type MQTTClient struct {
	client mqtt.Client
	topic  string
}

func NewMQTTClient(config Config) (*MQTTClient, error) {
	opts := mqtt.NewClientOptions()
	brokerURL := fmt.Sprintf("tcp://%s:%d", config.MQTTBroker, config.MQTTPort)
	opts.AddBroker(brokerURL)
	opts.SetClientID("ds18b20-sensor")
	opts.SetUsername(config.MQTTUsername)
	opts.SetPassword(config.MQTTPassword)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)

	// Connection lost handler
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	// On connect handler
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("Connected to MQTT broker")
	})

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}

	return &MQTTClient{
		client: client,
		topic:  config.MQTTTopic,
	}, nil
}

func (mc *MQTTClient) PublishTemperature(temperature float64) error {
	faren := (temperature * 1.8) + 32
	payload := fmt.Sprintf(`{"temperature": %.2f, "fahrenheit": %.2fF, "unit": "C", "timestamp": "%s"}`,
		temperature, faren, time.Now().Format(time.RFC3339))

	token := mc.client.Publish(mc.topic, 0, false, payload)
	token.Wait()

	if token.Error() != nil {
		return fmt.Errorf("failed to publish temperature: %v", token.Error())
	}

	log.Printf("Published temperature: %.2f°C, fahrenheit: %.2fF", temperature, faren)
	return nil
}

func (mc *MQTTClient) Disconnect() {
	mc.client.Disconnect(250)
}

func loadConfig() Config {
	config := Config{
		MQTTBroker:   getEnvOrDefault("MQTT_BROKER", "localhost"),
		MQTTPort:     getEnvIntOrDefault("MQTT_PORT", 1883),
		MQTTUsername: getEnvOrDefault("MQTT_USERNAME", ""),
		MQTTPassword: getEnvOrDefault("MQTT_PASSWORD", ""),
		MQTTTopic:    getEnvOrDefault("MQTT_TOPIC", "sensors/temperature"),
		ReadInterval: time.Duration(getEnvIntOrDefault("READ_INTERVAL_SECONDS", 30)) * time.Second,
	}

	return config
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func main() {
	log.Println("Starting DS18B20 Temperature Monitor")

	// Load configuration
	config := loadConfig()

	// Initialize temperature sensor
	sensor, err := NewTemperatureSensor()
	if err != nil {
		log.Fatalf("Failed to initialize temperature sensor: %v", err)
	}

	// Initialize MQTT client
	mqttClient, err := NewMQTTClient(config)
	if err != nil {
		log.Fatalf("Failed to initialize MQTT client: %v", err)
	}
	defer mqttClient.Disconnect()

	log.Printf("Reading temperature every %v", config.ReadInterval)
	log.Printf("Publishing to topic: %s", config.MQTTTopic)

	// Main loop
	ticker := time.NewTicker(config.ReadInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			temperature, err := sensor.ReadTemperature()
			if err != nil {
				log.Printf("Error reading temperature: %v", err)
				continue
			}
			faren := (temperature * 1.8) + 32
			log.Printf("Published temperature: %.2f°C, fahrenheit: %.2fF", temperature, faren)
			if int(temperature) == PrevTemp {
				continue // Skip publishing if the temperature hasn't changed
			} else {
				PrevTemp = int(temperature)
			}

			err = mqttClient.PublishTemperature(temperature)
			if err != nil {
				log.Printf("Error publishing temperature: %v", err)
			}
		}
	}
}
