/*
   Copyright 2018 the original author or authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package main

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-ble/ble"
	log "github.com/mgutz/logxi/v1"
	"github.com/sworisbreathing/go-ibbq/v2"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var logger = log.New("main")
var mqClient mqtt.Client

const topic = "home/iBBQ"
const broker = "tcp://192.168.1.196:1883"
const clientID = "go_ibbq"

var mqttPublishHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	logger.Info("TOPIC: ", msg.Topic())
	logger.Info("MSG: ", msg.Payload())
}

func temperatureReceived(temperatures []float64) {
	logger.Info("Received temperature data", "temperatures", temperatures)
	m, err := json.Marshal(temperatures)
	if err != nil {
		logger.Fatal("Can't encode to JSON", "err", err)
	}
	token := mqClient.Publish(topic+"/temperature", 0, false, m)
	token.Wait()
}
func batteryLevelReceived(batteryLevel int) {
	logger.Info("Received battery data", "batteryPct", strconv.Itoa(batteryLevel))
	m, err := json.Marshal(batteryLevel)
	if err != nil {
		logger.Fatal("Can't encode to JSON", "err", err)
	}
	token := mqClient.Publish(topic+"/battery", 0, false, m)
	token.Wait()
}
func statusUpdated(status ibbq.Status) {
	logger.Info("Status updated", "status", status)
}

func disconnectedHandler(cancel func(), done chan struct{}) func() {
	return func() {
		logger.Info("Disconnected")
		cancel()
		close(done)
	}
}

func main() {
	var err error
	ctx1, cancel := context.WithCancel(context.Background())
	defer cancel()
	registerInterruptHandler(cancel)

	logger.Info("instantiating mqtt config")
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientID)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(mqttPublishHandler)
	opts.SetPingTimeout(1 * time.Second)

	logger.Info("instatiating mqtt client")
	mqClient = mqtt.NewClient(opts)

	logger.Info("MQTT broker", "status", "connecting")
	if token := mqClient.Connect(); token.Wait() && token.Error() != nil {
		logger.Fatal("Error connecting to mqtt broker", "err", token.Error())
	}
	logger.Info("MQTT broker", "status", "connected")
	token := mqClient.Publish(topic+"/status", 0, false, "Connected")
	token.Wait()

	logger.Debug("initializing context")

	ctx := ble.WithSigHandler(ctx1, cancel)
	logger.Debug("context initialized")
	var bbq ibbq.Ibbq
	logger.Debug("instantiating ibbq struct")
	done := make(chan struct{})
	var config ibbq.Configuration
	if config, err = ibbq.NewConfiguration(60*time.Second, 5*time.Minute); err != nil {
		logger.Fatal("Error creating configuration", "err", err)
	}
	if bbq, err = ibbq.NewIbbq(ctx, config, disconnectedHandler(cancel, done), temperatureReceived, batteryLevelReceived, statusUpdated); err != nil {
		logger.Fatal("Error creating iBBQ", "err", err)
	}
	logger.Debug("instantiated ibbq struct")
	logger.Info("Connecting to device")
	if err = bbq.Connect(); err != nil {
		logger.Fatal("Error connecting to device", "err", err)
	}
	logger.Info("Connected to device")
	<-ctx.Done()
	<-done
	mqClient.Disconnect(250)
	logger.Info("Exiting")
}
