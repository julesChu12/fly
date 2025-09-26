# Repository Guidelines

## Project Structure & Module Organization
Mora is a Go capability library with three top-level workspaces: `pkg/` for framework-neutral modules, `adapters/` for HTTP framework bindings, and `starter/` for runnable demos. Populate each subpackage with cohesive logic (e.g., `pkg/auth` for JWT helpers, `pkg/logger` for tracing-friendly logging) and keep cross-imports minimal. Place shared documentation, ADRs, and configuration samples under `docs/` as they are produced.

## Build, Test, and Development Commands
- `go mod tidy` keeps `go.mod` stable after you add dependencies.
- `go test ./...` runs all package tests; prefer running it before every push.
- `go vet ./...` catches common anti-patterns and should accompany feature work.
- `go run ./starter/gin-starter` boots the sample service once the entrypoint is in place; mirror this pattern for new starters.

## Coding Style & Naming Conventions
Target Go 1.25.1 and rely on `gofmt` or `go fmt ./...` before committing. Favor short, lower-case package names, PascalCase for exported APIs, and camelCase for internals. Keep modules decoupled from frameworks; expose constructors like `NewLogger` and accept interface dependencies. Document any non-obvious behavior inline with concise comments.

## Testing Guidelines
Co-locate `*_test.go` files with the code they verify, using table-driven tests and subtests (`t.Run`) for coverage. Mocks should live in the same package when only used there, or under `_test` helper subfolders if shared. Run `go test -cover ./pkg/...` to track coverage trends and capture regression gaps in PR notes.

## Commit & Pull Request Guidelines
Follow Conventional Commits (see `chore: init Mora project`) and keep messages under 72 characters. Each PR should link relevant issues, outline capability changes, and attach console output for `go test` runs. Include configuration updates in the same PR, and request reviewers from maintainers of the touched packages.

## Configuration & Security Tips
Store sample YAML or `.env.example` files under `config/` and never commit real secrets. Document required environment variables in `docs/` so adapters and starters stay reproducible. Rotate any leaked credentials immediately and update onboarding docs.
