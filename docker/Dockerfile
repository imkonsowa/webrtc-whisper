FROM golang:1.25-bookworm AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY backend/ ./backend/

RUN CGO_ENABLED=0 go build -o transcription-server ./backend

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/transcription-server .
COPY frontend/ ./frontend

ENV PORT=8080
ENV HOST=0.0.0.0
ENV WHISPER_ENDPOINT=http://whisper:8000/v1/audio/transcriptions

EXPOSE 8080

CMD ["./transcription-server"]
