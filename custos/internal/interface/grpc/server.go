package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	custosv1 "github.com/julesChu12/fly/custos/api/proto/custos/v1"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"github.com/julesChu12/fly/custos/internal/domain/service/token"
	"github.com/julesChu12/fly/custos/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CustosGRPCServer implements the gRPC server for Custos service
type CustosGRPCServer struct {
	custosv1.UnimplementedCustosServiceServer
	userRepo        repository.UserRepository
	userProfileRepo repository.UserProfileRepository
	sessionRepo     repository.SessionRepository
	tokenService    *token.TokenService
	grpcServer      *grpc.Server
}

// NewCustosGRPCServer creates a new gRPC server instance
func NewCustosGRPCServer(
	userRepo repository.UserRepository,
	userProfileRepo repository.UserProfileRepository,
	sessionRepo repository.SessionRepository,
	tokenService *token.TokenService,
) *CustosGRPCServer {
	return &CustosGRPCServer{
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
		sessionRepo:     sessionRepo,
		tokenService:    tokenService,
	}
}

// GetUser retrieves user information by user ID
func (s *CustosGRPCServer) GetUser(ctx context.Context, req *custosv1.GetUserRequest) (*custosv1.GetUserResponse, error) {
	log.Printf("gRPC GetUser called with UserID: %d", req.UserId)

	// Get user from repository
	user, err := s.userRepo.GetByID(ctx, uint(req.UserId))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get profile data
	var firstName, avatar string
	profile, err := s.userProfileRepo.GetByUserID(ctx, user.ID)
	if err == nil {
		firstName = profile.Nickname
		avatar = profile.Avatar
	}

	// Convert entity to gRPC message
	grpcUser := &custosv1.User{
		Id:        int64(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: firstName, // Use Nickname as FirstName
		LastName:  "",        // LastName not available in entity
		Avatar:    avatar,
		Bio:       "", // Bio not available in entity
		Phone:     "", // Phone not available in entity
		Location:  "", // Location not available in entity
		Website:   "", // Website not available in entity
		UserType:  convertUserTypeToEnum(user.UserType),
		TenantId:  convertTenantIDToInt64(user.TenantID),
		Status:    convertUserStatusToEnum(user.Status),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	return &custosv1.GetUserResponse{User: grpcUser}, nil
}

// ValidateToken validates a JWT token and returns user information
func (s *CustosGRPCServer) ValidateToken(ctx context.Context, req *custosv1.ValidateTokenRequest) (*custosv1.ValidateTokenResponse, error) {
	log.Printf("gRPC ValidateToken called")

	// Validate token using token service
	claims, err := s.tokenService.ValidateToken(req.Token)
	if err != nil {
		return &custosv1.ValidateTokenResponse{
			IsValid:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Get user ID from claims
	userID := claims.UserID

	// Get user information
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return &custosv1.ValidateTokenResponse{
			IsValid:      false,
			ErrorMessage: "user not found",
		}, nil
	}

	// Get profile data
	var firstName, avatar string
	profile, err := s.userProfileRepo.GetByUserID(ctx, user.ID)
	if err == nil {
		firstName = profile.Nickname
		avatar = profile.Avatar
	}

	// Convert entity to gRPC message
	grpcUser := &custosv1.User{
		Id:        int64(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: firstName, // Use Nickname as FirstName
		LastName:  "",        // LastName not available in entity
		Avatar:    avatar,
		Bio:       "", // Bio not available in entity
		Phone:     "", // Phone not available in entity
		Location:  "", // Location not available in entity
		Website:   "", // Website not available in entity
		UserType:  convertUserTypeToEnum(user.UserType),
		TenantId:  convertTenantIDToInt64(user.TenantID),
		Status:    convertUserStatusToEnum(user.Status),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	return &custosv1.ValidateTokenResponse{
		User:    grpcUser,
		IsValid: true,
	}, nil
}

// Start starts the gRPC server on the specified port
func (s *CustosGRPCServer) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	s.grpcServer = grpc.NewServer()

	// Register the Custos service
	custosv1.RegisterCustosServiceServer(s.grpcServer, s)

	// Enable reflection for debugging
	reflection.Register(s.grpcServer)

	log.Printf("gRPC server starting on port %s", port)

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *CustosGRPCServer) Stop() {
	if s.grpcServer != nil {
		log.Println("Stopping gRPC server...")
		s.grpcServer.GracefulStop()
	}
}

// HTTP-to-gRPC adapter methods for easy integration

// ServeHTTP provides an HTTP interface to gRPC methods for debugging
func (s *CustosGRPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// This would implement HTTP-to-gRPC bridging
	// For production, consider using grpc-gateway
}

// Helper functions for type conversion

// convertUserTypeToEnum converts UserType to protobuf enum
func convertUserTypeToEnum(userType types.UserType) custosv1.UserType {
	switch userType {
	case types.UserTypeStaff:
		return custosv1.UserType_USER_TYPE_ADMIN
	case types.UserTypePartner:
		return custosv1.UserType_USER_TYPE_OPERATOR
	default:
		return custosv1.UserType_USER_TYPE_CUSTOMER
	}
}

// convertUserStatusToEnum converts UserStatus to protobuf enum
func convertUserStatusToEnum(status types.UserStatus) custosv1.UserStatus {
	switch status {
	case types.UserStatusActive:
		return custosv1.UserStatus_USER_STATUS_ACTIVE
	case types.UserStatusInactive:
		return custosv1.UserStatus_USER_STATUS_INACTIVE
	case types.UserStatusFrozen:
		return custosv1.UserStatus_USER_STATUS_SUSPENDED
	case types.UserStatusDeleted:
		return custosv1.UserStatus_USER_STATUS_DELETED
	default:
		return custosv1.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

// convertTenantIDToInt64 converts *uint to int64 for gRPC
func convertTenantIDToInt64(tenantID *uint) int64 {
	if tenantID == nil {
		return 0
	}
	return int64(*tenantID)
}
