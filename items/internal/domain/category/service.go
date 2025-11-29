package category

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service 分类领域服务
type Service struct {
	repo Repository
}

// NewService 创建分类服务
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateCategory 创建分类
func (s *Service) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*Category, error) {
	// 验证请求
	if err := s.validateCreateCategoryRequest(req); err != nil {
		return nil, err
	}

	// 检查名称唯一性
	existing, err := s.repo.GetByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, ErrDuplicateCategory
	}

	// 验证父分类
	if req.ParentID != nil {
		if err := s.validateParent(ctx, *req.ParentID); err != nil {
			return nil, err
		}
	}

	// 创建分类
	category := NewCategory(req.Name, req.Description, req.ParentID)
	category.Icon = req.Icon
	category.SortOrder = req.SortOrder

	// 保存到仓储
	if err := s.repo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("create category failed: %w", err)
	}

	return category, nil
}

// UpdateCategory 更新分类
func (s *Service) UpdateCategory(ctx context.Context, id uuid.UUID, req *UpdateCategoryRequest) (*Category, error) {
	// 获取现有分类
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get category failed: %w", err)
	}

	// 检查名称唯一性
	if req.Name != nil && *req.Name != category.Name {
		existing, err := s.repo.GetByName(ctx, *req.Name)
		if err == nil && existing != nil {
			return nil, ErrDuplicateCategory
		}
	}

	// 验证父分类变更
	if req.ParentID != nil && !s.equalParentID(req.ParentID, category.ParentID) {
		if err := s.validateParentChange(ctx, id, *req.ParentID); err != nil {
			return nil, err
		}
	}

	// 更新字段
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.Icon != nil {
		category.Icon = req.Icon
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}

	// 保存更新
	if err := s.repo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("update category failed: %w", err)
	}

	return category, nil
}

// GetCategory 获取分类
func (s *Service) GetCategory(ctx context.Context, id uuid.UUID) (*Category, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get category failed: %w", err)
	}
	return category, nil
}

// ListCategories 获取分类列表
func (s *Service) ListCategories(ctx context.Context, req *ListCategoriesRequest) ([]*Category, error) {
	// 构建过滤器
	filter := &Filter{
		ParentID: req.ParentID,
		Status:   req.Status,
		Search:   req.Search,
	}

	// 查询分类
	categories, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list categories failed: %w", err)
	}

	return categories, nil
}

// GetCategoryTree 获取分类树
func (s *Service) GetCategoryTree(ctx context.Context, req *GetCategoryTreeRequest) ([]*CategoryTree, error) {
	// 构建过滤器
	filter := &Filter{
		Status: req.Status,
		Search: req.Search,
	}

	// 获取树结构
	trees, err := s.repo.GetTree(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get category tree failed: %w", err)
	}

	return trees, nil
}

// DeleteCategory 删除分类
func (s *Service) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	// 获取分类
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get category failed: %w", err)
	}

	// 检查是否可以删除
	if !category.CanDelete() {
		if category.ItemCount > 0 {
			return ErrHasItems
		}
		if hasChildren, _ := s.repo.HasChildren(ctx, id); hasChildren {
			return ErrHasChildren
		}
		return ErrCategoryInUse
	}

	// 检查是否为根分类
	if category.IsRoot() {
		return ErrCannotDeleteRoot
	}

	// 删除分类
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete category failed: %w", err)
	}

	return nil
}

// UpdateCategoryStatus 更新分类状态
func (s *Service) UpdateCategoryStatus(ctx context.Context, id uuid.UUID, status CategoryStatus) error {
	// 获取分类
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get category failed: %w", err)
	}

	// 更新状态
	switch status {
	case CategoryStatusActive:
		category.Activate()
	case CategoryStatusInactive:
		category.Deactivate()
	}

	// 保存更新
	if err := s.repo.Update(ctx, category); err != nil {
		return fmt.Errorf("update category status failed: %w", err)
	}

	return nil
}

// GetCategoryPath 获取分类路径
func (s *Service) GetCategoryPath(ctx context.Context, id uuid.UUID) ([]*CategoryPath, error) {
	// 获取路径
	paths, err := s.repo.GetPath(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get category path failed: %w", err)
	}

	return paths, nil
}

// MoveCategory 移动分类到新的父分类下
func (s *Service) MoveCategory(ctx context.Context, id, newParentID uuid.UUID) error {
	// 获取分类
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get category failed: %w", err)
	}

	// 验证新的父分类
	if err := s.validateParentChange(ctx, id, newParentID); err != nil {
		return err
	}

	// 更新父分类
	category.ParentID = &newParentID
	category.UpdatedAt = category.UpdatedAt // 这里使用当前时间

	// 保存更新
	if err := s.repo.Update(ctx, category); err != nil {
		return fmt.Errorf("move category failed: %w", err)
	}

	return nil
}

// BatchUpdateItemCounts 批量更新商品数量
func (s *Service) BatchUpdateItemCounts(ctx context.Context, updates map[uuid.UUID]int) error {
	for categoryID, count := range updates {
		if err := s.repo.UpdateItemCount(ctx, categoryID, count); err != nil {
			return fmt.Errorf("update item count for category %s failed: %w", categoryID, err)
		}
	}
	return nil
}

// 验证方法
func (s *Service) validateCreateCategoryRequest(req *CreateCategoryRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.SortOrder < 0 {
		return fmt.Errorf("sort_order must be non-negative")
	}
	return nil
}

func (s *Service) validateParent(ctx context.Context, parentID uuid.UUID) error {
	parent, err := s.repo.GetByID(ctx, parentID)
	if err != nil {
		return ErrInvalidParent
	}
	if !parent.IsActive() {
		return fmt.Errorf("parent category is not active")
	}
	return nil
}

func (s *Service) validateParentChange(ctx context.Context, categoryID, newParentID uuid.UUID) error {
	// 不能将分类移动到自己的子分类下
	descendantIDs, err := s.repo.GetDescendantIDs(ctx, categoryID)
	if err != nil {
		return err
	}
	for _, id := range descendantIDs {
		if id == newParentID {
			return ErrCircularReference
		}
	}

	// 验证新的父分类
	if err := s.validateParent(ctx, newParentID); err != nil {
		return err
	}

	return nil
}

func (s *Service) equalParentID(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a != nil && b != nil {
		return *a == *b
	}
	return false
}

// 请求和响应结构体

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Icon        *string    `json:"icon,omitempty"`
	SortOrder   int        `json:"sort_order"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Icon        *string    `json:"icon,omitempty"`
	SortOrder   *int       `json:"sort_order,omitempty"`
}

// ListCategoriesRequest 获取分类列表请求
type ListCategoriesRequest struct {
	ParentID *uuid.UUID      `json:"parent_id,omitempty"`
	Status   *CategoryStatus `json:"status,omitempty"`
	Search   *string         `json:"search,omitempty"`
}

// GetCategoryTreeRequest 获取分类树请求
type GetCategoryTreeRequest struct {
	Status *CategoryStatus `json:"status,omitempty"`
	Search *string         `json:"search,omitempty"`
}