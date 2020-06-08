package ringing

import (
	"errors"

	"github.com/nsqio/go-nsq"
)

type publish struct {
	consumers *nsq.Consumer
}

func (p *publish) close() {
	p.consumers.Stop()
}

type msgHandler struct {
	topic *Topic
}

func (h *msgHandler) HandleMessage(message *nsq.Message) error {
	if len(message.Body) > 0 {
		select {
		case h.topic.sink <- message.Body:
		default:
			return errors.New("consumer channel was full")
		}
	}
	return nil
}

func newNSQ(addr string, topic string, channel string, handler *msgHandler) (*nsq.Consumer, error) {
	config := nsq.NewConfig()
	c, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return nil, err
	}
	c.AddHandler(handler)
	err = c.ConnectToNSQLookupd(addr)
	if err != nil {
		return nil, err
	}

	return c, nil
}
