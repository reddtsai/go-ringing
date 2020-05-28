package ringing

import (
	"sync"
)

// Topic pub/pub controller
type Topic struct {
	name        string
	sessions    map[*Session]bool
	publish     *publish
	subscribe   chan *Session
	unsubscribe chan *Session
	sink        chan []byte
	stop        chan struct{}
	rwmutex     *sync.RWMutex
}

func newTopic(name string) *Topic {
	return &Topic{
		name:        name,
		sessions:    make(map[*Session]bool),
		subscribe:   make(chan *Session),
		unsubscribe: make(chan *Session),
		sink:        make(chan []byte),
		stop:        make(chan struct{}),
		rwmutex:     &sync.RWMutex{},
	}
}

func (t *Topic) controller() {
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
		case signal := <-t.sink:
			t.rwmutex.RLock()
			for s := range t.sessions {
				s.push(signal)
			}
			t.rwmutex.RUnlock()
		case <-t.stop:
			break Loop
		}
	}
}

func (t *Topic) len() int {
	t.rwmutex.RLock()
	defer t.rwmutex.RUnlock()
	return len(t.sessions)
}

func (t *Topic) close() {
	t.stop <- struct{}{}
	t.rwmutex.Lock()
	for s := range t.sessions {
		delete(t.sessions, s)
	}
	close(t.subscribe)
	close(t.unsubscribe)
	close(t.sink)
	close(t.stop)
	if t.publish != nil {
		t.publish.close()
	}
	t.rwmutex.Unlock()
}
