package gozero

import (
	"context"

	"github.com/julesChu12/mora/pkg/auth"
)

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return userID
	}
	return ""
}

// WithClaims adds claims to context
func WithClaims(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, ContextKeyClaims, claims)
}

// GetClaims extracts claims from context
func GetClaims(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(ContextKeyClaims).(*auth.Claims); ok {
		return claims
	}
	return nil
}
