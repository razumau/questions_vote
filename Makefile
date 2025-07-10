.PHONY: build test clean run dev

# Build the application
build:
	go build -o bin/bot ./cmd/bot
	go build -o bin/admin ./cmd/admin
	go build -o bin/importer ./cmd/importer
	go build -o bin/tournament_manager ./cmd/tournament_manager

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f bin/bot bin/admin bin/importer bin/tournament_manager

# Run the bot locally
run:
	go run ./cmd/bot

# Run in development mode with auto-reload
dev:
	go run ./cmd/bot

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy
