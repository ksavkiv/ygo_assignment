# Destination Data Aggregation API

Go REST API for managing travel destination data, backed by PostgreSQL and Redis.

## Prerequisites

- Go 1.23+
- Docker & Docker Compose

## Quick Start

```bash
# Start Postgres and Redis
docker compose up -d

# Apply migrations
psql "postgres://postgres:postgres@localhost:5432/destinations?sslmode=disable" -f migrations/000001_create_destinations.up.sql

# Run the server
go run ./cmd/server
```

The API starts on `:8080` by default.

## API Endpoints

| Method | Path                | Description            |
|--------|---------------------|------------------------|
| GET    | /destinations       | List all destinations  |
| POST   | /destinations       | Create a destination   |
| GET    | /destinations/{id}  | Get a destination      |
| PUT    | /destinations/{id}  | Update a destination   |
| DELETE | /destinations/{id}  | Delete a destination   |

## Configuration

| Variable      | Default                                                                  |
|---------------|--------------------------------------------------------------------------|
| SERVER_ADDR   | :8080                                                                    |
| DATABASE_URL  | postgres://postgres:postgres@localhost:5432/destinations?sslmode=disable  |
| REDIS_ADDR    | localhost:6379                                                           |

## Project Structure

```
├── cmd/server/          # Application entrypoint
├── internal/
│   ├── api/             # HTTP handlers & routing
│   ├── destination/     # Core business logic & models
│   ├── storage/         # PostgreSQL repository
│   ├── cache/           # Redis caching layer
│   └── config/          # Configuration loading
├── migrations/          # SQL migration files
├── ai-logs/             # Claude Code conversation logs
├── docker-compose.yml
├── go.mod
└── README.md
```
