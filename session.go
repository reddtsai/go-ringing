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
	sink    chan *signal
	ringing *Ringing
	state   bool
	rwmutex *sync.RWMutex
}

type signal struct {
	msgType int
	msg     []byte
}

func (session *Session) ping() {
	session.writeMessage(websocket.PingMessage, []byte{})
}

func (session *Session) writeMessage(msgType int, msg []byte) error {
	if !session.isAvailable() {
		return errors.New("session is not available to send")
	}
	session.conn.SetWriteDeadline(time.Now().Add(session.ringing.Config.SendWait))
	err := session.conn.WriteMessage(msgType, msg)
	if err != nil {
		return err
	}
	return nil
}

func (session *Session) push(s *signal) {
	if !session.isAvailable() {
		session.ringing.errHandler(session, errors.New("session is not available to send"))
		return
	}
	select {
	case session.sink <- s:
	default:
		session.ringing.errHandler(session, errors.New("session message buffer is full"))
	}
}

func (session *Session) isAvailable() bool {
	session.rwmutex.RLock()
	defer session.rwmutex.RUnlock()

	return session.state
}

func (session *Session) close() {
	if session.isAvailable() {
		session.rwmutex.Lock()
		session.state = false
		session.conn.Close()
		close(session.sink)
		session.rwmutex.Unlock()
		// TODO
		// unregister
	}
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
			err := session.writeMessage(s.msgType, s.msg)
			if err != nil {
				session.ringing.errHandler(session, err)
				break Loop
			}

			if s.msgType == websocket.CloseMessage {
				// TODO
				// print close
				break Loop
			}
			session.ringing.sendHandler(session, s)
		case <-ticker.C:
			session.ping()
		}
	}
}

func (session *Session) receiver() {
	session.conn.SetReadLimit(session.ringing.Config.ReadLimit)
	session.conn.SetReadDeadline(time.Now().Add(session.ringing.Config.PongWait))
	session.conn.SetPongHandler(func(string) error {
		session.conn.SetReadDeadline(time.Now().Add(session.ringing.Config.PongWait))
		session.ringing.pongHandler(session)
		return nil
	})
	if session.ringing.closeHandler != nil {
		session.conn.SetCloseHandler(func(code int, text string) error {
			return session.ringing.closeHandler(session, code, text)
		})
	}
	for {
		msgType, msg, err := session.conn.ReadMessage()
		if err != nil {
			// if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			// }
			session.ringing.errHandler(session, err)
			break
		}
		// message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		session.ringing.receiveHandler(session, &signal{msgType: msgType, msg: msg})
		// TODO
		// close
	}
}

// TODO
// Subscribe
