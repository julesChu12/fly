# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Custos** is a Golang monorepo containing:
- **Mora**: A capability library providing reusable technical modules (auth, logger, config, db, cache, mq, utils) and framework adapters
- **User Domain**: A domain service managing user identity, lifecycle, security, and authorization
- **API Layer** (future): Orchestrator that enforces trust/zero-trust policies

This is currently a documentation-only repository with no Go code yet implemented.

## Architecture Philosophy

Follow clean architecture principles:
- **Mora** = Capability library (no business logic)
- **User Domain** = Business domain service
- **API Layer** = Orchestrator (glue layer)

The User Domain handles:
- User lifecycle (register, activate, freeze, delete, profile management)
- Authentication (username+password, OTP, OAuth2.0, token rotation, multi-session)
- Security (password hashing, 2FA/MFA, audit logs, anomaly detection)
- Authorization (RBAC via Casbin, ABAC, multi-tenant support)
- Optional OAuth2.0 Provider capabilities
- Audit & Observability

The User Domain does NOT handle trust/zero-trust (API layer responsibility) or infrastructure capabilities (Mora responsibility).

## Planned Directory Structure

Based on AGENTS.md guidance:
```
custos/
├── cmd/              # Main applications (userd, etc.)
├── domains/user/     # User Domain service code
├── mora/            # Capability library modules
├── configs/         # Configuration files (.env.example, local.env)
├── testdata/        # Testing fixtures
└── docs/           # Architecture decisions
```

## Development Commands

When Go module is initialized:
- `go mod init custos` - Initialize Go module
- `go mod tidy` - Clean dependencies before commits
- `go build ./cmd/...` - Build all binaries
- `go run ./cmd/userd` - Run main user service
- `go test ./...` - Run all tests
- `go test -race ./...` - Run tests with race detection
- `go fmt ./...` - Format code
- `golangci-lint run ./...` - Lint code

## Key Development Principles

- Keep functions small and orchestrate through use-case services
- Use package-scoped interfaces that describe behavior (`TokenSigner`, `UserRepository`)
- Follow table-driven tests with subtests
- Keep test fixtures deterministic
- Use Conventional Commits (`feat:`, `fix:`)
- Store secrets in environment variables only
- Log user-identifying data sparingly

## Security Considerations

- Password hashing with bcrypt/argon2
- Refresh token rotation with state table
- Forced logout via `token_version` strategy
- Integration with Casbin for RBAC
- Audit logging for security events
- No secrets in commits (use .env files)

## Integration Points

- **Mora integration**: Auth token signing/validation, logging, config management
- **Casbin integration**: RBAC policy enforcement
- **External services**: OAuth2.0 providers (Google, WeChat, Apple), MQ/ES/Prometheus for audit events