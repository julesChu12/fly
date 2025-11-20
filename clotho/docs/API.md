# Clotho API Documentation

## Overview

Clotho is an API orchestration layer in the Fly monorepo ecosystem. It serves as the HTTP/REST API gateway that orchestrates calls to internal domain services via gRPC. Clotho does not implement business logic - it only handles request routing, authentication middleware, and response aggregation.

## Base URL

```
Development: http://localhost:8080
Production:  https://api.clotho.fly.com
```

## Authentication

All API endpoints (except `/health`, `/swagger/*`, and `/metrics`) require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <jwt_token>
```

## API Endpoints

### Health Check

#### GET /health
Check the health status of the Clotho service.

**Response:**
```json
{
  "service": "clotho",
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "uptime": "24h30m15s",
  "version": "1.0.0"
}
```

### User Management

#### GET /api/v1/users/me
Get the current authenticated user's information.

**Authentication:** Required

**Response:**
```json
{
  "id": "user-uuid",
  "username": "johndoe",
  "email": "john@example.com",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

#### GET /api/v1/users/:id
Get user information by ID.

**Authentication:** Required

**Parameters:**
- `id` (path): User UUID

**Response:**
```json
{
  "id": "user-uuid",
  "username": "johndoe",
  "email": "john@example.com",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Profile Management

#### GET /api/v1/profile
Get the current user's complete profile.

**Authentication:** Required

**Response:**
```json
{
  "id": "user-uuid",
  "username": "johndoe",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "bio": "Software developer",
  "avatar": "https://example.com/avatar.jpg",
  "location": "San Francisco, CA",
  "website": "https://johndoe.dev",
  "phone": "+1-555-0123",
  "preferences": {
    "theme": "dark",
    "language": "en",
    "notifications": "enabled"
  },
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

#### PUT /api/v1/profile
Update the current user's profile.

**Authentication:** Required

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "bio": "Software developer",
  "avatar": "https://example.com/avatar.jpg",
  "location": "San Francisco, CA",
  "website": "https://johndoe.dev",
  "phone": "+1-555-0123"
}
```

**Response:**
```json
{
  "id": "user-uuid",
  "username": "johndoe",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "bio": "Software developer",
  "avatar": "https://example.com/avatar.jpg",
  "location": "San Francisco, CA",
  "website": "https://johndoe.dev",
  "phone": "+1-555-0123",
  "preferences": {
    "theme": "dark",
    "language": "en",
    "notifications": "enabled"
  },
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:30:00Z"
}
```

#### PUT /api/v1/profile/preferences
Update user preferences.

**Authentication:** Required

**Request Body:**
```json
{
  "theme": "light",
  "language": "es",
  "notifications": "disabled"
}
```

**Response:**
```json
{
  "preferences": {
    "theme": "light",
    "language": "es",
    "notifications": "disabled"
  }
}
```

#### GET /api/v1/profile/users/:id
Get another user's public profile.

**Authentication:** Required

**Parameters:**
- `id` (path): User UUID

**Response:**
```json
{
  "id": "user-uuid",
  "username": "johndoe",
  "first_name": "John",
  "last_name": "Doe",
  "bio": "Software developer",
  "avatar": "https://example.com/avatar.jpg",
  "location": "San Francisco, CA",
  "website": "https://johndoe.dev"
}
```

### Monitoring & Statistics

#### GET /api/v1/monitoring/stats
Get comprehensive middleware statistics.

**Authentication:** Required

**Response:**
```json
{
  "rate_limiter": {
    "global_requests": 10000,
    "blocked_requests": 500,
    "per_ip_stats": {...},
    "per_user_stats": {...}
  },
  "circuit_breaker": {
    "total_breakers": 5,
    "open_breakers": 1,
    "breaker_states": {...}
  },
  "metrics": {
    "request_count": 10000,
    "error_count": 50,
    "avg_response_time": "150ms"
  }
}
```

#### GET /api/v1/monitoring/rate-limiter
Get rate limiter statistics.

**Authentication:** Required

**Response:**
```json
{
  "global_rps": 1000.0,
  "global_burst": 2000,
  "current_global_requests": 850,
  "per_ip_rps": 10.0,
  "per_ip_burst": 20,
  "per_user_rps": 100.0,
  "per_user_burst": 200,
  "blocked_requests_last_minute": 25
}
```

#### GET /api/v1/monitoring/circuit-breaker
Get circuit breaker statistics.

**Authentication:** Required

**Response:**
```json
{
  "total_breakers": 5,
  "open_breakers": 1,
  "half_open_breakers": 0,
  "closed_breakers": 4,
  "breaker_details": {
    "custos_client": {
      "state": "closed",
      "requests": 100,
      "failures": 2,
      "success_rate": 0.98
    }
  }
}
```

#### POST /api/v1/monitoring/circuit-breaker/reset
Reset all circuit breakers to closed state.

**Authentication:** Required

**Response:**
```json
{
  "message": "All circuit breakers have been reset",
  "reset_count": 5
}
```

### System Endpoints

#### GET /swagger/*
Access Swagger UI for API documentation.
**Authentication:** Not required

#### GET /metrics
Access Prometheus metrics endpoint.
**Authentication:** Not required

## Error Responses

All endpoints return errors in the following format:

```json
{
  "error": "error_type",
  "message": "Human-readable error message"
}
```

### Common Error Types

- `invalid_request`: The request is malformed or missing required parameters
- `unauthorized`: Invalid or missing authentication token
- `forbidden`: User does not have permission to access the resource
- `not_found`: The requested resource does not exist
- `rate_limited`: Too many requests (rate limit exceeded)
- `service_unavailable`: Upstream service is unavailable (circuit breaker open)
- `internal_error`: Internal server error

## Rate Limiting

The API implements multiple layers of rate limiting:

- **Global**: 1000 requests per second, burst of 2000
- **Per IP**: 10 requests per second, burst of 20
- **Per User**: 100 requests per second, burst of 200

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Request limit for the window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time when the rate limit window resets

## Circuit Breaker

The service implements circuit breakers for upstream gRPC services. When a circuit breaker is open, requests will fail fast with `service_unavailable` errors rather than waiting for timeouts.

## HTTP Status Codes

- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required or failed
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `502 Bad Gateway`: Upstream service error
- `503 Service Unavailable`: Service temporarily unavailable (circuit breaker open)

## SDKs and Libraries

Currently, Clotho does not provide official SDKs. Use standard HTTP client libraries to interact with the API.

## Support

For API support and questions:
- Documentation: Available via `/swagger` endpoint
- Health Status: Available via `/health` endpoint
- Monitoring: Available via `/api/v1/monitoring/*` endpoints