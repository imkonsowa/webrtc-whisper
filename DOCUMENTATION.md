# Real-Time Arabic Speech Transcription System

A real-time speech-to-text system that captures audio from a browser microphone and transcribes it using Whisper AI, optimized for Arabic language.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         BROWSER                                  │
│  ┌──────────────┐    ┌─────────────────┐    ┌────────────────┐ │
│  │ Microphone   │───▶│ AudioWorklet    │───▶│ WebSocket      │ │
│  │ (48kHz)      │    │ (PCM chunks)    │    │ Client         │ │
│  └──────────────┘    └─────────────────┘    └───────┬────────┘ │
└─────────────────────────────────────────────────────┼──────────┘
                                                      │ PCM Binary
                                                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                      GO BACKEND                                  │
│  ┌────────────────┐    ┌────────────────┐    ┌───────────────┐ │
│  │ WebSocket      │───▶│ Resampler      │───▶│ Audio Buffer  │ │
│  │ Handler        │    │ (48k→16k)      │    │ + VAD         │ │
│  └────────────────┘    └────────────────┘    └───────┬───────┘ │
│                                                      │ WAV      │
│                                                      ▼          │
│                                              ┌───────────────┐  │
│                                              │ Whisper       │  │
│                                              │ Client        │  │
│                                              └───────┬───────┘  │
└──────────────────────────────────────────────────────┼──────────┘
                                                       │ HTTP POST
                                                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                    WHISPER SERVER                                │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ faster-whisper-server (Systran/faster-whisper-base)        │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Data Flow

1. Browser captures audio at 48kHz via `getUserMedia`
2. AudioWorklet converts float32 samples to 16-bit PCM, sends 100ms chunks
3. WebSocket transmits raw PCM binary to backend
4. Backend resamples 48kHz → 16kHz (Whisper's preferred rate)
5. Audio buffer accumulates 3 seconds of samples
6. VAD checks if speech is present; silent chunks are discarded
7. Speech chunks are converted to WAV format
8. WAV sent to Whisper API for transcription
9. Transcription text returned to browser via WebSocket

## Project Structure

```
webrtc-transcription/
├── backend/
│   ├── main.go              # Application entry point
│   ├── config.go            # Environment configuration
│   ├── audio/
│   │   ├── buffer.go        # Audio buffering and WAV generation
│   │   ├── resampler.go     # Sample rate conversion
│   │   └── vad.go           # Voice Activity Detection
│   ├── signaling/
│   │   └── handler.go       # WebSocket connection handler
│   └── transcription/
│       └── whisper.go       # Whisper API client
├── frontend/
│   ├── index.html           # UI markup
│   ├── app.js               # Main application logic
│   └── audio-processor.js   # AudioWorklet processor
├── docker/
│   ├── Dockerfile           # Backend container build
│   └── docker-compose.yml   # Service orchestration
└── Makefile                 # Build and run commands
```

## Component Documentation

### Frontend

#### `audio-processor.js` - AudioWorklet Processor

Runs in a dedicated audio thread for low-latency capture.

| Property | Type | Description |
|----------|------|-------------|
| `bufferSize` | number | Samples per chunk (4800 = 100ms at 48kHz) |
| `buffer` | Float32Array | Accumulates incoming samples |
| `bufferIndex` | number | Current write position |

| Method | Description |
|--------|-------------|
| `process(inputs, outputs, parameters)` | Called by audio system ~every 128 samples. Accumulates samples, converts to Int16, sends via postMessage when buffer is full. |

#### `app.js` - TranscriptionApp Class

Main application controller.

| Property | Type | Description |
|----------|------|-------------|
| `ws` | WebSocket | Connection to backend |
| `audioContext` | AudioContext | Web Audio API context |
| `workletNode` | AudioWorkletNode | Reference to audio processor |
| `stream` | MediaStream | Microphone stream |
| `sessionId` | string | Unique session identifier |

| Method | Description |
|--------|-------------|
| `start()` | Requests microphone access, initiates WebSocket connection |
| `connectWebSocket()` | Establishes WebSocket, sends join message |
| `startAudioCapture()` | Loads AudioWorklet, connects audio graph |
| `addTranscription(text)` | Appends transcribed text to UI |
| `stop()` | Terminates recording session |
| `cleanup()` | Releases all resources |

### Backend

#### `config.go` - Configuration

Environment-based configuration with defaults.

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | HTTP server port |
| `HOST` | 0.0.0.0 | Bind address |
| `WHISPER_ENDPOINT` | http://localhost:9000/v1/audio/transcriptions | Whisper API URL |
| `LANGUAGE` | ar | Target language for transcription |
| `CHUNK_DURATION` | 3 | Seconds of audio per transcription request |
| `SAMPLE_RATE` | 16000 | Output sample rate for Whisper |
| `FRONTEND_DIR` | ./frontend | Static files directory |

#### `main.go` - Entry Point

- Loads configuration
- Creates Whisper client
- Sets up HTTP routes (`/ws` for WebSocket, `/` for static files)
- Starts server with graceful shutdown

#### `audio/buffer.go` - Buffer

Accumulates PCM samples and generates WAV files.

| Property | Type | Description |
|----------|------|-------------|
| `samples` | []int16 | Accumulated audio samples |
| `sampleRate` | int | Output sample rate |
| `chunkSize` | int | Samples per chunk (sampleRate * seconds) |
| `vad` | *VAD | Voice activity detector |
| `hasSpeech` | bool | Whether current chunk contains speech |

| Method | Description |
|--------|-------------|
| `NewBuffer(sampleRate, chunkDurationMs)` | Creates buffer with specified parameters |
| `Write(pcm []byte)` | Adds PCM data, returns WAV if chunk ready and has speech |
| `Flush()` | Returns remaining audio as WAV if contains speech |
| `toWAV(samples []int16)` | Generates WAV file bytes with proper headers |

#### `audio/resampler.go` - Resampler

Linear interpolation resampler for sample rate conversion.

| Function | Description |
|----------|-------------|
| `Resample(input []int16, inputRate, outputRate)` | Resamples audio using linear interpolation |
| `ResampleBytes(input []byte, inputRate, outputRate)` | Byte-level wrapper for Resample |

#### `audio/vad.go` - Voice Activity Detection

Energy-based speech detection to filter silence.

| Property | Type | Description |
|----------|------|-------------|
| `threshold` | float64 | RMS energy threshold (default: 500) |
| `minSpeechMs` | int | Minimum speech duration to trigger (100ms) |
| `minSilenceMs` | int | Silence duration to end speech (300ms) |
| `isSpeaking` | bool | Current speech state |

| Method | Description |
|--------|-------------|
| `NewVAD(sampleRate)` | Creates VAD with default thresholds |
| `Process(samples []int16)` | Updates state, returns true if speech detected |
| `calculateEnergy(samples)` | Computes RMS energy of samples |
| `Reset()` | Clears speech/silence counters |

#### `signaling/handler.go` - WebSocket Handler

Manages WebSocket connections and audio processing pipeline.

| Type | Description |
|------|-------------|
| `Msg` | JSON message structure for WebSocket protocol |
| `AudioHandler` | Callback function type for processed audio |
| `Handler` | HTTP handler implementing WebSocket upgrade |

| Message Type | Direction | Description |
|--------------|-----------|-------------|
| `join` | client→server | Session initiation with sessionId |
| `ready` | server→client | Acknowledgment, start sending audio |
| `transcription` | server→client | Transcribed text result |
| `error` | server→client | Error notification |

#### `transcription/whisper.go` - Whisper Client

HTTP client for Whisper speech-to-text API.

| Method | Description |
|--------|-------------|
| `NewWhisper(endpoint, lang)` | Creates client with API endpoint and language |
| `Transcribe(audio []byte)` | Sends WAV to API, returns transcribed text |

## Running the System

### Prerequisites
- Docker and Docker Compose
- Go 1.25+ (for local development)

### Quick Start
```bash
make up      # Start all services
make down    # Stop all services
make logs    # View logs
```

### Local Development
```bash
make build   # Build backend binary
make run     # Run locally (requires Whisper server)
make lint    # Run golangci-lint
```

### Access
- Web UI: http://localhost:8080
- Whisper API: http://localhost:9000

## Configuration

All configuration via environment variables in `docker-compose.yml`:

```yaml
environment:
  - PORT=8080
  - WHISPER_ENDPOINT=http://whisper:8000/v1/audio/transcriptions
  - LANGUAGE=ar
  - CHUNK_DURATION=3
```

## Performance Characteristics

| Metric | Value |
|--------|-------|
| Audio chunk interval | 100ms |
| Transcription chunk | 3 seconds |
| Sample rate (browser) | 48kHz |
| Sample rate (Whisper) | 16kHz |
| Expected latency | 3-4 seconds |
| VAD speech threshold | 100ms minimum |
| VAD silence threshold | 300ms to end |
