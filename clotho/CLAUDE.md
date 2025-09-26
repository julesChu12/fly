# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Clotho is an API orchestration layer in the Fly monorepo ecosystem. It serves as the HTTP/REST API gateway that orchestrates calls to internal domain services via gRPC. Clotho does not implement business logic - it only handles request routing, authentication middleware, and response aggregation.

## Monorepo Architecture

This is part of a monorepo with three main services:
- **Clotho** (current): API orchestration layer exposing HTTP/REST APIs externally
- **Custos**: User domain service handling authentication, authorization, and user lifecycle management
- **Mora**: Capability library providing common utilities (auth, logging, config, database, cache, message queue)

## Key Architectural Principles

- Clotho is NOT a gateway (no rate limiting, circuit breaking, traffic control)
- Clotho is NOT a domain service (no business data maintenance)
- Clotho IS an orchestration layer (request forwarding, aggregation, unified external interface)
- All business logic resides in domain services (Custos, Orders, Billing, etc.)
- Internal communication uses gRPC for high performance
- External APIs use HTTP/REST with JWT token validation via Mora Auth middleware

## Project Structure

```
clotho/
├── cmd/
│   └── clotho/            # CLI entry point using Cobra (serve, version commands)
├── configs/
│   └── clotho.yaml        # Configuration files (loaded via Mora config loader)
├── internal/
│   ├── application/
│   │   └── usecase/       # API orchestration logic
│   │       ├── user_proxy.go
│   │       ├── order_proxy.go
│   │       └── payment_proxy.go
│   ├── infrastructure/
│   │   ├── client/        # gRPC clients for domain services
│   │   │   ├── custos_grpc.go
│   │   │   └── orders_grpc.go
│   │   └── http/          # External HTTP API
│   │       ├── handler/   # HTTP handlers for external routes
│   │       └── router.go  # Route setup and middleware integration
│   └── middleware/
│       └── auth.go        # Mora Auth middleware for Access Token validation
├── docs/
│   └── README.md          # Additional architecture documentation
└── go.mod
```

## Development Commands

Since Clotho is currently being initialized, refer to the related services for common patterns:

### Related Service Commands (Custos example):
- `make help` - Show available make targets
- `make build` - Build the application
- `make test` - Run tests with coverage
- `make run` - Build and run the application
- `make dev` - Setup development environment
- `make lint` - Run linter (golangci-lint if available)
- `make clean` - Clean build artifacts

### Go Dependencies:
- Go version: 1.25.1
- Key dependencies will likely include:
  - `gin-gonic/gin` for HTTP framework
  - `golang-jwt/jwt/v5` for JWT handling
  - `google/uuid` for UUID generation
  - `gorm.io/gorm` and `gorm.io/driver/mysql` for database operations
  - gRPC libraries for internal service communication

## Request Flow

1. External client calls Clotho's HTTP API
2. Clotho uses Mora Auth Middleware to validate Access Token
3. Based on routing, Clotho calls domain services (Custos/Orders/etc.) via gRPC
4. Clotho aggregates results and returns HTTP response

## Authentication & Authorization

- Uses Mora Auth middleware for Access Token validation
- Custos service handles user authentication and RBAC (via Casbin)
- JWT tokens are signed and validated using Mora's auth utilities
- All external APIs must use Mora Auth middleware
- Internal service communication uses gRPC (no additional auth middleware needed)

## Key Implementation Guidelines

- Always call domain services via gRPC clients in `internal/infrastructure/client/`
- Never implement business/domain logic in Clotho - delegate to appropriate domain services
- Use Cobra for CLI implementation in `cmd/clotho/`
- Expose only HTTP endpoints - no GraphQL or gateway logic
- Provide minimal but clear implementations (health check, example proxies)
- Use Gin framework for HTTP server implementation
- All configuration should be loaded via Mora config loader

## Essential Endpoints to Implement

- `GET /health` - Health check endpoint
- Authentication proxy endpoints calling Custos
- User management proxy endpoints calling Custos
- Future: Order proxy endpoints, Payment proxy endpoints

## Development Notes

- This service is currently being scaffolded
- Follow patterns established in Custos and Mora services
- Maintain clear separation between orchestration (Clotho) and business logic (domain services)
- Use gRPC for all internal service communication
- Clotho should be stateless - all state managed by domain services