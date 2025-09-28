// Package docs Clotho API Documentation
//
// Clotho API provides a unified HTTP/REST interface for external clients in the Fly ecosystem.
// It acts as an orchestration layer that forwards requests to domain services via gRPC.
//
//	Version: 0.1.0
//	Host: localhost:8080
//	BasePath: /
//	License: MIT
//	Schemes: http, https
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	SecurityDefinitions:
//	  BearerAuth:
//	    type: apiKey
//	    in: header
//	    name: Authorization
//	    description: Bearer token authentication. Enter "Bearer <token>" in the value field.
//
// swagger:meta
package docs

import "github.com/julesChu12/fly/clotho/internal/infrastructure/http/handler"

// Error response model
type ErrorResponse struct {
	Error   string `json:"error" example:"invalid_request"`
	Message string `json:"message" example:"The request is invalid"`
}

// Aliases for handler types to avoid duplication
type (
	HealthResponse  = handler.HealthResponse
	UserResponse    = handler.UserResponse
	ProfileResponse = handler.ProfileResponse
	ProfileRequest  = handler.ProfileRequest
)