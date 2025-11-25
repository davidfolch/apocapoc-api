# Habit Tracker API

Self-hosted habit tracking service with a clean, hexagonal architecture.

## Features

- **Multiple habit types**: Boolean (check), Counter, Value
- **Flexible scheduling**: Daily, Weekly, Monthly with specific days
- **Carry-over support**: Choose if incomplete habits persist or expire
- **Full history tracking**: Complete audit trail of all interactions
- **Charts & Analytics**: Heatmaps, line charts, and statistics
- **Self-hosted first**: Easy deployment with SQLite or PostgreSQL

## Quick Start

### Using the binary

1. Download the latest release
2. Copy `.env.example` to `.env` and configure
3. Run: `./habit-tracker-api`

The API will be available at `http://localhost:8080`

## Development

### Prerequisites

- Go 1.23+
- SQLite (or PostgreSQL)

### Running locally

```bash
cp .env.example .env
go run cmd/api/main.go
```

## API Documentation

Once running, visit `http://localhost:8080/api/v1/docs` for interactive API documentation.

## Architecture

This project follows hexagonal (ports & adapters) architecture:

- `domain/`: Core business logic and entities
- `application/`: Use cases (commands & queries)
- `infrastructure/`: External adapters (database, HTTP, etc.)
- `shared/`: Common utilities and errors

## License

MIT
