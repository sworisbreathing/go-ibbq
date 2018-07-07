package main

import (
	"context"
	"os"
	"strconv"

	"github.com/go-ble/ble"
	"github.com/mgutz/logxi/v1"
	"github.com/sworisbreathing/go-iBBQ/ibbq"
)

var logger = log.New("main")

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
	temperatureReceived := func(temperatures []float64) {
		logger.Info("Received temperature data", "temperatures", temperatures)
	}
	batteryLevelReceived := func(batteryLevel int) {
		logger.Info("Received battery data", "batteryPct", strconv.Itoa(batteryLevel))
	}
	if bbq, err = ibbq.NewIbbq(ctx, cancel, temperatureReceived, batteryLevelReceived); err != nil {
		logger.Fatal("Error creating iBBQ", "err", err)
		os.Exit(-1)
	}
	logger.Debug("instantiated ibbq struct")
	logger.Debug("connecting to device")
	if err = bbq.Connect(); err != nil {
		logger.Fatal("Error connecting to device", "err", err)
		os.Exit(-1)
	}
	logger.Debug("Connected to device")
	<-ctx.Done()
	if err = bbq.Disconnect(); err != nil {
		logger.Fatal("Error disconnecting from device", "err", err)
		os.Exit(-1)
	}
	logger.Debug("waiting for device to send disconnect signal")
	<-bbq.Disconnected()
}
