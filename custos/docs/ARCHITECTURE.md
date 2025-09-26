# Custos Architecture

## Overview

Custos is the User Domain service in the Mora ecosystem, responsible for user identity lifecycle, authentication flows, security controls, RBAC/ABAC, and optional OAuth provider endpoints.

## Architecture Principles

- **Domain-Driven Design (DDD)**: Clear domain boundaries and business logic encapsulation
- **Clean Architecture**: Dependency inversion with clear separation of concerns
- **AI-Assistant Friendly**: Well-structured code with clear naming and documentation

## Project Structure

```
custos/
├── cmd/
│   └── userd/                 # Main application entry point
├── configs/
│   ├── local.env.example      # Environment configuration template
│   └── migrations/            # Database migration files
├── docs/
│   ├── README.md              # Project documentation
│   ├── ARCHITECTURE.md        # This architecture document
│   └── ai/                    # AI assistant documentation
│       ├── AGENTS.md
│       ├── CLAUDE.md
│       └── README_AI.md
├── internal/
│   ├── application/           # Application layer (use cases)
│   │   ├── dto/              # Data Transfer Objects
│   │   └── usecase/          # Use case implementations
│   │       ├── auth/         # Authentication use cases
│   │       ├── user/         # User management use cases
│   │       ├── oauth/        # OAuth integration use cases
│   │       └── session/      # Session management use cases
│   ├── config/               # Configuration management
│   ├── domain/               # Domain layer (business logic)
│   │   ├── entity/           # Domain entities
│   │   ├── repository/       # Repository interfaces
│   │   └── service/          # Domain services
│   │       ├── auth/         # Authentication service
│   │       ├── token/        # Token management service
│   │       └── rbac/         # Role-based access control service
│   ├── infrastructure/       # Infrastructure layer
│   │   ├── persistence/      # Data persistence
│   │   │   ├── mysql/        # MySQL implementation
│   │   │   └── redis/        # Redis implementation (future)
│   │   ├── security/         # Security implementations
│   │   │   └── casbin/       # Casbin RBAC implementation (future)
│   │   └── messaging/        # Message queue implementations
│   │       └── kafka/        # Kafka implementation (future)
│   └── interface/            # Interface layer (HTTP handlers)
│       └── http/
│           ├── handler/      # HTTP request handlers
│           ├── middleware/   # HTTP middleware
│           └── router/       # Route configuration
├── pkg/                      # Custos-specific packages
│   ├── constants/            # Domain-specific constants
│   ├── errors/               # Domain-specific errors
│   └── types/                # Domain-specific types
├── scripts/                  # Build and deployment scripts
├── test/                     # Test files and fixtures
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Layer Responsibilities

### Domain Layer (`internal/domain/`)

**Entities** (`entity/`):
- `User`: Core user entity with business rules
- Contains validation logic and business constraints

**Repositories** (`repository/`):
- Interface definitions for data access
- `UserRepository`: User data access contract

**Services** (`service/`):
- `auth/`: Authentication business logic
- `token/`: JWT token management
- `rbac/`: Role-based access control (future: Casbin integration)

### Application Layer (`internal/application/`)

**Use Cases** (`usecase/`):
- `auth/`: Registration and login flows
- `user/`: User management operations
- `oauth/`: OAuth provider integration
- `session/`: Session management and revocation

**DTOs** (`dto/`):
- Data transfer objects for API communication
- Request/response models

### Infrastructure Layer (`internal/infrastructure/`)

**Persistence** (`persistence/`):
- `mysql/`: MySQL database implementation
- `redis/`: Redis cache implementation (future)

**Security** (`security/`):
- `casbin/`: Casbin RBAC policies (future)

**Messaging** (`messaging/`):
- `kafka/`: Event publishing (future)

### Interface Layer (`internal/interface/`)

**HTTP** (`http/`):
- `handler/`: HTTP request handlers
- `middleware/`: Authentication and authorization middleware
- `router/`: Route configuration

## Key Features

### Authentication
- User registration and login
- Password hashing with bcrypt
- JWT token generation and validation
- Session management

### Authorization
- Role-based access control (RBAC)
- Token-based authentication
- Middleware for route protection

### User Management
- User CRUD operations
- Account activation/deactivation
- Profile management

### OAuth Integration
- Google, GitHub, Microsoft OAuth providers
- OAuth token management
- User linking with OAuth accounts

### Session Management
- Session creation and validation
- Session revocation
- Multi-session support

## Future Enhancements

### RBAC with Casbin
- Policy-based access control
- Dynamic permission management
- Multi-tenant support

### Event-Driven Architecture
- Kafka integration for events
- User lifecycle events
- Audit logging

### Caching
- Redis integration for session storage
- Token blacklisting
- Performance optimization

## Dependencies

### External
- **Gin**: HTTP web framework
- **GORM**: ORM for database operations
- **JWT**: Token management
- **bcrypt**: Password hashing

### Internal (Mora)
- **Mora**: Shared capabilities library
- **Clotho**: API orchestration (future)

## Configuration

Environment variables are managed through `configs/local.env.example`:
- Database connection settings
- JWT configuration
- OAuth provider settings
- Application settings

## Testing

- Unit tests for domain logic
- Integration tests for use cases
- HTTP handler tests
- Database migration tests

## Development Workflow

1. **Domain First**: Start with domain entities and business rules
2. **Repository Pattern**: Define data access interfaces
3. **Use Cases**: Implement application logic
4. **Infrastructure**: Add persistence and external service implementations
5. **Interface**: Create HTTP handlers and middleware
6. **Testing**: Add comprehensive test coverage

## AI Assistant Guidelines

- Use clear, descriptive function and variable names
- Add comprehensive comments for complex business logic
- Follow Go conventions and best practices
- Maintain consistent error handling patterns
- Document architectural decisions and trade-offs
