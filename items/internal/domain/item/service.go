package item

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service 商品领域服务
type Service struct {
	repo Repository
}

// NewService 创建商品服务
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateItem 创建商品
func (s *Service) CreateItem(ctx context.Context, req *CreateItemRequest) (*Item, error) {
	// 验证请求
	if err := s.validateCreateItemRequest(req); err != nil {
		return nil, err
	}

	// 检查 SKU 唯一性（如果提供）
	if req.Type == ItemTypeProduct && req.SKU != nil && *req.SKU != "" {
		existing, err := s.repo.GetBySKU(ctx, *req.SKU)
		if err == nil && existing != nil {
			return nil, ErrDuplicateSKU
		}
	}

	// 创建商品
	item := NewItem(req.Name, req.Description, req.Type, req.Price, req.CategoryID)
	item.ImageURL = req.ImageURL
	item.Tags = req.Tags

	// 设置特有字段
	if req.Type == ItemTypeService {
		item.UpdateServiceFields(req.Duration, req.StaffRequired, req.Capacity)
	} else if req.Type == ItemTypeProduct {
		item.UpdateProductFields(req.Stock, req.CostPrice, req.Weight, req.SKU)
	}

	// 保存到仓储
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("create item failed: %w", err)
	}

	return item, nil
}

// UpdateItem 更新商品
func (s *Service) UpdateItem(ctx context.Context, id uuid.UUID, req *UpdateItemRequest) (*Item, error) {
	// 获取现有商品
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get item failed: %w", err)
	}

	// 更新基础信息
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.Price != nil {
		if *req.Price < 0 {
			return nil, ErrInvalidPrice
		}
		item.Price = *req.Price
	}
	if req.ImageURL != nil {
		item.ImageURL = req.ImageURL
	}
	if req.Tags != nil {
		item.Tags = req.Tags
	}

	// 更新特有字段
	if item.IsService() {
		item.UpdateServiceFields(req.Duration, req.StaffRequired, req.Capacity)
	} else if item.IsProduct() {
		item.UpdateProductFields(req.Stock, req.CostPrice, req.Weight, req.SKU)

		// 检查 SKU 唯一性
		if req.SKU != nil && *req.SKU != "" {
			existing, err := s.repo.GetBySKU(ctx, *req.SKU)
			if err == nil && existing != nil && existing.ID != id {
				return nil, ErrDuplicateSKU
			}
		}
	}

	// 保存更新
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("update item failed: %w", err)
	}

	return item, nil
}

// GetItem 获取商品
func (s *Service) GetItem(ctx context.Context, id uuid.UUID) (*Item, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get item failed: %w", err)
	}
	return item, nil
}

// ListItems 获取商品列表
func (s *Service) ListItems(ctx context.Context, req *ListItemsRequest) (*ListItemsResponse, error) {
	// 验证分页参数
	req.Pagination.Validate()

	// 构建过滤器
	filter := &Filter{
		Type:       req.Type,
		Status:     req.Status,
		CategoryID: req.CategoryID,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		Search:     req.Search,
	}

	// 查询商品
	items, total, err := s.repo.List(ctx, filter, req.Pagination)
	if err != nil {
		return nil, fmt.Errorf("list items failed: %w", err)
	}

	return &ListItemsResponse{
		Items: items,
		Total: total,
		Page:  req.Pagination.Page,
		Size:  req.Pagination.PageSize,
	}, nil
}

// SearchItems 搜索商品
func (s *Service) SearchItems(ctx context.Context, req *SearchItemsRequest) ([]*Item, error) {
	// 构建搜索查询
	query := &SearchQuery{
		Query:      req.Query,
		Type:       req.Type,
		CategoryID: req.CategoryID,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		Limit:      req.Limit,
	}

	if query.Limit <= 0 {
		query.Limit = 20
	}

	items, err := s.repo.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search items failed: %w", err)
	}

	return items, nil
}

// DeleteItem 删除商品
func (s *Service) DeleteItem(ctx context.Context, id uuid.UUID) error {
	// 获取商品
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get item failed: %w", err)
	}

	// 检查是否可以删除
	if !item.CanDelete() {
		return ErrCannotDeleteActive
	}

	// 删除商品
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete item failed: %w", err)
	}

	return nil
}

// UpdateItemStatus 更新商品状态
func (s *Service) UpdateItemStatus(ctx context.Context, id uuid.UUID, status ItemStatus) error {
	// 获取商品
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get item failed: %w", err)
	}

	// 更新状态
	switch status {
	case StatusActive:
		item.Activate()
	case StatusInactive:
		item.Deactivate()
	case StatusArchived:
		item.Archive()
	default:
		return ErrInvalidItemStatus
	}

	// 保存更新
	if err := s.repo.Update(ctx, item); err != nil {
		return fmt.Errorf("update item status failed: %w", err)
	}

	return nil
}

// UpdateStock 更新库存
func (s *Service) UpdateStock(ctx context.Context, id uuid.UUID, stock int) error {
	// 获取商品
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get item failed: %w", err)
	}

	// 检查是否为产品
	if !item.IsProduct() {
		return ErrNotProductOrNoStock
	}

	// 更新库存
	item.UpdateStock(stock)

	// 保存更新
	if err := s.repo.Update(ctx, item); err != nil {
		return fmt.Errorf("update stock failed: %w", err)
	}

	return nil
}

// DecreaseStock 减少库存
func (s *Service) DecreaseStock(ctx context.Context, id uuid.UUID, quantity int) error {
	// 获取商品
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get item failed: %w", err)
	}

	// 减少库存
	if err := item.DecreaseStock(quantity); err != nil {
		return err
	}

	// 保存更新
	if err := s.repo.Update(ctx, item); err != nil {
		return fmt.Errorf("decrease stock failed: %w", err)
	}

	return nil
}

// IncreaseStock 增加库存
func (s *Service) IncreaseStock(ctx context.Context, id uuid.UUID, quantity int) error {
	// 获取商品
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get item failed: %w", err)
	}

	// 增加库存
	if err := item.IncreaseStock(quantity); err != nil {
		return err
	}

	// 保存更新
	if err := s.repo.Update(ctx, item); err != nil {
		return fmt.Errorf("increase stock failed: %w", err)
	}

	return nil
}

// GetStatistics 获取统计信息
func (s *Service) GetStatistics(ctx context.Context, req *GetStatisticsRequest) (*Statistics, error) {
	filter := &StatisticsFilter{
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Type:       req.Type,
		CategoryID: req.CategoryID,
	}

	stats, err := s.repo.GetStatistics(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get statistics failed: %w", err)
	}

	return stats, nil
}

// 验证方法
func (s *Service) validateCreateItemRequest(req *CreateItemRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Type != ItemTypeService && req.Type != ItemTypeProduct {
		return ErrInvalidItemType
	}
	if req.Price < 0 {
		return ErrInvalidPrice
	}

	// 验证服务特有字段
	if req.Type == ItemTypeService {
		if req.Duration != nil && *req.Duration <= 0 {
			return ErrInvalidDuration
		}
		if req.Capacity != nil && *req.Capacity <= 0 {
			return fmt.Errorf("capacity must be positive")
		}
	}

	// 验证产品特有字段
	if req.Type == ItemTypeProduct {
		if req.Stock != nil && *req.Stock < 0 {
			return ErrInvalidStock
		}
		if req.CostPrice != nil && *req.CostPrice < 0 {
			return fmt.Errorf("cost price must be positive")
		}
		if req.Weight != nil && *req.Weight <= 0 {
			return fmt.Errorf("weight must be positive")
		}
	}

	return nil
}

// 请求和响应结构体

// CreateItemRequest 创建商品请求
type CreateItemRequest struct {
	Name          string     `json:"name" binding:"required"`
	Description   string     `json:"description"`
	Type          ItemType  `json:"type" binding:"required"`
	Price         float64    `json:"price" binding:"required,min=0"`
	CategoryID    uuid.UUID  `json:"category_id" binding:"required"`
	ImageURL      *string    `json:"image_url,omitempty"`
	Tags          *string    `json:"tags,omitempty"`

	// 服务字段
	Duration      *int  `json:"duration,omitempty"`
	StaffRequired *bool `json:"staff_required,omitempty"`
	Capacity      *int  `json:"capacity,omitempty"`

	// 产品字段
	Stock     *int     `json:"stock,omitempty"`
	CostPrice *float64 `json:"cost_price,omitempty"`
	Weight    *float64 `json:"weight,omitempty"`
	SKU       *string  `json:"sku,omitempty"`
}

// UpdateItemRequest 更新商品请求
type UpdateItemRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	ImageURL    *string  `json:"image_url,omitempty"`
	Tags        *string  `json:"tags,omitempty"`

	// 服务字段
	Duration      *int  `json:"duration,omitempty"`
	StaffRequired *bool `json:"staff_required,omitempty"`
	Capacity      *int  `json:"capacity,omitempty"`

	// 产品字段
	Stock     *int     `json:"stock,omitempty"`
	CostPrice *float64 `json:"cost_price,omitempty"`
	Weight    *float64 `json:"weight,omitempty"`
	SKU       *string  `json:"sku,omitempty"`
}

// ListItemsRequest 获取商品列表请求
type ListItemsRequest struct {
	Type       *ItemType  `json:"type,omitempty"`
	Status     *ItemStatus `json:"status,omitempty"`
	CategoryID *uuid.UUID  `json:"category_id,omitempty"`
	MinPrice   *float64    `json:"min_price,omitempty"`
	MaxPrice   *float64    `json:"max_price,omitempty"`
	Search     *string     `json:"search,omitempty"`
	Pagination *Pagination `json:"pagination"`
}

// ListItemsResponse 获取商品列表响应
type ListItemsResponse struct {
	Items []*Item `json:"items"`
	Total int64   `json:"total"`
	Page  int     `json:"page"`
	Size  int     `json:"size"`
}

// SearchItemsRequest 搜索商品请求
type SearchItemsRequest struct {
	Query      string     `json:"query" binding:"required"`
	Type       *ItemType  `json:"type,omitempty"`
	CategoryID *uuid.UUID `json:"category_id,omitempty"`
	MinPrice   *float64   `json:"min_price,omitempty"`
	MaxPrice   *float64   `json:"max_price,omitempty"`
	Limit      int        `json:"limit,omitempty"`
}

// GetStatisticsRequest 获取统计信息请求
type GetStatisticsRequest struct {
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Type       *ItemType  `json:"type,omitempty"`
	CategoryID *uuid.UUID `json:"category_id,omitempty"`
}