# Custos User Domain Service - MVP

A minimal viable product (MVP) implementation of the Custos User Domain Service with basic user management and authentication capabilities.

## 🚀 Features

### ✅ Implemented (MVP)
- **User Registration**: Username/email registration with password
- **User Authentication**: Username + password login with JWT tokens
- **Basic RBAC**: Simple admin/user role system
- **RESTful API**: HTTP endpoints for authentication and user management
- **Database Storage**: MySQL with GORM ORM
- **Security**: Password hashing with bcrypt, JWT token validation

### 🔄 Future Releases
- OAuth2.0 third-party login (Google, WeChat, Apple)
- 2FA/MFA multi-factor authentication
- Refresh token rotation
- Audit logging system
- Multi-session management
- Fine-grained permission system

## 📋 Prerequisites

- Go 1.21+
- MySQL 8.0+
- Git

## 🛠️ Quick Start

### 1. Clone and Setup
```bash
git clone <repository-url>
cd custos
make dev
```

### 2. Configure Database
Update `.env` file with your database credentials:
```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=custos
DB_PASSWORD=your_password
DB_DATABASE=custos_dev
```

Create MySQL database:
```sql
CREATE DATABASE custos_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 3. Run the Service
```bash
make run
```

The service will start on `http://localhost:8080`

## 📡 API Endpoints

### Health Check
```bash
GET /api/v1/health
```

### Authentication
```bash
# Register
POST /api/v1/auth/register
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "securepassword123"
}

# Login
POST /api/v1/auth/login
{
  "username": "john_doe",
  "password": "securepassword123"
}
```

### User Management (Requires Authentication)
```bash
# Get user profile
GET /api/v1/user/profile
Authorization: Bearer <your_jwt_token>
```

## 🗂️ Project Structure

```
custos/
├── cmd/userd/           # Main application
├── internal/
│   ├── domain/          # Domain layer (entities, repositories, services)
│   ├── infrastructure/  # Infrastructure layer (database, external services)
│   ├── application/     # Application layer (use cases, DTOs)
│   ├── interface/       # Interface layer (HTTP handlers, middleware)
│   └── config/          # Configuration management
├── pkg/                 # Public packages (errors, types, constants)
├── configs/             # Configuration files and migrations
├── scripts/             # Development and deployment scripts
└── test/                # Test fixtures and integration tests
```

## 🧪 Development

### Available Commands
```bash
make build    # Build the application
make test     # Run tests with coverage
make run      # Run the service
make clean    # Clean build artifacts
make dev      # Setup development environment
make lint     # Run code linter
```

### Testing
```bash
# Run all tests
make test

# Run specific tests
go test ./internal/domain/service/...
```

### Database Migrations
Database schema is automatically migrated on service startup using GORM AutoMigrate.

## 🔐 Security Features

- **Password Security**: bcrypt hashing with default cost
- **JWT Authentication**: HS256 signed tokens with configurable expiration
- **Input Validation**: Request parameter validation and sanitization
- **CORS Support**: Configurable cross-origin resource sharing
- **Error Handling**: Structured error responses without sensitive information

## ⚙️ Configuration

Key environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `APP_ENV` | Environment (development/production) | `development` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `3306` |
| `JWT_SECRET` | JWT signing secret | - |
| `JWT_ACCESS_TTL` | Access token TTL (minutes) | `15` |

## 🐳 Docker Support (Future)

Docker configuration will be added in future releases for easy deployment.

## 📈 Roadmap

- **v0.2.0**: Security enhancements (2FA, audit logs, refresh tokens)
- **v0.3.0**: OAuth2.0 integration (Google, WeChat, Apple)
- **v0.4.0**: Advanced RBAC with Casbin integration
- **v0.5.0**: Multi-tenant support for B-end usage

## 🤝 Contributing

1. Follow Clean Architecture principles
2. Write tests for new features
3. Use conventional commits (`feat:`, `fix:`, `docs:`)
4. Update documentation for API changes

## 📄 License

See [LICENSE](LICENSE) file for details.