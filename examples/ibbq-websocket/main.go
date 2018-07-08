package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/staert"
	"github.com/gin-gonic/gin"
	"github.com/go-ble/ble"
	"github.com/gorilla/websocket"
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

	gin.SetMode(gin.ReleaseMode)
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
	router.GET("/ws", func(c *gin.Context) {
		if err := handleWebsocket(c.Writer, c.Request); err != nil {
			logger.Error(err.Error())
		}
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
					logger.Info("temps channel closed")
					return nil
				}
				temps = t
				go updateWebsockets(batteryLevel, temps)
			case bl := <-batteryLevelChannel:
				if bl == nil {
					logger.Info("battery level channel closed")
					return nil
				}
				batteryLevel = bl[0]
				go updateWebsockets(batteryLevel, temps)
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var connections = []*websocket.Conn{}

var connectionsMutex = &sync.RWMutex{}

func handleWebsocket(w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	connectionsMutex.Lock()
	connections = append(connections, conn)
	logger.Debug("Connection added", "connections", connections)
	connectionsMutex.Unlock()
	return nil
}

func connectionClosed(conn *websocket.Conn) {
	logger.Debug("Connection closed", "conn", conn)
	connectionsMutex.Lock()
	for i, c := range connections {
		if c == conn {
			copy(connections[i:], connections[i+1:])
			connections[len(connections)-1] = nil
			connections = connections[:len(connections)-1]
		}
	}
	logger.Debug("Connection removed", "connections", connections)
	connectionsMutex.Unlock()
}

func updateWebsockets(batteryLevel int, temps []float64) {
	connectionsMutex.RLock()
	for _, conn := range connections {
		go func(conn *websocket.Conn) {
			if err := conn.WriteJSON(
				gin.H{
					"batteryLevel": batteryLevel,
					"temps":        temps,
				},
			); err != nil {
				if isClosedError(err) {
					connectionClosed(conn)
				} else {
					logger.Error("Error writing to websocket", "err", err)
				}
			}
		}(conn)
	}
	connectionsMutex.RUnlock()
}

func isClosedError(err error) bool {
	logger.Debug(reflect.TypeOf(err).String())
	if websocket.IsUnexpectedCloseError(err) {
		return true
	}
	switch err.(type) {
	default:
		return false
	case syscall.Errno:
		if err.(syscall.Errno) == syscall.EPIPE {
			return true
		}
		return false
	case *net.OpError:
		err1 := err.(*net.OpError).Err
		return isClosedError(err1)
	case *os.SyscallError:
		err1 := err.(*os.SyscallError).Err
		return isClosedError(err1)
	}
}
