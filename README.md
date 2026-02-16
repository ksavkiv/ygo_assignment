# Destination Data Aggregation API

Go REST API for aggregating travel destination data from multiple external sources, backed by PostgreSQL and Redis.

## Prerequisites

- Go 1.24+
- Docker & Docker Compose

## Quick Start

```bash
# Start Postgres and Redis
docker compose up -d

# Apply migrations
psql "postgres://postgres:postgres@localhost:5432/destinations?sslmode=disable" -f migrations/000001_create_destinations.up.sql
psql "postgres://postgres:postgres@localhost:5432/destinations?sslmode=disable" -f migrations/000002_add_metadata_jsonb.up.sql

# Run the server
go run ./cmd/server
```

The API starts on `:8080` by default.

## API Endpoints

| Method | Path                                  | Description                                        |
|--------|---------------------------------------|----------------------------------------------------|
| GET    | /api/v1/destinations/{city}           | Get cached/stored destination data                 |
| POST   | /api/v1/destinations/{city}/refresh   | Fetch fresh data from external sources, store/cache |
| GET    | /api/v1/health                        | Health check (DB + Redis connectivity)             |

## Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
```

## Configuration

| Variable     | Default                                                                 |
|--------------|-------------------------------------------------------------------------|
| SERVER_ADDR  | :8080                                                                   |
| DATABASE_URL | postgres://postgres:postgres@localhost:5432/destinations?sslmode=disable |
| REDIS_ADDR   | localhost:6379                                                          |

## Project Structure

```
├── cmd/server/              # Application entrypoint
├── internal/
│   ├── api/                 # HTTP handlers & routing
│   ├── destination/         # Core business logic, models & fetcher
│   ├── storage/             # PostgreSQL repository
│   ├── cache/               # Redis caching layer
│   └── config/              # Environment-based configuration
├── migrations/              # SQL migration files (up/down pairs)
├── docs/                    # API research & reference docs
├── ai-logs/                 # Claude Code conversation logs
├── docker-compose.yml
├── go.mod
└── README.md
```
