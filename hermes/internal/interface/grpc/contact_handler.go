package grpc

import (
	"context"

	pb "github.com/julesChu12/fly/hermes/api/proto"
	"github.com/julesChu12/fly/hermes/internal/domain/entity"
	"github.com/julesChu12/fly/hermes/internal/domain/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ContactGRPCHandler implements the gRPC ContactService interface
// 联系方式服务gRPC处理器
type ContactGRPCHandler struct {
	pb.UnimplementedContactServiceServer
	contactRepo repository.ContactRepository
}

// NewContactGRPCHandler creates a new ContactGRPCHandler
// 创建新的联系方式gRPC处理器
func NewContactGRPCHandler(contactRepo repository.ContactRepository) *ContactGRPCHandler {
	return &ContactGRPCHandler{
		contactRepo: contactRepo,
	}
}

// CreateContact creates a new contact via gRPC
// 通过gRPC创建新联系方式
func (h *ContactGRPCHandler) CreateContact(ctx context.Context, req *pb.CreateContactRequest) (*pb.CreateContactResponse, error) {
	contact := &entity.Contact{
		CustomerID: uint(req.CustomerId),
		Type:       req.Type,
		Value:      req.Value,
		IsPrimary:  req.IsPrimary,
	}

	if err := h.contactRepo.Create(ctx, contact); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create contact: %v", err)
	}

	return &pb.CreateContactResponse{
		Contact: h.toProtoContact(contact),
	}, nil
}

// GetContact retrieves a contact by ID via gRPC
// 通过gRPC根据ID获取联系方式
func (h *ContactGRPCHandler) GetContact(ctx context.Context, req *pb.GetContactRequest) (*pb.GetContactResponse, error) {
	contact, err := h.contactRepo.GetByID(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "contact not found: %v", err)
	}

	return &pb.GetContactResponse{
		Contact: h.toProtoContact(contact),
	}, nil
}

// UpdateContact updates a contact via gRPC
// 通过gRPC更新联系方式
func (h *ContactGRPCHandler) UpdateContact(ctx context.Context, req *pb.UpdateContactRequest) (*pb.UpdateContactResponse, error) {
	contact, err := h.contactRepo.GetByID(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "contact not found: %v", err)
	}

	if req.Type != "" {
		contact.Type = req.Type
	}
	if req.Value != "" {
		contact.Value = req.Value
	}
	contact.IsPrimary = req.IsPrimary

	if err := h.contactRepo.Update(ctx, contact); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update contact: %v", err)
	}

	return &pb.UpdateContactResponse{
		Contact: h.toProtoContact(contact),
	}, nil
}

// DeleteContact deletes a contact via gRPC
// 通过gRPC删除联系方式
func (h *ContactGRPCHandler) DeleteContact(ctx context.Context, req *pb.DeleteContactRequest) (*pb.DeleteContactResponse, error) {
	if err := h.contactRepo.Delete(ctx, uint(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete contact: %v", err)
	}

	return &pb.DeleteContactResponse{
		Success: true,
	}, nil
}

// ListContactsByCustomer lists contacts by customer ID via gRPC
// 通过gRPC获取客户的所有联系方式
func (h *ContactGRPCHandler) ListContactsByCustomer(ctx context.Context, req *pb.ListContactsByCustomerRequest) (*pb.ListContactsByCustomerResponse, error) {
	contacts, err := h.contactRepo.GetByCustomerID(ctx, uint(req.CustomerId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list contacts: %v", err)
	}

	protoContacts := make([]*pb.Contact, 0, len(contacts))
	for _, contact := range contacts {
		protoContacts = append(protoContacts, h.toProtoContact(contact))
	}

	return &pb.ListContactsByCustomerResponse{
		Contacts: protoContacts,
	}, nil
}

// toProtoContact converts entity.Contact to proto Contact
// 将内部类型转换为protobuf类型
func (h *ContactGRPCHandler) toProtoContact(contact *entity.Contact) *pb.Contact {
	if contact == nil {
		return nil
	}

	return &pb.Contact{
		Id:         uint32(contact.ID),
		CustomerId: uint32(contact.CustomerID),
		Type:       contact.Type,
		Value:      contact.Value,
		IsPrimary:  contact.IsPrimary,
		CreatedAt:  timestamppb.New(contact.CreatedAt),
		UpdatedAt:  timestamppb.New(contact.UpdatedAt),
	}
}
