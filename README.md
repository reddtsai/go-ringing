# goRinging

[![Actions Status](https://github.com/reddtsai/goRinging/workflows/Go/badge.svg)](https://github.com/reddtsai/goRinging/actions)

goRinging is Go package that for notify message. It's based on [Gorilla WebSocket](https://github.com/gorilla/websocket) and [nsqio go-nsq](https://github.com/nsqio/go-nsq).

![Alt text](/examples/Ringing.png)

## Features

* [x] Ping/Pong
* [x] Publish/Subscribe
* [x] Load Balance Cluster

## Installation

``` bash
go get github.com/reddtsai/goRinging
```

## Example

Using Gin:
``` go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/reddtsai/go-ringing"
)

func main() {
	g := gin.Default()
	c := ringing.NewConfig("localhost:4161")
	r, err := ringing.New(c)
	if err != nil {
		return
	}
	g.GET("/ws", func(c *gin.Context) {
		r.HandleRequest(c.Writer, c.Request)
	})

	g.Run(":5000")
}

```