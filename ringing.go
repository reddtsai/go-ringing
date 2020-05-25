package ringing

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type handleSessionFunc func(*Session)
type handleMessageFunc func(*Session, *signal)
type handleCloseFunc func(*Session, int, string) error
type handleErrorFunc func(*Session, error)

// Ringing implement websocket
type Ringing struct {
	Config               *Config
	Upgrader             *websocket.Upgrader
	topic                map[*Topic]bool
	session              map[*Session]bool
	connHandler          handleSessionFunc
	disconnHandler       handleSessionFunc
	receiveHandler       handleMessageFunc
	receiveBinaryHandler handleMessageFunc
	pongHandler          handleSessionFunc
	sendHandler          handleMessageFunc
	sendBinaryHandler    handleMessageFunc
	closeHandler         handleCloseFunc
	errHandler           handleErrorFunc
}

// New create a new Ringing instance
func New() *Ringing {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	c := newConfig()
	topic := make(map[*Topic]bool)
	for _, n := range c.Topic {
		t := newTopic(n)
		topic[t] = true
		go t.task()
	}

	return &Ringing{
		Config:   c,
		Upgrader: upgrader,
		topic:    topic,
	}
}

// HandleConn session connect event
func (r *Ringing) HandleConn(fn func(*Session)) {
	r.connHandler = fn
}

// HandleDisconn session disconnect event
func (r *Ringing) HandleDisconn(fn func(*Session)) {
	r.disconnHandler = fn
}

// HandleRequest upgrade http request to websocket connection
func (r *Ringing) HandleRequest(resp http.ResponseWriter, req *http.Request) error {
	// if r.publish.isClose()

	conn, err := r.Upgrader.Upgrade(resp, req, resp.Header())
	if err != nil {
		return err
	}
	session := &Session{
		req:     req,
		conn:    conn,
		sink:    make(chan *signal, r.Config.MsgBufSize),
		ringing: r,
		state:   true,
		rwmutex: &sync.RWMutex{},
	}
	r.session[session] = true

	defer func() {
		delete(r.session, session)
		session.close()
		session = nil
	}()

	return nil
}

// TODO
// close
// topic.close <- struct{}{}
