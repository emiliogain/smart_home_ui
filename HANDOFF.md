# Project Handoff Context

## What Was Already Done (Phase 0 -- completed)

- Deleted `backend/internal/model/` (dead duplicate of domain entities)
- Removed `.idea/` from tracking
- Fixed `go.mod` -- resolved real dependency tree via `go mod tidy`. Only `gin` is a direct dependency currently. Other planned deps (pgx, goose, viper, zap, testify, websocket) are listed as comments in `go.mod` and will auto-add when imported.
- Created `backend/.golangci.yml` with linter config (errcheck, govet with shadow checking, staticcheck, goimports, etc.)
- Created `backend/scripts/pre-commit.sh` (runs `go vet`, `golangci-lint`, `go test`)
- Created `backend/migrations/` directory for goose migrations
- Updated `backend/Makefile` with migration targets (`migrate-up`, `migrate-down`, `migrate-create`, `migrate-status`), `vet`, `test-race`, and `check` (combined pre-commit)
- Updated `.gitignore` with Go binary patterns
- Wrote **skeleton** `backend/cmd/smart-home-backend/main.go` -- has all 10 composition steps as TODO comments, graceful shutdown is fully working
- Wrote **skeleton** `backend/internal/config/config.go` -- `Config` struct with fields, `Load()` function signature with TODO for Viper implementation
- Project compiles cleanly (`go build ./cmd/smart-home-backend` succeeds)
- Go is installed at `/usr/local/go/bin/go` (not in default PATH in shell, needs `export PATH="/usr/local/go/bin:$PATH"`)

## What I Want To Do Next

- I want to **write the code myself** with guidance/hints from you
- Phase 1 is next: implement `config.go` with Viper, set up PostgreSQL connection, create goose migrations, implement real pgx repository adapters
- The plan file (attached separately) has the full schema SQL, architecture details, and workload distribution

## Key Architecture Decision

Pure hexagonal architecture with `domain/` (no deps), `ports/` (interfaces), `app/` (orchestration), `adapters/` (infrastructure). Domain entities use unexported fields with getters and constructor validation.

## Module Path

`github.com/emiliogain/smart-home-backend`

---

## Key Files Reference

| File | What it contains |
|---|---|
| `backend/cmd/smart-home-backend/main.go` | Skeleton with all TODO steps |
| `backend/internal/config/config.go` | Config struct + `Load()` signature to implement |
| `backend/internal/domain/sensor/entity.go` | Full sensor domain entity (~217 lines) |
| `backend/internal/domain/device/entity.go` | Full device domain entity (~419 lines) |
| `backend/internal/ports/secondary/sensor_repository.go` | Repository interface contract |
| `backend/internal/ports/secondary/device_repository.go` | Repository interface contract |
| `backend/internal/adapters/secondary/database/sensor_repository.go` | Current in-memory impl (to be replaced with pgx) |
| `backend/configs/config.example.yaml` | Example config file format |
| `backend/Makefile` | All build/migrate/test commands |
