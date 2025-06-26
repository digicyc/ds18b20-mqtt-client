#!/bin/sh

echo "Setting up the GPIO"

echo "dtoverlay=w1-gpio" | sudo tee -a /boot/config.txt

# Reboot?
sudo modprobe w1-gpio
sudo modprobe w1-therm
