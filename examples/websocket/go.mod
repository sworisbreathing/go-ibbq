module github.com/sworisbreathing/go-ibbq/v2/examples/websocket

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/abronan/valkeyrie v0.0.0-20190503213338-20861cd6729e // indirect
	github.com/containous/flaeg v1.4.1
	github.com/containous/staert v3.1.2+incompatible
	github.com/gin-gonic/gin v1.4.0
	github.com/go-ble/ble v0.0.0-20181002102605-e78417b510a3
	github.com/gorilla/websocket v1.4.0
	github.com/mgutz/logxi v0.0.0-20161027140823-aebf8a7d67ab
	github.com/ogier/pflag v0.0.1 // indirect
	github.com/sworisbreathing/go-ibbq/v2 v2.0.0
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
)

replace github.com/sworisbreathing/go-ibbq/v2 v2.0.0 => ../../
