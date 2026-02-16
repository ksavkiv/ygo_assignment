# Claude Code Rules

## Conversation Logging

Log every user prompt and assistant response to `ai-logs/conversation.md`. Each entry should be numbered sequentially and include the full prompt and a summary of the response.

## Development Workflow

### Test-Driven Development (TDD)

All new features and bug fixes **must** follow the TDD red-green-refactor cycle:

1. **Red** — Write a failing test that defines the expected behavior
2. **Green** — Write the minimum code to make the test pass
3. **Refactor** — Clean up while keeping tests green

Never write implementation code without a corresponding test first.

### Test Conventions

- Use Go's standard `testing` package only (no testify, gomock, etc.)
- Hand-roll mock implementations of interfaces using struct fields with injectable `func` types
- Each test case gets its own `Test*` function (no table-driven tests unless clearly beneficial)
- Use `t.Errorf` / `t.Fatalf` for assertions
- Test files live alongside the code they test (e.g., `service.go` → `service_test.go`)

### Test Coverage

- **Minimum 80% coverage** is required for all packages
- Check coverage with:
  ```bash
  go test -coverprofile=coverage.out ./...
  go tool cover -func=coverage.out
  ```
- Generate HTML coverage report:
  ```bash
  go tool cover -html=coverage.out -o coverage.html
  ```

## Project Structure

```
├── cmd/server/              # Application entrypoint (main.go)
├── internal/
│   ├── api/                 # HTTP handlers, router, handler tests
│   ├── destination/         # Core business logic, models, fetcher, service tests
│   ├── storage/             # PostgreSQL repository (pgx)
│   ├── cache/               # Redis caching layer (go-redis)
│   └── config/              # Environment-based configuration, config tests
├── migrations/              # SQL migration files (up/down pairs)
├── docs/                    # API research and reference docs
├── ai-logs/                 # Claude Code conversation logs
├── docker-compose.yml       # Postgres 16 + Redis 7
├── go.mod                   # Module: destination-data-aggregation-api (Go 1.24)
└── CLAUDE.md                # This file
```

## Infrastructure & Services

### Docker Compose

Start infrastructure (Postgres + Redis):
```bash
docker compose up -d
```

Stop infrastructure:
```bash
docker compose down
```

Stop and remove volumes:
```bash
docker compose down -v
```

Services:
| Service  | Image              | Port  | Credentials         |
|----------|--------------------|-------|---------------------|
| postgres | postgres:16-alpine | 5432  | postgres / postgres |
| redis    | redis:7-alpine     | 6379  | —                   |

### Database Migrations

Apply migrations in order:
```bash
psql "postgres://postgres:postgres@localhost:5432/destinations?sslmode=disable" -f migrations/000001_create_destinations.up.sql
psql "postgres://postgres:postgres@localhost:5432/destinations?sslmode=disable" -f migrations/000002_add_metadata_jsonb.up.sql
```

### Running the Server

```bash
go run ./cmd/server
```

The API starts on `:8080` by default. Configure via environment variables:

| Variable     | Default                                                                 |
|--------------|-------------------------------------------------------------------------|
| SERVER_ADDR  | :8080                                                                   |
| DATABASE_URL | postgres://postgres:postgres@localhost:5432/destinations?sslmode=disable |
| REDIS_ADDR   | localhost:6379                                                          |

## Running Tests

Run all tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

Run tests for a specific package:
```bash
go test -v ./internal/destination/...
go test -v ./internal/api/...
go test -v ./internal/config/...
```

Run tests with coverage summary:
```bash
go test -cover ./...
```

Run tests with full coverage report (must show >80%):
```bash
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
```

**Note:** Handler tests (`internal/api/handler_test.go`) for health endpoints require Docker Compose services to be running (they ping localhost:5432 and localhost:6379).

## API Endpoints

| Method | Path                                  | Description                            |
|--------|---------------------------------------|----------------------------------------|
| GET    | /api/v1/destinations/{city}           | Get cached/stored destination data     |
| POST   | /api/v1/destinations/{city}/refresh   | Fetch fresh data from external sources |
| GET    | /api/v1/health                        | Health check (DB + Redis connectivity) |
