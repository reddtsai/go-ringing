package ringing

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Session wrap websocket conn
type Session struct {
	req     *http.Request
	conn    *websocket.Conn
	topics  map[string]bool
	state   bool
	sink    chan []byte
	ringing *Ringing
	stop    chan struct{}
	rwmutex *sync.RWMutex
}

func (session *Session) ping() {
	session.writeMessage(websocket.PingMessage, []byte{})
}

func (session *Session) writeMessage(msgType int, msg []byte) error {
	session.conn.SetWriteDeadline(time.Now().Add(session.ringing.Config.SendWait))
	err := session.conn.WriteMessage(msgType, msg)
	if err != nil {
		return err
	}
	return nil
}

func (session *Session) push(s []byte) {
	if !session.isAvailable() {
		session.ringing.handleErr(session, errors.New("session is not available to send"))
		return
	}
	select {
	case session.sink <- s:
	default:
		session.ringing.handleErr(session, errors.New("session message buffer is full"))
	}
}

func (session *Session) isAvailable() bool {
	session.rwmutex.RLock()
	defer session.rwmutex.RUnlock()
	return session.state
}

func (session *Session) sender() {
	ticker := time.NewTicker(session.ringing.Config.PingPeriod)
	defer ticker.Stop()
Loop:
	for {
		select {
		case s, ok := <-session.sink:
			if !ok {
				break Loop
			}
			err := session.writeMessage(websocket.TextMessage, s)
			if err != nil {
				session.ringing.handleErr(session, err)
				break Loop
			}
			// session.ringing.sendHandler(session, s)
		case <-session.stop:
			break Loop
		case <-ticker.C:
			session.ping()
		}
	}
}

func (session *Session) receiver() {
	defer session.close()

	session.conn.SetReadLimit(session.ringing.Config.ReadLimit)
	session.conn.SetReadDeadline(time.Now().Add(session.ringing.Config.PongWait))

	// pong
	session.conn.SetPongHandler(func(string) error {
		session.conn.SetReadDeadline(time.Now().Add(session.ringing.Config.PongWait))
		session.ringing.handlePong(session)
		return nil
	})

	// read
	for {
		_, msg, err := session.conn.ReadMessage()
		if err != nil {
			// if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			// }
			session.ringing.handleErr(session, err)
			break
		}
		// message := bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
		session.ringing.handleReceive(session, msg)
		// session.ringing.subscribe(session, msg)
	}
}

func (session *Session) close() {
	if session.state {
		session.rwmutex.Lock()
		session.state = false
		session.stop <- struct{}{}
		session.conn.Close()
		close(session.sink)
		close(session.stop)
		for k := range session.topics {
			if t, ok := session.ringing.topics[k]; ok {
				t.unsubscribe <- session
			}
		}
		session.rwmutex.Unlock()

		session.ringing.handleDisconn(session)
		delete(session.ringing.sessions, session)
	}
}
