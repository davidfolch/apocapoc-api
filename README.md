# Apocapoc

Self-hosted habit tracking service with a clean, hexagonal architecture.

## Features

- **Multiple habit types**: Boolean (check), Counter, Value
- **Flexible scheduling**: Daily, Weekly, Monthly with specific days
- **Carry-over support**: Choose if incomplete habits persist or expire
- **Full history tracking**: Complete audit trail of all interactions
- **Statistics**: Track streaks, completion rates, and progress
- **Self-hosted first**: Easy deployment with SQLite
- **Security**: JWT authentication, rate limiting on auth endpoints
- **API Documentation**: Interactive Swagger UI

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

**Available image tags:**
- `latest`: Latest stable release (recommended for production)
- `1`, `1.0`, `1.0.0`: Specific version tags
- `edge`: Latest development build from main branch (unstable)
- `sha-abc123`: Specific commit (for debugging)

**Configuration options:**
- `JWT_SECRET`: **Required**. Use a long random string
- `JWT_EXPIRY`: Token expiration (e.g., `24h`, `48h`)
- `REFRESH_TOKEN_EXPIRY`: Refresh token expiration (e.g., `168h` = 7 days)
- `CORS_ORIGINS`: Comma-separated list of allowed origins
- `DEFAULT_TIMEZONE`: Timezone for date calculations (e.g., `UTC`, `Europe/Madrid`)

### Using the binary

1. Download the latest release from [GitHub Releases](https://github.com/davidfolch/apocapoc-api/releases)
2. Extract the archive:
   ```bash
   tar -xzf apocapoc-api_*_linux_amd64.tar.gz
   ```
3. Copy `.env.example` to `.env` and configure
4. Run the binary:
   ```bash
   ./apocapoc-api
   ```

The API will be available at `http://localhost:8080`

**Note:** Linux binaries only (amd64 and arm64). For other platforms, use Docker.

## Development

### Prerequisites

- Go 1.23+
- SQLite
- Docker (optional)

### Running locally with Docker (Recommended)

```bash
cp docker-compose.example.yml docker-compose.yml
docker-compose up --build
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
