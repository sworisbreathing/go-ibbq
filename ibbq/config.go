package ibbq

import (
	"errors"
	"time"
)

// Configuration configures our ibbq session
type Configuration struct {
	ConnectTimeout         time.Duration `description:"Connection timeout"`
	BatteryPollingInterval time.Duration `description:"Battery level polling interval"`
}

// DefaultConfiguration is a somewhat sane default.
var DefaultConfiguration = Configuration{
	ConnectTimeout:         60 * time.Second,
	BatteryPollingInterval: 5 * time.Minute,
}

// NewConfiguration creates a configuration
func NewConfiguration(connectTimeout time.Duration, batteryPollingInterval time.Duration) (Configuration, error) {
	if connectTimeout < 0 {
		return Configuration{}, errors.New("connect timeout must not be negative")
	}
	return Configuration{
		ConnectTimeout:         connectTimeout,
		BatteryPollingInterval: batteryPollingInterval,
	}, nil
}
