package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/staert"
	"github.com/gin-gonic/gin"
	"github.com/go-ble/ble"
	"github.com/mgutz/logxi/v1"
	"github.com/sworisbreathing/go-iBBQ/ibbq"
	"golang.org/x/sync/errgroup"
)

var logger = log.New("main")

func main() {
	command := newCommand(run)
	s := staert.NewStaert(command)
	toml := staert.NewTomlSource("ibbq-websocket", []string{"."})
	f := flaeg.New(command, os.Args[1:])
	if _, err := f.Parse(command); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	s.AddSource(toml)
	s.AddSource(f)
	if _, err := s.LoadConfig(); err != nil {
		logger.Fatal(err.Error())
	}

	if err := s.Run(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err.Error())
	}
	logger.Info("Exiting")
}

func run(config *Configuration) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	registerInterruptHandler(cancel)
	router := gin.Default()
	var g errgroup.Group
	temps := []float64{}
	tempsChannel := make(chan []float64)
	batteryLevel := 0
	batteryLevelChannel := make(chan []int)
	router.GET("/temperatureData", func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{
				"temperatures": temps,
			},
		)
	})
	router.GET("/batteryLevel", func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{
				"batteryLevel": batteryLevel,
			},
		)
	})
	router.GET("/allData", func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{
				"batteryLevel": batteryLevel,
				"temperatures": temps,
			},
		)
	})
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: router,
	}
	g.Go(func() error {
		for {
			select {
			case t := <-tempsChannel:
				if t == nil {
					logger.Warn("temps channel closed")
					return nil
				}
				temps = t
			case bl := <-batteryLevelChannel:
				if bl == nil {
					logger.Warn("battery level channel closed")
					return nil
				}
				batteryLevel = bl[0]
			}
		}
	})
	g.Go(func() error {
		if err := startIbbq(ctx, cancel, config.IbbqConfiguration, tempsChannel, batteryLevelChannel); err != nil {
			close(batteryLevelChannel)
			close(tempsChannel)
			return err
		}
		return nil
	})
	g.Go(func() error { return srv.ListenAndServe() })
	go func() {
		<-ctx.Done()
		logger.Info("shutting down server")
		sdc, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(sdc); err != nil {
			logger.Fatal("Shutdown failed")
		}
	}()
	return g.Wait()
}

func startIbbq(ctx1 context.Context, cancel func(), config IbbqConfiguration, tempsChannel chan []float64, batteryLevelChannel chan []int) error {
	ctx := ble.WithSigHandler(ctx1, cancel)
	var bbq ibbq.Ibbq
	done := make(chan struct{})
	var ibbqConfig ibbq.Configuration
	var err error
	if ibbqConfig, err = config.asConfig(); err != nil {
		return err
	}
	disconnectedHandler := func() {
		logger.Info("Disconnected")
		cancel()
		close(done)
		close(tempsChannel)
		close(batteryLevelChannel)
	}
	temperatureReceived := func(temps []float64) {
		tempsChannel <- temps
	}
	batteryLevelReceived := func(batteryLevel int) {
		batteryLevelChannel <- []int{batteryLevel}
	}
	if bbq, err = ibbq.NewIbbq(ctx, ibbqConfig, disconnectedHandler, temperatureReceived, batteryLevelReceived); err != nil {
		return err
	}
	if err = bbq.Connect(); err != nil {
		return err
	}
	<-ctx.Done()
	<-done
	logger.Info("all done")
	return nil
}
