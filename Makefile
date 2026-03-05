.PHONY: run fmt vet tidy up down web

run:
	GEMINI_API_KEY=$(GEMINI_API_KEY) go run ./cmd/server

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

up:
	docker compose up -d

down:
	docker compose down

web:
	cd web && npm install && npm run dev
