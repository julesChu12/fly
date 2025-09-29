package grpc

import (
	"context"

	"hermes/internal/application/service"
	"hermes/pkg/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CustomerGRPCHandler implements the gRPC CustomerService interface
// 客户服务gRPC处理器，提供客户管理的所有gRPC接口
type CustomerGRPCHandler struct {
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
func (h *CustomerGRPCHandler) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CreateCustomerResponse, error) {
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
	return &CreateCustomerResponse{
		Customer: h.toProtoCustomer(customer),
	}, nil
}

// GetCustomer retrieves a customer by ID via gRPC
// 通过gRPC根据ID获取客户信息
func (h *CustomerGRPCHandler) GetCustomer(ctx context.Context, req *GetCustomerRequest) (*GetCustomerResponse, error) {
	customer, err := h.customerService.GetCustomer(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "customer not found: %v", err)
	}

	return &GetCustomerResponse{
		Customer: h.toProtoCustomer(customer),
	}, nil
}

// GetCustomerWithContacts retrieves a customer with contacts via gRPC
// 通过gRPC获取客户及其联系方式
func (h *CustomerGRPCHandler) GetCustomerWithContacts(ctx context.Context, req *GetCustomerRequest) (*GetCustomerWithContactsResponse, error) {
	customer, err := h.customerService.GetCustomerWithContacts(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "customer not found: %v", err)
	}

	protoCustomer := h.toProtoCustomer(customer)

	// 添加联系方式
	for _, contact := range customer.Contacts {
		protoCustomer.Contacts = append(protoCustomer.Contacts, &Contact{
			Id:         uint32(contact.ID),
			CustomerId: uint32(contact.CustomerID),
			Type:       contact.Type,
			Value:      contact.Value,
			IsPrimary:  contact.IsPrimary,
			CreatedAt:  timestamppb.New(contact.CreatedAt),
			UpdatedAt:  timestamppb.New(contact.UpdatedAt),
		})
	}

	return &GetCustomerWithContactsResponse{
		Customer: protoCustomer,
	}, nil
}

// UpdateCustomer updates a customer via gRPC
// 通过gRPC更新客户信息
func (h *CustomerGRPCHandler) UpdateCustomer(ctx context.Context, req *UpdateCustomerRequest) (*UpdateCustomerResponse, error) {
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

	return &UpdateCustomerResponse{
		Customer: h.toProtoCustomer(customer),
	}, nil
}

// DeleteCustomer deletes a customer via gRPC
// 通过gRPC删除客户
func (h *CustomerGRPCHandler) DeleteCustomer(ctx context.Context, req *DeleteCustomerRequest) (*DeleteCustomerResponse, error) {
	err := h.customerService.DeleteCustomer(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete customer: %v", err)
	}

	return &DeleteCustomerResponse{
		Success: true,
	}, nil
}

// ListCustomers lists customers via gRPC
// 通过gRPC获取客户列表
func (h *CustomerGRPCHandler) ListCustomers(ctx context.Context, req *ListCustomersRequest) (*ListCustomersResponse, error) {
	listReq := &types.ListRequest{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	result, err := h.customerService.ListCustomers(ctx, listReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list customers: %v", err)
	}

	customers := make([]*Customer, 0)
	if customerList, ok := result.Data.([]types.CustomerResponse); ok {
		for _, customer := range customerList {
			customers = append(customers, &Customer{
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

	return &ListCustomersResponse{
		Customers: customers,
		Total:     result.Total,
		Page:      int32(result.Page),
		Size:      int32(result.Size),
	}, nil
}

// toProtoCustomer converts types.CustomerResponse to proto Customer
// 将内部类型转换为protobuf类型
func (h *CustomerGRPCHandler) toProtoCustomer(customer *types.CustomerResponse) *Customer {
	return &Customer{
		Id:        uint32(customer.ID),
		Name:      customer.Name,
		Phone:     customer.Phone,
		Email:     customer.Email,
		Tags:      customer.Tags,
		CreatedAt: timestamppb.New(customer.CreatedAt),
		UpdatedAt: timestamppb.New(customer.UpdatedAt),
	}
}

// Proto message definitions (normally generated by protoc)
// 以下是protobuf消息定义（通常由protoc生成）

type CreateCustomerRequest struct {
	Name  string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Phone string `protobuf:"bytes,2,opt,name=phone,proto3" json:"phone,omitempty"`
	Email string `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	Tags  string `protobuf:"bytes,4,opt,name=tags,proto3" json:"tags,omitempty"`
}

type CreateCustomerResponse struct {
	Customer *Customer `protobuf:"bytes,1,opt,name=customer,proto3" json:"customer,omitempty"`
}

type GetCustomerRequest struct {
	Id uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

type GetCustomerResponse struct {
	Customer *Customer `protobuf:"bytes,1,opt,name=customer,proto3" json:"customer,omitempty"`
}

type GetCustomerWithContactsResponse struct {
	Customer *Customer `protobuf:"bytes,1,opt,name=customer,proto3" json:"customer,omitempty"`
}

type UpdateCustomerRequest struct {
	Id    uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Phone string `protobuf:"bytes,3,opt,name=phone,proto3" json:"phone,omitempty"`
	Email string `protobuf:"bytes,4,opt,name=email,proto3" json:"email,omitempty"`
	Tags  string `protobuf:"bytes,5,opt,name=tags,proto3" json:"tags,omitempty"`
}

type UpdateCustomerResponse struct {
	Customer *Customer `protobuf:"bytes,1,opt,name=customer,proto3" json:"customer,omitempty"`
}

type DeleteCustomerRequest struct {
	Id uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

type DeleteCustomerResponse struct {
	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

type ListCustomersRequest struct {
	Page     int32 `protobuf:"varint,1,opt,name=page,proto3" json:"page,omitempty"`
	PageSize int32 `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
}

type ListCustomersResponse struct {
	Customers []*Customer `protobuf:"bytes,1,rep,name=customers,proto3" json:"customers,omitempty"`
	Total     int64       `protobuf:"varint,2,opt,name=total,proto3" json:"total,omitempty"`
	Page      int32       `protobuf:"varint,3,opt,name=page,proto3" json:"page,omitempty"`
	Size      int32       `protobuf:"varint,4,opt,name=size,proto3" json:"size,omitempty"`
}

type Customer struct {
	Id        uint32                 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name      string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Phone     string                 `protobuf:"bytes,3,opt,name=phone,proto3" json:"phone,omitempty"`
	Email     string                 `protobuf:"bytes,4,opt,name=email,proto3" json:"email,omitempty"`
	Tags      string                 `protobuf:"bytes,5,opt,name=tags,proto3" json:"tags,omitempty"`
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Contacts  []*Contact             `protobuf:"bytes,8,rep,name=contacts,proto3" json:"contacts,omitempty"`
}

type Contact struct {
	Id         uint32                 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	CustomerId uint32                 `protobuf:"varint,2,opt,name=customer_id,json=customerId,proto3" json:"customer_id,omitempty"`
	Type       string                 `protobuf:"bytes,3,opt,name=type,proto3" json:"type,omitempty"`
	Value      string                 `protobuf:"bytes,4,opt,name=value,proto3" json:"value,omitempty"`
	IsPrimary  bool                   `protobuf:"varint,5,opt,name=is_primary,json=isPrimary,proto3" json:"is_primary,omitempty"`
	CreatedAt  *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt  *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}