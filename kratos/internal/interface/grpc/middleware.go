package grpc

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/julesChu12/fly/kratos/pkg/constants"
	"github.com/julesChu12/fly/mora/pkg/auth"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a new unary server interceptor for logging and recovery
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Recovery from panic
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] %v", r)
			}
		}()

		// Log request
		log.Printf("[gRPC] %s started", info.FullMethod)

		// Call handler
		resp, err := handler(ctx, req)

		// Log response
		duration := time.Since(start)
		if err != nil {
			log.Printf("[gRPC] %s failed in %v: %v", info.FullMethod, duration, err)
		} else {
			log.Printf("[gRPC] %s completed in %v", info.FullMethod, duration)
		}

		return resp, err
	}
}

// ContextInjectorInterceptor extracts metadata and injects into context
func ContextInjectorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		ctx = injectUintFromMetadata(ctx, md, constants.ContextKeyTenantID, "tenant_id", "x-tenant-id")
		ctx = injectUintFromMetadata(ctx, md, constants.ContextKeyUserID, "user_id", "x-user-id")

		return handler(ctx, req)
	}
}

// AuthInterceptor validates incoming JWT tokens and enriches context with claims.
func AuthInterceptor(secret string) grpc.UnaryServerInterceptor {
	if secret == "" {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		token := extractBearer(md)
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "missing bearer token")
		}

		claims, err := auth.ValidateToken(token, secret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		if claims.UserID != "" {
			if userID, err := strconv.ParseUint(claims.UserID, 10, 32); err == nil {
				ctx = context.WithValue(ctx, constants.ContextKeyUserID, uint(userID))
			}
		}

		return handler(ctx, req)
	}
}

// RateLimitInterceptor enforces a global rate limit for incoming RPC calls.
func RateLimitInterceptor(limiter *rate.Limiter) grpc.UnaryServerInterceptor {
	if limiter == nil {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !limiter.Allow() {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(ctx, req)
	}
}

func injectUintFromMetadata(ctx context.Context, md metadata.MD, key constants.ContextKey, aliases ...string) context.Context {
	for _, alias := range aliases {
		if values := md.Get(alias); len(values) > 0 {
			if id, err := strconv.ParseUint(values[0], 10, 32); err == nil {
				return context.WithValue(ctx, key, uint(id))
			}
		}
	}
	return ctx
}

func extractBearer(md metadata.MD) string {
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return ""
	}

	header := authHeaders[0]
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return ""
	}

	return strings.TrimSpace(header[len("bearer "):])
}
