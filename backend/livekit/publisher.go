package livekit

import (
	"encoding/json"
	"log"
	"time"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

type TranscriptionMessage struct {
	Type            string `json:"type"`
	ParticipantID   string `json:"participantId"`
	ParticipantName string `json:"participantName"`
	Text            string `json:"text"`
	Language        string `json:"language"`
	Timestamp       int64  `json:"timestamp"`
	IsFinal         bool   `json:"isFinal"`
}

func PublishTranscription(room *lksdk.Room, participantID, participantName, text, language string) error {
	if room == nil || text == "" {
		return nil
	}

	msg := TranscriptionMessage{
		Type:            "transcription",
		ParticipantID:   participantID,
		ParticipantName: participantName,
		Text:            text,
		Language:        language,
		Timestamp:       time.Now().UnixMilli(),
		IsFinal:         true,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = room.LocalParticipant.PublishDataPacket(lksdk.UserData(data), lksdk.WithDataPublishReliable(true))
	if err != nil {
		log.Printf("publish transcription error: %v", err)
		return err
	}

	return nil
}

func PublishTranscriptionToDestinations(room *lksdk.Room, participantID, participantName, text, language string, destSIDs []string) error {
	if room == nil || text == "" {
		return nil
	}

	msg := TranscriptionMessage{
		Type:            "transcription",
		ParticipantID:   participantID,
		ParticipantName: participantName,
		Text:            text,
		Language:        language,
		Timestamp:       time.Now().UnixMilli(),
		IsFinal:         true,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	opts := []lksdk.DataPublishOption{
		lksdk.WithDataPublishReliable(true),
	}
	if len(destSIDs) > 0 {
		opts = append(opts, lksdk.WithDataPublishDestination(destSIDs))
	}

	err = room.LocalParticipant.PublishDataPacket(lksdk.UserData(data), opts...)
	if err != nil {
		log.Printf("publish transcription error: %v", err)
		return err
	}

	return nil
}

var _ = livekit.DataPacket_Kind_name
