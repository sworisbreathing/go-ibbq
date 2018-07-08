# go-iBBQ Example WebSocket Server

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

In one shell:

```bash
$ LOGXI=main=INF ./ibbq-websocket
02:54:42.923948 INF main Starting websocket server port: 8080
```

In another shell:

```bash
$ curl --include \
    --no-buffer \
    --header "Connection: Upgrade" \
    --header "Upgrade: websocket" \
    --header "Host: localhost:8080" \
    --header "Origin: http://localhost:8080" \
    --header "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" \
    --header "Sec-WebSocket-Version: 13" \
    http://localhost:8080/ws
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: qGEgH3En71di5rrssAZTmtRTyFk=

?${"batteryLevel":93,"temps":[21,21]}
?${"batteryLevel":93,"temps":[21,21]}
?${"batteryLevel":93,"temps":[21,21]}
?${"batteryLevel":93,"temps":[21,21]}
?${"batteryLevel":93,"temps":[21,20]}
?${"batteryLevel":93,"temps":[21,21]}
...
```