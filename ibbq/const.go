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
