package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"webrtc-transcription/backend/livekit"

	"github.com/google/uuid"
)

type RoomsAPI struct {
	bot        *livekit.Bot
	tokenGen   *livekit.TokenGenerator
	liveKitURL string
}

func NewRoomsAPI(bot *livekit.Bot, tokenGen *livekit.TokenGenerator, liveKitURL string) *RoomsAPI {
	return &RoomsAPI{
		bot:        bot,
		tokenGen:   tokenGen,
		liveKitURL: liveKitURL,
	}
}

type CreateRoomRequest struct {
	Name string `json:"name,omitempty"`
}

type CreateRoomResponse struct {
	Name string `json:"name"`
}

type JoinRoomRequest struct {
	ParticipantName string `json:"participantName"`
}

type JoinRoomResponse struct {
	Token      string `json:"token"`
	LiveKitURL string `json:"livekitUrl"`
}

func (a *RoomsAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/rooms")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	switch {
	case r.Method == "POST" && len(parts) == 1 && parts[0] == "":
		a.createRoom(w, r)
	case r.Method == "POST" && len(parts) == 2 && parts[1] == "join":
		a.joinRoom(w, r, parts[0])
	case r.Method == "DELETE" && len(parts) == 1 && parts[0] != "":
		a.deleteRoom(w, r, parts[0])
	default:
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	}
}

func (a *RoomsAPI) createRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	json.NewDecoder(r.Body).Decode(&req)

	roomName := req.Name
	if roomName == "" {
		roomName = "room-" + uuid.New().String()[:8]
	}

	if err := a.bot.JoinRoom(roomName); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(CreateRoomResponse{Name: roomName})
}

func (a *RoomsAPI) joinRoom(w http.ResponseWriter, r *http.Request, roomName string) {
	var req JoinRoomRequest
	json.NewDecoder(r.Body).Decode(&req)

	if req.ParticipantName == "" {
		req.ParticipantName = "User"
	}

	participantID := "user-" + uuid.New().String()[:8]

	if err := a.bot.JoinRoom(roomName); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	token, err := a.tokenGen.Generate(roomName, participantID, req.ParticipantName)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	wsURL := strings.Replace(a.liveKitURL, "ws://livekit:", "ws://localhost:", 1)

	json.NewEncoder(w).Encode(JoinRoomResponse{
		Token:      token,
		LiveKitURL: wsURL,
	})
}

func (a *RoomsAPI) deleteRoom(w http.ResponseWriter, r *http.Request, roomName string) {
	a.bot.LeaveRoom(roomName)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
