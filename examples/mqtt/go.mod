module github.com/sworisbreathing/go-ibbq/v2/examples/datalogger

go 1.12

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/go-ble/ble v0.0.0-20181002102605-e78417b510a3
	github.com/mgutz/logxi v0.0.0-20161027140823-aebf8a7d67ab
	github.com/sworisbreathing/go-ibbq/v2 v2.0.0
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8 // indirect
)

replace github.com/sworisbreathing/go-ibbq/v2 v2.0.0 => ../../
