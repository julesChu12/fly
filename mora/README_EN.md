# Mora

**Mora** derives from the Greek mythology of *Moirai* (the Fates), who control the threads of destiny for all beings.  
As a Golang capability library, Mora carries the meaning of "allocation and order":  
It provides common foundational capability modules for all services, allowing projects to set sail quickly under rules and clear boundaries.

Mora is not a specific gateway or framework, but rather a **capability source**:  
- Accumulates common modules in `pkg/` (auth/logger/config/...)  
- Provides framework adaptation layer in `adapters/`  
- Demonstrates how the API layer orchestrates capabilities and domain services in `starter/`

---

## Project Structure
```
mora/
  ├── go.mod
  ├── pkg/                    # Core capability packages (framework-agnostic) ✅
  │   ├── auth/              # JWT token generation and validation ✅
  │   │   ├── claims.go      # JWT Claims structure definition
  │   │   ├── jwt.go         # JWT generation and validation
  │   │   └── jwks.go        # JWK support
  │   ├── logger/            # Logging wrapper ✅
  │   │   ├── logger.go      # Zap logging wrapper
  │   │   └── context.go     # Context tracing support
  │   ├── config/            # Configuration loading ✅
  │   │   └── loader.go      # YAML/ENV configuration loading
  │   ├── db/                # Database wrapper ✅
  │   │   ├── gorm.go        # GORM wrapper
  │   │   └── sqlx.go        # SQLX wrapper
  │   ├── cache/             # Redis cache wrapper ✅
  │   │   ├── redis.go       # Redis basic operations
  │   │   └── lock.go        # Distributed locks
  │   ├── mq/                # Message queue wrapper ✅
  │   │   ├── mq.go          # Message queue interface
  │   │   ├── memory.go      # Memory queue implementation
  │   │   └── redis.go       # Redis queue implementation
  │   └── utils/             # Common utilities ✅
  │       ├── crypto.go      # Cryptographic utilities
  │       ├── string.go      # String utilities
  │       └── time.go        # Time utilities
  │
  ├── adapters/              # Framework adaptation layer ✅
  │   ├── gin/               # Gin framework adaptation ✅
  │   │   ├── auth_middleware.go # JWT authentication middleware
  │   │   └── otel_middleware.go # OpenTelemetry middleware
  │   └── gozero/            # Go-Zero framework adaptation ✅
  │       ├── auth_middleware.go # JWT authentication middleware
  │       ├── context.go     # Context utilities
  │       └── otel_middleware.go # OpenTelemetry gRPC interceptor
  │
  ├── starter/               # Example applications ✅
  │   ├── gin-starter/       # Gin demo application
  │   │   ├── main.go        # Complete REST API example
  │   │   └── docs/          # Swagger documentation
  │   └── gozero-starter/    # Go-Zero demo application
  │       ├── main.go        # Go-Zero service example
  │       ├── api/           # API definitions
  │       ├── etc/           # Configuration files
  │       └── internal/      # Internal implementation
  │
  └── docs/
      └── usage-examples.md
```

---

## Module Description

### pkg/
- **auth/**  
  Provides JWT/JWK generation and validation utility methods:  
  - `GenerateToken(userID, secret, ttl)`  
  - `ValidateToken(token, secret)` → returns `Claims` (containing userID)  
  - **No DB dependency, no User Service dependency**  
  - **Only provides JWT/JWK utility methods, not responsible for user authentication or state management**  

- **logger/**  
  Wraps logging libraries (zap/logx), unifies output format, supports traceId.  

- **config/**  
  Supports YAML/ENV configuration loading, can be extended to remote configuration center in the future.  

- **observability/**  
  OpenTelemetry observability support, provides distributed tracing, metrics collection, and log correlation.

- **db/**  
  Database wrapper, based on sqlx or gorm.  

- **cache/**  
  Redis utilities, supports common patterns (cache aside, distributed locks).  

- **mq/**  
  Message queue wrapper, supports memory and Redis implementations.  

- **utils/**  
  Utility functions (string, time, crypto, etc.).  

---

### adapters/
- **gin/**  
  Provides gin middleware wrapper, such as:  
  - `AuthMiddleware(secret)`: calls `pkg/auth` to validate token, injects userID into gin.Context.  
  - `ObservabilityMiddleware(serviceName)`: adds OpenTelemetry distributed tracing support.

- **gozero/**  
  Provides go-zero middleware wrapper:  
  - `AuthMiddleware(secret)`: JWT authentication middleware  
  - `ServerOption()` / `ClientOption()`: gRPC OpenTelemetry interceptors  

---

### starter/
- **gin-starter/**  
  Demonstrates how the API layer orchestrates User Service and Auth modules:  
  - `/login`: simulates calling User Service to validate username/password, then uses `pkg/auth` to issue token upon success.  
  - `/ping`: protected endpoint, uses `AuthMiddleware` to validate token, returns userID.  

Run with:
```bash
# Gin demo application
cd starter/gin-starter
go run main.go
# Visit http://localhost:8080/swagger/ for API documentation

# Go-Zero demo application
cd starter/gozero-starter
go run main.go -f etc/mora-api.yaml
# Default runs on http://localhost:8888
```

### API Endpoints Examples
- **Public endpoints**:
  - `GET /health` - Health check
  - `POST /login` - User login (returns JWT Token)
- **Authenticated endpoints**:
  - `GET /profile` - Get user profile
  - `GET /protected` - Protected example endpoint
  - `GET /api/v1/orders` - Get orders list
  - `POST /api/v1/orders` - Create order
  - `GET /api/v1/users` - Get users list

---

## Implementation Status

### ✅ Completed
- **Core capability packages (pkg/)**:
  - `auth/` - JWT Token generation and validation, JWKS support
  - `logger/` - Structured logging based on Zap, with trace context
  - `config/` - Unified configuration loading (YAML + ENV)
  - `observability/` - OpenTelemetry observability support
  - `db/` - Database abstraction layer (GORM + SQLX)
  - `cache/` - Redis cache and distributed locks
  - `mq/` - Message queue abstraction (Memory + Redis implementations)
  - `utils/` - Common utility set (crypto, string, time)

- **Framework adapters (adapters/)**:
  - `gin/` - Gin framework authentication middleware + OpenTelemetry middleware
  - `gozero/` - Go-Zero framework authentication middleware + OpenTelemetry middleware

- **Demo applications (starter/)**:
  - `gin-starter/` - Complete Gin REST API (with Swagger docs)
  - `gozero-starter/` - Go-Zero microservice example

- **Test Coverage**:
  - All core packages have comprehensive unit tests
  - 100% test pass rate (50 Go files)
  - Support for multiple database and cache backends

### 📋 Development Roadmap
- Extend more MQ implementations (Kafka/RabbitMQ)
- Add more database driver support (MongoDB, ClickHouse, etc.)
- Improve CI/CD scaffolding and automated testing
- Add deployment examples and best practices
- Add more framework adapters (Echo, Fiber, etc.)
- Enhance monitoring and alerting integration

---

## Design Principles
- **Core capability packages (pkg/) are framework-agnostic**  
- **adapters/** serves as an anti-corruption layer, responsible for integrating capability packages into gin/go-zero and other frameworks  
- **starter/** demonstrates complete scenarios, where the API layer is an orchestrator, connecting Auth and User Service  
- **User Service belongs to domain services**, responsible for user tables/permission tables, not coupled with Auth modules  
- **Mora is not responsible for user authentication logic (login/refresh/state management), these belong to UserService (Custos)**

---

## Next Steps
- Continue improving core modules' functionality and performance optimization
- Extend more framework adapters (Echo, Fiber, etc.)
- Add more business scenario examples and best practices
- Optimize documentation and developer experience  

---

## Language Support
- [中文版 README](README.md)
- [English README](README_EN.md)
