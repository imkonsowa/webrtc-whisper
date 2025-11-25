package transcription

import (
	"bytes"
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

func (w *Whisper) Transcribe(audio []byte) (string, error) {
	var body bytes.Buffer
	wr := multipart.NewWriter(&body)
	part, _ := wr.CreateFormFile("file", "audio.wav")
	_, _ = part.Write(audio)
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
