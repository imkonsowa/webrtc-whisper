package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"webrtc-transcription/backend/signaling"
	"webrtc-transcription/backend/transcription"
)

func main() {
	cfg := LoadConfig()
	whisper := transcription.NewWhisper(cfg.WhisperEndpoint, cfg.Language)

	handler := func(sessionID string, wav []byte, onText func(string)) {
		go func() {
			text, err := whisper.Transcribe(wav)
			if err != nil {
				log.Printf("transcribe error: %v", err)
				return
			}
			if text != "" {
				onText(text)
			}
		}()
	}

	inputRate := 48000
	outputRate := cfg.SampleRate
	chunkMs := cfg.ChunkDuration * 1000

	http.Handle("/ws", signaling.NewHandler(handler, inputRate, outputRate, chunkMs))
	http.Handle("/", http.FileServer(http.Dir(cfg.FrontendDir)))

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("listening on %s", addr)

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
