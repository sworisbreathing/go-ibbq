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
