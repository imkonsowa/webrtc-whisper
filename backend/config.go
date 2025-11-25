package main

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port            int
	Host            string
	WhisperEndpoint string
	Language        string
	ChunkDuration   int
	STUNServers     []string
	SampleRate      int
	FrontendDir     string
}

func LoadConfig() *Config {
	return &Config{
		Port:            envInt("PORT", 8080),
		Host:            env("HOST", "0.0.0.0"),
		WhisperEndpoint: env("WHISPER_ENDPOINT", "http://localhost:9000/v1/audio/transcriptions"),
		Language:        env("LANGUAGE", "ar"),
		ChunkDuration:   envInt("CHUNK_DURATION", 3),
		STUNServers:     strings.Split(env("STUN_SERVERS", "stun:stun.l.google.com:19302"), ","),
		SampleRate:      envInt("SAMPLE_RATE", 16000),
		FrontendDir:     env("FRONTEND_DIR", "./frontend"),
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
