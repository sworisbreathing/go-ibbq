# Go-BLE library for iBBQ Devices

This library builds on top of [go-ble](https://github.com/go-ble/ble) to read temperatures and battery level from
bluetooth thermometers such as the [Inkbird IBT-2X](http://www.ink-bird.com/products-bluetooth-thermometer-ibt2x.html).

# Usage

The below sections are taken from the [datalogger](./examples/datalogger) example app

## Configuration

```go
connectTimeout := 60*time.Second
batteryPollingInterval := 5*time.Minute
config, err := ibbq.NewConfiguration(connectTimeout, batteryPollingInterval)
```

## Notification Handlers / Callbacks

Data received from the device is sent asynchronously to registered callback functions.
There is also a callback fired when the device disconnects.

```go
logger := log.New("main")

temperatureReceived := func(temperatures []float64) {
	logger.Info("Received temperature data", "temperatures", temperatures)
}
batteryLevelReceived := func (batteryLevel int) {
	logger.Info("Received battery data", "batteryPct", strconv.Itoa(batteryLevel))
}

disconnectedHandler := func(cancel func(), done chan struct{}) func() {
	return func() {
		logger.Info("Disconnected")
		cancel()
		close(done)
	}
}
```

## Instantiating and Connecting

```go
ctx1, cancel := context.WithCancel(context.Background())
defer cancel()

ctx := ble.WithSigHandler(ctx1, cancel)

if bbq, err := ibbq.NewIbbq(ctx, config, disconnectedHandler(cancel, done), temperatureReceived, batteryLevelReceived); err != nil {
    return err
}

err = bbq.Connect()
```