# mfi-mqtt

Simple MQTT Publisher for Ubiquity's mFI Power Strips. No need for mFi controller - run it directly on the mFi power strip itself.

# Installation

Install [golang](https://golang.org/doc/install)

`env GOOS=linux GOARCH=mips go build`

Then `scp` binary to mfi device.

Install crontab on mFi

`* * * * * if [ -e /var/etc/persistent/mfi-mqtt ]; then /var/etc/persistent/mfi-mqtt -server tcp://mqtt_address:1883 -topic "home/mfi-unit"; fi`
