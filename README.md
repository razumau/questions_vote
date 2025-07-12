# Questions Vote

A Telegram bot for voting on questions with an ELO rating system, converted from Python to Go.

## Build Instructions

### Prerequisites

- Go 1.24.4 or later
- SQLite3 development libraries

### Building All Binaries

To build all binaries at once:

```bash
go build -o bin/bot ./cmd/bot
go build -o bin/admin ./cmd/admin
go build -o bin/importer ./cmd/importer
go build -o bin/tournament_manager ./cmd/tournament_manager
```

### Building Individual Binaries

#### Bot Binary
The main Telegram bot application:
```bash
go build -o bin/bot ./cmd/bot
```

#### Admin Binary
Administrative tools (currently in development):
```bash
go build -o bin/admin ./cmd/admin
```

#### Importer Binary
Tool for importing questions from external sources:
```bash
go build -o bin/importer ./cmd/importer
```

#### Tournament Manager Binary
Tool for managing tournaments:
```bash
go build -o bin/tournament_manager ./cmd/tournament_manager
```

### Dependencies

Install dependencies:
```bash
go mod download
```

### Running the Applications

#### Running the Bot
```bash
./bin/bot
```

#### Running the Admin Tool
```bash
./bin/admin
```

#### Running the Importer
```bash
# List packages
./bin/importer -command=list-packages

# Import specific package
./bin/importer -command=import-package -package-id=5220

# Import all packages for a year
./bin/importer -command=import-year -year=2022
```

#### Running the Tournament Manager
```bash
# List all tournaments
./bin/tournament_manager -command=list-tournaments

# Create a new tournament
./bin/tournament_manager -command=create-tournament -earliest-date=2023-01-01 -last-date=2023-12-31 -title="2023 Tournament"

# Activate a tournament by ID
./bin/tournament_manager -command=activate-tournament -id=1

# Deactivate a tournament by ID
./bin/tournament_manager -command=deactivate-tournament -id=1
```

### Development

For development, you can run directly with Go:

```bash
# Run bot
go run ./cmd/bot

# Run admin tool
go run ./cmd/admin

# Run importer
go run ./cmd/importer -command=list-packages

# Run tournament manager
go run ./cmd/tournament_manager -command=list-tournaments
```

### Cross-compilation

To build for different platforms:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o bin/bot-linux-amd64 ./cmd/bot

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o bin/bot-windows-amd64.exe ./cmd/bot

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o bin/bot-darwin-arm64 ./cmd/bot
```