package custos

import (
	"context"
	"fmt"
	"time"

	custosv1 "github.com/julesChu12/fly/custos/api/proto/custos/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps the Custos gRPC client
type Client struct {
	conn   *grpc.ClientConn
	client custosv1.CustosServiceClient
}

// TokenValidationResult represents the result of token validation
type TokenValidationResult struct {
	IsValid      bool
	UserID       uint
	Username     string
	Email        string
	TenantID     uint
	UserType     string
	ErrorMessage string
}

// NewClient creates a new Custos gRPC client
func NewClient(endpoint string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Custos at %s: %w", endpoint, err)
	}

	return &Client{
		conn:   conn,
		client: custosv1.NewCustosServiceClient(conn),
	}, nil
}

// ValidateToken validates a JWT token via Custos
func (c *Client) ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error) {
	req := &custosv1.ValidateTokenRequest{
		Token: token,
	}

	resp, err := c.client.ValidateToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Custos ValidateToken: %w", err)
	}

	if !resp.IsValid {
		return &TokenValidationResult{
			IsValid:      false,
			ErrorMessage: resp.ErrorMessage,
		}, nil
	}

	user := resp.User
	if user == nil {
		return &TokenValidationResult{
			IsValid:      false,
			ErrorMessage: "user information not found",
		}, nil
	}

	return &TokenValidationResult{
		IsValid:  true,
		UserID:   uint(user.Id),
		Username: user.Username,
		Email:    user.Email,
		TenantID: uint(user.TenantId),
		UserType: user.UserType.String(),
	}, nil
}

// GetUser retrieves user information by user ID
func (c *Client) GetUser(ctx context.Context, userID uint) (*custosv1.User, error) {
	req := &custosv1.GetUserRequest{
		UserId: int64(userID),
	}

	resp, err := c.client.GetUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Custos GetUser: %w", err)
	}

	return resp.User, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
