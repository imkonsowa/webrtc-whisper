.PHONY: up down build run clean lint test

up:
	docker compose -f docker/docker-compose.yml up --build -d

down:
	docker compose -f docker/docker-compose.yml down

build:
	go build -o backend/transcription-server ./backend

run: build
	cd backend && FRONTEND_DIR=../frontend ./transcription-server

clean:
	rm -f backend/transcription-server

lint:
	golangci-lint run ./...

test:
	go test -v ./...
