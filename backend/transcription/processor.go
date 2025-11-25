package transcription

import (
	"log"
	"sync"
)

type Processor struct {
	whisper   *Whisper
	buf       []float32
	chunkSize int
	overlap   int
	onResult  func(string)
	mu        sync.Mutex
	busy      bool
}

func NewProcessor(w *Whisper, sampleRate, chunkSec int, onResult func(string)) *Processor {
	return &Processor{
		whisper:   w,
		chunkSize: sampleRate * chunkSec,
		overlap:   sampleRate,
		onResult:  onResult,
	}
}

func (p *Processor) Add(samples []float32) {
	p.mu.Lock()
	p.buf = append(p.buf, samples...)
	ready := len(p.buf) >= p.chunkSize && !p.busy
	p.mu.Unlock()
	if ready {
		go p.process()
	}
}

func (p *Processor) process() {
	p.mu.Lock()
	if p.busy || len(p.buf) < p.chunkSize {
		p.mu.Unlock()
		return
	}
	p.busy = true
	chunk := append([]float32{}, p.buf[:p.chunkSize]...)
	p.buf = append([]float32{}, p.buf[p.chunkSize-p.overlap:]...)
	p.mu.Unlock()

	if text, err := p.whisper.Transcribe(chunk); err != nil {
		log.Printf("transcribe: %v", err)
	} else if text != "" {
		p.onResult(text)
	}

	p.mu.Lock()
	p.busy = false
	more := len(p.buf) >= p.chunkSize
	p.mu.Unlock()
	if more {
		go p.process()
	}
}

func (p *Processor) Flush() {
	p.mu.Lock()
	if len(p.buf) < p.chunkSize/3 {
		p.mu.Unlock()
		return
	}
	chunk := append([]float32{}, p.buf...)
	p.buf = nil
	p.mu.Unlock()

	if text, err := p.whisper.Transcribe(chunk); err == nil && text != "" {
		p.onResult(text)
	}
}
