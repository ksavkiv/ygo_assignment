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

# Build the server binary
go build -o server ./cmd/server

# Run it
./server
```

The API starts on `:8080` by default.

## Authentication

All endpoints require a Bearer token via the `Authorization` header. The token is configured with the `API_TOKEN` environment variable.

Requests without a valid token receive `401 Unauthorized`.

## API Endpoints

| Method | Path                                  | Description                                        |
|--------|---------------------------------------|----------------------------------------------------|
| GET    | /api/v1/destinations                  | List all stored cities                             |
| GET    | /api/v1/destinations/{city}           | Get cached/stored destination data                 |
| POST   | /api/v1/destinations/{city}/refresh   | Fetch fresh data from external sources, store/cache |
| GET    | /api/v1/health                        | Health check (DB + Redis connectivity)             |

### Example curl Commands

```bash
# Set your token
export TOKEN="super-secret-token-2025"

# Health check
curl -s http://localhost:8080/api/v1/health \
  -H "Authorization: Bearer $TOKEN" | jq

# List all stored cities
curl -s http://localhost:8080/api/v1/destinations \
  -H "Authorization: Bearer $TOKEN" | jq

# Get destination data for a city
curl -s http://localhost:8080/api/v1/destinations/brussels \
  -H "Authorization: Bearer $TOKEN" | jq

# Refresh destination data from external sources
curl -s -X POST http://localhost:8080/api/v1/destinations/brussels/refresh \
  -H "Authorization: Bearer $TOKEN" | jq

# Request without token (returns 401)
curl -s -w "\nHTTP Status: %{http_code}\n" http://localhost:8080/api/v1/health
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage summary
go test -cover ./...

# Run with full coverage report
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

# Generate HTML coverage report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
```

## Test Coverage

| Package                      | Coverage |
|------------------------------|----------|
| internal/config              | 100.0%   |
| internal/destination         | 100.0%   |
| internal/destination/pipeline| 92.9%    |
| internal/destination/feed    | 88.4%    |
| internal/api                 | 85.9%    |

Key function coverage:

| Function              | Coverage |
|-----------------------|----------|
| BearerAuth            | 100.0%   |
| GetByCity (handler)   | 100.0%   |
| Refresh (handler)     | 100.0%   |
| Health                | 92.3%    |
| Config.Load           | 100.0%   |
| Service.GetByCity     | 100.0%   |
| Service.Refresh       | 100.0%   |
| APIFetcher.Fetch      | 89.7%    |
| FetchGeocode          | 93.8%    |
| Pipeline.Start        | 91.7%    |
| Pipeline.pollCity     | 100.0%   |
| Pipeline.listen       | 87.0%    |
| Pipeline.flush        | 87.5%    |

## Configuration

| Variable     | Default                                                                 |
|--------------|-------------------------------------------------------------------------|
| SERVER_ADDR  | :8080                                                                   |
| DATABASE_URL | postgres://postgres:postgres@localhost:5432/destinations?sslmode=disable |
| REDIS_ADDR   | localhost:6379                                                          |
| API_TOKEN    | _(required)_                                                            |
| LOG_LEVEL    | info                                                                    |

Copy `.env.example` to `.env` and adjust values for local development.

## Project Structure

```
├── cmd/server/              # Application entrypoint
├── internal/
│   ├── api/                 # HTTP handlers & routing
│   ├── destination/         # Domain model, interfaces & service
│   │   ├── feed/            # External API clients & fetcher implementations
│   │   └── pipeline/        # Background polling & batched upserts
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
