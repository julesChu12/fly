# Repository Guidelines

## Project Structure & Module Organization
This Go 1.25.1 monorepo is wired together through `go.work` and shared capability module `mora`. Service directories (`custos`, `clotho`, `hermes`, `kratos`, `plutus`) follow the same layout: binaries in `cmd/<service>`, HTTP/gRPC adapters in `internal/interface`, domain logic in `internal/domain`, application services in `internal/application`, and persistence/adapters inside `internal/infrastructure`. Config files, migrations, and swagger specs live under each module’s `configs/` and `docs/`. Infra helpers live in `observability/` and `scripts/`; `docker-compose.yaml` plus `fly.sh` orchestrate local MySQL/Redis.

## Build, Test, and Development Commands
Run `go work sync` after cloning to keep module replacements aligned. Within a module use `make build` (generates swagger first) to compile `./bin/<service>`, `make test` for coverage-enabled unit tests, and `make run` to launch the binary against `configs/<service>.yaml`. `./scripts/test-compile.sh` provides a fast compile sanity check for Hermes/Kratos/Plutus, and `./scripts/verify-system.sh` audits required folders, proto files, and docker wiring. For a full stack, bring up stateful deps with `docker compose up -d mysql redis` or simply call `./fly.sh start` to boot the platform.

## Coding Style & Naming Conventions
Format Go code with `gofmt` (4-space indents) before committing and prefer `goimports`/`golangci-lint run` (available via each module’s `make lint`). Keep package names lowercase and short (e.g., `internal/interface/http`). Public structs/interfaces are CamelCase, private members camelCase, and configuration keys use snake_case to match the YAML files in `configs/`. Swagger comments must follow the existing `// @` pattern so `make swagger` stays deterministic. Proto definitions belong in `api/proto` and generate code into `internal/interface/grpc`.

## Testing Guidelines
Unit and integration tests live alongside the code as `<name>_test.go` and should assert behavior with `testify`. Mock via `pkg/` abstractions to avoid real databases. Aim for covering all domain services plus adapters touched by a change, and keep `go test ./... -cover` green before pushing. When altering service boundaries or schemas, add a smoke test to `scripts/verify-system.sh` or extend the module’s Makefile target (e.g., `make dev`) so reviewers can validate end to end.

## Commit & Pull Request Guidelines
Follow the existing `type(scope): summary` style (`feat:`, `fix:`, `chore(mora):`, `refactor:`). Reference the impacted service or capability in the scope whenever possible and keep bodies in English or bilingual. PRs should describe the motivation, list test evidence (`make test`, `./fly.sh status`), link to tracking issues, and attach screenshots or swagger diffs for API/UI changes. Tag owners of affected modules (per README) and call out migration or config steps explicitly so deployers can order rollouts safely.
