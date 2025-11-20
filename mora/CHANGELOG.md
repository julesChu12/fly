# Changelog

All notable changes to the Mora project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-20

### Added
- **JWT Authentication**: Complete JWT/JWK token generation and validation support
  - Token generation with custom claims
  - Token validation and parsing
  - JWKS (JSON Web Key Set) support
- **Logger**: Structured logging wrapper based on Zap
  - Context-aware logging with trace ID support
  - Multiple output formats (JSON, Console)
  - File rotation and log management
- **Configuration**: Unified configuration loading
  - YAML configuration support
  - Environment variable support
  - Configuration merging and validation
- **Database Adapters**: Database abstraction layer
  - GORM wrapper for ORM operations
  - SQLX wrapper for raw SQL queries
  - Support for MySQL, PostgreSQL, SQLite
- **Cache**: Redis cache utilities
  - Basic cache operations
  - Distributed lock support
  - Connection pooling
- **Message Queue**: Message queue abstraction
  - Memory queue implementation
  - Redis queue implementation
  - Kafka support (via Sarama)
- **Observability**: OpenTelemetry integration
  - Distributed tracing support
  - Metrics collection
  - Multiple exporter backends (Jaeger, OTLP, stdout)
- **Framework Adapters**: Framework-specific integrations
  - Gin framework middleware (auth, observability, prometheus)
  - Go-Zero framework middleware (auth, observability)
- **Service Discovery**: Service discovery utilities
  - Consul integration
  - Environment-based discovery
  - Load balancing support
- **Utilities**: Common utility functions
  - Cryptographic utilities
  - String manipulation
  - Time utilities

### Documentation
- Complete README with usage examples
- API documentation
- Starter applications (Gin and Go-Zero examples)

### Testing
- Comprehensive unit tests for all core packages
- 100% test pass rate
- Mock implementations for testing

---

## Version History

- **v1.0.0** (2025-01-20): Initial stable release

