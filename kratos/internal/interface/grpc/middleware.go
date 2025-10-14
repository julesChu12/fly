package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/julesChu12/fly/kratos/pkg/constants"
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
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		// Extract tenant_id
		if tenantIDs := md.Get("tenant_id"); len(tenantIDs) > 0 {
			var tenantID uint
			if _, err := fmt.Sscanf(tenantIDs[0], "%d", &tenantID); err == nil {
				ctx = context.WithValue(ctx, constants.ContextKeyTenantID, tenantID)
			}
		}

		// Extract user_id
		if userIDs := md.Get("user_id"); len(userIDs) > 0 {
			var userID uint
			if _, err := fmt.Sscanf(userIDs[0], "%d", &userID); err == nil {
				ctx = context.WithValue(ctx, constants.ContextKeyUserID, userID)
			}
		}

		return handler(ctx, req)
	}
}
