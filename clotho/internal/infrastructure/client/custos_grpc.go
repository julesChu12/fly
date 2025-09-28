package client

import (
	"context"
	"fmt"
	"time"

	custosv1 "github.com/julesChu12/fly/clotho/api/proto"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// CustosClient represents a gRPC client for the Custos service
type CustosClient struct {
	conn   *grpc.ClientConn
	client custosv1.CustosServiceClient
	logger *logger.Logger
}

// UserInfo represents user information from Custos
type UserInfo struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Avatar    string `json:"avatar"`
	Bio       string `json:"bio"`
	Phone     string `json:"phone"`
	Location  string `json:"location"`
	Website   string `json:"website"`
	UserType  string `json:"user_type"`
	TenantID  int64  `json:"tenant_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// NewCustosClient creates a new Custos gRPC client
func NewCustosClient(address string, timeout time.Duration) (*CustosClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log := logger.NewDefault()
	log.Info("Connecting to Custos service", "address", address)

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Error("Failed to connect to Custos service", "address", address, "error", err.Error())
		return nil, fmt.Errorf("failed to connect to Custos service at %s: %w", address, err)
	}

	client := custosv1.NewCustosServiceClient(conn)
	log.Info("Successfully connected to Custos service", "address", address)

	return &CustosClient{
		conn:   conn,
		client: client,
		logger: log,
	}, nil
}

// GetUser retrieves user information by user ID
func (c *CustosClient) GetUser(ctx context.Context, userID int64) (*UserInfo, error) {
	req := &custosv1.GetUserRequest{
		UserId: userID,
	}

	c.logger.Debug("Calling Custos GetUser", "user_id", userID)

	resp, err := c.client.GetUser(ctx, req)
	if err != nil {
		c.logger.Error("GetUser gRPC call failed", "user_id", userID, "error", err.Error())

		// Handle specific gRPC error codes
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.NotFound:
				return nil, fmt.Errorf("user not found: %d", userID)
			case codes.PermissionDenied:
				return nil, fmt.Errorf("permission denied for user: %d", userID)
			case codes.Unavailable:
				return nil, fmt.Errorf("custos service unavailable")
			default:
				return nil, fmt.Errorf("failed to get user %d: %s", userID, s.Message())
			}
		}
		return nil, fmt.Errorf("failed to get user %d: %w", userID, err)
	}

	if resp.User == nil {
		c.logger.Warn("GetUser returned nil user", "user_id", userID)
		return nil, fmt.Errorf("user not found: %d", userID)
	}

	userInfo := convertProtoUserToUserInfo(resp.User)
	c.logger.Debug("GetUser successful", "user_id", userID, "username", userInfo.Username)

	return userInfo, nil
}

// ValidateToken validates a JWT token with the Custos service
func (c *CustosClient) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	req := &custosv1.ValidateTokenRequest{
		Token: token,
	}

	c.logger.Debug("Calling Custos ValidateToken")

	resp, err := c.client.ValidateToken(ctx, req)
	if err != nil {
		c.logger.Error("ValidateToken gRPC call failed", "error", err.Error())

		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.Unauthenticated:
				return nil, fmt.Errorf("invalid token")
			case codes.Unavailable:
				return nil, fmt.Errorf("custos service unavailable")
			default:
				return nil, fmt.Errorf("token validation failed: %s", s.Message())
			}
		}
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !resp.IsValid {
		c.logger.Warn("Token validation failed", "error", resp.ErrorMessage)
		return nil, fmt.Errorf("invalid token: %s", resp.ErrorMessage)
	}

	if resp.User == nil {
		c.logger.Error("ValidateToken returned valid but nil user")
		return nil, fmt.Errorf("token validation failed: user data missing")
	}

	userInfo := convertProtoUserToUserInfo(resp.User)
	c.logger.Debug("ValidateToken successful", "user_id", userInfo.ID, "username", userInfo.Username)

	return userInfo, nil
}

// UpdateUser updates user profile information
func (c *CustosClient) UpdateUser(ctx context.Context, userID int64, userInfo *UserInfo, updateMask []string) (*UserInfo, error) {
	protoUser := convertUserInfoToProtoUser(userInfo)

	req := &custosv1.UpdateUserRequest{
		UserId:     userID,
		User:       protoUser,
		UpdateMask: updateMask,
	}

	c.logger.Debug("Calling Custos UpdateUser", "user_id", userID, "update_mask", updateMask)

	resp, err := c.client.UpdateUser(ctx, req)
	if err != nil {
		c.logger.Error("UpdateUser gRPC call failed", "user_id", userID, "error", err.Error())

		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.NotFound:
				return nil, fmt.Errorf("user not found: %d", userID)
			case codes.PermissionDenied:
				return nil, fmt.Errorf("permission denied for user: %d", userID)
			case codes.InvalidArgument:
				return nil, fmt.Errorf("invalid user data: %s", s.Message())
			case codes.Unavailable:
				return nil, fmt.Errorf("custos service unavailable")
			default:
				return nil, fmt.Errorf("failed to update user %d: %s", userID, s.Message())
			}
		}
		return nil, fmt.Errorf("failed to update user %d: %w", userID, err)
	}

	updatedUserInfo := convertProtoUserToUserInfo(resp.User)
	c.logger.Debug("UpdateUser successful", "user_id", userID, "username", updatedUserInfo.Username)

	return updatedUserInfo, nil
}

// GetUserPreferences retrieves user preferences
func (c *CustosClient) GetUserPreferences(ctx context.Context, userID int64) (map[string]string, error) {
	req := &custosv1.GetUserPreferencesRequest{
		UserId: userID,
	}

	c.logger.Debug("Calling Custos GetUserPreferences", "user_id", userID)

	resp, err := c.client.GetUserPreferences(ctx, req)
	if err != nil {
		c.logger.Error("GetUserPreferences gRPC call failed", "user_id", userID, "error", err.Error())

		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.NotFound:
				return nil, fmt.Errorf("user preferences not found: %d", userID)
			case codes.PermissionDenied:
				return nil, fmt.Errorf("permission denied for user: %d", userID)
			case codes.Unavailable:
				return nil, fmt.Errorf("custos service unavailable")
			default:
				return nil, fmt.Errorf("failed to get user preferences %d: %s", userID, s.Message())
			}
		}
		return nil, fmt.Errorf("failed to get user preferences %d: %w", userID, err)
	}

	c.logger.Debug("GetUserPreferences successful", "user_id", userID, "preferences_count", len(resp.Preferences))
	return resp.Preferences, nil
}

// UpdateUserPreferences updates user preferences
func (c *CustosClient) UpdateUserPreferences(ctx context.Context, userID int64, preferences map[string]string) (map[string]string, error) {
	req := &custosv1.UpdateUserPreferencesRequest{
		UserId:      userID,
		Preferences: preferences,
	}

	c.logger.Debug("Calling Custos UpdateUserPreferences", "user_id", userID, "preferences_count", len(preferences))

	resp, err := c.client.UpdateUserPreferences(ctx, req)
	if err != nil {
		c.logger.Error("UpdateUserPreferences gRPC call failed", "user_id", userID, "error", err.Error())

		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.NotFound:
				return nil, fmt.Errorf("user not found: %d", userID)
			case codes.PermissionDenied:
				return nil, fmt.Errorf("permission denied for user: %d", userID)
			case codes.InvalidArgument:
				return nil, fmt.Errorf("invalid preferences data: %s", s.Message())
			case codes.Unavailable:
				return nil, fmt.Errorf("custos service unavailable")
			default:
				return nil, fmt.Errorf("failed to update user preferences %d: %s", userID, s.Message())
			}
		}
		return nil, fmt.Errorf("failed to update user preferences %d: %w", userID, err)
	}

	c.logger.Debug("UpdateUserPreferences successful", "user_id", userID, "updated_preferences_count", len(resp.Preferences))
	return resp.Preferences, nil
}

// GetUserStatistics retrieves user statistics
func (c *CustosClient) GetUserStatistics(ctx context.Context, userID int64) (map[string]int64, error) {
	req := &custosv1.GetUserStatisticsRequest{
		UserId: userID,
	}

	c.logger.Debug("Calling Custos GetUserStatistics", "user_id", userID)

	resp, err := c.client.GetUserStatistics(ctx, req)
	if err != nil {
		c.logger.Error("GetUserStatistics gRPC call failed", "user_id", userID, "error", err.Error())

		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.NotFound:
				return nil, fmt.Errorf("user statistics not found: %d", userID)
			case codes.PermissionDenied:
				return nil, fmt.Errorf("permission denied for user: %d", userID)
			case codes.Unavailable:
				return nil, fmt.Errorf("custos service unavailable")
			default:
				return nil, fmt.Errorf("failed to get user statistics %d: %s", userID, s.Message())
			}
		}
		return nil, fmt.Errorf("failed to get user statistics %d: %w", userID, err)
	}

	c.logger.Debug("GetUserStatistics successful", "user_id", userID, "statistics_count", len(resp.Statistics))
	return resp.Statistics, nil
}

// Close closes the gRPC connection
func (c *CustosClient) Close() error {
	c.logger.Info("Closing Custos gRPC connection")
	return c.conn.Close()
}

// Helper functions to convert between protobuf and internal types

func convertProtoUserToUserInfo(protoUser *custosv1.User) *UserInfo {
	userType := "unknown"
	switch protoUser.UserType {
	case custosv1.UserType_USER_TYPE_CUSTOMER:
		userType = "customer"
	case custosv1.UserType_USER_TYPE_ADMIN:
		userType = "admin"
	case custosv1.UserType_USER_TYPE_OPERATOR:
		userType = "operator"
	}

	status := "unknown"
	switch protoUser.Status {
	case custosv1.UserStatus_USER_STATUS_ACTIVE:
		status = "active"
	case custosv1.UserStatus_USER_STATUS_INACTIVE:
		status = "inactive"
	case custosv1.UserStatus_USER_STATUS_SUSPENDED:
		status = "suspended"
	case custosv1.UserStatus_USER_STATUS_DELETED:
		status = "deleted"
	}

	createdAt := ""
	if protoUser.CreatedAt != nil {
		createdAt = protoUser.CreatedAt.AsTime().Format(time.RFC3339)
	}

	updatedAt := ""
	if protoUser.UpdatedAt != nil {
		updatedAt = protoUser.UpdatedAt.AsTime().Format(time.RFC3339)
	}

	return &UserInfo{
		ID:        protoUser.Id,
		Username:  protoUser.Username,
		Email:     protoUser.Email,
		FirstName: protoUser.FirstName,
		LastName:  protoUser.LastName,
		Avatar:    protoUser.Avatar,
		Bio:       protoUser.Bio,
		Phone:     protoUser.Phone,
		Location:  protoUser.Location,
		Website:   protoUser.Website,
		UserType:  userType,
		TenantID:  protoUser.TenantId,
		Status:    status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func convertUserInfoToProtoUser(userInfo *UserInfo) *custosv1.User {
	var userType custosv1.UserType
	switch userInfo.UserType {
	case "customer":
		userType = custosv1.UserType_USER_TYPE_CUSTOMER
	case "admin":
		userType = custosv1.UserType_USER_TYPE_ADMIN
	case "operator":
		userType = custosv1.UserType_USER_TYPE_OPERATOR
	default:
		userType = custosv1.UserType_USER_TYPE_UNSPECIFIED
	}

	var status custosv1.UserStatus
	switch userInfo.Status {
	case "active":
		status = custosv1.UserStatus_USER_STATUS_ACTIVE
	case "inactive":
		status = custosv1.UserStatus_USER_STATUS_INACTIVE
	case "suspended":
		status = custosv1.UserStatus_USER_STATUS_SUSPENDED
	case "deleted":
		status = custosv1.UserStatus_USER_STATUS_DELETED
	default:
		status = custosv1.UserStatus_USER_STATUS_UNSPECIFIED
	}

	return &custosv1.User{
		Id:        userInfo.ID,
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
		Avatar:    userInfo.Avatar,
		Bio:       userInfo.Bio,
		Phone:     userInfo.Phone,
		Location:  userInfo.Location,
		Website:   userInfo.Website,
		UserType:  userType,
		TenantId:  userInfo.TenantID,
		Status:    status,
	}
}