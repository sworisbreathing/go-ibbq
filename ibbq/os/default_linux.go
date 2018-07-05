package os

import (
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
)

// DefaultDevice ...
func DefaultDevice() (d ble.Device, err error) {
	return linux.NewDevice()
}
