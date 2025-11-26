package main

import (
	"os"
	"strconv"
)

type Config struct {
	Port            int
	Host            string
	WhisperEndpoint string
	Language        string
	ChunkDuration   int
	SampleRate      int
	FrontendDir     string
	LiveKitURL      string
	LiveKitAPIKey   string
	LiveKitSecret   string
}

func LoadConfig() *Config {
	return &Config{
		Port:            envInt("PORT", 8080),
		Host:            env("HOST", "0.0.0.0"),
		WhisperEndpoint: env("WHISPER_ENDPOINT", "http://localhost:9000/v1/audio/transcriptions"),
		Language:        env("LANGUAGE", "ar"),
		ChunkDuration:   envInt("CHUNK_DURATION", 3),
		SampleRate:      envInt("SAMPLE_RATE", 16000),
		FrontendDir:     env("FRONTEND_DIR", "./frontend"),
		LiveKitURL:      env("LIVEKIT_URL", "ws://localhost:7880"),
		LiveKitAPIKey:   env("LIVEKIT_API_KEY", "devkey"),
		LiveKitSecret:   env("LIVEKIT_API_SECRET", "devsecret1234567890abcdefghijklmn"),
	}
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envInt(k string, d int) int {
	if v, err := strconv.Atoi(os.Getenv(k)); err == nil {
		return v
	}
	return d
}
