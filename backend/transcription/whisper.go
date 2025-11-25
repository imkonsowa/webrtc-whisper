package transcription

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type Whisper struct {
	endpoint, lang string
}

func NewWhisper(endpoint, lang string) *Whisper {
	return &Whisper{endpoint, lang}
}

func (w *Whisper) Transcribe(samples []float32) (string, error) {
	var body bytes.Buffer
	wr := multipart.NewWriter(&body)
	part, _ := wr.CreateFormFile("file", "a.wav")
	_, _ = part.Write(toWav(samples))
	_ = wr.WriteField("language", w.lang)
	_ = wr.Close()

	req, _ := http.NewRequest("POST", w.endpoint, &body)
	req.Header.Set("Content-Type", wr.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("whisper: %d %s", resp.StatusCode, b)
	}

	var r struct{ Text string }
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	return strings.TrimSpace(r.Text), nil
}

func toWav(samples []float32) []byte {
	n := len(samples)
	buf := new(bytes.Buffer)
	buf.WriteString("RIFF")
	_ = binary.Write(buf, binary.LittleEndian, uint32(36+n*2))
	buf.WriteString("WAVEfmt ")
	_ = binary.Write(buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(buf, binary.LittleEndian, uint32(16000))
	_ = binary.Write(buf, binary.LittleEndian, uint32(32000))
	_ = binary.Write(buf, binary.LittleEndian, uint16(2))
	_ = binary.Write(buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	_ = binary.Write(buf, binary.LittleEndian, uint32(n*2))
	for _, s := range samples {
		_ = binary.Write(buf, binary.LittleEndian, int16(s*32767))
	}
	return buf.Bytes()
}
