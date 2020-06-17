# go-iBBQ MQTT Example

## Building

### Linux

```bash
$ GOOS=linux go build
```

### OS X

```bash
$ GOOS=darwin go build
```

## Usage

```bash
$ LOGXI=main=INF ./mqtt
10:53:05.583697 INF main instantiating mqtt config
10:53:05.587428 INF main instatiating mqtt client
10:53:05.590406 INF main MQTT broker status: connecting
10:53:05.607564 INF main MQTT broker status: connected
10:53:05.788197 INF main Connecting to device
10:53:05.793017 INF main Status updated status: Connecting
10:53:07.677380 INF main Connected to device
10:53:07.679722 INF main Status updated status: Connected
^C10:53:20.105120 INF main Status updated status: Disconnecting
10:53:22.721103 INF main Disconnected
10:53:22.724603 INF main Status updated status: Disconnected # <- ctrl-C was pressed (SIGINT)
10:53:22.745772 INF main Exiting
$
```
### MQTT broker output - subscribed to: home/iBBQ/#
```
2020-03-24 21:23:05	home/iBBQ/status	Connected
2020-03-24 21:23:07	home/iBBQ/battery	91
2020-03-24 21:23:08	home/iBBQ/temperature	[6552.6,19,6552.6,6552.6,6552.6,6552.6]
2020-03-24 21:23:10	home/iBBQ/temperature	[6552.6,19,6552.6,6552.6,6552.6,6552.6]
2020-03-24 21:23:12	home/iBBQ/temperature	[6552.6,19,6552.6,6552.6,6552.6,6552.6]
2020-03-24 21:23:14	home/iBBQ/temperature	[6552.6,19,6552.6,6552.6,6552.6,6552.6]
2020-03-24 21:23:16	home/iBBQ/temperature	[6552.6,19,6552.6,6552.6,6552.6,6552.6]
2020-03-24 21:23:18	home/iBBQ/temperature	[6552.6,19,6552.6,6552.6,6552.6,6552.6]
2020-03-24 21:23:20	home/iBBQ/temperature	[6552.6,19,6552.6,6552.6,6552.6,6552.6]
```