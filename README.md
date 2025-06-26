# DS18B20 MQTT Client

## Setup

- `sudo raspi-config` Interfacing Options->1-Wire->Yes
- `echo "dtoverlay=w1-gpio" | sudo tee -a /boot/config.txt`
- `sudo reboot`

## Setup Driver

- `sudo modprobe w1-gpio`
- `sudo modprobe w1-therm`
