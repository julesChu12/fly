package gozero

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julesChu12/mora/pkg/auth"
)

const (
	// ContextKeyUserID is the key used to store user ID in go-zero context
	ContextKeyUserID = "user_id"
	// ContextKeyClaims is the key used to store claims in go-zero context
	ContextKeyClaims = "claims"
)

// AuthMiddlewareConfig holds the configuration for auth middleware
type AuthMiddlewareConfig struct {
	Secret string
	// SkipPaths contains paths that should skip authentication
	SkipPaths []string
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// writeErrorResponse writes an error response
func writeErrorResponse(w http.ResponseWriter, code int, err, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ErrorResponse{
		Error:   err,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// AuthMiddleware creates a new authentication middleware for go-zero
func AuthMiddleware(config AuthMiddlewareConfig) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check if current path should skip authentication
			currentPath := r.URL.Path
			for _, path := range config.SkipPaths {
				// Support exact matching
				if path == currentPath {
					next(w, r)
					return
				}
				// Support path/* patterns
				if strings.HasSuffix(path, "/*") {
					prefix := strings.TrimSuffix(path, "/*")
					if strings.HasPrefix(currentPath, prefix) {
						next(w, r)
						return
					}
				}
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "missing authorization header")
				return
			}

			// Check Bearer token format
			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "invalid authorization header format")
				return
			}

			// Extract token
			token := strings.TrimPrefix(authHeader, bearerPrefix)
			if token == "" {
				writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "missing token")
				return
			}

			// Validate token
			claims, err := auth.ValidateToken(token, config.Secret)
			if err != nil {
				var message string
				switch err {
				case auth.ErrExpiredToken:
					message = "token expired"
				case auth.ErrMalformedToken:
					message = "malformed token"
				default:
					message = "invalid token"
				}

				writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", message)
				return
			}

			// Store claims and user ID in context
			ctx := r.Context()
			ctx = WithClaims(ctx, claims)
			ctx = WithUserID(ctx, claims.UserID)

			// Continue with the modified context
			next(w, r.WithContext(ctx))
		}
	}
}
