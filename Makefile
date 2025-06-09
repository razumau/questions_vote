.PHONY: build test clean run dev

# Build the application
build:
	go build -o bin/bot ./cmd/bot
	go build -o bin/admin ./cmd/admin

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f bin/bot bin/admin

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
