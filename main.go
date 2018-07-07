package main

import (
	"context"
	"os"

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
	done := make(chan struct{})
	if bbq, err = ibbq.NewIbbq(ctx, done, cancel); err != nil {
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
	<-done
}
