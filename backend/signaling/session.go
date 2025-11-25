package signaling

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type Session struct {
	ID       string
	Conn     *websocket.Conn
	PeerConn *webrtc.PeerConnection
	mu       sync.Mutex
}

func (s *Session) Send(v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Conn.WriteJSON(v)
}

func (s *Session) Close() {
	if s.PeerConn != nil {
		_ = s.PeerConn.Close()
	}
	if s.Conn != nil {
		_ = s.Conn.Close()
	}
}

type Sessions struct {
	m  map[string]*Session
	mu sync.RWMutex
}

func NewSessions() *Sessions {
	return &Sessions{m: make(map[string]*Session)}
}

func (s *Sessions) Add(id string, conn *websocket.Conn) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := &Session{ID: id, Conn: conn}
	s.m[id] = sess
	return sess
}

func (s *Sessions) Get(id string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.m[id]
}

func (s *Sessions) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess := s.m[id]; sess != nil {
		sess.Close()
		delete(s.m, id)
	}
}
