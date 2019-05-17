/*
   Copyright 2018 the original author or authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package main

import (
	"time"

	"github.com/sworisbreathing/go-ibbq/v2"
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
		TemperatureUnits:       "f",
	},
	Port: 8080,
}

// IbbqConfiguration is our ibbq configuration
type IbbqConfiguration struct {
	ConnectTimeout         int    `description:"Connect timeout (in seconds)"`
	BatteryPollingInterval int    `description:"Battery polling interval (in seconds)"`
	TemperatureUnits       string `description:"Temperature units ('c'/'celsius' or 'f'/'fahrenheit', case-insensitive)"`
}

func (c *IbbqConfiguration) asConfig() (ibbq.Configuration, error) {
	return ibbq.NewConfiguration(
		time.Duration(c.ConnectTimeout)*time.Second,
		time.Duration(c.BatteryPollingInterval)*time.Second,
	)
}
