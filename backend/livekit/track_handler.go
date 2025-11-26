package livekit

import (
	"log"

	"webrtc-transcription/backend/audio"

	"gopkg.in/hraban/opus.v2"
)

type TrackHandler struct {
	participantID   string
	participantName string
	decoder         *opus.Decoder
	inputRate       int
	outputRate      int
	buffer          *audio.Buffer
	onWav           TranscriptionHandler
	decodeBuffer    []int16
	packetCount     int
}

func NewTrackHandler(participantID, participantName string, inputRate, outputRate, chunkMs int, onWav TranscriptionHandler) *TrackHandler {
	decoder, err := opus.NewDecoder(inputRate, 1)
	if err != nil {
		log.Printf("failed to create opus decoder: %v", err)
		return nil
	}

	return &TrackHandler{
		participantID:   participantID,
		participantName: participantName,
		decoder:         decoder,
		inputRate:       inputRate,
		outputRate:      outputRate,
		buffer:          audio.NewBuffer(outputRate, chunkMs),
		onWav:           onWav,
		decodeBuffer:    make([]int16, 5760),
	}
}

func (h *TrackHandler) ProcessRTP(payload []byte) {
	if h == nil || len(payload) == 0 {
		return
	}

	h.packetCount++
	if h.packetCount%500 == 1 {
		log.Printf("[%s] processing RTP packet %d, payload size: %d", h.participantName, h.packetCount, len(payload))
	}

	n, err := h.decoder.Decode(payload, h.decodeBuffer)
	if err != nil {
		if h.packetCount%500 == 1 {
			log.Printf("[%s] opus decode error: %v", h.participantName, err)
		}
		return
	}
	if n == 0 {
		return
	}

	resampled := audio.Resample(h.decodeBuffer[:n], h.inputRate, h.outputRate)
	pcmBytes := int16ToBytes(resampled)

	wav := h.buffer.Write(pcmBytes)
	if wav != nil {
		log.Printf("[%s] sending WAV to transcription (%d bytes)", h.participantName, len(wav))
		h.onWav(h.participantID, h.participantName, wav)
	}
}

func (h *TrackHandler) Close() {
	if h == nil {
		return
	}
	wav := h.buffer.Flush()
	if wav != nil {
		log.Printf("[%s] flushing final WAV", h.participantName)
		h.onWav(h.participantID, h.participantName, wav)
	}
}

func int16ToBytes(samples []int16) []byte {
	b := make([]byte, len(samples)*2)
	for i, s := range samples {
		b[i*2] = byte(s)
		b[i*2+1] = byte(s >> 8)
	}
	return b
}
