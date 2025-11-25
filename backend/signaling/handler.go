package signaling

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type Msg struct {
	Type      string          `json:"type"`
	SessionID string          `json:"sessionId,omitempty"`
	SDP       string          `json:"sdp,omitempty"`
	Candidate json.RawMessage `json:"candidate,omitempty"`
	Text      string          `json:"text,omitempty"`
	Language  string          `json:"language,omitempty"`
	Message   string          `json:"message,omitempty"`
	Code      string          `json:"code,omitempty"`
}

type PCFactory func(sid string, onText func(string)) (*webrtc.PeerConnection, error)

type Handler struct {
	sessions *Sessions
	factory  PCFactory
	upgrader websocket.Upgrader
}

func NewHandler(factory PCFactory) *Handler {
	return &Handler{
		sessions: NewSessions(),
		factory:  factory,
		upgrader: websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }},
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.handle(conn)
}

func (h *Handler) handle(conn *websocket.Conn) {
	var sid string
	defer func() {
		if sid != "" {
			h.sessions.Remove(sid)
		}
		_ = conn.Close()
	}()

	for {
		var msg Msg
		if err := conn.ReadJSON(&msg); err != nil {
			return
		}

		switch msg.Type {
		case "join":
			sid = msg.SessionID
			sess := h.sessions.Add(sid, conn)
			pc, err := h.factory(sid, func(text string) {
				_ = sess.Send(Msg{Type: "transcription", SessionID: sid, Text: text, Language: "ar"})
			})
			if err != nil {
				_ = sess.Send(Msg{Type: "error", Code: "PEER_ERROR", Message: err.Error()})
				return
			}
			sess.PeerConn = pc
			_ = sess.Send(Msg{Type: "ready", SessionID: sid})

		case "offer":
			sess := h.sessions.Get(msg.SessionID)
			if sess == nil || sess.PeerConn == nil {
				continue
			}
			_ = sess.PeerConn.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: msg.SDP})
			ans, _ := sess.PeerConn.CreateAnswer(nil)
			_ = sess.PeerConn.SetLocalDescription(ans)
			_ = sess.Send(Msg{Type: "answer", SessionID: msg.SessionID, SDP: ans.SDP})

		case "candidate":
			sess := h.sessions.Get(msg.SessionID)
			if sess == nil || sess.PeerConn == nil {
				continue
			}
			var c webrtc.ICECandidateInit
			_ = json.Unmarshal(msg.Candidate, &c)
			_ = sess.PeerConn.AddICECandidate(c)
		}
	}
}
