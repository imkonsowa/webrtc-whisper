package webrtc

import "github.com/pion/webrtc/v4"

func HandleTrack(track *webrtc.TrackRemote, onSamples func([]float32)) {
	if track.Kind() != webrtc.RTPCodecTypeAudio {
		return
	}
	dec := NewDecoder()
	buf := make([]byte, 1500)
	for {
		n, _, err := track.Read(buf)
		if err != nil {
			return
		}
		if samples, err := dec.Decode(buf[:n]); err == nil && len(samples) > 0 {
			onSamples(samples)
		}
	}
}
