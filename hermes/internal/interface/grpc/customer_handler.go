package grpc

import (
	"context"

	pb "github.com/julesChu12/fly/hermes/api/proto"
	"github.com/julesChu12/fly/hermes/internal/application/service"
	"github.com/julesChu12/fly/hermes/pkg/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CustomerGRPCHandler implements the gRPC CustomerService interface
// 客户服务gRPC处理器，提供客户管理的所有gRPC接口
type CustomerGRPCHandler struct {
	pb.UnimplementedCustomerServiceServer
	customerService service.CustomerService
}

// NewCustomerGRPCHandler creates a new CustomerGRPCHandler
// 创建新的客户gRPC处理器
func NewCustomerGRPCHandler(customerService service.CustomerService) *CustomerGRPCHandler {
	return &CustomerGRPCHandler{
		customerService: customerService,
	}
}

// CreateCustomer creates a new customer via gRPC
// 通过gRPC创建新客户
func (h *CustomerGRPCHandler) CreateCustomer(ctx context.Context, req *pb.CreateCustomerRequest) (*pb.CreateCustomerResponse, error) {
	// 转换请求参数
	createReq := &types.CreateCustomerRequest{
		Name:  req.Name,
		Phone: req.Phone,
		Email: req.Email,
		Tags:  req.Tags,
	}

	// 调用业务服务
	customer, err := h.customerService.CreateCustomer(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create customer: %v", err)
	}

	// 转换响应
	return &pb.CreateCustomerResponse{
		Customer: h.toProtoCustomer(customer),
	}, nil
}

// GetCustomer retrieves a customer by ID via gRPC
// 通过gRPC根据ID获取客户信息
func (h *CustomerGRPCHandler) GetCustomer(ctx context.Context, req *pb.GetCustomerRequest) (*pb.GetCustomerResponse, error) {
	customer, err := h.customerService.GetCustomer(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "customer not found: %v", err)
	}

	return &pb.GetCustomerResponse{
		Customer: h.toProtoCustomer(customer),
	}, nil
}

// GetCustomerWithContacts retrieves a customer with contacts via gRPC
// 通过gRPC获取客户及其联系方式
func (h *CustomerGRPCHandler) GetCustomerWithContacts(ctx context.Context, req *pb.GetCustomerRequest) (*pb.GetCustomerWithContactsResponse, error) {
	customer, err := h.customerService.GetCustomerWithContacts(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "customer not found: %v", err)
	}

	protoCustomer := h.toProtoCustomer(customer)

	// 添加联系方式
	for _, contact := range customer.Contacts {
		protoCustomer.Contacts = append(protoCustomer.Contacts, &pb.Contact{
			Id:         uint32(contact.ID),
			CustomerId: uint32(contact.CustomerID),
			Type:       contact.Type,
			Value:      contact.Value,
			IsPrimary:  contact.IsPrimary,
			CreatedAt:  timestamppb.New(contact.CreatedAt),
			UpdatedAt:  timestamppb.New(contact.UpdatedAt),
		})
	}

	return &pb.GetCustomerWithContactsResponse{
		Customer: protoCustomer,
	}, nil
}

// UpdateCustomer updates a customer via gRPC
// 通过gRPC更新客户信息
func (h *CustomerGRPCHandler) UpdateCustomer(ctx context.Context, req *pb.UpdateCustomerRequest) (*pb.UpdateCustomerResponse, error) {
	updateReq := &types.UpdateCustomerRequest{
		Name:  req.Name,
		Phone: req.Phone,
		Email: req.Email,
		Tags:  req.Tags,
	}

	customer, err := h.customerService.UpdateCustomer(ctx, uint(req.Id), updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update customer: %v", err)
	}

	return &pb.UpdateCustomerResponse{
		Customer: h.toProtoCustomer(customer),
	}, nil
}

// DeleteCustomer deletes a customer via gRPC
// 通过gRPC删除客户
func (h *CustomerGRPCHandler) DeleteCustomer(ctx context.Context, req *pb.DeleteCustomerRequest) (*pb.DeleteCustomerResponse, error) {
	err := h.customerService.DeleteCustomer(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete customer: %v", err)
	}

	return &pb.DeleteCustomerResponse{
		Success: true,
	}, nil
}

// ListCustomers lists customers via gRPC
// 通过gRPC获取客户列表
func (h *CustomerGRPCHandler) ListCustomers(ctx context.Context, req *pb.ListCustomersRequest) (*pb.ListCustomersResponse, error) {
	listReq := &types.ListRequest{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	result, err := h.customerService.ListCustomers(ctx, listReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list customers: %v", err)
	}

	customers := make([]*pb.Customer, 0)
	if customerList, ok := result.Data.([]types.CustomerResponse); ok {
		for _, customer := range customerList {
			customers = append(customers, &pb.Customer{
				Id:        uint32(customer.ID),
				Name:      customer.Name,
				Phone:     customer.Phone,
				Email:     customer.Email,
				Tags:      customer.Tags,
				CreatedAt: timestamppb.New(customer.CreatedAt),
				UpdatedAt: timestamppb.New(customer.UpdatedAt),
			})
		}
	}

	return &pb.ListCustomersResponse{
		Customers: customers,
		Total:     result.Total,
		Page:      int32(result.Page),
		Size:      int32(result.Size),
	}, nil
}

// toProtoCustomer converts types.CustomerResponse to proto Customer
// 将内部类型转换为protobuf类型
func (h *CustomerGRPCHandler) toProtoCustomer(customer *types.CustomerResponse) *pb.Customer {
	return &pb.Customer{
		Id:        uint32(customer.ID),
		Name:      customer.Name,
		Phone:     customer.Phone,
		Email:     customer.Email,
		Tags:      customer.Tags,
		CreatedAt: timestamppb.New(customer.CreatedAt),
		UpdatedAt: timestamppb.New(customer.UpdatedAt),
	}
}
