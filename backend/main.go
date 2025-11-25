package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"webrtc-transcription/backend/signaling"
	"webrtc-transcription/backend/transcription"
	rtc "webrtc-transcription/backend/webrtc"

	"github.com/pion/webrtc/v4"
)

func main() {
	cfg := LoadConfig()
	whisper := transcription.NewWhisper(cfg.WhisperEndpoint, cfg.Language)

	var (
		procs = make(map[string]*transcription.Processor)
		mu    sync.Mutex
	)

	factory := func(sid string, onText func(string)) (*webrtc.PeerConnection, error) {
		proc := transcription.NewProcessor(whisper, cfg.SampleRate, cfg.ChunkDuration, onText)
		mu.Lock()
		procs[sid] = proc
		mu.Unlock()

		pc, err := rtc.NewPeerConnection(cfg.STUNServers)
		if err != nil {
			return nil, err
		}

		pc.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
			go rtc.HandleTrack(t, proc.Add)
		})

		pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
			if s == webrtc.PeerConnectionStateClosed || s == webrtc.PeerConnectionStateDisconnected {
				mu.Lock()
				if p := procs[sid]; p != nil {
					p.Flush()
					delete(procs, sid)
				}
				mu.Unlock()
			}
		})
		return pc, nil
	}

	http.Handle("/ws", signaling.NewHandler(factory))
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
