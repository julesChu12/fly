# Repository Guidelines

## Project Structure & Module Organization
- `cmd/userd` holds the main entrypoint; keep startup wiring light and delegate logic to internal packages.
- `internal` contains domain services for identity, security, and persistence; add new modules under clear subdirectories (e.g., `internal/auth`, `internal/session`).
- `pkg` is for shared utilities that may be reused outside Custos; avoid coupling it to internal state.
- `configs` stores environment templates such as `local.env.example`; never check secrets into the repo.
- `scripts`, `docs`, and `test` host automation helpers, architectural references, and integration fixtures respectively.

## Build, Test, and Development Commands
- `make build` cleans `./bin` and compiles the `userd` binary via `scripts/build.sh`.
- `make test` runs `go test -v -race -coverprofile=coverage.out ./...` and emits `coverage.html`.
- `make run` executes the freshly built service (`./bin/userd`).
- `make dev` copies `.env` from `configs/local.env.example`, tidies modules, and rebuilds.
- `make lint` calls `golangci-lint run` when installed; install locally for consistent checks.

## Coding Style & Naming Conventions
- Format Go code with `gofmt` (tabs, grouped imports); run `goimports` if available to maintain canonical ordering.
- Prefer explicit package names (`auth`, `tokens`, `sessions`) and clear filenames (`handler.go`, `repository.go`).
- Follow Go idioms for exported identifiers (`CamelCase`) and keep test files suffixed with `_test.go`.
- Resist adding binaries (e.g., `userd`) to version control—artifacts belong in `bin/`.

## Testing Guidelines
- Place unit tests alongside source files in `internal/...` with descriptive `TestXxx` names.
- Use the `test` directory for integration data, mocks, or longer-running suites; document setup in `docs/` if complex.
- Aim to maintain race-free tests (`-race`) and investigate coverage dips flagged by `coverage.out`.
- Generate artifacts via `make test`; open `coverage.html` locally to spot gaps before review.

## Commit & Pull Request Guidelines
- Use Conventional Commits (`feat: add session revocation`, `fix: handle expired refresh token`) to clarify intent and assist changelog tooling.
- Keep commits scoped; include migration scripts or config bumps in the same commit only when tightly coupled.
- Pull requests should summarize behavior changes, list test evidence (`make test` output), and reference tracking issues or tickets.
- Add screenshots or API traces when touching user-facing flows or HTTP contracts to ease reviewer context.
