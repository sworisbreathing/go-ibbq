package main

import (
	"context"
	"strconv"

	"github.com/go-ble/ble"
	"github.com/mgutz/logxi/v1"
	"github.com/sworisbreathing/go-iBBQ/ibbq"
)

var logger = log.New("main")

func temperatureReceived(temperatures []float64) {
	logger.Info("Received temperature data", "temperatures", temperatures)
}
func batteryLevelReceived(batteryLevel int) {
	logger.Info("Received battery data", "batteryPct", strconv.Itoa(batteryLevel))
}

func main() {
	var err error
	logger.Debug("initializing context")
	ctx1, cancel := context.WithCancel(context.Background())
	defer cancel()
	registerInterruptHandler(cancel)
	ctx := ble.WithSigHandler(ctx1, cancel)
	logger.Debug("context initialized")
	var bbq ibbq.Ibbq
	logger.Debug("instantiating ibbq struct")
	if bbq, err = ibbq.NewIbbq(ctx, temperatureReceived, batteryLevelReceived); err != nil {
		logger.Fatal("Error creating iBBQ", "err", err)
	}
	logger.Debug("instantiated ibbq struct")
	logger.Info("Connecting to device")
	if err = bbq.Connect(); err != nil {
		logger.Fatal("Error connecting to device", "err", err)
	}
	logger.Info("Connected to device")
	<-ctx.Done()
	if err = bbq.Disconnect(); err != nil {
		logger.Fatal("Error disconnecting from device", "err", err)
	}
	logger.Debug("waiting for device to send disconnect signal")
	<-bbq.Disconnected()
	cancel()
}
