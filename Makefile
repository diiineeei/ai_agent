.PHONY: run fmt vet tidy up down web mcp-weather build-mcp run-mcp

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

mcp-weather:
	go run ./cmd/mcp-weather

run-mcp:
	MCP_PORT=3001 go run ./cmd/mcp-weather

build-mcp:
	go build -o ./bin/mcp-weather ./cmd/mcp-weather
	@echo "Binário gerado em: $$(pwd)/bin/mcp-weather"
