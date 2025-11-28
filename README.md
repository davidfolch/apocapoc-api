# Apocapoc API - Self-Hosted Habit Tracker

**Lightweight REST API for habit tracking built with Go.** Designed for developers who want full control over their data without relying on third-party services. Deploy in minutes with Docker and start building your own productivity tools.

[![Docker](https://img.shields.io/docker/v/ghcr.io/davidfolch/apocapoc-api?label=docker&logo=docker)](https://github.com/davidfolch/apocapoc-api/pkgs/container/apocapoc-api)
[![Go Version](https://img.shields.io/github/go-mod/go-version/davidfolch/apocapoc-api)](https://golang.org/)
[![License](https://img.shields.io/github/license/davidfolch/apocapoc-api)](LICENSE)

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
  - [Docker Compose (Recommended)](#using-docker-compose-recommended)
  - [Binary Installation](#using-the-binary)
- [Use Cases](#use-cases)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Architecture](#architecture)

## Features

- **Multiple habit types**: Boolean (daily check-ins), Counter (track numbers), Value (measurements)
- **Flexible scheduling**: Daily, Weekly, Monthly with custom day selection
- **Statistics endpoints**: Streaks, completion rates, and progress tracking
- **Complete history**: Full audit trail of all interactions
- **Easy deployment**: Single Docker container or binary with embedded SQLite
- **Security**: JWT authentication, rate limiting, optional email verification
- **Registration modes**: Open or closed for controlled access
- **Interactive docs**: Built-in Swagger UI for testing endpoints

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
      - PORT=8080
      - APP_URL=http://localhost:8080
      - JWT_SECRET=YOUR_SECRET_HERE
      - JWT_EXPIRY=1h
      - REFRESH_TOKEN_EXPIRY=7d
      - DEFAULT_TIMEZONE=UTC
      - REGISTRATION_MODE=open
      # Email configuration (optional)
      # - SMTP_HOST=smtp.example.com
      # - SMTP_PORT=587
      # - SMTP_USER=your-email@example.com
      # - SMTP_PASSWORD=your-password
      # - SMTP_FROM=noreply@example.com
      # - SUPPORT_EMAIL=contact@apocapoc.app
      # - SEND_WELCOME_EMAIL=false
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

API available at `http://localhost:8080`

**Image tags:**
- `latest`: Stable release (recommended)
- `1`, `1.0`, `1.0.0`: Specific versions
- `edge`: Development build (unstable)

**Configuration:**

*Required:*
- `JWT_SECRET`: Long random string (required)
- `DB_PATH`: Database path (default: `./data/apocapoc.db`)

*Application:*
- `PORT`: HTTP port (default: `8080`)
- `APP_URL`: Public URL for email links
- `DEFAULT_TIMEZONE`: e.g., `UTC`, `Europe/Madrid`

*Authentication:*
- `JWT_EXPIRY`: e.g., `1h`, `24h`
- `REFRESH_TOKEN_EXPIRY`: e.g., `7d`, `168h`
- `REGISTRATION_MODE`: `open` or `closed`

*Email (optional):*
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`
- `SUPPORT_EMAIL`: Default `contact@apocapoc.app`
- `SEND_WELCOME_EMAIL`: `true`/`false`

Without SMTP config, users are auto-verified.

### Using the binary

1. Download from [GitHub Releases](https://github.com/davidfolch/apocapoc-api/releases)
2. Extract: `tar -xzf apocapoc-api_*_linux_amd64.tar.gz`
3. Configure: `cp .env.example .env` (edit as needed)
4. Run: `./apocapoc-api`

API available at `http://localhost:8080`

*Note: Linux only (amd64/arm64). Use Docker for other platforms.*

## Development

**Prerequisites:** Go 1.23+, SQLite, Docker (optional)

**With Docker:**
```bash
cp docker-compose.example.yml docker-compose.yml
docker-compose up --build
```

**With Go:**
```bash
cp .env.example .env
go run cmd/api/main.go
```

API runs on `http://localhost:8080`

## Use Cases

Perfect for:

- **Custom mobile/web apps**: Build your own interface without backend complexity
- **Personal dashboards**: Integrate with Grafana, Nextcloud, or Home Assistant
- **Automation workflows**: Connect to n8n, Zapier, or custom scripts
- **Privacy-focused teams**: Keep sensitive productivity data on your infrastructure
- **API learning projects**: Clean architecture with real-world examples
- **Offline-first tools**: SQLite backend works without cloud dependencies

## API Documentation

Access the interactive Swagger UI at `http://localhost:8080/api/v1/docs`

Includes endpoint reference, schemas, authentication examples, and live testing.

## Architecture

This project follows hexagonal (ports & adapters) architecture:

- `domain/`: Core business logic and entities
- `application/`: Use cases (commands & queries)
- `infrastructure/`: External adapters (database, HTTP, etc.)
- `shared/`: Common utilities and errors

## Support

- 📧 Email: contact@apocapoc.app
- 🐛 Issues: [GitHub Issues](https://github.com/davidfolch/apocapoc-api/issues)

## Keywords

`habit-tracker` `habit-tracking` `rest-api` `self-hosted` `golang` `api` `habits` `productivity` `docker` `sqlite` `hexagonal-architecture` `clean-architecture` `habit-tracker-api` `self-hosted-api` `personal-analytics` `privacy` `open-source`

## License

MIT
