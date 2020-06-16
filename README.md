# goRinging

[![Actions Status](https://github.com/reddtsai/goRinging/workflows/Go/badge.svg)](https://github.com/reddtsai/goRinging/actions)

goRinging is Go package that for notify message. It's based on [Gorilla WebSocket](https://github.com/gorilla/websocket) and [nsqio go-nsq](https://github.com/nsqio/go-nsq).

![Alt text](https://github.com/reddtsai/static/blob/master/Ringing.png)

## Features

* [x] Ping/Pong
* [x] Publish/Subscribe
* [x] Load Balance Cluster

## Installation

``` bash
go get github.com/reddtsai/goRinging
```

## Example

[Using Gin](https://github.com/reddtsai/go-ringing/blob/master/examples/main.go)

test
``` bash
go test -v
```