package main

import (
	"time"

	"github.com/sworisbreathing/go-ibbq/ibbq"
)

// Configuration is our app configuration
type Configuration struct {
	IbbqConfiguration
	Port int `short:"p" description:"Web server port number"`
}

// DefaultConfiguration is a somewhat sane set of default values.
var DefaultConfiguration = &Configuration{
	IbbqConfiguration: IbbqConfiguration{
		ConnectTimeout:         int(ibbq.DefaultConfiguration.ConnectTimeout / time.Second),
		BatteryPollingInterval: int(ibbq.DefaultConfiguration.BatteryPollingInterval / time.Second),
	},
	Port: 8080,
}

// IbbqConfiguration is our ibbq configuration
type IbbqConfiguration struct {
	ConnectTimeout         int `description:"Connect timeout (in seconds)"`
	BatteryPollingInterval int `description:"Battery polling interval (in seconds)"`
}

func (c *IbbqConfiguration) asConfig() (ibbq.Configuration, error) {
	return ibbq.NewConfiguration(
		time.Duration(c.ConnectTimeout)*time.Second,
		time.Duration(c.BatteryPollingInterval)*time.Second,
	)
}
