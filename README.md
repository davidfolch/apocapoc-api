# Apocapoc

Self-hosted habit tracking service with a clean, hexagonal architecture.

## Features

- **Multiple habit types**: Boolean (check), Counter, Value
- **Flexible scheduling**: Daily, Weekly, Monthly with specific days
- **Carry-over support**: Choose if incomplete habits persist or expire
- **Full history tracking**: Complete audit trail of all interactions
- **Statistics**: Track streaks, completion rates, and progress
- **Self-hosted first**: Easy deployment with SQLite
- **Security**: JWT authentication, rate limiting on auth endpoints, optional email verification
- **Email notifications**: Optional welcome emails and verification emails
- **Registration control**: Open or closed registration modes
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

The API will be available at `http://localhost:8080`

**Available image tags:**
- `latest`: Latest stable release (recommended for production)
- `1`, `1.0`, `1.0.0`: Specific version tags
- `edge`: Latest development build from main branch (unstable)
- `sha-abc123`: Specific commit (for debugging)

**Configuration options:**

**Required:**
- `JWT_SECRET`: **Required**. Use a long random string
- `DB_PATH`: Database file path (default: `./data/apocapoc.db`)

**Application:**
- `PORT`: HTTP port (default: `8080`)
- `APP_URL`: Public URL for email links (e.g., `https://habits.yourdomain.com`)
- `DEFAULT_TIMEZONE`: Timezone for date calculations (e.g., `UTC`, `Europe/Madrid`)

**Authentication:**
- `JWT_EXPIRY`: Token expiration (e.g., `1h`, `24h`)
- `REFRESH_TOKEN_EXPIRY`: Refresh token expiration (e.g., `7d`, `168h`)

**Registration:**
- `REGISTRATION_MODE`: `open` (anyone can register) or `closed` (registration disabled)

**Email (optional - all or none):**
- `SMTP_HOST`: SMTP server hostname
- `SMTP_PORT`: SMTP port (`587` for STARTTLS, `465` for SSL)
- `SMTP_USER`: SMTP username
- `SMTP_PASSWORD`: SMTP password (use `$$` to escape `$` in passwords)
- `SMTP_FROM`: From address for emails
- `SUPPORT_EMAIL`: Support email shown in emails (default: `contact@apocapoc.app`)
- `SEND_WELCOME_EMAIL`: Send welcome email after verification (`true`/`false`)

**Note:** If SMTP is not configured, email verification is skipped and users are auto-verified.

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

## Support

- 📧 Email: contact@apocapoc.app
- 🐛 Issues: [GitHub Issues](https://github.com/davidfolch/apocapoc-api/issues)

## License

MIT
