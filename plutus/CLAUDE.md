# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Plutus** is the Payment & Wallet Base Service for the Fly platform, named after the Greek god of wealth. It handles all payment processing, wallet management, and transaction recording with strong consistency guarantees and full auditability.

### Core Responsibilities
- Wallet account management with multi-tenancy (`tenant_id + customer_id`)
- Transaction recording (recharge, consume, refund) with idempotency
- External payment channel integration and logging
- Atomic balance updates with transactional safety
- Provides both REST and gRPC APIs

### What Plutus Does NOT Handle
- Order lifecycle management (handled by **Kratos**)
- Customer information management (handled by **Hermes**)
- CRM or membership system logic

## Build and Development Commands

### Basic Commands
```bash
# Build the service
make build

# Run the service (uses configs/plutus.yaml)
make run

# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Format code
make fmt

# Lint code
make lint

# Run all checks (format + lint + test)
make check

# Clean build artifacts
make clean
```

### Development Setup
```bash
# First-time setup: install dependencies and tools
make dev-setup

# Generate Swagger documentation
make swagger

# Install required tools separately
make install-tools
```

### Docker Commands
```bash
# Build Docker image
make docker-build

# Run in Docker
make docker-run
```

### Configuration
- Main config: `configs/plutus.yaml`
- Default ports: HTTP 8085, gRPC 9085
- Database migrations: `configs/migrations/`
- Service runs from `cmd/plutus/main.go`

## Architecture

### Project Structure (DDD Clean Architecture)

```
plutus/
├── cmd/plutus/           # Application entry point
├── configs/              # Configuration files and SQL migrations
├── internal/
│   ├── domain/          # Core business entities and repository interfaces
│   │   ├── entity/      # Wallet, Transaction, PaymentChannel entities
│   │   └── repository/  # Repository interface definitions
│   ├── application/     # Business logic layer
│   │   └── service/     # WalletService with transaction operations
│   ├── infrastructure/  # External dependencies implementation
│   │   └── database/    # Repository implementations with GORM
│   └── interface/       # API layer (HTTP/gRPC handlers)
│       └── http/        # Gin-based REST API handlers and routing
└── pkg/                 # Shared utilities
    ├── constants/       # Application-wide constants
    ├── errors/          # Custom error types
    └── types/           # Request/Response DTOs
```

### Data Flow
1. **HTTP Handler** (`internal/interface/http/`) receives REST requests
2. **Service Layer** (`internal/application/service/`) contains business logic
3. **Repository Layer** (`internal/infrastructure/database/`) performs database operations
4. **Domain Entities** (`internal/domain/entity/`) represent core business objects

### Database Schema

**wallets table:**
- Unique per `(tenant_id, customer_id)`
- Stores balance as DECIMAL(14,2) with currency (default: CNY)
- Status: `active` or `frozen`

**transactions table:**
- Immutable transaction log
- Types: `recharge`, `consume`, `refund`
- Includes `idempotency_key` for duplicate prevention
- Links to `order_id` when applicable
- Channels: wallet, wechat, alipay, stripe, paypal, bank, other
- Status: pending, success, failed

**payment_channels table:**
- Configuration for external payment gateways
- Stores channel-specific config as JSON

### Transaction Safety Rules

**Critical:** All balance modifications MUST follow these rules:
1. Use database transactions (implicit in GORM operations)
2. Lock wallet rows using `SELECT ... FOR UPDATE` for consume operations
3. Always create a transaction record BEFORE updating balance
4. Never directly modify `wallets.balance` without a corresponding transaction entry
5. Use idempotency keys to prevent duplicate operations
6. Validate wallet status (must be `active`) before any operation

## Multi-Tenancy

All operations are scoped by `tenant_id` extracted from the request context (`constants.ContextKeyTenantID`). Services automatically:
- Filter queries by tenant_id
- Validate tenant ownership before operations
- Return `ErrUnauthorized` if tenant context is missing

## Key Design Patterns

### Auto-create Wallets
When a recharge operation targets a non-existent customer, the service automatically creates a new wallet. This simplifies client integration.

### Idempotency
All transaction endpoints support optional `idempotency_key`. If provided and a matching transaction exists, the service returns the existing transaction instead of creating a duplicate.

### Balance Validation
- Consume operations check balance sufficiency and return `InsufficientBalanceError` with current/required amounts
- Amount validation: Must be between `constants.MinAmount` and `constants.MaxAmount`

## REST API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/wallets` | Create wallet |
| GET | `/api/wallets/{id}` | Get wallet by ID |
| GET | `/api/wallets/customer/{customer_id}` | Get wallet by customer |
| GET | `/api/wallets` | List wallets (paginated) |
| POST | `/api/transactions/recharge` | Add funds to wallet |
| POST | `/api/transactions/consume` | Deduct funds from wallet |
| POST | `/api/transactions/refund` | Return funds to wallet |
| GET | `/api/transactions/{id}` | Get transaction by ID |
| GET | `/api/transactions` | List transactions (paginated) |

## Integration with Other Services

### Service Communication
- **Clotho**: API Gateway layer that routes requests to Plutus
- **Kratos**: Order service that triggers consume/refund operations via order_id
- **Hermes**: Customer service that provides customer_id validation
- **Custos**: Authentication service (JWT validation via endpoint at `configs/plutus.yaml`)

### Context Propagation
Services use context to pass:
- `tenant_id` (required for all operations)
- Trace IDs for distributed tracing
- User/customer identity information

## Testing Practices

### Test Organization
- Unit tests for service logic: Test business rules and error conditions
- Repository tests: Test database operations and query correctness
- Integration tests: Test full request/response flow

### Running Specific Tests
```bash
# Run tests for a specific package
go test -v ./internal/application/service/...

# Run a specific test function
go test -v ./internal/application/service -run TestWalletService_Recharge

# Run with race detection
go test -race ./...
```

## Configuration Management

Service uses Viper for configuration with the following precedence:
1. Config file (`configs/plutus.yaml`)
2. Environment variables
3. Default values in code

Key configuration sections:
- `server`: HTTP/gRPC ports and timeouts
- `database`: MySQL connection with pool settings
- `cache`: Redis configuration for distributed locking
- `observability`: OpenTelemetry tracing, metrics, health checks
- `dtm`: Distributed transaction manager settings
- `auth`: JWT secrets and Custos endpoint

## Future Enhancements (from README)

Planned features to be aware of:
- Third-party payment gateway integration (WeChat, Alipay, Stripe)
- Reconciliation and settlement systems
- Multi-party settlement and profit sharing
- Pre-authorization and fund freezing
- Async callbacks with retry mechanisms
- Webhook notification system
