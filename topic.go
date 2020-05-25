package ringing

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
)

// Topic pub/pub controller
type Topic struct {
	name        string
	sessions    map[*Session]bool
	subscribe   chan *Session
	unsubscribe chan *Session
	publish     chan *signal
	close       chan struct{}
	rwmutex     *sync.RWMutex
}

func newTopic(name string) *Topic {
	// TODO
	config := nsq.NewConfig()
	_, err := nsq.NewConsumer(name, "ch1", config)
	if err != nil {

	}

	return &Topic{
		name:        name,
		sessions:    make(map[*Session]bool),
		subscribe:   make(chan *Session),
		unsubscribe: make(chan *Session),
		publish:     make(chan *signal),
		close:       make(chan struct{}),
		rwmutex:     &sync.RWMutex{},
	}
}

func (t *Topic) task() {
Loop:
	for {
		select {
		case s := <-t.subscribe:
			t.rwmutex.Lock()
			t.sessions[s] = true
			t.rwmutex.Unlock()
		case s := <-t.unsubscribe:
			if _, ok := t.sessions[s]; ok {
				t.rwmutex.Lock()
				delete(t.sessions, s)
				t.rwmutex.Unlock()
			}
		case signal := <-t.publish:
			t.rwmutex.RLock()
			for s := range t.sessions {
				s.push(signal)
			}
			t.rwmutex.RUnlock()
		case <-t.close:
			t.rwmutex.Lock()
			for s := range t.sessions {
				s.push(&signal{msgType: websocket.CloseMessage, msg: []byte{}})
				delete(t.sessions, s)
			}
			t.rwmutex.Unlock()
			break Loop
		}
	}
}

func (t *Topic) len() int {
	t.rwmutex.RLock()
	defer t.rwmutex.RUnlock()

	return len(t.sessions)
}

// TODO
// close
