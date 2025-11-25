package webrtc

import (
	"github.com/pion/opus"
	"github.com/pion/rtp"
)

type Decoder struct {
	dec opus.Decoder
	buf []float32
}

func NewDecoder() *Decoder {
	return &Decoder{dec: opus.NewDecoder(), buf: make([]float32, 5760)}
}

func (d *Decoder) Decode(data []byte) ([]float32, error) {
	pkt := &rtp.Packet{}
	if err := pkt.Unmarshal(data); err != nil {
		return nil, err
	}
	_, stereo, err := d.dec.DecodeFloat32(pkt.Payload, d.buf)
	if err != nil {
		return nil, err
	}
	n := 960
	out := make([]float32, n/3)
	if stereo {
		for i := 0; i < n; i += 3 {
			out[i/3] = (d.buf[i*2] + d.buf[i*2+1]) / 2
		}
	} else {
		for i := 0; i < n; i += 3 {
			out[i/3] = d.buf[i]
		}
	}
	return out, nil
}
