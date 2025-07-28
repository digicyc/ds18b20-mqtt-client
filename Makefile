BINARY_NAME=ds18b20-mqtt
SERVICE_NAME=ds18b20-mqtt.service

build:
	go mod tidy
	go build -o $(BINARY_NAME) .

install: build
	sudo cp $(BINARY_NAME) /opt/ds18b20-mqtt/
	sudo cp $(BINARY_NAME) /usr/local/bin/
	sudo cp $(SERVICE_NAME) /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable $(SERVICE_NAME)

start:
	sudo systemctl start $(SERVICE_NAME)

stop:
	sudo systemctl stop $(SERVICE_NAME)

status:
	sudo systemctl status $(SERVICE_NAME)

logs:
	sudo journalctl -u $(SERVICE_NAME) -f

clean:
	rm -f $(BINARY_NAME)

run:
	go run .
