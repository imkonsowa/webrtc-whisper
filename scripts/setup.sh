#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR/.."

echo "installing go dependencies..."
cd "$PROJECT_DIR"
go mod tidy

echo "building server..."
go build -o backend/transcription-server ./backend

echo "setup complete."
echo ""
echo "to run locally:"
echo "  1. start a whisper server on port 9000"
echo "  2. cd backend && ./transcription-server"
echo ""
echo "or use docker-compose:"
echo "  cd docker && docker-compose up --build"
