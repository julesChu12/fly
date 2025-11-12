# Clotho Project Structure

## Overview

Clotho is a Go-based API orchestration layer that follows Clean Architecture principles with clear separation of concerns. The project is organized to handle HTTP/REST API requests and orchestrate calls to internal domain services via gRPC.

## Directory Structure

```
clotho/
├── cmd/                          # Command-line interface (CLI)
│   ├── clotho/                  # Main application command
│   │   └── main.go             # Application entry point
│   ├── root.go                 # Root CLI command setup
│   ├── serve.go                # Server command configuration
│   └── version.go              # Version command implementation
├── configs/                     # Configuration files
│   ├── clotho.yaml            # Main configuration file
│   └── clotho.env.yaml        # Environment-specific configuration
├── internal/                    # Private application code
│   ├── application/            # Application layer (use cases)
│   │   └── usecase/           # Business use case implementations
│   │       └── user_proxy.go  # User management proxy use case
│   ├── infrastructure/         # Infrastructure layer
│   │   ├── client/           # gRPC client implementations
│   │   │   └── custos_grpc.go # Custos service gRPC client
│   │   └── http/             # HTTP infrastructure
│   │       ├── handler/      # HTTP request handlers
│   │       │   ├── health.go      # Health check handler
│   │       │   ├── monitoring.go  # Monitoring statistics handler
│   │       │   ├── profile.go     # Profile management handler
│   │       │   └── user.go        # User management handler
│   │       └── router.go     # HTTP router setup and middleware
│   ├── middleware/            # HTTP middleware implementations
│   │   ├── auth.go           # JWT authentication middleware
│   │   ├── circuit_breaker.go # Circuit breaker middleware
│   │   ├── cors.go           # CORS handling middleware
│   │   ├── error.go          # Error handling middleware
│   │   ├── logging.go        # Request logging middleware
│   │   ├── metrics.go        # Prometheus metrics middleware
│   │   └── rate_limiter.go   # Rate limiting middleware
│   └── validation/           # Request validation logic
│       └── profile.go        # Profile data validation
├── api/                       # API definitions and generated code
│   └── proto/                # Protocol Buffer definitions
│       ├── custos.proto      # Custos service gRPC definition
│       ├── custos.pb.go      # Generated protobuf code
│       └── custos_grpc.pb.go # Generated gRPC client code
├── docs/                      # Documentation files
│   ├── API.md               # API documentation
│   ├── STRUCTURE.md         # Project structure documentation
│   ├── OBSERVABILITY.md     # Observability and monitoring guide
│   ├── swagger.json         # OpenAPI/Swagger specification
│   ├── swagger.yaml         # OpenAPI/Swagger specification
│   ├── doc.go               # Swagger documentation metadata
│   └── docs.go              # Swagger documentation generation
├── docker-compose.yml        # Docker Compose configuration
├── Dockerfile               # Docker image configuration
├── Dockerfile.dev           # Development Docker image
├── Dockerfile.prod          # Production Docker image
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
├── Makefile                 # Build and development commands
├── main.go                  # Application entry point
├── README.md                # Project overview and getting started
└── CLAUDE.md               # Claude Code assistant instructions
```

## Architecture Layers

### 1. CLI Layer (`cmd/`)
- **Purpose**: Application entry point and command-line interface
- **Key Components**:
  - `cmd/clotho/main.go`: Main application bootstrap
  - `cmd/serve.go`: Server startup and configuration
  - `cmd/version.go`: Version information command
- **Frameworks**: Cobra for CLI, Viper for configuration

### 2. Application Layer (`internal/application/`)
- **Purpose**: Business use case orchestration
- **Key Components**:
  - `usecase/user_proxy.go`: User management proxy implementation
- **Responsibilities**: Coordinate between HTTP handlers and gRPC clients

### 3. Infrastructure Layer (`internal/infrastructure/`)
- **Purpose**: External interface implementations
- **Sub-layers**:
  - **HTTP Layer**: REST API handlers and routing
  - **Client Layer**: gRPC clients for domain services
- **Key Components**:
  - `http/handler/`: HTTP request handlers by domain
  - `http/router.go`: Route configuration and middleware setup
  - `client/custos_grpc.go`: Custos service gRPC client

### 4. Cross-Cutting Concerns (`internal/middleware/`)
- **Purpose**: Reusable HTTP middleware
- **Middleware Components**:
  - **Authentication**: JWT token validation
  - **Rate Limiting**: Multi-level request throttling
  - **Circuit Breaking**: Upstream service protection
  - **Logging**: Request/response logging
  - **Metrics**: Prometheus metrics collection
  - **CORS**: Cross-origin resource sharing
  - **Error Handling**: Centralized error processing

### 5. API Definitions (`api/proto/`)
- **Purpose**: gRPC service definitions and generated code
- **Key Components**:
  - `custos.proto`: Custos service interface definition
  - Generated Go code for protobuf and gRPC

## Data Flow Architecture

```
External Client
       ↓
HTTP Request
       ↓
[Middleware Chain]
├── CORS
├── Logging
├── Metrics
├── Rate Limiter
├── Circuit Breaker
└── Authentication
       ↓
HTTP Handler
       ↓
Use Case (Application Layer)
       ↓
gRPC Client (Infrastructure Layer)
       ↓
Domain Service (e.g., Custos)
       ↓
Response Flow (reverse order)
```

## Key Design Patterns

### 1. Dependency Injection
- gRPC clients are injected into use cases via factory functions
- Configuration is injected through Viper
- Middleware dependencies are managed through constructors

### 2. Middleware Chain Pattern
- Request processing flows through a series of middleware
- Each middleware handles a specific cross-cutting concern
- Middleware are composable and configurable

### 3. Proxy Pattern
- Use cases act as proxies to domain services
- HTTP handlers delegate to use cases for business logic
- gRPC clients abstract service communication details

### 4. Factory Pattern
- gRPC client creation via factory functions
- Lazy initialization for performance
- Connection pooling and lifecycle management

## Configuration Architecture

### Configuration Sources
1. **Default Values**: Hardcoded defaults in `router.go`
2. **Config Files**: YAML configuration in `configs/`
3. **Environment Variables**: Runtime overrides
4. **Command Line Flags**: CLI argument overrides

### Configuration Categories
- **App**: Basic application settings (mode, version)
- **Observability**: Service name, tracing configuration
- **Services**: gRPC service addresses and timeouts
- **Rate Limiter**: Multi-level rate limiting parameters
- **Circuit Breaker**: Fault tolerance configuration
- **JWT**: Authentication token settings

## Observability Integration

### Metrics Collection
- **Prometheus**: Custom metrics for request count, latency, errors
- **Middleware Metrics**: Rate limiter and circuit breaker statistics
- **Health Checks**: Service health and dependency status

### Logging Strategy
- **Structured Logging**: JSON format with consistent fields
- **Request Correlation**: Unique request IDs for tracing
- **Log Levels**: Configurable log verbosity

### Tracing
- **OpenTelemetry**: Distributed tracing integration
- **Context Propagation**: Span context across service boundaries

## Security Architecture

### Authentication Flow
1. Client presents JWT token in `Authorization` header
2. Middleware validates token signature and expiration
3. User context extracted and attached to request
4. Authorization checks performed as needed

### Rate Limiting Strategy
- **Global**: Service-wide request limits
- **Per IP**: Client IP-based throttling
- **Per User**: Authenticated user-based limits
- **Token Bucket Algorithm**: Fair resource allocation

### Circuit Breaking Logic
- **Failure Detection**: Configurable failure thresholds
- **Circuit States**: Closed, Open, Half-Open
- **Automatic Recovery**: Periodic health checks
- **Graceful Degradation**: Fast failure on upstream issues

## Development Workflow

### Local Development
```bash
# Development environment
make dev          # Setup development dependencies
make run          # Run with hot reload
make test         # Run test suite
make lint         # Code quality checks
```

### Build Process
```bash
# Production builds
make build        # Build production binary
make docker       # Build Docker images
make release      # Full release pipeline
```

### Configuration Management
- Development: `configs/clotho.yaml` with debug settings
- Production: Environment-specific overrides
- Secrets: Environment variables or secure stores

## Integration Points

### Internal Services
- **Custos**: User authentication and management service
- **Mora**: Shared capability library (auth, logging, config)

### External Integrations
- **Prometheus**: Metrics collection and monitoring
- **OpenTelemetry**: Distributed tracing
- **Swagger/OpenAPI**: API documentation generation

## Scalability Considerations

### Horizontal Scaling
- **Stateless Design**: All state managed in external services
- **Load Balancing**: Multiple instances behind load balancer
- **Database**: External database connections via gRPC

### Performance Optimizations
- **Connection Pooling**: gRPC connection reuse
- **Circuit Breaking**: Fast failure for overloaded services
- **Rate Limiting**: Protection against traffic spikes
- **Metrics**: Low-overhead performance monitoring

## Testing Strategy

### Unit Tests
- Handler logic testing
- Middleware functionality
- Use case orchestration
- Client communication

### Integration Tests
- End-to-end API flows
- gRPC service integration
- Middleware chain behavior
- Configuration validation

### Performance Tests
- Load testing with concurrent requests
- Rate limiter effectiveness
- Circuit breaker behavior
- Memory and CPU profiling