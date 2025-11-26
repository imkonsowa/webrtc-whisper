package livekit

import (
	"time"

	"github.com/livekit/protocol/auth"
)

type TokenGenerator struct {
	apiKey    string
	apiSecret string
}

func NewTokenGenerator(apiKey, apiSecret string) *TokenGenerator {
	return &TokenGenerator{
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

func (t *TokenGenerator) Generate(roomName, participantID, participantName string) (string, error) {
	at := auth.NewAccessToken(t.apiKey, t.apiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	at.AddGrant(grant).
		SetIdentity(participantID).
		SetName(participantName).
		SetValidFor(24 * time.Hour)

	return at.ToJWT()
}
