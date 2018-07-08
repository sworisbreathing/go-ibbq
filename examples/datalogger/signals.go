package main

import (
	"context"
	"os"
	"os/signal"
)

func registerInterruptHandler(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
	}()
}
