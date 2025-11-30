package handlers

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	itemsv1 "github.com/julesChu12/fly/items/api/proto/items/v1"
	"github.com/julesChu12/fly/items/internal/domain/category"
	"github.com/julesChu12/fly/items/internal/interface/grpc/converters"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// CategoryService 定义分类服务接口
type CategoryService interface {
	CreateCategory(ctx context.Context, req *category.CreateCategoryRequest) (*category.Category, error)
	UpdateCategory(ctx context.Context, id uuid.UUID, req *category.UpdateCategoryRequest) (*category.Category, error)
	GetCategory(ctx context.Context, id uuid.UUID) (*category.Category, error)
	ListCategories(ctx context.Context, req *category.ListCategoriesRequest) ([]*category.Category, error)
	GetCategoryTree(ctx context.Context, req *category.GetCategoryTreeRequest) ([]*category.CategoryTree, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	UpdateCategoryStatus(ctx context.Context, id uuid.UUID, status category.CategoryStatus) error
	GetCategoryPath(ctx context.Context, id uuid.UUID) ([]*category.CategoryPath, error)
	MoveCategory(ctx context.Context, id, newParentID uuid.UUID) error
}

// CategoryServer 实现 gRPC CategoryService
type CategoryServer struct {
	itemsv1.UnimplementedCategoryServiceServer
	categoryService CategoryService
	logger          *logger.Logger
}

// NewCategoryServer 创建新的 Category gRPC 服务器
func NewCategoryServer(categoryService CategoryService, logger *logger.Logger) *CategoryServer {
	return &CategoryServer{
		categoryService: categoryService,
		logger:          logger,
	}
}

// CreateCategory 创建分类
func (s *CategoryServer) CreateCategory(ctx context.Context, req *itemsv1.CreateCategoryRequest) (*itemsv1.CreateCategoryResponse, error) {
	s.logger.Infof("CreateCategory request received: %s", req.Name)

	// 转换请求
	createReq, err := converters.ProtoToCreateCategoryRequest(req)
	if err != nil {
		s.logger.Errorf("Failed to convert CreateCategoryRequest: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	// 调用服务
	domainCategory, err := s.categoryService.CreateCategory(ctx, createReq)
	if err != nil {
		s.logger.Errorf("Failed to create category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create category: %v", err)
	}

	// 转换响应
	protoCategory := converters.DomainCategoryToProto(domainCategory)

	s.logger.Infof("Category created successfully: %s", protoCategory.Id)
	return &itemsv1.CreateCategoryResponse{
		Category: protoCategory,
	}, nil
}

// GetCategory 获取分类详情
func (s *CategoryServer) GetCategory(ctx context.Context, req *itemsv1.GetCategoryRequest) (*itemsv1.GetCategoryResponse, error) {
	s.logger.Infof("GetCategory request received: %s", req.Id)

	// 解析ID
	categoryID, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid category ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid category ID: %v", err)
	}

	// 调用服务
	domainCategory, err := s.categoryService.GetCategory(ctx, categoryID)
	if err != nil {
		s.logger.Errorf("Failed to get category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get category: %v", err)
	}

	// 转换响应
	protoCategory := converters.DomainCategoryToProto(domainCategory)

	// 如果请求包含子分类，获取子分类列表
	var protoChildren []*itemsv1.Category
	if req.IncludeChildren {
		// 这里可以调用 ListCategories 来获取子分类
		// 但为了简化，暂时不实现
		protoChildren = []*itemsv1.Category{}
	}

	s.logger.Infof("Category retrieved successfully: %s", protoCategory.Id)
	return &itemsv1.GetCategoryResponse{
		Category: protoCategory,
		Children: protoChildren,
	}, nil
}

// ListCategories 获取分类列表
func (s *CategoryServer) ListCategories(ctx context.Context, req *itemsv1.ListCategoriesRequest) (*itemsv1.ListCategoriesResponse, error) {
	s.logger.Infof("ListCategories request received: page=%d, page_size=%d", req.Page, req.PageSize)

	// 解析父级ID（如果提供）
	var parentID *uuid.UUID
	if req.ParentId != "" {
		id, err := uuid.Parse(req.ParentId)
		if err != nil {
			s.logger.Errorf("Invalid parent ID: %v", err)
			return nil, status.Errorf(codes.InvalidArgument, "invalid parent ID: %v", err)
		}
		parentID = &id
	}

	// 构建请求
	categoryStatus := converters.CategoryStatusFromProto(req.Status)
	search := req.Search

	listReq := &category.ListCategoriesRequest{
		ParentID: parentID,
		Status:   &categoryStatus,
		Search:   &search,
	}

	// 调用服务
	categories, err := s.categoryService.ListCategories(ctx, listReq)
	if err != nil {
		s.logger.Errorf("Failed to list categories: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list categories: %v", err)
	}

	// 转换响应
	protoCategories := converters.BatchDomainCategoriesToProto(categories)

	s.logger.Infof("Categories listed successfully: total=%d, returned=%d", len(categories), len(protoCategories))
	return &itemsv1.ListCategoriesResponse{
		Categories: protoCategories,
		Total:      int32(len(categories)),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: 1, // 由于服务不返回分页信息，简化处理
	}, nil
}

// GetCategoryTree 获取分类树
func (s *CategoryServer) GetCategoryTree(ctx context.Context, req *itemsv1.GetCategoryTreeRequest) (*itemsv1.GetCategoryTreeResponse, error) {
	s.logger.Infof("GetCategoryTree request received")

	// 构建请求
	categoryStatus := converters.CategoryStatusFromProto(req.Status)
	search := "" // 空搜索

	treeReq := &category.GetCategoryTreeRequest{
		Status: &categoryStatus,
		Search: &search,
	}

	// 调用服务
	categoryTrees, err := s.categoryService.GetCategoryTree(ctx, treeReq)
	if err != nil {
		s.logger.Errorf("Failed to get category tree: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get category tree: %v", err)
	}

	// 手动转换 CategoryTree 到 protobuf CategoryTree
	protoCategories := make([]*itemsv1.CategoryTree, len(categoryTrees))
	for i, tree := range categoryTrees {
		protoCategories[i] = s.convertCategoryTreeToProto(tree)
	}

	s.logger.Infof("Category tree retrieved successfully: total=%d", len(protoCategories))
	return &itemsv1.GetCategoryTreeResponse{
		Categories: protoCategories,
	}, nil
}

// UpdateCategory 更新分类
func (s *CategoryServer) UpdateCategory(ctx context.Context, req *itemsv1.UpdateCategoryRequest) (*itemsv1.UpdateCategoryResponse, error) {
	s.logger.Infof("UpdateCategory request received: %s", req.Id)

	// 解析ID
	categoryID, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid category ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid category ID: %v", err)
	}

	// 转换请求
	updateReq, err := converters.ProtoToUpdateCategoryRequest(req, req.Id)
	if err != nil {
		s.logger.Errorf("Failed to convert UpdateCategoryRequest: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	// 调用服务
	domainCategory, err := s.categoryService.UpdateCategory(ctx, categoryID, updateReq)
	if err != nil {
		s.logger.Errorf("Failed to update category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update category: %v", err)
	}

	// 转换响应
	protoCategory := converters.DomainCategoryToProto(domainCategory)

	s.logger.Infof("Category updated successfully: %s", protoCategory.Id)
	return &itemsv1.UpdateCategoryResponse{
		Category: protoCategory,
	}, nil
}

// DeleteCategory 删除分类
func (s *CategoryServer) DeleteCategory(ctx context.Context, req *itemsv1.DeleteCategoryRequest) (*emptypb.Empty, error) {
	s.logger.Infof("DeleteCategory request received: %s", req.Id)

	// 解析ID
	categoryID, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid category ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid category ID: %v", err)
	}

	// 调用服务
	err = s.categoryService.DeleteCategory(ctx, categoryID)
	if err != nil {
		s.logger.Errorf("Failed to delete category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete category: %v", err)
	}

	s.logger.Infof("Category deleted successfully: %s", req.Id)
	return &emptypb.Empty{}, nil
}

// ActivateCategory 激活分类
func (s *CategoryServer) ActivateCategory(ctx context.Context, req *itemsv1.ActivateCategoryRequest) (*itemsv1.ActivateCategoryResponse, error) {
	s.logger.Infof("ActivateCategory request received: %s", req.Id)

	// 解析ID
	categoryID, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid category ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid category ID: %v", err)
	}

	// 调用服务
	err = s.categoryService.UpdateCategoryStatus(ctx, categoryID, category.CategoryStatusActive)
	if err != nil {
		s.logger.Errorf("Failed to activate category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to activate category: %v", err)
	}

	// 获取更新后的分类
	domainCategory, err := s.categoryService.GetCategory(ctx, categoryID)
	if err != nil {
		s.logger.Errorf("Failed to get updated category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get updated category: %v", err)
	}

	// 转换响应
	protoCategory := converters.DomainCategoryToProto(domainCategory)

	s.logger.Infof("Category activated successfully: %s", protoCategory.Id)
	return &itemsv1.ActivateCategoryResponse{
		Category: protoCategory,
	}, nil
}

// DeactivateCategory 停用分类
func (s *CategoryServer) DeactivateCategory(ctx context.Context, req *itemsv1.DeactivateCategoryRequest) (*itemsv1.DeactivateCategoryResponse, error) {
	s.logger.Infof("DeactivateCategory request received: %s", req.Id)

	// 解析ID
	categoryID, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid category ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid category ID: %v", err)
	}

	// 调用服务
	err = s.categoryService.UpdateCategoryStatus(ctx, categoryID, category.CategoryStatusInactive)
	if err != nil {
		s.logger.Errorf("Failed to deactivate category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to deactivate category: %v", err)
	}

	// 获取更新后的分类
	domainCategory, err := s.categoryService.GetCategory(ctx, categoryID)
	if err != nil {
		s.logger.Errorf("Failed to get updated category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get updated category: %v", err)
	}

	// 转换响应
	protoCategory := converters.DomainCategoryToProto(domainCategory)

	s.logger.Infof("Category deactivated successfully: %s", protoCategory.Id)
	return &itemsv1.DeactivateCategoryResponse{
		Category: protoCategory,
	}, nil
}

// MoveCategory 移动分类
func (s *CategoryServer) MoveCategory(ctx context.Context, req *itemsv1.MoveCategoryRequest) (*itemsv1.MoveCategoryResponse, error) {
	s.logger.Infof("MoveCategory request received: %s -> %s", req.Id, req.NewParentId)

	// 解析ID
	categoryID, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid category ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid category ID: %v", err)
	}

	// 解析新的父级ID
	var newParentID *uuid.UUID
	if req.NewParentId != "" {
		id, err := uuid.Parse(req.NewParentId)
		if err != nil {
			s.logger.Errorf("Invalid new parent ID: %v", err)
			return nil, status.Errorf(codes.InvalidArgument, "invalid new parent ID: %v", err)
		}
		newParentID = &id
	}

	// 调用服务
	err = s.categoryService.MoveCategory(ctx, categoryID, *newParentID)
	if err != nil {
		s.logger.Errorf("Failed to move category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to move category: %v", err)
	}

	// 获取更新后的分类
	domainCategory, err := s.categoryService.GetCategory(ctx, categoryID)
	if err != nil {
		s.logger.Errorf("Failed to get updated category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get updated category: %v", err)
	}

	// 转换响应
	protoCategory := converters.DomainCategoryToProto(domainCategory)

	s.logger.Infof("Category moved successfully: %s", protoCategory.Id)
	return &itemsv1.MoveCategoryResponse{
		Category:             protoCategory,
		AffectedCategories:   []*itemsv1.Category{}, // 简化实现，不返回影响的分类列表
	}, nil
}

// GetCategoryStats 获取分类统计信息
func (s *CategoryServer) GetCategoryStats(ctx context.Context, req *itemsv1.GetCategoryStatsRequest) (*itemsv1.GetCategoryStatsResponse, error) {
	s.logger.Infof("GetCategoryStats request received: %s", req.Id)

	// 解析ID
	_, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid category ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid category ID: %v", err)
	}

	// 注意：域服务中没有 GetCategoryStats 方法，这里返回一个基本的响应
	// 在实际实现中，需要添加相应的统计功能或调用其他服务

	response := &itemsv1.GetCategoryStatsResponse{
		CategoryId:     req.Id,
		CategoryName:   "Category", // 需要从域服务获取
		TotalItems:     0,
		ActiveItems:    0,
		InactiveItems:  0,
		TotalServices:  0,
		TotalProducts:  0,
		TotalValue:     0,
		ChildStats:     []*itemsv1.DetailedCategoryStats{},
	}

	s.logger.Infof("Category stats retrieved successfully: %s", req.Id)
	return response, nil
}

// convertCategoryTreeToProto 将 domain CategoryTree 转换为 protobuf CategoryTree
func (s *CategoryServer) convertCategoryTreeToProto(tree *category.CategoryTree) *itemsv1.CategoryTree {
	if tree == nil {
		return nil
	}

	proto := &itemsv1.CategoryTree{
		Id:          tree.ID.String(),
		Name:        tree.Name,
		Description: tree.Description,
		SortOrder:   int32(tree.SortOrder),
		Status:      converters.CategoryStatusToProto(tree.Status),
		ItemCount:   int32(tree.ItemCount),
		Level:       int32(tree.Level),
		Path:        tree.Path,
		CreatedAt:   timestamppb.New(tree.CreatedAt),
		UpdatedAt:   timestamppb.New(tree.UpdatedAt),
	}

	// 处理父级ID
	if tree.ParentID != nil {
		proto.ParentId = tree.ParentID.String()
	}

	// 处理图标
	if tree.Icon != nil {
		proto.Icon = *tree.Icon
	}

	// 递归转换子分类
	if len(tree.Children) > 0 {
		proto.Children = make([]*itemsv1.CategoryTree, len(tree.Children))
		for i, child := range tree.Children {
			proto.Children[i] = s.convertCategoryTreeToProto(child)
		}
	}

	return proto
}