# Custos Agent Handbook

## Repository Snapshot
- Repo currently contains documentation only (`README.md`, `AIPrompt.md`, `AGENTS.md`, `LICENSE`). No Go module, binaries, or source directories exist yet.
- `.gitignore` follows Go defaults with workspace-level exclusions. `.vscode/settings.json` disables ChatGPT code lenses.
- `README.md` describes the User Domain vision. `AIPrompt.md` captures the broader Mora vs User Domain context.

## System Vision
- **Mora capability library**: reusable technical modules (auth, config, log, db, cache, mq, utils) plus framework adapters.
- **User Domain service**: owns user identity lifecycle, authentication flows, security controls, RBAC/ABAC, optional OAuth provider endpoints, and observability events.
- **API layer (future)**: orchestrates domains + Mora, enforces trust and zero-trust policies.

## Immediate Priorities
1. Scaffold Go module (`go mod init ...`), establish `cmd/`, `domains/user/`, and `mora/` folders per guidelines.
2. Capture architecture decisions in `docs/` and keep `README.md` in sync as the codebase grows.
3. Define preliminary data models and interfaces for the User Domain (domain entities, repositories, services).
4. Prepare baseline configuration (`configs/.env.example`, `configs/local.env` ignored) and testing fixtures under `testdata/`.
5. Introduce CI-ready workflows: `go fmt`, `go build ./cmd/...`, `go test ./...`, `golangci-lint run ./...`.

## Development Workflow
- Run `go mod tidy` before commits to keep dependencies clean.
- Ensure binaries under `cmd/` compile (`go build ./cmd/...`) and run main services locally (`go run ./cmd/userd`).
- Use table-driven tests with subtests; keep fixtures deterministic.
- Execute `go test -race ./...` at least once before merging substantial changes.
- Favor package-scoped interfaces that describe behavior (`TokenSigner`, `UserRepository`). Keep functions small and orchestrate through use-case services.

## Collaboration Guidelines
- Follow Conventional Commits (`feat: ...`, `fix: ...`). Keep commits focused and include matching tests.
- Document security-sensitive changes and update architecture docs alongside code.
- Seek domain owner review for `domains/user/` and platform owner review for `mora/` changes.
- Store secrets in environment variables only; commit sanitized examples. Rotate signing keys through Mora's vault integration.
- Log user-identifying data sparingly and scrub before persistence.

## When Updating This File
- Reflect new directories, services, or tooling as they land.
- Note outstanding gaps or TODOs so the next agent can pick up quickly.
- Keep instructions concise and action-oriented.
