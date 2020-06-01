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
