package audio

import (
	"bytes"
	"encoding/binary"
	"sync"
)

type Buffer struct {
	mu         sync.Mutex
	samples    []int16
	sampleRate int
	chunkSize  int
	vad        *VAD
	hasSpeech  bool
}

func NewBuffer(sampleRate, chunkDurationMs int) *Buffer {
	chunkSize := sampleRate * chunkDurationMs / 1000
	return &Buffer{
		sampleRate: sampleRate,
		chunkSize:  chunkSize,
		samples:    make([]int16, 0, chunkSize),
		vad:        NewVAD(sampleRate),
	}
}

func (b *Buffer) Write(pcm []byte) []byte {
	b.mu.Lock()
	defer b.mu.Unlock()

	incoming := make([]int16, len(pcm)/2)
	for i := 0; i < len(incoming); i++ {
		incoming[i] = int16(pcm[i*2]) | int16(pcm[i*2+1])<<8
	}

	if b.vad.Process(incoming) {
		b.hasSpeech = true
	}

	b.samples = append(b.samples, incoming...)

	if len(b.samples) >= b.chunkSize {
		if b.hasSpeech {
			wav := b.toWAV(b.samples[:b.chunkSize])
			b.samples = b.samples[b.chunkSize:]
			b.hasSpeech = false
			return wav
		}
		b.samples = b.samples[b.chunkSize:]
		b.hasSpeech = false
	}

	return nil
}

func (b *Buffer) Flush() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.samples) == 0 || !b.hasSpeech {
		b.samples = b.samples[:0]
		b.hasSpeech = false
		return nil
	}

	wav := b.toWAV(b.samples)
	b.samples = b.samples[:0]
	b.hasSpeech = false
	return wav
}

func (b *Buffer) toWAV(samples []int16) []byte {
	var buf bytes.Buffer

	dataSize := len(samples) * 2
	fileSize := 36 + dataSize

	buf.WriteString("RIFF")
	_ = binary.Write(&buf, binary.LittleEndian, int32(fileSize))
	buf.WriteString("WAVE")

	buf.WriteString("fmt ")
	_ = binary.Write(&buf, binary.LittleEndian, int32(16))
	_ = binary.Write(&buf, binary.LittleEndian, int16(1))
	_ = binary.Write(&buf, binary.LittleEndian, int16(1))
	_ = binary.Write(&buf, binary.LittleEndian, int32(b.sampleRate))
	_ = binary.Write(&buf, binary.LittleEndian, int32(b.sampleRate*2))
	_ = binary.Write(&buf, binary.LittleEndian, int16(2))
	_ = binary.Write(&buf, binary.LittleEndian, int16(16))

	buf.WriteString("data")
	_ = binary.Write(&buf, binary.LittleEndian, int32(dataSize))

	for _, s := range samples {
		_ = binary.Write(&buf, binary.LittleEndian, s)
	}

	return buf.Bytes()
}
