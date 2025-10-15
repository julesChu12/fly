package middleware

import (
	"context"
	"time"

	"github.com/julesChu12/fly/kratos/internal/infrastructure/custos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// AuthorizationMetadataKey is the metadata key for authorization token
	AuthorizationMetadataKey = "authorization"
)

// GRPCAuthInterceptor handles JWT authentication for gRPC requests
type GRPCAuthInterceptor struct {
	custosClient *custos.Client
}

// NewGRPCAuthInterceptor creates a new gRPC authentication interceptor
func NewGRPCAuthInterceptor(custosClient *custos.Client) *GRPCAuthInterceptor {
	return &GRPCAuthInterceptor{
		custosClient: custosClient,
	}
}

// UnaryInterceptor returns a unary server interceptor for authentication
func (i *GRPCAuthInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract token from metadata
		token, err := extractTokenFromMetadata(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "missing authorization token")
		}

		// Validate token via Custos
		authCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		result, err := i.custosClient.ValidateToken(authCtx, token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "failed to validate token: %v", err)
		}

		if !result.IsValid {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %s", result.ErrorMessage)
		}

		// Inject user context
		ctx = injectUserContext(ctx, result)

		// Call the handler
		return handler(ctx, req)
	}
}

// StreamInterceptor returns a stream server interceptor for authentication
func (i *GRPCAuthInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Extract token from metadata
		token, err := extractTokenFromMetadata(ss.Context())
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "missing authorization token")
		}

		// Validate token via Custos
		authCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		result, err := i.custosClient.ValidateToken(authCtx, token)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "failed to validate token: %v", err)
		}

		if !result.IsValid {
			return status.Errorf(codes.Unauthenticated, "invalid token: %s", result.ErrorMessage)
		}

		// Create new stream with injected context
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          injectUserContext(ss.Context(), result),
		}

		// Call the handler
		return handler(srv, wrapped)
	}
}

// extractTokenFromMetadata extracts JWT token from gRPC metadata
func extractTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	tokens := md.Get(AuthorizationMetadataKey)
	if len(tokens) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "missing authorization token")
	}

	token := tokens[0]
	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	return token, nil
}

// injectUserContext injects user information into context
func injectUserContext(ctx context.Context, result *custos.TokenValidationResult) context.Context {
	ctx = context.WithValue(ctx, "user_id", result.UserID)
	ctx = context.WithValue(ctx, "username", result.Username)
	ctx = context.WithValue(ctx, "tenant_id", result.TenantID)
	ctx = context.WithValue(ctx, "user_type", result.UserType)
	ctx = context.WithValue(ctx, "email", result.Email)
	return ctx
}

// wrappedServerStream wraps grpc.ServerStream with a custom context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
