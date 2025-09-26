# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Mora is a Go capability library inspired by the Greek Moirai (Fates), providing foundational modules for all services. It's organized into three workspaces:
- `pkg/` - Framework-neutral capability modules (auth, logger, config, db, cache, mq, utils)
- `adapters/` - Framework adapters (gin, gozero) for integrating pkg modules
- `starter/` - Runnable demo applications showing how to orchestrate capabilities

Go version: 1.25.1

## Development Commands

### Essential Commands
```bash
# Dependency management
go mod tidy

# Testing
go test ./...                    # Run all tests
go test -cover ./pkg/...         # Run tests with coverage for pkg modules
go vet ./...                     # Check for common mistakes

# Formatting
go fmt ./...                     # Format all Go files (always run before commit)

# Running starter demos
go run ./starter/gin-starter     # Run the gin starter demo (once implemented)
```

## Architecture Principles

1. **Framework Independence**: Core modules in `pkg/` must remain framework-agnostic
2. **Adapter Pattern**: `adapters/` serves as an anti-corruption layer, bridging pkg capabilities to specific frameworks
3. **API Layer as Orchestrator**: `starter/` demonstrates how API layers orchestrate Auth modules with domain services
4. **Service Separation**: User Service handles domain logic (user tables, permissions) and should not couple with Auth module

### Auth Module Design
- Located in `pkg/auth/`
- Provides JWT token generation and validation
- Functions: `GenerateToken(userID, secret, ttl)` and `ValidateToken(token, secret)`
- **No database dependencies, no User Service dependencies**

## Code Conventions

- Target Go 1.25.1 syntax and features
- Use `gofmt` formatting (enforced via `go fmt ./...`)
- Package naming: short, lowercase names
- Exported APIs: PascalCase (e.g., `NewLogger`, `GenerateToken`)
- Internal symbols: camelCase
- Keep modules decoupled; use interface dependencies
- Co-locate `*_test.go` files with implementation
- Use table-driven tests and subtests (`t.Run`)

## Commit Style

Follow Conventional Commits format (see existing commits like "chore: init Mora project"):
- Keep subject lines under 72 characters
- Types: feat, fix, chore, docs, refactor, test
- Example: `feat: add JWT token generation to pkg/auth`