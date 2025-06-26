# DS18B20 MQTT Client

Running the DS18B20 Temp sensor on a Raspberry PI 4 B with Raspbian.

## Setup

- `sudo raspi-config` Interfacing Options->1-Wire->Yes
- `echo "dtoverlay=w1-gpio" | sudo tee -a /boot/config.txt`
- `sudo reboot`
- `cd ds18b20-mqtt-client && cp example.env .env`
- Update `.env` values with what you get from Dashboard->Add Device

## Driver Setup/Install

- `sudo modprobe w1-gpio`
- `sudo modprobe w1-therm`

## Features

- Automatic DS18B20 Detection: Automatically finds connected DS18B20 sensors
- MQTT Authentication: Supports username/password authentication
- JSON Payload: Sends structured JSON data with temperature, unit, and timestamp
- Error Handling: Robust error handling with automatic reconnection
- Configurable: Environment-based configuration
- System Service: Can run as a systemd service for automatic startup
- Logging: Comprehensive logging for monitoring and debugging
