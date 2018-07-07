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
	logger.Info("initializing context")
	ctx1, cancel := context.WithCancel(context.Background())
	defer cancel()
	registerInterruptHandler(cancel)
	ctx := ble.WithSigHandler(ctx1, cancel)
	logger.Info("context initialized")
	var bbq ibbq.Ibbq
	logger.Info("instantiating ibbq struct")
	if bbq, err = ibbq.NewIbbq(ctx); err != nil {
		logger.Fatal("fatal", err)
		os.Exit(-1)
	}
	logger.Info("instantiated ibbq struct")
	logger.Info("connecting to device")
	done := make(chan struct{})
	if err = bbq.Connect(done, cancel); err != nil {
		logger.Fatal("fatal", err)
		os.Exit(-1)
	}
	logger.Info("Connected to device")
	<-ctx.Done()
	if err = bbq.Disconnect(); err != nil {
		logger.Fatal("fatal", err)
		os.Exit(-1)
	}
	logger.Info("waiting for device to send disconnect signal")
	<-done
}
