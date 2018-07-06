package ibbq

import (
	"context"
	"errors"
	"fmt"
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

func (ibbq *Ibbq) disconnectHandler(done chan struct{}) func() {
	return func() {
		fmt.Println("waiting for disconnect")
		<-ibbq.client.Disconnected()
		fmt.Printf("\n%s disconnected\n", ibbq.client.Addr().String())
		ibbq.client = nil
		ibbq.profile = nil
		close(done)
	}
}

// Connect connects to an ibbq
func (ibbq *Ibbq) Connect(done chan struct{}) error {
	var client ble.Client
	var err error
	timeoutContext, cancel := context.WithTimeout(ibbq.ctx, 15*time.Second)
	defer cancel()
	c := make(chan struct{})
	go func() {
		if client, err = ble.Connect(timeoutContext, filter()); err == nil {
			fmt.Print("Connected to device: ")
			fmt.Println(client.Addr())
			ibbq.client = client
			fmt.Println("Setting up disconnect handler")
			go ibbq.disconnectHandler(done)
			err = ibbq.discoverProfile()
		}
		if err == nil {
			err = ibbq.login()
		}
		if err == nil {
			err = ibbq.subscribeToRealTimeData()
		}
		close(c)
	}()
	select {
	case <-timeoutContext.Done():
		fmt.Println("done connecting")
		err = timeoutContext.Err()
	case err := <-c:
		fmt.Println(err)
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
		fmt.Print("logging in to ")
		fmt.Println(uuid)
		characteristic := ble.NewCharacteristic(uuid)
		if c := ibbq.profile.FindCharacteristic(characteristic); c != nil {
			err = ibbq.client.WriteCharacteristic(c, Credentials, false)
			fmt.Println("credentials written")
		}
	}
	return err
}

func (ibbq *Ibbq) subscribeToRealTimeData() error {
	var err error
	var uuid ble.UUID
	fmt.Println("Subscribing to real-time data")
	if uuid, err = ble.Parse(RealTimeData); err == nil {
		characteristic := ble.NewCharacteristic(uuid)
		if c := ibbq.profile.FindCharacteristic(characteristic); c != nil {
			err = ibbq.client.Subscribe(c, false, ibbq.realTimeDataReceived())
			if err == nil {
				fmt.Println("subscribed")
			} else {
				fmt.Print("error subscribing: ")
				fmt.Println(err)
			}
		} else {
			err = errors.New("can't find characteristic for real-time data")
		}
	}
	return err
}

func (ibbq *Ibbq) realTimeDataReceived() ble.NotificationHandler {
	return func(data []byte) {
		fmt.Print("received real-time data ")
		fmt.Println(data)
	}
}

// Disconnect disconnects from an ibbq
func (ibbq *Ibbq) Disconnect() error {
	var err error
	if ibbq.client == nil {
		err = errors.New("Not connected")
	} else {
		fmt.Println("Disconnecting")
		err = ibbq.client.CancelConnection()
		fmt.Println("Disconnected")
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
			fmt.Printf("[%s] C %3d:", a.Addr(), a.RSSI())
		} else {
			fmt.Printf("[%s] N %3d:", a.Addr(), a.RSSI())
		}
		comma := ""
		if len(a.LocalName()) > 0 {
			fmt.Printf(" Name: %s", a.LocalName())
			comma = ","
		}
		if len(a.Services()) > 0 {
			fmt.Printf("%s Svcs: %v", comma, a.Services())
			comma = ","
		}
		if len(a.ManufacturerData()) > 0 {
			fmt.Printf("%s MD: %X", comma, a.ManufacturerData())
		}
		fmt.Printf("\n")
	}
}
