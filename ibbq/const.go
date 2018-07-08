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

import "github.com/mgutz/logxi/v1"

// SettingResult NOTIFY
const SettingResult = "fff1"

// AccountAndVerify WRITE
const AccountAndVerify = "fff2"

// HistoryData NOTIFY
const HistoryData = "fff3"

// RealTimeData NOTIFY
const RealTimeData = "fff4"

// SettingData WRITE
const SettingData = "fff5"

// DeviceName is the name we look for when we scan.
const DeviceName = "iBBQ"

var (
	// Credentials stores our login credentials for the thermometer.
	Credentials = []byte{0x21, 0x07, 0x06,
		0x05, 0x04, 0x03, 0x02, 0x01, 0xb8, 0x22,
		0x00, 0x00, 0x00, 0x00, 0x00}

	realTimeDataEnable = []byte{0x0B, 0x01, 0x00, 0x00, 0x00, 0x00}

	unitsFahrenheit = []byte{0x02, 0x01, 0x00, 0x00, 0x00, 0x00}

	unitsCelsius = []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00}

	batteryLevel = []byte{0x08, 0x24, 0x00, 0x00, 0x00, 0x00}

	logger = log.New("ibbq")
)
