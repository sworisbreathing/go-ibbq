package ibbq

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/go-ble/ble"
)

// Ibbq is an instance of the thermometer
type Ibbq struct {
	ctx     context.Context
	device  ble.Device
	client  ble.Client
	profile *ble.Profile
}

// NewIbbq creates a new Ibbq
func NewIbbq(ctx context.Context) (ibbq Ibbq, err error) {
	d, err := NewDevice("default")
	ble.SetDefaultDevice(d)
	return Ibbq{ctx, d, nil, nil}, err
}

func (ibbq *Ibbq) disconnectHandler(done chan struct{}, cancelFunc func()) func() {
	return func() {
		logger.Debug("waiting for disconnect")
		<-ibbq.client.Disconnected()
		logger.Debug(ibbq.client.Addr().String(), "disconnected")
		ibbq.client = nil
		ibbq.profile = nil
		close(done)
		cancelFunc()
	}
}

// Connect connects to an ibbq
func (ibbq *Ibbq) Connect(done chan struct{}, cancelFunc func()) error {
	var client ble.Client
	var err error
	timeoutContext, cancel := context.WithTimeout(ibbq.ctx, 15*time.Second)
	defer cancel()
	c := make(chan interface{})
	go func() {
		if client, err = ble.Connect(timeoutContext, filter()); err == nil {
			logger.Info("Connected to device:", client.Addr())
			ibbq.client = client
			logger.Info("Setting up disconnect handler")
			go ibbq.disconnectHandler(done, cancelFunc)()
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
		c <- err
		close(c)
	}()
	select {
	case <-timeoutContext.Done():
		logger.Error("timeout while connecting")
		err = timeoutContext.Err()
	case err := <-c:
		if err != nil {
			logger.Error("Error received while connecting:", err)
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
		logger.Info("logging in to", uuid)
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
				logger.Info("subscribed")
			} else {
				logger.Error("error subscribing:", err)
			}
		} else {
			err = errors.New("can't find characteristic for real-time data")
		}
	}
	return err
}

func (ibbq *Ibbq) realTimeDataReceived() ble.NotificationHandler {
	return func(data []byte) {
		logger.Info("received real-time data", hex.EncodeToString(data))
		switch data[0] {
		default:
			// temperature
			probe1 := binary.LittleEndian.Uint16(data[0:2]) / 10.0
			probe2 := binary.LittleEndian.Uint16(data[2:4]) / 10.0
			logger.Info("probe 1 temp", probe1)
			logger.Info("probe 2 temp", probe2)
		}
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
				logger.Info("subscribed")
			} else {
				logger.Error("error subscribing:", err)
			}
		} else {
			err = errors.New("can't find characteristic for history data")
		}
	}
	return err
}

func (ibbq *Ibbq) historyDataReceived() ble.NotificationHandler {
	return func(data []byte) {
		logger.Info("received history data", hex.EncodeToString(data))
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
				logger.Info("subscribed")
			} else {
				logger.Error("error subscribing:", err)
			}
		} else {
			err = errors.New("can't find characteristic for setting results")
		}
	}
	return err
}

func (ibbq *Ibbq) settingResultReceived() ble.NotificationHandler {
	return func(data []byte) {
		logger.Info("received setting result:", hex.EncodeToString(data))
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
		logger.Info("Disconnected")
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
		if a.Connectable() {
			logger.Debug("[", a.Addr(), "] C", a.RSSI())
		} else {
			logger.Debug("[", a.Addr(), "] N ", a.RSSI())
		}
		if len(a.LocalName()) > 0 {
			logger.Debug(" Name:", a.LocalName())
		}
		if len(a.Services()) > 0 {
			logger.Debug("Svcs:", a.Services())
		}
		if len(a.ManufacturerData()) > 0 {
			logger.Debug("MD:", a.ManufacturerData())
		}
	}
}
