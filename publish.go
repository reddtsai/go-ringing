package ringing

import (
	"github.com/nsqio/go-nsq"
)

type publish struct {
	consumers *nsq.Consumer
}

func registerPublish(topic *Topic, host string, channel string) (*publish, error) {
	h := &msgHandler{
		topic: topic,
	}
	c, err := newNSQ(topic.name, host, channel, h)
	if err != nil {
		return nil, err
	}
	return &publish{
		consumers: c,
	}, nil
}

func (p *publish) close() {
	p.consumers.Stop()
}

type msgHandler struct {
	topic *Topic
}

func (h *msgHandler) HandleMessage(message *nsq.Message) error {
	if len(message.Body) > 0 {
		h.topic.sink <- message.Body
	}
	return nil
}

func newNSQ(name string, host string, channel string, handler *msgHandler) (*nsq.Consumer, error) {
	config := nsq.NewConfig()
	c, err := nsq.NewConsumer(name, channel, config)
	if err != nil {
		return nil, err
	}
	c.AddHandler(handler)
	err = c.ConnectToNSQLookupd(host)
	if err != nil {
		return nil, err
	}

	return c, nil
}
