package signaling

import (
	"encoding/json"
	"log"
	"net/http"

	"webrtc-transcription/backend/audio"

	"github.com/gorilla/websocket"
)

type Msg struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId,omitempty"`
	Text      string `json:"text,omitempty"`
	Language  string `json:"language,omitempty"`
	Message   string `json:"message,omitempty"`
	Code      string `json:"code,omitempty"`
}

type AudioHandler func(sessionID string, wav []byte, onText func(string))

type Handler struct {
	upgrader     websocket.Upgrader
	audioHandler AudioHandler
	inputRate    int
	outputRate   int
	chunkMs      int
}

func NewHandler(audioHandler AudioHandler, inputRate, outputRate, chunkMs int) *Handler {
	return &Handler{
		audioHandler: audioHandler,
		inputRate:    inputRate,
		outputRate:   outputRate,
		chunkMs:      chunkMs,
		upgrader:     websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }},
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	var sessionID string
	buf := audio.NewBuffer(h.outputRate, h.chunkMs)

	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		if msgType == websocket.TextMessage {
			var msg Msg
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}

			if msg.Type == "join" {
				sessionID = msg.SessionID
				log.Printf("session %s joined", sessionID)
				_ = conn.WriteJSON(Msg{Type: "ready", SessionID: sessionID})
			}
		} else if msgType == websocket.BinaryMessage && sessionID != "" {
			resampled := audio.ResampleBytes(data, h.inputRate, h.outputRate)
			wav := buf.Write(resampled)

			if wav != nil {
				h.audioHandler(sessionID, wav, func(text string) {
					_ = conn.WriteJSON(Msg{Type: "transcription", SessionID: sessionID, Text: text, Language: "ar"})
				})
			}
		}
	}
}
