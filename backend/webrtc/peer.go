package webrtc

import (
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
)

func NewPeerConnection(stunServers []string) (*webrtc.PeerConnection, error) {
	m := &webrtc.MediaEngine{}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeOpus,
			ClockRate:   48000,
			Channels:    2,
			SDPFmtpLine: "minptime=10;useinbandfec=1",
		},
		PayloadType: 111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		return nil, err
	}

	i := &interceptor.Registry{}
	pli, _ := intervalpli.NewReceiverInterceptor()
	i.Add(pli)
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		return nil, err
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i))

	iceServers := make([]webrtc.ICEServer, len(stunServers))
	for idx, s := range stunServers {
		iceServers[idx] = webrtc.ICEServer{URLs: []string{s}}
	}

	return api.NewPeerConnection(webrtc.Configuration{ICEServers: iceServers})
}
