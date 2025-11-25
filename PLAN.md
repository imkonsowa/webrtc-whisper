# WebRTC Real-Time Arabic Transcription System
## Technical Specification Document

---

## 1. Project Overview

Build a full-stack real-time transcription system for Arabic audio using WebRTC. The system captures audio from web browsers, streams it to a Go backend, and transcribes it using local Whisper AI models.

### Key Features
- Real-time audio capture from browser
- WebRTC peer-to-peer connection with signaling server
- Audio stream interception and buffering
- Local Arabic transcription using Whisper
- Live transcription display in web interface
- Support for multiple concurrent sessions

### Target Performance (M1 Pro)
- Latency: 3-4 seconds from speech to text
- Throughput: 30-35 concurrent sessions
- Model: Whisper base (multilingual, Arabic support)
- Real-time factor: ~0.07x (15x faster than real-time)

---

## 2. System Architecture

```
┌─────────────────┐
│   Web Browser   │
│   (SimplePeer)  │
│                 │
│  ┌───────────┐  │
│  │Microphone │  │
│  └─────┬─────┘  │
│        │ Audio  │
└────────┼────────┘
         │ WebRTC
         │
         ▼
┌─────────────────────────────────────────┐
│         Go Backend Server               │
│                                         │
│  ┌──────────────┐   ┌───────────────┐  │
│  │   Signaling  │   │ Pion WebRTC   │  │
│  │   Server     │──▶│  (SFU Mode)   │  │
│  │  (WebSocket) │   └───────┬───────┘  │
│  └──────────────┘           │          │
│                             │ RTP      │
│                    ┌────────▼────────┐ │
│                    │  RTP Interceptor│ │
│                    │  (Audio Buffer) │ │
│                    └────────┬────────┘ │
│                             │ PCM      │
│                    ┌────────▼────────┐ │
│                    │ Whisper Engine  │ │
│                    │  (CoreML/Metal) │ │
│                    └────────┬────────┘ │
│                             │ Text     │
│                    ┌────────▼────────┐ │
│                    │   WebSocket     │ │
│                    │  (Send Results) │ │
│                    └─────────────────┘ │
└─────────────────────────────────────────┘
         │ Transcription
         ▼
┌─────────────────┐
│   Web Browser   │
│ (Display Text)  │
└─────────────────┘
```

---

## 3. Technology Stack

### Frontend
- **Framework**: Plain HTML/JavaScript (or React if preferred)
- **WebRTC**: SimplePeer (https://github.com/feross/simple-peer)
- **UI**: Tailwind CSS
- **WebSocket**: Native WebSocket API

### Backend
- **Language**: Go 1.21+
- **WebRTC**: Pion WebRTC v4 (https://github.com/pion/webrtc)
- **Signaling**: Gorilla WebSocket
- **Transcription**: Whisper.cpp Go bindings (https://github.com/ggerganov/whisper.cpp)
- **Audio Processing**: Pion RTP, Pion Interceptor

### Infrastructure
- **Model Storage**: Local filesystem (`./models/`)
- **Audio Codec**: Opus (decode to PCM for Whisper)
- **Sample Rate**: 48kHz (WebRTC) → 16kHz (Whisper)

---

## 4. Project Structure

```
webrtc-transcription/
├── backend/
│   ├── main.go                    # Server entry point
│   ├── signaling/
│   │   ├── handler.go             # WebSocket signaling
│   │   └── session.go             # Session management
│   ├── webrtc/
│   │   ├── peer.go                # Pion peer connection setup
│   │   ├── interceptor.go         # RTP audio interceptor
│   │   └── track_handler.go       # Track processing
│   ├── transcription/
│   │   ├── whisper.go             # Whisper integration
│   │   ├── audio_buffer.go        # Audio accumulation
│   │   └── processor.go           # Audio processing pipeline
│   ├── models/                    # Whisper models directory
│   │   └── .gitkeep
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── index.html                 # Main page
│   ├── app.js                     # SimplePeer logic
│   ├── transcription.js           # Transcription display
│   └── styles.css                 # Tailwind/custom styles
│
├── scripts/
│   ├── download-models.sh         # Download Whisper models
│   └── setup.sh                   # Setup script
│
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── README.md
└── .gitignore
```

---

## 5. API Specifications

### 5.1 WebSocket Signaling API

**Endpoint**: `ws://localhost:8080/ws`

#### Message Types

**1. Client → Server: Join Session**
```json
{
  "type": "join",
  "sessionId": "unique-session-id",
  "userId": "user-123"
}
```

**2. Server → Client: Ready**
```json
{
  "type": "ready",
  "sessionId": "unique-session-id"
}
```

**3. Client → Server: WebRTC Offer**
```json
{
  "type": "offer",
  "sessionId": "unique-session-id",
  "sdp": "v=0\r\no=- ... (SDP string)"
}
```

**4. Server → Client: WebRTC Answer**
```json
{
  "type": "answer",
  "sessionId": "unique-session-id",
  "sdp": "v=0\r\no=- ... (SDP string)"
}
```

**5. Bidirectional: ICE Candidate**
```json
{
  "type": "candidate",
  "sessionId": "unique-session-id",
  "candidate": {
    "candidate": "candidate:...",
    "sdpMid": "0",
    "sdpMLineIndex": 0
  }
}
```

**6. Server → Client: Transcription Result**
```json
{
  "type": "transcription",
  "sessionId": "unique-session-id",
  "text": "مرحبا، كيف حالك؟",
  "timestamp": "2025-01-15T10:30:45Z",
  "language": "ar",
  "confidence": 0.92
}
```

**7. Server → Client: Error**
```json
{
  "type": "error",
  "message": "Failed to process audio",
  "code": "TRANSCRIPTION_ERROR"
}
```

---

## 6. Implementation Requirements

### 6.1 Frontend Implementation

#### Requirements
1. **Audio Capture**
    - Request microphone access via `getUserMedia`
    - Use Opus codec (48kHz, mono)
    - Handle permission errors gracefully

2. **SimplePeer Setup**
   ```javascript
   const peer = new SimplePeer({
     initiator: true,
     stream: audioStream,
     config: {
       iceServers: [
         { urls: 'stun:stun.l.google.com:19302' }
       ]
     },
     offerOptions: {
       offerToReceiveAudio: false,
       offerToReceiveVideo: false
     }
   });
   ```

3. **Signaling**
    - Connect to WebSocket on page load
    - Send offer/answer through signaling server
    - Handle ICE candidates
    - Reconnect on disconnect

4. **UI Components**
    - Start/Stop recording button
    - Real-time transcription display (auto-scroll)
    - Connection status indicator
    - Language selector (Arabic/Auto)
    - Clear transcription button
    - Session ID display

5. **Transcription Display**
    - Show timestamps with each segment
    - Highlight new text briefly
    - Support RTL for Arabic text
    - Word-by-word appearance animation

#### Example Frontend Structure
```javascript
class TranscriptionApp {
  constructor() {
    this.ws = null;
    this.peer = null;
    this.sessionId = this.generateSessionId();
    this.isRecording = false;
  }

  async init() {
    await this.connectSignaling();
    await this.setupAudio();
    this.setupUI();
  }

  async connectSignaling() {
    // WebSocket connection
  }

  async setupAudio() {
    // getUserMedia + SimplePeer
  }

  handleTranscription(data) {
    // Display transcription
  }
}
```

---

### 6.2 Backend Implementation

#### 6.2.1 Signaling Server

**File**: `backend/signaling/handler.go`

Requirements:
- Gorilla WebSocket for connections
- Session management (map of sessionId → connection)
- Handle concurrent connections safely (use sync.Mutex)
- Broadcast messages to specific sessions
- Cleanup on disconnect

```go
type SignalingServer struct {
    sessions map[string]*Session
    mu       sync.RWMutex
}

type Session struct {
    ID         string
    Connection *websocket.Conn
    PeerConn   *webrtc.PeerConnection
    Created    time.Time
}
```

#### 6.2.2 Pion WebRTC Setup

**File**: `backend/webrtc/peer.go`

Requirements:
1. Create MediaEngine with Opus codec support
2. Setup InterceptorRegistry for RTP packet capture
3. Create API with custom MediaEngine and Interceptor
4. Configure ICE servers
5. Handle track events
6. Setup datachannel for metadata (optional)

```go
func CreatePeerConnection(audioHandler func([]byte)) (*webrtc.PeerConnection, error) {
    // MediaEngine setup
    m := &webrtc.MediaEngine{}
    if err := m.RegisterCodec(...); err != nil {
        return nil, err
    }

    // Interceptor registry
    i := &interceptor.Registry{}
    
    // Add custom audio interceptor
    audioInterceptor := NewAudioInterceptor(audioHandler)
    i.Add(audioInterceptor)

    // Create API
    api := webrtc.NewAPI(
        webrtc.WithMediaEngine(m),
        webrtc.WithInterceptorRegistry(i),
    )

    // Create PeerConnection
    return api.NewPeerConnection(config)
}
```

#### 6.2.3 RTP Audio Interceptor

**File**: `backend/webrtc/interceptor.go`

Requirements:
1. Implement `interceptor.Interceptor` interface
2. Capture RTP packets from audio track
3. Extract Opus payload
4. Decode Opus to PCM
5. Resample 48kHz → 16kHz
6. Buffer audio for Whisper processing
7. Handle packet loss gracefully

```go
type AudioInterceptor struct {
    interceptor.NoOp
    audioBuffer chan []float32
}

func (a *AudioInterceptor) BindRemoteStream(info *interceptor.StreamInfo, reader interceptor.RTPReader) interceptor.RTPReader {
    if info.MimeType == "audio/opus" {
        return interceptor.RTPReaderFunc(func(b []byte, attr interceptor.Attributes) (int, interceptor.Attributes, error) {
            n, attr, err := reader.Read(b, attr)
            if err != nil {
                return n, attr, err
            }

            // Parse RTP packet
            packet := &rtp.Packet{}
            if err := packet.Unmarshal(b[:n]); err == nil {
                // Decode Opus payload
                pcmData := decodeOpus(packet.Payload)
                
                // Resample to 16kHz
                resampled := resample48to16(pcmData)
                
                // Send to buffer
                select {
                case a.audioBuffer <- resampled:
                default:
                    // Buffer full, skip
                }
            }

            return n, attr, err
        })
    }
    return reader
}
```

#### 6.2.4 Opus Decoder

**Libraries needed**:
- `github.com/pion/opus` or
- `gopkg.in/hraban/opus.v2`

Requirements:
- Decode Opus frames (48kHz, 20ms frames)
- Convert to float32 PCM
- Handle decoder errors

#### 6.2.5 Audio Resampling

**Library**: `github.com/zaf/resample` or custom implementation

Requirements:
- Resample from 48kHz to 16kHz
- Maintain audio quality
- Low latency processing

#### 6.2.6 Whisper Integration

**File**: `backend/transcription/whisper.go`

Requirements:
1. Load Whisper model on startup (cache in memory)
2. Configure for Arabic language
3. Process audio in chunks (3-5 seconds)
4. Return transcription segments with timestamps
5. Handle processing errors
6. Thread-safe for concurrent sessions

```go
type WhisperEngine struct {
    model whisper.Model
    mu    sync.Mutex
}

func NewWhisperEngine(modelPath string) (*WhisperEngine, error) {
    model, err := whisper.New(modelPath)
    if err != nil {
        return nil, err
    }
    
    return &WhisperEngine{model: model}, nil
}

func (w *WhisperEngine) Transcribe(audioSamples []float32, language string) (string, error) {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    ctx, err := w.model.NewContext()
    if err != nil {
        return "", err
    }
    
    // Configure
    ctx.SetLanguage(language) // "ar" or "auto"
    ctx.SetTranslate(false)
    
    // Process
    if err := ctx.Process(audioSamples, nil, nil); err != nil {
        return "", err
    }
    
    // Collect results
    var result strings.Builder
    for {
        segment, err := ctx.NextSegment()
        if err != nil {
            break
        }
        result.WriteString(segment.Text)
        result.WriteString(" ")
    }
    
    return result.String(), nil
}
```

#### 6.2.7 Audio Buffer and Processor

**File**: `backend/transcription/audio_buffer.go`

Requirements:
1. Accumulate audio samples
2. Process when buffer reaches threshold (3-5 seconds)
3. Maintain overlap between chunks for context
4. Send results via WebSocket

```go
type AudioProcessor struct {
    whisper     *WhisperEngine
    buffer      []float32
    chunkSize   int // samples (e.g., 16000 * 3 = 3 seconds)
    overlap     int // samples (e.g., 16000 * 1 = 1 second)
    sessionId   string
    sendResult  func(string) error
}

func (p *AudioProcessor) AddSamples(samples []float32) {
    p.buffer = append(p.buffer, samples...)
    
    // Process if buffer is full
    if len(p.buffer) >= p.chunkSize {
        go p.processChunk()
    }
}

func (p *AudioProcessor) processChunk() {
    chunk := p.buffer[:p.chunkSize]
    
    // Transcribe
    text, err := p.whisper.Transcribe(chunk, "ar")
    if err != nil {
        log.Printf("Transcription error: %v", err)
        return
    }
    
    // Send result
    if text != "" {
        p.sendResult(text)
    }
    
    // Keep overlap
    p.buffer = p.buffer[p.chunkSize-p.overlap:]
}
```

---

## 7. Configuration

### 7.1 Server Configuration

**File**: `backend/config.go`

```go
type Config struct {
    // Server
    Port            int    `env:"PORT" envDefault:"8080"`
    Host            string `env:"HOST" envDefault:"0.0.0.0"`
    
    // Whisper
    ModelPath       string `env:"MODEL_PATH" envDefault:"./models/ggml-base.bin"`
    Language        string `env:"LANGUAGE" envDefault:"ar"`
    ChunkDuration   int    `env:"CHUNK_DURATION" envDefault:"3"` // seconds
    
    // WebRTC
    STUNServers     []string `env:"STUN_SERVERS" envDefault:"stun:stun.l.google.com:19302"`
    
    // Audio
    SampleRate      int `env:"SAMPLE_RATE" envDefault:"16000"`
}
```

---

## 8. Setup and Deployment

### 8.1 Prerequisites

```bash
# Install Go 1.21+
# Install Whisper dependencies
brew install opus portaudio

# Clone Whisper.cpp
git clone https://github.com/ggerganov/whisper.cpp
cd whisper.cpp
make
```

### 8.2 Download Models

**File**: `scripts/download-models.sh`

```bash
#!/bin/bash

mkdir -p backend/models

# Download base model (recommended)
wget https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin \
  -O backend/models/ggml-base.bin

# Download small model (optional, better accuracy)
wget https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin \
  -O backend/models/ggml-small.bin

echo "Models downloaded successfully!"
```

### 8.3 Build and Run

```bash
# Backend
cd backend
go mod download
go build -o transcription-server

# Download models
chmod +x ../scripts/download-models.sh
../scripts/download-models.sh

# Run server
./transcription-server

# Frontend (separate terminal)
cd frontend
python3 -m http.server 3000
# Or use: npx serve
```

---

## 9. Testing Requirements

### 9.1 Unit Tests

**Backend tests** (`*_test.go`):
1. Signaling message parsing
2. Opus decoding accuracy
3. Resampling quality
4. Whisper transcription with sample audio
5. Audio buffer management

### 9.2 Integration Tests

1. End-to-end WebRTC connection
2. Audio streaming and reception
3. Transcription accuracy with test audio files
4. Concurrent session handling
5. Reconnection scenarios

### 9.3 Performance Tests

1. Measure processing latency per chunk
2. Memory usage per session
3. Maximum concurrent sessions
4. Audio quality degradation test
5. Network interruption handling

---

## 10. Error Handling

### Critical Error Scenarios

1. **Microphone access denied**
    - Display clear error message
    - Provide instructions to enable

2. **WebRTC connection failure**
    - Retry with exponential backoff
    - Fall back to different ICE servers

3. **Whisper model not found**
    - Check model path on startup
    - Provide download instructions

4. **Audio buffer overflow**
    - Drop oldest chunks
    - Log warning

5. **WebSocket disconnect**
    - Auto-reconnect
    - Preserve session state

6. **High CPU usage**
    - Implement rate limiting
    - Skip frames if necessary

---

## 11. Monitoring and Logging

### Required Metrics

```go
type Metrics struct {
    ActiveSessions    int64
    TotalTranscribed  int64
    AvgLatency        time.Duration
    AudioBytesRecv    int64
    ErrorCount        int64
}
```

### Logging Levels

- **INFO**: Session start/end, transcription results
- **WARN**: Buffer overflow, high latency
- **ERROR**: Connection failures, transcription errors
- **DEBUG**: RTP packets, WebRTC state changes

---

## 12. Security Considerations

1. **Rate Limiting**: Max 5 sessions per IP
2. **Session Timeout**: Auto-close after 1 hour
3. **Input Validation**: Sanitize all WebSocket messages
4. **CORS**: Configure allowed origins
5. **TLS**: Use HTTPS in production (Let's Encrypt)

---

## 13. Performance Optimization

### Backend
1. Pool Whisper contexts (avoid recreation)
2. Use buffered channels for audio
3. Goroutine pool for processing
4. Profile with pprof

### Frontend
1. Use Web Workers for audio processing
2. Implement virtual scrolling for long transcripts
3. Debounce UI updates
4. Lazy load SimplePeer

---

## 14. Future Enhancements

1. **Speaker Diarization**: Who spoke when
2. **Punctuation Restoration**: Add punctuation to output
3. **Translation**: Arabic → English
4. **Recording**: Save audio + transcription
5. **Multi-language**: Support language switching
6. **Cloud Deployment**: Docker + Kubernetes
7. **Mobile App**: React Native client
8. **Dashboard**: Analytics and monitoring UI

---

## 15. Example Usage

### Starting a Session

```javascript
// Frontend
const app = new TranscriptionApp();
await app.init();
await app.startRecording();

// User speaks: "مرحبا، كيف حالك؟"
// After ~3 seconds, transcription appears on screen
```

### Expected Flow

```
t=0s:   User clicks "Start Recording"
t=0.1s: Microphone access granted
t=0.2s: WebRTC connection established
t=0.3s: Audio streaming begins
t=3.3s: First transcription chunk processed
t=3.4s: "مرحبا" appears on screen
t=6.4s: "كيف حالك؟" appears on screen
```

---

## 16. Dependencies

### Backend Go Modules

```go
module github.com/yourusername/webrtc-transcription

go 1.21

require (
    github.com/ggerganov/whisper.cpp/bindings/go v0.0.0-latest
    github.com/gorilla/websocket v1.5.1
    github.com/pion/interceptor v0.1.25
    github.com/pion/opus v0.0.0-latest
    github.com/pion/rtcp v1.2.12
    github.com/pion/rtp v1.8.3
    github.com/pion/webrtc/v4 v4.0.0
    github.com/zaf/resample v1.0.0
)
```

### Frontend

```html
<!-- SimplePeer -->
<script src="https://cdn.jsdelivr.net/npm/simple-peer@9.11.1/simplepeer.min.js"></script>

<!-- Tailwind CSS -->
<script src="https://cdn.tailwindcss.com"></script>
```

---

## 17. Quick Start Checklist

- [ ] Clone repository
- [ ] Install Go dependencies: `go mod download`
- [ ] Download Whisper models: `./scripts/download-models.sh`
- [ ] Build backend: `go build -o server`
- [ ] Start server: `./server`
- [ ] Open frontend: Navigate to `http://localhost:3000`
- [ ] Grant microphone permission
- [ ] Click "Start Recording"
- [ ] Speak in Arabic
- [ ] See transcription appear in real-time

---

## 18. Support and Documentation

### Useful Links

- Pion WebRTC: https://github.com/pion/webrtc/tree/master/examples
- SimplePeer: https://github.com/feross/simple-peer#api
- Whisper.cpp: https://github.com/ggerganov/whisper.cpp
- Go Opus: https://github.com/pion/opus

### Common Issues

**Issue**: Audio sounds distorted
- **Solution**: Check sample rate conversion, ensure 48kHz → 16kHz

**Issue**: High latency (>5 seconds)
- **Solution**: Reduce chunk size, use smaller Whisper model

**Issue**: WebRTC connection fails
- **Solution**: Check firewall, verify STUN server accessibility

**Issue**: Empty transcriptions
- **Solution**: Verify Opus decoding, check audio levels

---

## 19. Success Criteria

The project is complete when:

1. ✅ Browser captures audio and streams via WebRTC
2. ✅ Backend receives RTP packets and decodes Opus
3. ✅ Whisper transcribes Arabic speech accurately (>80%)
4. ✅ Transcription appears in browser within 4 seconds
5. ✅ System handles 5+ concurrent sessions
6. ✅ Audio quality is clear and intelligible
7. ✅ Error handling gracefully manages failures
8. ✅ Code is documented and testable

---

## 20. Contact and Questions

For implementation questions:
- Check Pion examples for WebRTC patterns
- Refer to Whisper.cpp issues for transcription problems
- Test audio pipeline incrementally (capture → decode → transcribe)

---

**Document Version**: 1.0  
**Last Updated**: 2025-01-15  
**Target Platform**: macOS (M1 Pro), but portable to Linux/Windows