# Quinta Edizione API

REST API for D&D 5th Edition (Quinta Edizione) game data.

## Build and Run

```bash
make build          # Build binary to bin/api
make run            # Build and run with PostgreSQL (docker-compose)
go build -o bin/api ./cmd/api  # Manual build
```

## Testing

```bash
make test           # Unit tests (./internal/...)
make test-e2e       # E2E tests with testcontainers (-tags e2e)
make bench          # Run benchmarks
make bench-compare  # Compare benchmarks with baseline using benchstat
```

## Docker

```bash
make docker-up      # Start PostgreSQL container
make docker-down    # Stop containers
```

## Project Structure

- `cmd/api/` - Application entrypoint
- `internal/` - Core application code
  - `app/` - Application setup
  - `config/` - Configuration loading
  - `health/` - Health check handler
  - `middleware/` - HTTP middleware (auth, rate limiting, CORS)
  - `shared/` - Shared utilities
  - Domain modules (classi, incantesimi, mostri, oggetti, specie, talenti, etc.)
- `migrations/` - SQL migrations (golang-migrate)
- `swagger/` - OpenAPI documentation

## Tech Stack

- Go 1.25.2
- chi router (github.com/go-chi/chi/v5)
- PostgreSQL 16+ with sqlx
- golang-migrate for migrations
- testcontainers for E2E tests

## API Endpoints

Base URL: `http://localhost:8080`

| Endpoint | Description |
|----------|-------------|
| GET /health | Health check |
| GET /swagger | OpenAPI docs |
| GET /v1/classi | List classes |
| GET /v1/classi/{id} | Class details |
| GET /v1/classi/{id}/sotto-classi | Subclasses |

Query params: `nome`, `sort` (asc/desc), `$limit` (1-100), `$offset`

## Authentication

API key via `X-API-Key` header. Optional - leave `API_KEY` empty in `.env` to disable.

## Configuration

Copy `.env.example` to `.env`. Key variables:
- `API_KEY` - API key for authentication (optional)
- `RATE_LIMIT_ENABLED`, `RATE_LIMIT_RPM` - Rate limiting config
- `CORS_ALLOWED_ORIGINS`, `CORS_ALLOWED_METHODS` - CORS config

## Code Quality

```bash
make fmt            # Format code
make vet            # Run go vet (includes fmt)
go test -race ...   # Always use race detector in tests
```

CI runs: format check, vet, build, unit tests, E2E tests, and benchmark comparison.
