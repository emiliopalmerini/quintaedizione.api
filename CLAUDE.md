# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
make run          # Build binary + start Docker postgres + run API
make build        # Compile binary to bin/api
make fmt          # go fmt ./...
make vet          # go vet ./...
make docker-up    # Start postgres + pgadmin containers
make docker-down  # Stop containers
```

## Test Commands

```bash
make test                              # Unit tests (./internal/...)
make test-e2e                          # E2E tests (requires Docker, uses testcontainers)
make bench                             # Run benchmarks
go test ./internal/classi/...          # Run tests for a single module
go test -run TestListClassi ./internal/classi/  # Run a single test
go test -tags e2e -v .                 # Run E2E tests directly
```

## Architecture

Go REST API for D&D 5th Edition, following **hexagonal architecture**. The `classi` module is the reference implementation for all domain modules.

### Module structure (ports & adapters)

```
internal/<module>/
├── models.go            # Domain entities and enums
├── interfaces.go        # Repository port (interface defined by domain)
├── service.go           # Business logic, depends on Repository interface
├── errors.go            # Domain-specific error helpers
├── responses.go         # API response DTOs
├── persistence/
│   └── postgres.go      # Repository adapter (implements interface)
└── transports/
    └── http.go          # HTTP adapter (depends on a Service interface, not concrete type)
```

### Wiring flow

`cmd/api/main.go` → `internal/app/app.go` creates `Dependencies{DB, Logger, Config}`, then for each module: `repo := persistence.New(db)` → `svc := module.NewService(repo, logger)` → `handler := transports.NewHandler(svc)` → `r.Mount("/v1/<module>", handler.Routes())`.

### Shared utilities (`internal/shared/`)

- **errors.go**: `AppError` wraps HTTP status + JSON error response. Constructors: `NewBadRequestError`, `NewNotFoundError`, `NewInternalError`.
- **httputil.go**: `WriteJSON(w, status, data)`, `WriteError(w, err)` (auto-logs 5xx as Error, 4xx as Warn).
- **pagination.go**: `ListFilter` parsed from query params (`nome`, `sort`, `$limit`, `$offset`). Default limit 20, max 100.
- **validation.go**: Uses `go-playground/validator`. `ValidateID(name, value)` for URL params. Custom `slug` validator.
- **dbutil.go**: `EscapeLike(s)` for safe LIKE/ILIKE queries.

### Error response format

```json
{"errors": [{"code": "NOT_FOUND", "title": "Not Found", "detail": "Classe with id 'x' not found"}]}
```

## Conventions

- **JSON tags**: kebab-case (`json:"id-classe"`). **DB tags**: snake_case (`db:"id_classe"`).
- **Pagination response DTOs** embed `shared.PaginationMeta` and expose `pagina`, `numero-di-elementi` fields.
- **Repository returns `nil`** (not error) when entity not found; service converts to `NewNotFoundError`.
- **Service constructors** accept a nil-safe logger (fallback to discard handler).
- **URL param names** use kebab-case matching JSON field names.
- **Migrations** in `migrations/` as `000NNN_description.{up,down}.sql`, auto-run in dev mode (`APP_VERSION=dev`).

## Testing patterns

- **Unit tests**: Mock repositories/services via struct with function fields (no framework). See `mock_repository_test.go`.
- **Integration tests** (`persistence/*_test.go`): Real PostgreSQL via `testcontainers-go`. Helper `setupTestDB(t)` + seed functions.
- **E2E tests** (`e2e_test.go`, build tag `e2e`): Full HTTP flow with `TestMain` sharing a single container. Helper `get(t, path)` for HTTP requests.
- **Benchmarks**: `*_bench_test.go` files for service and handler performance.
- Use stdlib `testing` only — no assertion libraries.
