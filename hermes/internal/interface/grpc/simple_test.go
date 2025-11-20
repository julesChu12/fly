package grpc

import (
	"testing"

	pb "github.com/julesChu12/fly/hermes/api/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestContactGRPCHandler_NewContactGRPCHandler tests constructor
func TestContactGRPCHandler_NewContactGRPCHandler(t *testing.T) {
	// Arrange & Act
	handler := NewContactGRPCHandler(nil)

	// Assert
	assert.NotNil(t, handler)
}

// TestCustomerGRPCHandler_NewCustomerGRPCHandler tests constructor
func TestCustomerGRPCHandler_NewCustomerGRPCHandler(t *testing.T) {
	// Arrange & Act
	handler := NewCustomerGRPCHandler(nil)

	// Assert
	assert.NotNil(t, handler)
}

// TestGRPCErrors tests gRPC error handling
func TestGRPCErrors(t *testing.T) {
	// Test status error creation
	err := status.Errorf(codes.Internal, "test error: %v", assert.AnError)

	// Assert
	assert.Error(t, err)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcStatus.Code())
	assert.Contains(t, grpcStatus.Message(), "test error")
}

// TestProtoConversion tests protobuf message creation
func TestProtoConversion(t *testing.T) {
	// Test proto message creation
	req := &pb.CreateCustomerRequest{
		Name:  "Test Customer",
		Phone: "+1234567890",
		Email: "test@example.com",
		Tags:  "tag1,tag2",
	}

	// Assert
	assert.NotNil(t, req)
	assert.Equal(t, "Test Customer", req.Name)
	assert.Equal(t, "+1234567890", req.Phone)
	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "tag1,tag2", req.Tags)
}

// TestContactProtoConversion tests contact protobuf message creation
func TestContactProtoConversion(t *testing.T) {
	// Test proto message creation
	req := &pb.CreateContactRequest{
		CustomerId: 1,
		Type:       "email",
		Value:      "test@example.com",
		IsPrimary:  true,
	}

	// Assert
	assert.NotNil(t, req)
	assert.Equal(t, uint32(1), req.CustomerId)
	assert.Equal(t, "email", req.Type)
	assert.Equal(t, "test@example.com", req.Value)
	assert.True(t, req.IsPrimary)
}