package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"webrtc-transcription/backend/api"
	"webrtc-transcription/backend/livekit"
	"webrtc-transcription/backend/transcription"
)

func main() {
	cfg := LoadConfig()
	whisper := transcription.NewWhisper(cfg.WhisperEndpoint, cfg.Language)
	tokenGen := livekit.NewTokenGenerator(cfg.LiveKitAPIKey, cfg.LiveKitSecret)

	var bot *livekit.Bot

	transcriptionHandler := func(participantID, participantName string, wav []byte) {
		go func() {
			text, err := whisper.Transcribe(wav)
			if err != nil {
				log.Printf("transcribe error for %s: %v", participantID, err)
				return
			}
			if text == "" {
				return
			}
			log.Printf("[%s] %s", participantName, text)

			for _, roomName := range getRoomNames(bot) {
				room := bot.GetRoom(roomName)
				if room != nil {
					livekit.PublishTranscription(room, participantID, participantName, text, cfg.Language)
				}
			}
		}()
	}

	bot = livekit.NewBot(cfg.LiveKitURL, cfg.LiveKitAPIKey, cfg.LiveKitSecret, transcriptionHandler)

	roomsAPI := api.NewRoomsAPI(bot, tokenGen, cfg.LiveKitURL)

	http.Handle("/api/rooms", roomsAPI)
	http.Handle("/api/rooms/", roomsAPI)
	http.Handle("/", http.FileServer(http.Dir(cfg.FrontendDir)))

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("listening on %s", addr)
	log.Printf("LiveKit URL: %s", cfg.LiveKitURL)

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}

func getRoomNames(bot *livekit.Bot) []string {
	return bot.GetRoomNames()
}
