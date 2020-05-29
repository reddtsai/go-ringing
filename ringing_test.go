package ringing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type TestServer struct {
	ringing *Ringing
}

func NewTestServer() *TestServer {
	// ws
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	// topic
	topics := make(map[string]*Topic)
	t := newTopic("test")
	topics["test"] = t
	go t.controller()

	r := &Ringing{
		Config:   NewConfig("localhost:4161"),
		Upgrader: upgrader,
		topics:   topics,
		sessions: make(map[*Session]bool),
	}

	r.SetConnHandler(nil)
	r.SetDisconnHandler(nil)
	r.SetReceiveHandler(subscribe)
	r.SetPongHandler(nil)
	r.SetErrorHandler(nil)
	return &TestServer{
		ringing: r,
	}
}

func NewDialer(url string) (*websocket.Conn, error) {
	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial(strings.Replace(url, "http", "ws", 1), nil)
	return conn, err
}

func (s *TestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.ringing.HandleRequest(w, r)
}

func TestSubscribe(t *testing.T) {
	ringing := NewTestServer()
	server := httptest.NewServer(ringing)
	defer func() {
		ringing.ringing.close()
		server.Close()
	}()

	fn := func() bool {
		conn, err := NewDialer(server.URL)
		defer conn.Close()
		if err != nil {
			t.Error(err)
			return false
		}
		sub := &SubscribeReq{
			Topic: "test",
		}
		req, _ := json.Marshal(sub)
		conn.WriteMessage(websocket.TextMessage, req)
		_, buf, err := conn.ReadMessage()
		if err != nil {
			t.Error(err)
			return false
		}
		resp := SubscribeResp{}
		if err := json.Unmarshal(buf, &resp); err != nil {
			t.Error(err)
			return false
		}
		if "ok" != resp.Status {
			t.Errorf("response status %s", resp.Status)
			return false
		}
		return true
	}

	fn()
}

func TestPublish(t *testing.T) {
	ringing := NewTestServer()
	server := httptest.NewServer(ringing)
	defer func() {
		ringing.ringing.close()
		server.Close()
	}()

	fakePublish := func(r *Ringing) {
		topic, _ := r.topics["test"]
		for s := range r.sessions {
			topic.sessions[s] = true
		}
		b := []byte("hi")
		topic.sink <- b
	}
	conn, err := NewDialer(server.URL)
	defer conn.Close()
	if err != nil {
		t.Error(err)
	}
	fakePublish(ringing.ringing)
	_, buf, err := conn.ReadMessage()
	if err != nil {
		t.Error(err)
	}
	if "hi" != string(buf) {
		t.Errorf("read returned %s, want hi", string(buf))
	}
}

func TestPingPong(t *testing.T) {
	ringing := NewTestServer()
	ringing.ringing.Config.PongWait = time.Second
	ringing.ringing.Config.PingPeriod = time.Second * 9 / 10
	server := httptest.NewServer(ringing)
	defer func() {
		ringing.ringing.close()
		server.Close()
	}()

	conn, err := NewDialer(server.URL)
	conn.SetPingHandler(func(string) error {
		t.Log("ping")
		if (time.Now().Second() % 10) != 0 {
			conn.WriteMessage(websocket.PongMessage, nil)
		}
		return nil
	})
	defer conn.Close()
	if err != nil {
		t.Error(err)
	}
	_, _, err = conn.ReadMessage()
	t.Log(err)
	if err == nil {
		t.Error("there should be an error")
	}
}
