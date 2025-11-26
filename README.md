# Apocapoc

Self-hosted habit tracking service with a clean, hexagonal architecture.

## Features

- **Multiple habit types**: Boolean (check), Counter, Value
- **Flexible scheduling**: Daily, Weekly, Monthly with specific days
- **Carry-over support**: Choose if incomplete habits persist or expire
- **Full history tracking**: Complete audit trail of all interactions
- **Charts & Analytics**: Heatmaps, line charts, and statistics
- **Self-hosted first**: Easy deployment with SQLite

## Quick Start

### Using Docker Compose (Recommended)

1. Create a `docker-compose.yml` file:

```yaml
services:
  api:
    image: ghcr.io/davidfolch/apocapoc-api:latest
    ports:
      - "8080:8080"
    environment:
      - DB_PATH=/data/apocapoc.db
      - JWT_SECRET=YOUR_SECRET_HERE
      - JWT_EXPIRY=24h
      - REFRESH_TOKEN_EXPIRY=168h
      - CORS_ORIGINS=http://localhost:3000
      - DEFAULT_TIMEZONE=UTC
    volumes:
      - habit-data:/data
    restart: unless-stopped

volumes:
  habit-data:
```

2. **Important**: Replace `YOUR_SECRET_HERE` with a secure random string for `JWT_SECRET`

3. Start the service:

```bash
docker-compose up -d
```

The API will be available at `http://localhost:8080`

**Configuration options:**
- `JWT_SECRET`: **Required**. Use a long random string
- `JWT_EXPIRY`: Token expiration (e.g., `24h`, `48h`)
- `REFRESH_TOKEN_EXPIRY`: Refresh token expiration (e.g., `168h` = 7 days)
- `CORS_ORIGINS`: Comma-separated list of allowed origins
- `DEFAULT_TIMEZONE`: Timezone for date calculations (e.g., `UTC`, `Europe/Madrid`)

### Using the binary

1. Download the latest release
2. Copy `.env.example` to `.env` and configure
3. Run: `./apocapoc-api`

The API will be available at `http://localhost:8080`

## Development

### Prerequisites

- Go 1.23+
- SQLite
- Docker (optional)

### Running locally with Docker (Recommended)

```bash
cp docker-compose.dev.example.yml docker-compose.dev.yml
docker-compose -f docker-compose.dev.yml up --build
```

The API will be available at `http://localhost:8080`

### Running locally with Go

```bash
cp .env.example .env
go run cmd/api/main.go
```

## API Documentation

Once running, visit `http://localhost:8080/api/v1/docs` for interactive Swagger documentation.

## Architecture

This project follows hexagonal (ports & adapters) architecture:

- `domain/`: Core business logic and entities
- `application/`: Use cases (commands & queries)
- `infrastructure/`: External adapters (database, HTTP, etc.)
- `shared/`: Common utilities and errors

## License

MIT
