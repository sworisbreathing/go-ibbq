package ibbq

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-ble/ble"
)

// Ibbq is an instance of the thermometer
type Ibbq struct {
	ctx        context.Context
	device     ble.Device
	done       chan struct{}
	cancelFunc func()
	client     ble.Client
	profile    *ble.Profile
}

// NewIbbq creates a new Ibbq
func NewIbbq(ctx context.Context, done chan struct{}, cancelFunc func()) (ibbq Ibbq, err error) {
	d, err := NewDevice("default")
	ble.SetDefaultDevice(d)
	return Ibbq{ctx, d, done, cancelFunc, nil, nil}, err
}

func (ibbq *Ibbq) disconnectHandler() {
	logger.Debug("waiting for disconnect")
	<-ibbq.client.Disconnected()
	logger.Info("Disconnected", "addr", ibbq.client.Addr().String())
	ibbq.client = nil
	ibbq.profile = nil
	close(ibbq.done)
	ibbq.cancelFunc()
}

// Connect connects to an ibbq
func (ibbq *Ibbq) Connect() error {
	var client ble.Client
	var err error
	timeoutContext, cancel := context.WithTimeout(ibbq.ctx, 15*time.Second)
	defer cancel()
	c := make(chan interface{})
	logger.Info("Connecting to device")
	go func() {
		if client, err = ble.Connect(timeoutContext, filter()); err == nil {
			logger.Info("Connected to device", "addr", client.Addr())
			ibbq.client = client
			logger.Debug("Setting up disconnect handler")
			go ibbq.disconnectHandler()
			err = ibbq.discoverProfile()
		}
		if err == nil {
			err = ibbq.login()
		}
		if err == nil {
			err = ibbq.subscribeToSettingResults()
		}
		if err == nil {
			err = ibbq.configureTemperatureCelsius()
		}
		if err == nil {
			err = ibbq.subscribeToRealTimeData()
		}
		if err == nil {
			err = ibbq.subscribeToHistoryData()
		}
		if err == nil {
			err = ibbq.enableRealTimeData()
		}
		if err == nil {
			err = ibbq.enableBatteryData()
		}
		c <- err
		close(c)
	}()
	select {
	case <-timeoutContext.Done():
		logger.Error("timeout while connecting")
		err = timeoutContext.Err()
	case err := <-c:
		if err != nil {
			logger.Error("Error received while connecting", "err", err)
		}
	}
	return err
}

func (ibbq *Ibbq) discoverProfile() error {
	var profile *ble.Profile
	var err error
	if profile, err = ibbq.client.DiscoverProfile(true); err == nil {
		ibbq.profile = profile
	}
	return err
}

func (ibbq *Ibbq) login() error {
	var err error
	var uuid ble.UUID
	if uuid, err = ble.Parse(AccountAndVerify); err == nil {
		logger.Debug("logging in to device", "addr", ibbq.client.Addr(), "uuid", uuid)
		characteristic := ble.NewCharacteristic(uuid)
		if c := ibbq.profile.FindCharacteristic(characteristic); c != nil {
			err = ibbq.client.WriteCharacteristic(c, Credentials, false)
			logger.Debug("credentials written")
		}
	}
	return err
}

func (ibbq *Ibbq) subscribeToRealTimeData() error {
	var err error
	var uuid ble.UUID
	logger.Info("Subscribing to real-time data")
	if uuid, err = ble.Parse(RealTimeData); err == nil {
		characteristic := ble.NewCharacteristic(uuid)
		if c := ibbq.profile.FindCharacteristic(characteristic); c != nil {
			err = ibbq.client.Subscribe(c, false, ibbq.realTimeDataReceived())
			if err == nil {
				logger.Info("Subscribed to real-time data")
			} else {
				logger.Error("Error subscribing to real-time data", "err", err)
			}
		} else {
			err = errors.New("can't find characteristic for real-time data")
		}
	}
	return err
}

func (ibbq *Ibbq) realTimeDataReceived() ble.NotificationHandler {
	return func(data []byte) {
		logger.Debug("received real-time data", hex.EncodeToString(data))
		probe1 := float64(binary.LittleEndian.Uint16(data[0:2])) / 10
		probe2 := float64(binary.LittleEndian.Uint16(data[2:4])) / 10
		logger.Info("Temperature received", "probe1", probe1, "probe2", probe2)
	}
}

func (ibbq *Ibbq) subscribeToHistoryData() error {
	var err error
	var uuid ble.UUID
	logger.Info("Subscribing to history data")
	if uuid, err = ble.Parse(HistoryData); err == nil {
		characteristic := ble.NewCharacteristic(uuid)
		if c := ibbq.profile.FindCharacteristic(characteristic); c != nil {
			err = ibbq.client.Subscribe(c, false, ibbq.historyDataReceived())
			if err == nil {
				logger.Info("Subscribed to history data")
			} else {
				logger.Error("Error subscribing to history data", "err", err)
			}
		} else {
			err = errors.New("Can't find characteristic for history data")
		}
	}
	return err
}

func (ibbq *Ibbq) historyDataReceived() ble.NotificationHandler {
	return func(data []byte) {
		logger.Debug("received history data", hex.EncodeToString(data))
	}
}

func (ibbq *Ibbq) subscribeToSettingResults() error {
	var err error
	var uuid ble.UUID
	logger.Info("Subscribing to setting results")
	if uuid, err = ble.Parse(SettingResult); err == nil {
		characteristic := ble.NewCharacteristic(uuid)
		if c := ibbq.profile.FindCharacteristic(characteristic); c != nil {
			err = ibbq.client.Subscribe(c, false, ibbq.settingResultReceived())
			if err == nil {
				logger.Info("Subscribed to setting results")
			} else {
				logger.Error("Error subscribing to setting results", "err", err)
			}
		} else {
			err = errors.New("Can't find characteristic for setting results")
		}
	}
	return err
}

func (ibbq *Ibbq) settingResultReceived() ble.NotificationHandler {
	return func(data []byte) {
		logger.Debug("Received setting result", "data", hex.EncodeToString(data))
		switch data[0] {
		case 0x24:
			// battery
			currentVoltage := int(binary.LittleEndian.Uint16(data[1:3]))
			maxVoltage := int(binary.LittleEndian.Uint16(data[3:5]))
			if maxVoltage == 0 {
				maxVoltage = 6550
			}
			batteryPct := 100 * currentVoltage / maxVoltage
			logger.Info("Battery data", "currentVoltage", strconv.Itoa(currentVoltage), "maxVoltage", strconv.Itoa(maxVoltage), "batteryPct", batteryPct)
		}
	}
}

func (ibbq *Ibbq) enableRealTimeData() error {
	logger.Info("Enabling real-time data sending")
	err := ibbq.writeSetting(realTimeDataEnable)
	if err == nil {
		logger.Info("Enabled real-time data sending")
	}
	return err
}

func (ibbq *Ibbq) enableBatteryData() error {
	logger.Info("Enabling battery data sending")
	var err error
	if err = ibbq.writeSetting(batteryLevel); err == nil {
		ticker := time.NewTicker(60 * time.Second)
		go func() {
			for {
				select {
				case <-ticker.C:
					logger.Debug("Requesting battery data")
					err := ibbq.writeSetting(batteryLevel)
					if err != nil {
						logger.Error("Unable to request battery level", "err", err)
						ticker.Stop()
						return
					}
				case <-ibbq.done:
					ticker.Stop()
					return
				}
			}
		}()
	}
	return err
}

func (ibbq *Ibbq) configureTemperatureCelsius() error {
	logger.Info("Configuring temperature for Celsius")
	err := ibbq.writeSetting(unitsCelsius)
	if err == nil {
		logger.Info("Configured temperature for Celsius")
	}
	return err
}

func (ibbq *Ibbq) configureTemperatureFahrenheit() error {
	logger.Info("Configuring temperature for Fahrenheit")
	err := ibbq.writeSetting(unitsFahrenheit)
	if err == nil {
		logger.Info("Configured temperature for Fahrenheit")
	}
	return err
}

func (ibbq *Ibbq) writeSetting(settingValue []byte) error {
	var err error
	var uuid ble.UUID
	if uuid, err = ble.Parse(SettingData); err == nil {
		characteristic := ble.NewCharacteristic(uuid)
		if c := ibbq.profile.FindCharacteristic(characteristic); c != nil {
			err = ibbq.client.WriteCharacteristic(c, settingValue, false)
		} else {
			err = errors.New("Can't find characteristic for settings data")
		}
	}
	return err
}

// Disconnect disconnects from an ibbq
func (ibbq *Ibbq) Disconnect() error {
	var err error
	if ibbq.client == nil {
		err = errors.New("Not connected")
	} else {
		logger.Info("Disconnecting")
		err = ibbq.client.CancelConnection()
	}
	return err
}

func filter() ble.AdvFilter {
	return func(a ble.Advertisement) bool {
		return strings.ToLower(a.LocalName()) == strings.ToLower(DeviceName) && a.Connectable()
	}
}

func advHandler() ble.AdvHandler {
	return func(a ble.Advertisement) {
		logger.Debug("Found advertisement",
			"address", a.Addr(),
			"connectable", a.Connectable(),
			"rssi", a.RSSI(),
			"name", a.LocalName(),
			"svcs", a.Services(),
			"manufacturerData", a.ManufacturerData())
	}
}
