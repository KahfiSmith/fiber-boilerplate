# Development Commands

## Build & run

| Command | Purpose |
|---|---|
| `go build ./...` | Compile all packages |
| `go run ./cmd/api` | Run the API |
| `gofmt -w .` | Format code |

## Quality

| Command | Purpose |
|---|---|
| `go test ./...` | Run tests |
| `go vet ./...` | Static analysis |

## Migrations

| Command | Purpose |
|---|---|
| `scripts/migrate.sh` | Apply migrations |
| `scripts/migrate-status.sh` | Show migration status |
| `scripts/migrate-down.sh` | Roll back migrations |

## Swagger

| Command | Purpose |
|---|---|
| `scripts/swagger-generate.sh` | Regenerate swagger artifacts |

Swagger artifacts live in `docs/swagger.yaml` / `docs/swagger.json`.
