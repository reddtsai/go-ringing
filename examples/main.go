package main

import (
	"github.com/gin-gonic/gin"
	"github.com/reddtsai/go-ringing"
)

func main() {
	g := gin.Default()
	n := &ringing.NSQSetting{
		NSQAddr:      "localhost:4161",
		Topics:       []string{"all"},
		TopicChannel: "ch1",
	}
	c := ringing.NewConfig(n)
	r, err := ringing.New(c)
	if err != nil {
		return
	}
	g.GET("/ws", func(c *gin.Context) {
		r.HandleRequest(c.Writer, c.Request)
	})

	g.Run(":5000")
}
