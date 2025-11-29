package grpc

import (
	"context"
	"encoding/json"
	"time"

	staffv1 "github.com/julesChu12/fly/staff/api/proto/staff/v1"
	"github.com/julesChu12/fly/staff/internal/application/service"
	"github.com/julesChu12/fly/staff/internal/domain/dto"
	"github.com/julesChu12/fly/staff/internal/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		Gender:          entity.Gender(req.Gender),
		Department:      req.Department,
		Position:        req.Position,
		RoleID:          req.RoleId,
		Status:          entity.StaffStatus(req.Status),
		Skills:          req.Skills,
		IsAvailable:     req.IsAvailable,
	}

	// Handle optional fields
	if req.Avatar != nil {
		createReq.Avatar = &req.Avatar.Value
	}
	if req.HireDate != nil {
		if t, err := time.Parse("2006-01-02", req.HireDate.Value); err == nil {
			createReq.HireDate = &t
		}
	}
	if req.Salary != nil {
		createReq.Salary = &req.Salary.Value
	}
	if req.Address != nil {
		createReq.Address = &req.Address.Value
	}
	if req.EmergencyContact != nil {
		createReq.EmergencyContact = &req.EmergencyContact.Value
	}
	if req.Notes != nil {
		createReq.Notes = &req.Notes.Value
	}
	if req.Birthday != nil {
		if t, err := time.Parse("2006-01-02", req.Birthday.Value); err == nil {
			createReq.Birthday = &t
		}
	}
	if req.WorkingHours != nil {
		// Convert structpb.Value to JSON and then to WorkingHours
		if workingHoursJSON, err := req.WorkingHours.MarshalJSON(); err == nil {
			var workingHours dto.WorkingHours
			if err := json.Unmarshal(workingHoursJSON, &workingHours); err == nil {
				createReq.WorkingHours = &workingHours
			}
		}
	}

	// Call service
	staff, err := s.staffService.CreateStaff(createReq)
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
	staff, err := s.staffService.GetStaffByID(req.Id)
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
		Skills:  req.Skills,
		Page:    int(req.Page),
		Limit:   int(req.PageSize),
		Sort:    req.Sort,
		Order:   req.Order,
	}

	// Handle optional fields
	if req.Search != nil {
		filter.Search = &req.Search.Value
	}
	if req.Department != nil {
		filter.Department = &req.Department.Value
	}
	if req.RoleId != nil {
		filter.RoleID = &req.RoleId.Value
	}
	if req.Status != nil {
		status := entity.StaffStatus(req.Status.Value)
		filter.Status = &status
	}
	if req.IsAvailable != nil {
		filter.IsAvailable = &req.IsAvailable.Value
	}
	if req.MinAge != nil {
		age := int(req.MinAge.Value)
		filter.MinAge = &age
	}
	if req.MaxAge != nil {
		age := int(req.MaxAge.Value)
		filter.MaxAge = &age
	}
	if req.Gender != nil {
		gender := entity.Gender(req.Gender.Value)
		filter.Gender = &gender
	}

	// Call service
	staffMembers, total, err := s.staffService.ListStaff(filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list staff: %v", err)
	}

	// Convert response
	staffList := make([]*staffv1.Staff, len(staffMembers))
	for i := range staffMembers {
		staffList[i] = toProtoStaff(staffMembers[i])
	}

	return &staffv1.ListStaffResponse{
		Staff:    staffList,
		Total:    total,
		Page:     int32(filter.Page),
		PageSize: int32(filter.Limit),
	}, nil
}

// UpdateStaff updates an existing staff member
func (s *StaffServer) UpdateStaff(ctx context.Context, req *staffv1.UpdateStaffRequest) (*staffv1.UpdateStaffResponse, error) {
	// Convert protobuf request to DTO
	updateReq := &dto.UpdateStaffRequest{}

	// Handle optional string fields
	if req.Name != nil {
		updateReq.Name = &req.Name.Value
	}
	if req.Email != nil {
		updateReq.Email = &req.Email.Value
	}
	if req.Phone != nil {
		updateReq.Phone = &req.Phone.Value
	}
	if req.Avatar != nil {
		updateReq.Avatar = &req.Avatar.Value
	}
	if req.Department != nil {
		updateReq.Department = &req.Department.Value
	}
	if req.Position != nil {
		updateReq.Position = &req.Position.Value
	}
	if req.RoleId != nil {
		updateReq.RoleID = &req.RoleId.Value
	}
	if req.Address != nil {
		updateReq.Address = &req.Address.Value
	}
	if req.EmergencyContact != nil {
		updateReq.EmergencyContact = &req.EmergencyContact.Value
	}
	if req.Notes != nil {
		updateReq.Notes = &req.Notes.Value
	}

	// Handle optional enum fields
	if req.Gender != nil {
		gender := entity.Gender(req.Gender.Value)
		updateReq.Gender = &gender
	}
	if req.Status != nil {
		status := entity.StaffStatus(req.Status.Value)
		updateReq.Status = &status
	}

	// Handle optional date fields
	if req.HireDate != nil {
		if t, err := time.Parse("2006-01-02", req.HireDate.Value); err == nil {
			updateReq.HireDate = &t
		}
	}
	if req.Birthday != nil {
		if t, err := time.Parse("2006-01-02", req.Birthday.Value); err == nil {
			updateReq.Birthday = &t
		}
	}

	// Handle optional numeric fields
	if req.Salary != nil {
		updateReq.Salary = &req.Salary.Value
	}

	// Handle optional boolean fields
	if req.IsAvailable != nil {
		updateReq.IsAvailable = &req.IsAvailable.Value
	}

	// Handle slice and complex fields
	if req.Skills != nil {
		updateReq.Skills = req.Skills
	}
	if req.WorkingHours != nil {
		// Convert structpb.Value to JSON and then to WorkingHours
		if workingHoursJSON, err := req.WorkingHours.MarshalJSON(); err == nil {
			var workingHours dto.WorkingHours
			if err := json.Unmarshal(workingHoursJSON, &workingHours); err == nil {
				updateReq.WorkingHours = &workingHours
			}
		}
	}

	// Call service
	staff, err := s.staffService.UpdateStaff(req.Id, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update staff: %v", err)
	}

	return &staffv1.UpdateStaffResponse{
		Staff: toProtoStaff(staff),
	}, nil
}

// DeleteStaff deletes a staff member
func (s *StaffServer) DeleteStaff(ctx context.Context, req *staffv1.DeleteStaffRequest) (*staffv1.DeleteStaffResponse, error) {
	err := s.staffService.DeleteStaff(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete staff: %v", err)
	}

	return &staffv1.DeleteStaffResponse{}, nil
}

// UpdateStaffStatus updates the status of a staff member
func (s *StaffServer) UpdateStaffStatus(ctx context.Context, req *staffv1.UpdateStaffStatusRequest) (*staffv1.UpdateStaffStatusResponse, error) {
	// Handle optional reason
	var reason *string
	if req.Reason != nil {
		reason = &req.Reason.Value
	}

	updateReq := &dto.UpdateStatusRequest{
		Status: entity.StaffStatus(req.Status),
		Reason: reason,
	}

	staff, err := s.staffService.UpdateStaffStatus(req.Id, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update staff status: %v", err)
	}

	return &staffv1.UpdateStaffStatusResponse{
		Staff: toProtoStaff(staff),
	}, nil
}

// GetAvailableStaff gets available staff members
func (s *StaffServer) GetAvailableStaff(ctx context.Context, req *staffv1.GetAvailableStaffRequest) (*staffv1.GetAvailableStaffResponse, error) {
	filter := &dto.StaffFilter{}

	// Handle optional fields
	if req.Department != nil {
		filter.Department = &req.Department.Value
	}
	if req.Skills != nil {
		filter.Skills = req.Skills
	}

	staffMembers, err := s.staffService.GetAvailableStaff(filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get available staff: %v", err)
	}

	protoStaff := make([]*staffv1.Staff, len(staffMembers))
	for i := range staffMembers {
		protoStaff[i] = toProtoStaff(staffMembers[i])
	}

	return &staffv1.GetAvailableStaffResponse{
		Staff: protoStaff,
	}, nil
}

// ListRoles retrieves a list of staff roles
func (s *StaffServer) ListRoles(ctx context.Context, req *staffv1.ListRolesRequest) (*staffv1.ListRolesResponse, error) {
	filter := &dto.RoleFilter{
		Sort:  req.Sort,
		Order: req.Order,
	}

	// Handle optional status field
	if req.Status != nil {
		status := entity.StaffStatus(req.Status.Value)
		filter.Status = &status
	}

	roles, err := s.staffService.ListRoles(filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list roles: %v", err)
	}

	protoRoles := make([]*staffv1.StaffRole, len(roles))
	for i := range roles {
		protoRoles[i] = toProtoRole(roles[i])
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
		Permissions: req.Permissions,
		IsDefault:   req.IsDefault,
		SortOrder:   int(req.SortOrder),
		Status:      entity.StaffStatus(req.Status),
	}

	// Handle optional description
	if req.Description != nil {
		createReq.Description = &req.Description.Value
	}

	role, err := s.staffService.CreateRole(createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create role: %v", err)
	}

	return &staffv1.CreateRoleResponse{
		Role: toProtoRole(role),
	}, nil
}

// GetRole retrieves a staff role by ID
func (s *StaffServer) GetRole(ctx context.Context, req *staffv1.GetRoleRequest) (*staffv1.GetRoleResponse, error) {
	role, err := s.staffService.GetRoleByID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Role not found: %v", err)
	}

	return &staffv1.GetRoleResponse{
		Role: toProtoRole(role),
	}, nil
}

// UpdateRole updates an existing staff role
func (s *StaffServer) UpdateRole(ctx context.Context, req *staffv1.UpdateRoleRequest) (*staffv1.UpdateRoleResponse, error) {
	updateReq := &dto.UpdateRoleRequest{}

	// Handle optional string fields
	if req.Name != nil {
		updateReq.Name = &req.Name.Value
	}
	if req.Code != nil {
		updateReq.Code = &req.Code.Value
	}
	if req.Description != nil {
		updateReq.Description = &req.Description.Value
	}

	// Handle other fields
	if req.Permissions != nil {
		updateReq.Permissions = req.Permissions
	}
	if req.IsDefault != nil {
		updateReq.IsDefault = &req.IsDefault.Value
	}
	if req.SortOrder != nil {
		sortOrder := int(req.SortOrder.Value)
		updateReq.SortOrder = &sortOrder
	}
	if req.Status != nil {
		status := entity.StaffStatus(req.Status.Value)
		updateReq.Status = &status
	}

	role, err := s.staffService.UpdateRole(req.Id, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update role: %v", err)
	}

	return &staffv1.UpdateRoleResponse{
		Role: toProtoRole(role),
	}, nil
}

// DeleteRole deletes a staff role
func (s *StaffServer) DeleteRole(ctx context.Context, req *staffv1.DeleteRoleRequest) (*staffv1.DeleteRoleResponse, error) {
	err := s.staffService.DeleteRole(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete role: %v", err)
	}

	return &staffv1.DeleteRoleResponse{}, nil
}

// GetAvailability gets staff availability
func (s *StaffServer) GetAvailability(ctx context.Context, req *staffv1.GetAvailabilityRequest) (*staffv1.GetAvailabilityResponse, error) {
	availabilities, err := s.staffService.GetAvailability(req.StaffId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get availability: %v", err)
	}

	protoAvailabilities := make([]*staffv1.StaffAvailability, len(availabilities.Availabilities))
	for i, avail := range availabilities.Availabilities {
		protoAvailabilities[i] = toProtoAvailability(&avail)
	}

	return &staffv1.GetAvailabilityResponse{
		Availabilities: protoAvailabilities,
	}, nil
}

// SetAvailability sets staff availability
func (s *StaffServer) SetAvailability(ctx context.Context, req *staffv1.SetAvailabilityRequest) (*staffv1.SetAvailabilityResponse, error) {
	// Convert protobuf request to DTO
	availabilities := make([]dto.AvailabilityItem, len(req.Availabilities))
	for i, item := range req.Availabilities {
		availabilities[i] = dto.AvailabilityItem{
			DayOfWeek:   int(item.DayOfWeek),
			StartTime:   item.StartTime,
			EndTime:     item.EndTime,
			IsAvailable: item.IsAvailable,
		}

		// Handle optional notes
		if item.Notes != nil {
			availabilities[i].Notes = &item.Notes.Value
		}
	}

	setReq := &dto.AvailabilityRequest{
		StaffID:       req.StaffId,
		Availabilities: availabilities,
	}

	err := s.staffService.SetAvailability(setReq.StaffID, setReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to set availability: %v", err)
	}

	// Return success response
	return &staffv1.SetAvailabilityResponse{
		Availabilities: []*staffv1.StaffAvailability{},
	}, nil
}

// GetStaffAvailability gets staff availability for a specific day
func (s *StaffServer) GetStaffAvailability(ctx context.Context, req *staffv1.GetStaffAvailabilityRequest) (*staffv1.GetStaffAvailabilityResponse, error) {
	// This method is not implemented in the service yet
	return &staffv1.GetStaffAvailabilityResponse{
		Availabilities: []*staffv1.StaffAvailability{},
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

	// Handle optional fields
	var avatar, hireDate, address, emergencyContact, notes *wrapperspb.StringValue
	if staff.Avatar != nil {
		avatar = wrapperspb.String(*staff.Avatar)
	}
	if staff.HireDate != nil {
		hireDate = wrapperspb.String(staff.HireDate.Format("2006-01-02"))
	}
	if staff.Address != nil {
		address = wrapperspb.String(*staff.Address)
	}
	if staff.EmergencyContact != nil {
		emergencyContact = wrapperspb.String(*staff.EmergencyContact)
	}
	if staff.Notes != nil {
		notes = wrapperspb.String(*staff.Notes)
	}

	var salary *wrapperspb.DoubleValue
	if staff.Salary != nil {
		salary = wrapperspb.Double(*staff.Salary)
	}

	// Handle UserID if present
	var userID *wrapperspb.StringValue
	if staff.UserID != nil {
		userID = wrapperspb.String(staff.UserID.String())
	}

	// Handle optional Birthday
	var birthday *wrapperspb.StringValue
	if staff.Birthday != nil {
		birthday = wrapperspb.String(staff.Birthday.Format("2006-01-02"))
	}

	// Parse Skills as JSON array
	var skills []string
	if staff.Skills != "" {
		if err := json.Unmarshal([]byte(staff.Skills), &skills); err != nil {
			// If parsing fails, treat as single skill
			skills = []string{staff.Skills}
		}
	}

	return &staffv1.Staff{
		Id:       staff.ID.String(),
		UserId:   userID,
		Name:     staff.Name,
		Email:    staff.Email,
		Phone:    staff.Phone,
		Gender:   string(staff.Gender),
		Birthday: birthday,
		Avatar:   avatar,
		Department: staff.Department,
		Position: staff.Position,
		RoleId:   staff.RoleID.String(),
		Status:   string(staff.Status),
		HireDate: hireDate,
		Salary:   salary,
		Address:  address,
		EmergencyContact: emergencyContact,
		Notes:    notes,
		Skills:   skills,
		WorkingHours: workingHours,
		IsAvailable: staff.IsAvailable,
		CreatedAt: timestamppb.New(staff.CreatedAt),
		UpdatedAt: timestamppb.New(staff.UpdatedAt),
	}
}

func toProtoRole(role *entity.StaffRole) *staffv1.StaffRole {
	// Handle optional description
	var description *wrapperspb.StringValue
	if role.Description != nil {
		description = wrapperspb.String(*role.Description)
	}

	// Parse Permissions as JSON array
	var permissions []string
	if role.Permissions != "" {
		if err := json.Unmarshal([]byte(role.Permissions), &permissions); err != nil {
			// If parsing fails, treat as single permission
			permissions = []string{role.Permissions}
		}
	}

	return &staffv1.StaffRole{
		Id:          role.ID.String(),
		Name:        role.Name,
		Code:        role.Code,
		Description: description,
		Permissions: permissions,
		IsDefault:   role.IsDefault,
		SortOrder:   int32(role.SortOrder),
		Status:      string(role.Status),
		StaffCount:  0, // This field is not in the entity, set to 0
		CreatedAt:   timestamppb.New(role.CreatedAt),
		UpdatedAt:   timestamppb.New(role.UpdatedAt),
	}
}

func toProtoAvailability(availability *entity.StaffAvailability) *staffv1.StaffAvailability {
	// Handle optional notes
	var notes *wrapperspb.StringValue
	if availability.Notes != nil {
		notes = wrapperspb.String(*availability.Notes)
	}

	return &staffv1.StaffAvailability{
		Id:          availability.ID.String(),
		StaffId:     availability.StaffID.String(),
		DayOfWeek:   int32(availability.DayOfWeek),
		StartTime:   availability.StartTime,
		EndTime:     availability.EndTime,
		IsAvailable: availability.IsAvailable,
		Notes:       notes,
		CreatedAt:   timestamppb.New(availability.CreatedAt),
		UpdatedAt:   timestamppb.New(availability.UpdatedAt),
	}
}