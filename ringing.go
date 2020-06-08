package ringing

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Ringing implement websocket
type Ringing struct {
	Config        *Config
	Upgrader      *websocket.Upgrader
	topics        map[string]*Topic
	sessions      map[*Session]bool
	handleConn    func(*Session)
	handleDisconn func(*Session)
	handlePong    func(*Session)
	handleReceive func(*Session, []byte)
	handleSend    func(*Session, []byte)
	handleErr     func(*Session, error)
}

// New create a new Ringing instance
func New(config *Config) (*Ringing, error) {
	r := &Ringing{
		sessions: make(map[*Session]bool),
		Config:   config,
	}

	// ws
	r.Upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	// topic
	r.topics = make(map[string]*Topic)
	for _, n := range r.Config.NSQSetting.Topics {
		t := newTopic(n)
		p, err := t.registerPublish(r.Config.NSQSetting.NSQAddr, r.Config.NSQSetting.TopicChannel)
		if err != nil {
			r.close()
			return nil, err
		}
		t.publish = p
		go t.controller()
		r.topics[n] = t
	}

	// handler
	r.SetConnHandler(nil)
	r.SetDisconnHandler(nil)
	r.SetReceiveHandler(subscribe)
	r.SetPongHandler(nil)
	r.SetErrorHandler(nil)
	return r, nil
}

// SetConnHandler session connect handle function
func (r *Ringing) SetConnHandler(fn func(*Session)) {
	if fn == nil {
		fn = func(*Session) {}
	}
	r.handleConn = fn
}

// SetDisconnHandler session disconnect handle function
func (r *Ringing) SetDisconnHandler(fn func(*Session)) {
	if fn == nil {
		fn = func(*Session) {}
	}
	r.handleDisconn = fn
}

// SetErrorHandler session error handle function
func (r *Ringing) SetErrorHandler(fn func(*Session, error)) {
	if fn == nil {
		fn = func(*Session, error) {}
	}
	r.handleErr = fn
}

// SetPongHandler session pong handle function
func (r *Ringing) SetPongHandler(fn func(*Session)) {
	if fn == nil {
		fn = func(*Session) {}
	}
	r.handlePong = fn
}

// SetReceiveHandler session receive handle function
func (r *Ringing) SetReceiveHandler(fn func(*Session, []byte)) {
	if fn == nil {
		fn = func(*Session, []byte) {}
	}
	r.handleReceive = fn
}

// HandleRequest upgrade http request to websocket connection
func (r *Ringing) HandleRequest(resp http.ResponseWriter, req *http.Request) error {
	conn, err := r.Upgrader.Upgrade(resp, req, resp.Header())
	if err != nil {
		return err
	}
	session := &Session{
		req:     req,
		conn:    conn,
		state:   true,
		topics:  make(map[string]bool),
		sink:    make(chan []byte, r.Config.MsgBufSize),
		ringing: r,
		stop:    make(chan struct{}),
		rwmutex: &sync.RWMutex{},
	}
	r.sessions[session] = true
	r.handleConn(session)
	go session.sender()
	go session.receiver()

	return nil
}

func subscribe(session *Session, msg []byte) {
	req := SubscribeReq{}
	resp := &SubscribeResp{
		Status: "err",
	}

	err := json.Unmarshal(msg, &req)
	if err == nil {
		if t, ok := session.ringing.topics[req.Topic]; ok {
			t.subscribe <- session
			session.topics[req.Topic] = true
			resp.Status = "ok"
		}
	}

	b, _ := json.Marshal(resp)
	session.sink <- b
}

func (r *Ringing) close() {
	for k, v := range r.topics {
		delete(r.topics, k)
		v.close()
		v = nil
	}
	for s := range r.sessions {
		s.close()
		s = nil
	}
}
