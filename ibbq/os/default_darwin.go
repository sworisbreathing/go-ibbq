package os

import (
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/darwin"
)

// DefaultDevice ...
func DefaultDevice() (d ble.Device, err error) {
	return darwin.NewDevice()
}
