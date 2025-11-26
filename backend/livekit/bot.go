package livekit

import (
	"context"
	"log"
	"sync"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4"
)

type TranscriptionHandler func(participantID, participantName string, wav []byte)

type Bot struct {
	url       string
	apiKey    string
	apiSecret string
	handler   TranscriptionHandler
	rooms     sync.Map
}

type roomConnection struct {
	room   *lksdk.Room
	cancel context.CancelFunc
}

func NewBot(url, apiKey, apiSecret string, handler TranscriptionHandler) *Bot {
	return &Bot{
		url:       url,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		handler:   handler,
	}
}

func (b *Bot) JoinRoom(roomName string) error {
	if _, exists := b.rooms.Load(roomName); exists {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	room, err := lksdk.ConnectToRoom(b.url, lksdk.ConnectInfo{
		APIKey:              b.apiKey,
		APISecret:           b.apiSecret,
		RoomName:            roomName,
		ParticipantIdentity: "transcription-bot",
		ParticipantName:     "Transcription Bot",
	}, &lksdk.RoomCallback{
		ParticipantCallback: lksdk.ParticipantCallback{
			OnTrackSubscribed: func(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, participant *lksdk.RemoteParticipant) {
				if track.Kind() == webrtc.RTPCodecTypeAudio {
					log.Printf("subscribed to audio track from %s", participant.Identity())
					go b.handleAudioTrack(ctx, track, participant)
				}
			},
			OnTrackUnsubscribed: func(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, participant *lksdk.RemoteParticipant) {
				log.Printf("unsubscribed from track of %s", participant.Identity())
			},
		},
		OnDisconnected: func() {
			log.Printf("bot disconnected from room %s", roomName)
			b.rooms.Delete(roomName)
		},
	})
	if err != nil {
		cancel()
		return err
	}

	b.rooms.Store(roomName, &roomConnection{room: room, cancel: cancel})
	log.Printf("bot joined room %s", roomName)
	return nil
}

func (b *Bot) LeaveRoom(roomName string) {
	if val, ok := b.rooms.LoadAndDelete(roomName); ok {
		conn := val.(*roomConnection)
		conn.cancel()
		conn.room.Disconnect()
		log.Printf("bot left room %s", roomName)
	}
}

func (b *Bot) GetRoom(roomName string) *lksdk.Room {
	if val, ok := b.rooms.Load(roomName); ok {
		return val.(*roomConnection).room
	}
	return nil
}

func (b *Bot) GetRoomNames() []string {
	var names []string
	b.rooms.Range(func(key, value any) bool {
		names = append(names, key.(string))
		return true
	})
	return names
}

func (b *Bot) handleAudioTrack(ctx context.Context, track *webrtc.TrackRemote, participant *lksdk.RemoteParticipant) {
	handler := NewTrackHandler(
		participant.Identity(),
		participant.Name(),
		48000,
		16000,
		3000,
		b.handler,
	)
	defer handler.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			pkt, _, err := track.ReadRTP()
			if err != nil {
				log.Printf("track read error for %s: %v", participant.Identity(), err)
				return
			}
			handler.ProcessRTP(pkt.Payload)
		}
	}
}
