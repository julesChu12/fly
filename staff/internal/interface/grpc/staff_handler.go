package grpc

import (
	"context"
	"fmt"
	"time"

	staffv1 "github.com/julesChu12/fly/staff/api/proto/staff/v1"
	"github.com/julesChu12/fly/staff/internal/application/service"
	"github.com/julesChu12/fly/staff/internal/domain/dto"
	"github.com/julesChu12/fly/staff/internal/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// StaffServer implements the gRPC StaffService
type StaffServer struct {
	staffv1.UnimplementedStaffServiceServer
	staffService service.StaffService
}

// NewStaffServer creates a new StaffServer
func NewStaffServer(staffService service.StaffService) *StaffServer {
	return &StaffServer{
		staffService: staffService,
	}
}

// CreateStaff creates a new staff member
func (s *StaffServer) CreateStaff(ctx context.Context, req *staffv1.CreateStaffRequest) (*staffv1.CreateStaffResponse, error) {
	// Convert protobuf request to DTO
	createReq := &dto.CreateStaffRequest{
		Name:            req.Name,
		Email:           req.Email,
		Phone:           req.Phone,
		Gender:          req.Gender,
		Avatar:          req.Avatar.Value,
		Department:      req.Department,
		Position:        req.Position,
		RoleID:          req.RoleId,
		Status:          req.Status,
		HireDate:        req.HireDate.Value,
		Salary:          req.Salary.Value,
		Address:         req.Address.Value,
		EmergencyContact: req.EmergencyContact.Value,
		Notes:           req.Notes.Value,
		Skills:          req.Skills,
		WorkingHours:    req.WorkingHours,
		IsAvailable:     req.IsAvailable,
		Birthday:        req.Birthday.Value,
	}

	// Call service
	staff, err := s.staffService.CreateStaff(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create staff: %v", err)
	}

	// Convert response to protobuf
	return &staffv1.CreateStaffResponse{
		Staff: toProtoStaff(staff),
	}, nil
}

// GetStaff retrieves a staff member by ID
func (s *StaffServer) GetStaff(ctx context.Context, req *staffv1.GetStaffRequest) (*staffv1.GetStaffResponse, error) {
	staff, err := s.staffService.GetStaff(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Staff not found: %v", err)
	}

	return &staffv1.GetStaffResponse{
		Staff: toProtoStaff(staff),
	}, nil
}

// ListStaff retrieves a paginated list of staff members
func (s *StaffServer) ListStaff(ctx context.Context, req *staffv1.ListStaffRequest) (*staffv1.ListStaffResponse, error) {
	// Convert request
	filter := &dto.StaffFilter{
		Search:      req.Search.Value,
		Department:  req.Department.Value,
		RoleID:      req.RoleId.Value,
		Status:      req.Status.Value,
		IsAvailable: req.IsAvailable.Value,
		MinAge:      req.MinAge.Value,
		MaxAge:      req.MaxAge.Value,
		Gender:      req.Gender.Value,
		Skills:      req.Skills,
		Page:        int(req.Page),
		PageSize:    int(req.PageSize),
		Sort:        req.Sort,
		Order:       req.Order,
	}

	// Call service
	result, err := s.staffService.ListStaff(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list staff: %v", err)
	}

	// Convert response
	staffMembers := make([]*staffv1.Staff, len(result.Staff))
	for i, staff := range result.Staff {
		staffMembers[i] = toProtoStaff(&staff)
	}

	return &staffv1.ListStaffResponse{
		Staff:    staffMembers,
		Total:    result.Total,
		Page:     int32(result.Page),
		PageSize: int32(result.PageSize),
	}, nil
}

// UpdateStaff updates an existing staff member
func (s *StaffServer) UpdateStaff(ctx context.Context, req *staffv1.UpdateStaffRequest) (*staffv1.UpdateStaffResponse, error) {
	// Convert protobuf request to DTO
	updateReq := &dto.UpdateStaffRequest{
		Name:            req.Name.Value,
		Email:           req.Email.Value,
		Phone:           req.Phone.Value,
		Gender:          req.Gender.Value,
		Avatar:          req.Avatar.Value,
		Department:      req.Department.Value,
		Position:        req.Position.Value,
		RoleID:          req.RoleId.Value,
		Status:          req.Status.Value,
		HireDate:        req.HireDate.Value,
		Salary:          req.Salary.Value,
		Address:         req.Address.Value,
		EmergencyContact: req.EmergencyContact.Value,
		Notes:           req.Notes.Value,
		Skills:          req.Skills,
		WorkingHours:    req.WorkingHours,
		IsAvailable:     req.IsAvailable.Value,
		Birthday:        req.Birthday.Value,
	}

	// Call service
	staff, err := s.staffService.UpdateStaff(ctx, req.Id, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update staff: %v", err)
	}

	return &staffv1.UpdateStaffResponse{
		Staff: toProtoStaff(staff),
	}, nil
}

// DeleteStaff deletes a staff member
func (s *StaffServer) DeleteStaff(ctx context.Context, req *staffv1.DeleteStaffRequest) (*emptypb.Empty, error) {
	err := s.staffService.DeleteStaff(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete staff: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// UpdateStaffStatus updates the status of a staff member
func (s *StaffServer) UpdateStaffStatus(ctx context.Context, req *staffv1.UpdateStaffStatusRequest) (*staffv1.UpdateStaffStatusResponse, error) {
	staff, err := s.staffService.UpdateStaffStatus(ctx, req.Id, req.Status, req.Reason.Value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update staff status: %v", err)
	}

	return &staffv1.UpdateStaffStatusResponse{
		Staff: toProtoStaff(staff),
	}, nil
}

// GetAvailableStaff gets available staff members
func (s *StaffServer) GetAvailableStaff(ctx context.Context, req *staffv1.GetAvailableStaffRequest) (*staffv1.GetAvailableStaffResponse, error) {
	filter := &dto.AvailableStaffFilter{
		Department: req.Department.Value,
		Skills:     req.Skills,
		TimeSlot:   req.TimeSlot.Value,
	}

	staffMembers, err := s.staffService.GetAvailableStaff(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get available staff: %v", err)
	}

	protoStaff := make([]*staffv1.Staff, len(staffMembers))
	for i, staff := range staffMembers {
		protoStaff[i] = toProtoStaff(&staff)
	}

	return &staffv1.GetAvailableStaffResponse{
		Staff: protoStaff,
	}, nil
}

// ListRoles retrieves a list of staff roles
func (s *StaffServer) ListRoles(ctx context.Context, req *staffv1.ListRolesRequest) (*staffv1.ListRolesResponse, error) {
	filter := &dto.RoleFilter{
		Status: req.Status.Value,
		Sort:   req.Sort,
		Order:  req.Order,
	}

	roles, err := s.staffService.ListRoles(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list roles: %v", err)
	}

	protoRoles := make([]*staffv1.StaffRole, len(roles))
	for i, role := range roles {
		protoRoles[i] = toProtoRole(&role)
	}

	return &staffv1.ListRolesResponse{
		Roles: protoRoles,
	}, nil
}

// CreateRole creates a new staff role
func (s *StaffServer) CreateRole(ctx context.Context, req *staffv1.CreateRoleRequest) (*staffv1.CreateRoleResponse, error) {
	createReq := &dto.CreateRoleRequest{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description.Value,
		Permissions: req.Permissions,
		IsDefault:   req.IsDefault,
		SortOrder:   int(req.SortOrder),
		Status:      req.Status,
	}

	role, err := s.staffService.CreateRole(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create role: %v", err)
	}

	return &staffv1.CreateRoleResponse{
		Role: toProtoRole(role),
	}, nil
}

// GetRole retrieves a staff role by ID
func (s *StaffServer) GetRole(ctx context.Context, req *staffv1.GetRoleRequest) (*staffv1.GetRoleResponse, error) {
	role, err := s.staffService.GetRole(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Role not found: %v", err)
	}

	return &staffv1.GetRoleResponse{
		Role: toProtoRole(role),
	}, nil
}

// UpdateRole updates an existing staff role
func (s *StaffServer) UpdateRole(ctx context.Context, req *staffv1.UpdateRoleRequest) (*staffv1.UpdateRoleResponse, error) {
	updateReq := &dto.UpdateRoleRequest{
		Name:        req.Name.Value,
		Code:        req.Code.Value,
		Description: req.Description.Value,
		Permissions: req.Permissions,
		IsDefault:   req.IsDefault.Value,
		SortOrder:   req.SortOrder.Value,
		Status:      req.Status.Value,
	}

	role, err := s.staffService.UpdateRole(ctx, req.Id, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update role: %v", err)
	}

	return &staffv1.UpdateRoleResponse{
		Role: toProtoRole(role),
	}, nil
}

// DeleteRole deletes a staff role
func (s *StaffServer) DeleteRole(ctx context.Context, req *staffv1.DeleteRoleRequest) (*emptypb.Empty, error) {
	err := s.staffService.DeleteRole(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete role: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// GetAvailability gets staff availability
func (s *StaffServer) GetAvailability(ctx context.Context, req *staffv1.GetAvailabilityRequest) (*staffv1.GetAvailabilityResponse, error) {
	availabilities, err := s.staffService.GetAvailability(ctx, req.StaffId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get availability: %v", err)
	}

	protoAvailabilities := make([]*staffv1.StaffAvailability, len(availabilities))
	for i, avail := range availabilities {
		protoAvailabilities[i] = toProtoAvailability(&avail)
	}

	return &staffv1.GetAvailabilityResponse{
		Availabilities: protoAvailabilities,
	}, nil
}

// SetAvailability sets staff availability
func (s *StaffServer) SetAvailability(ctx context.Context, req *staffv1.SetAvailabilityRequest) (*staffv1.SetAvailabilityResponse, error) {
	// Convert protobuf request to DTO
	availabilities := make([]*dto.AvailabilityItem, len(req.Availabilities))
	for i, item := range req.Availabilities {
		availabilities[i] = &dto.AvailabilityItem{
			DayOfWeek:    int(item.DayOfWeek),
			StartTime:    item.StartTime,
			EndTime:      item.EndTime,
			IsAvailable:  item.IsAvailable,
			Notes:        item.Notes.Value,
		}
	}

	setReq := &dto.SetAvailabilityRequest{
		StaffID:      req.StaffId,
		Availabilities: availabilities,
	}

	result, err := s.staffService.SetAvailability(ctx, setReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to set availability: %v", err)
	}

	protoAvailabilities := make([]*staffv1.StaffAvailability, len(result))
	for i, avail := range result {
		protoAvailabilities[i] = toProtoAvailability(&avail)
	}

	return &staffv1.SetAvailabilityResponse{
		Availabilities: protoAvailabilities,
	}, nil
}

// GetStaffAvailability gets staff availability for a specific day
func (s *StaffServer) GetStaffAvailability(ctx context.Context, req *staffv1.GetStaffAvailabilityRequest) (*staffv1.GetStaffAvailabilityResponse, error) {
	availabilities, err := s.staffService.GetStaffAvailability(ctx, req.StaffId, int(req.DayOfWeek))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get staff availability: %v", err)
	}

	protoAvailabilities := make([]*staffv1.StaffAvailability, len(availabilities))
	for i, avail := range availabilities {
		protoAvailabilities[i] = toProtoAvailability(&avail)
	}

	return &staffv1.GetStaffAvailabilityResponse{
		Availabilities: protoAvailabilities,
	}, nil
}

// Helper functions for converting between protobuf and internal types

func toProtoStaff(staff *entity.Staff) *staffv1.Staff {
	var workingHours *structpb.Value
	if staff.WorkingHours != nil {
		var err error
		workingHours, err = structpb.NewValue(staff.WorkingHours)
		if err != nil {
			// Log error but continue with nil
		}
	}

	return &staffv1.Staff{
		Id:              staff.ID,
		UserId:          wrapperspb.String(staff.UserID),
		Name:            staff.Name,
		Email:           staff.Email,
		Phone:           staff.Phone,
		Gender:          staff.Gender,
		Birthday:        wrapperspb.String(staff.Birthday),
		Avatar:          wrapperspb.String(staff.Avatar),
		Department:      staff.Department,
		Position:        staff.Position,
		RoleId:          staff.RoleID,
		Status:          staff.Status,
		HireDate:        wrapperspb.String(staff.HireDate),
		Salary:          wrapperspb.Double(staff.Salary),
		Address:         wrapperspb.String(staff.Address),
		EmergencyContact: wrapperspb.String(staff.EmergencyContact),
		Notes:           wrapperspb.String(staff.Notes),
		Skills:          staff.Skills,
		WorkingHours:    workingHours,
		IsAvailable:     staff.IsAvailable,
		CreatedAt:       timestamppb.New(staff.CreatedAt),
		UpdatedAt:       timestamppb.New(staff.UpdatedAt),
	}
}

func toProtoRole(role *entity.StaffRole) *staffv1.StaffRole {
	return &staffv1.StaffRole{
		Id:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: wrapperspb.String(role.Description),
		Permissions: role.Permissions,
		IsDefault:   role.IsDefault,
		SortOrder:   int32(role.SortOrder),
		Status:      role.Status,
		StaffCount:  int32(role.StaffCount),
		CreatedAt:   timestamppb.New(role.CreatedAt),
		UpdatedAt:   timestamppb.New(role.UpdatedAt),
	}
}

func toProtoAvailability(availability *entity.StaffAvailability) *staffv1.StaffAvailability {
	return &staffv1.StaffAvailability{
		Id:          availability.ID,
		StaffId:     availability.StaffID,
		DayOfWeek:   int32(availability.DayOfWeek),
		StartTime:   availability.StartTime,
		EndTime:     availability.EndTime,
		IsAvailable: availability.IsAvailable,
		Notes:       wrapperspb.String(availability.Notes),
		CreatedAt:   timestamppb.New(availability.CreatedAt),
		UpdatedAt:   timestamppb.New(availability.UpdatedAt),
	}
}