package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/items/internal/domain/item"
)

// ItemService 商品服务接口
type ItemService interface {
	// 基础CRUD操作
	CreateItem(ctx context.Context, req *CreateItemRequest) (*item.Item, error)
	GetItemByID(ctx context.Context, id uuid.UUID) (*item.Item, error)
	UpdateItem(ctx context.Context, id uuid.UUID, req *UpdateItemRequest) (*item.Item, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error

	// 查询操作
	ListItems(ctx context.Context, filter *ListItemsRequest) ([]*item.Item, int64, error)
	GetItemsByCategory(ctx context.Context, categoryID uuid.UUID, filter *ListItemsRequest) ([]*item.Item, error)
	SearchItems(ctx context.Context, req *SearchItemsRequest) ([]*item.Item, error)

	// 业务操作
	UpdateItemStatus(ctx context.Context, id uuid.UUID, status item.ItemStatus) (*item.Item, error)
	ActivateItem(ctx context.Context, id uuid.UUID) (*item.Item, error)
	DeactivateItem(ctx context.Context, id uuid.UUID) (*item.Item, error)

	// 库存管理 (产品)
	UpdateStock(ctx context.Context, id uuid.UUID, quantity int) (*item.Item, error)
	GetLowStockItems(ctx context.Context, threshold int) ([]*item.Item, error)

	// 统计和分析
	GetItemsStats(ctx context.Context) (*ItemsStats, error)
	GetPopularItems(ctx context.Context, limit int) ([]*item.Item, error)
}

// CreateItemRequest 创建商品请求
type CreateItemRequest struct {
	Name        string     `json:"name" validate:"required,min=1,max=255"`
	Description string     `json:"description"`
	Type        item.ItemType `json:"type" validate:"required"`
	Price       float64    `json:"price" validate:"required,min=0"`
	CategoryID  uuid.UUID  `json:"category_id" validate:"required"`
	ImageURL    *string    `json:"image_url"`
	Tags        *string    `json:"tags"`

	// 服务特有字段
	Duration     *int `json:"duration,omitempty"`
	StaffRequired *bool `json:"staff_required,omitempty"`
	Capacity     *int `json:"capacity,omitempty"`

	// 产品特有字段
	Stock     *float64 `json:"stock,omitempty"`
	CostPrice *float64 `json:"cost_price,omitempty"`
	Weight    *float64 `json:"weight,omitempty"`
	SKU       *string  `json:"sku,omitempty"`
}

// UpdateItemRequest 更新商品请求
type UpdateItemRequest struct {
	Name        *string     `json:"name,omitempty"`
	Description *string     `json:"description,omitempty"`
	Price       *float64    `json:"price,omitempty"`
	ImageURL    *string     `json:"image_url,omitempty"`
	Tags        *string     `json:"tags,omitempty"`

	// 服务特有字段
	Duration     *int `json:"duration,omitempty"`
	StaffRequired *bool `json:"staff_required,omitempty"`
	Capacity     *int `json:"capacity,omitempty"`

	// 产品特有字段
	Stock     *float64 `json:"stock,omitempty"`
	CostPrice *float64 `json:"cost_price,omitempty"`
	Weight    *float64 `json:"weight,omitempty"`
	SKU       *string  `json:"sku,omitempty"`
}

// ListItemsRequest 列表查询请求
type ListItemsRequest struct {
	CategoryID *uuid.UUID    `json:"category_id,omitempty"`
	Type       *item.ItemType `json:"type,omitempty"`
	Status     *item.ItemStatus `json:"status,omitempty"`
	Limit      int           `json:"limit" validate:"min=1,max=100"`
	Offset     int           `json:"offset" validate:"min=0"`
	SortBy     string        `json:"sort_by,omitempty"`
	SortOrder  string        `json:"sort_order,omitempty"`
}

// SearchItemsRequest 搜索商品请求
type SearchItemsRequest struct {
	Query    string `json:"query" validate:"required,min=1,max=100"`
	Type     *item.ItemType `json:"type,omitempty"`
	Status   *item.ItemStatus `json:"status,omitempty"`
	Limit    int `json:"limit" validate:"min=1,max=50"`
}

// ItemsStats 商品统计
type ItemsStats struct {
	TotalItems     int64            `json:"total_items"`
	ActiveItems    int64            `json:"active_items"`
	InactiveItems  int64            `json:"inactive_items"`
	TypeDistribution map[item.ItemType]int64 `json:"type_distribution"`
	CategoryDistribution map[string]int64 `json:"category_distribution"`
	AveragePrice    float64          `json:"average_price"`
}

// itemService 商品服务实现
type itemService struct {
	itemRepo item.Repository
}

// NewItemService 创建商品服务实例
func NewItemService(itemRepo item.Repository) ItemService {
	return &itemService{
		itemRepo: itemRepo,
	}
}

// CreateItem 创建商品
func (s *itemService) CreateItem(ctx context.Context, req *CreateItemRequest) (*item.Item, error) {
	// 验证业务规则
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 生成ID
	id := uuid.New()

	// 创建领域实体
	domainItem := &item.Item{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
		Status:      item.StatusDraft,
		ImageURL:    req.ImageURL,
		Tags:        req.Tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 设置服务字段
	domainItem.Duration = req.Duration
	domainItem.StaffRequired = req.StaffRequired
	domainItem.Capacity = req.Capacity

	// 设置产品字段
	if req.Stock != nil {
		stock := int(*req.Stock)
		domainItem.Stock = &stock
	}
	domainItem.CostPrice = req.CostPrice
	domainItem.Weight = req.Weight
	domainItem.SKU = req.SKU

	// 保存到仓储
	if err := s.itemRepo.Create(ctx, domainItem); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	return domainItem, nil
}

// GetItemByID 根据ID获取商品
func (s *itemService) GetItemByID(ctx context.Context, id uuid.UUID) (*item.Item, error) {
	domainItem, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return domainItem, nil
}

// UpdateItem 更新商品
func (s *itemService) UpdateItem(ctx context.Context, id uuid.UUID, req *UpdateItemRequest) (*item.Item, error) {
	// 获取现有商品
	domainItem, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("item not found: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		domainItem.Name = *req.Name
	}
	if req.Description != nil {
		domainItem.Description = *req.Description
	}
	if req.Price != nil {
		domainItem.Price = *req.Price
	}
	if req.ImageURL != nil {
		domainItem.ImageURL = req.ImageURL
	}
	if req.Tags != nil {
		domainItem.Tags = req.Tags
	}

	// 更新服务字段
	if req.Duration != nil {
		domainItem.Duration = req.Duration
	}
	if req.StaffRequired != nil {
		domainItem.StaffRequired = req.StaffRequired
	}
	if req.Capacity != nil {
		domainItem.Capacity = req.Capacity
	}

	// 更新产品字段
	if req.Stock != nil {
		stock := int(*req.Stock)
		domainItem.Stock = &stock
	}
	if req.CostPrice != nil {
		domainItem.CostPrice = req.CostPrice
	}
	if req.Weight != nil {
		domainItem.Weight = req.Weight
	}
	if req.SKU != nil {
		domainItem.SKU = req.SKU
	}

	// 更新时间戳
	domainItem.UpdatedAt = time.Now()

	// 保存到仓储
	if err := s.itemRepo.Update(ctx, domainItem); err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	return domainItem, nil
}

// DeleteItem 删除商品
func (s *itemService) DeleteItem(ctx context.Context, id uuid.UUID) error {
	if err := s.itemRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

// ListItems 列出商品
func (s *itemService) ListItems(ctx context.Context, filter *ListItemsRequest) ([]*item.Item, int64, error) {
	// 构建查询过滤器
	domainFilter := &item.Filter{}
	if filter.Type != nil {
		domainFilter.Type = filter.Type
	}
	if filter.Status != nil {
		domainFilter.Status = filter.Status
	}
	if filter.CategoryID != nil {
		domainFilter.CategoryID = filter.CategoryID
	}

	// 构建分页 - 转换 offset/limit 到 page/pageSize
	page := filter.Offset/filter.Limit + 1
	if filter.Offset == 0 {
		page = 1
	}

	pagination := &item.Pagination{
		Page:     page,
		PageSize: filter.Limit,
	}

	items, total, err := s.itemRepo.List(ctx, domainFilter, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list items: %w", err)
	}

	return items, total, nil
}

// GetItemsByCategory 根据分类获取商品
func (s *itemService) GetItemsByCategory(ctx context.Context, categoryID uuid.UUID, filter *ListItemsRequest) ([]*item.Item, error) {
	// 使用仓储层的按分类查询方法
	return s.itemRepo.GetByCategory(ctx, item.ID(categoryID), item.StatusActive)
}

// SearchItems 搜索商品
func (s *itemService) SearchItems(ctx context.Context, req *SearchItemsRequest) ([]*item.Item, error) {
	// 构建搜索查询
	searchQuery := &item.SearchQuery{
		Query: req.Query,
		Limit: req.Limit,
	}

	if req.Type != nil {
		searchQuery.Type = req.Type
	}

	items, err := s.itemRepo.Search(ctx, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search items: %w", err)
	}

	// 过滤状态
	if req.Status != nil {
		var filtered []*item.Item
		for _, item := range items {
			if item.Status == *req.Status {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	return items, nil
}

// UpdateItemStatus 更新商品状态
func (s *itemService) UpdateItemStatus(ctx context.Context, id uuid.UUID, status item.ItemStatus) (*item.Item, error) {
	domainItem, err := s.itemRepo.GetByID(ctx, item.ID(id))
	if err != nil {
		return nil, fmt.Errorf("item not found: %w", err)
	}

	// 验证状态转换
	if !s.isValidStatusTransition(domainItem.Status, status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", domainItem.Status, status)
	}

	domainItem.Status = status
	domainItem.UpdatedAt = time.Now()

	if err := s.itemRepo.Update(ctx, domainItem); err != nil {
		return nil, fmt.Errorf("failed to update item status: %w", err)
	}

	return domainItem, nil
}

// ActivateItem 激活商品
func (s *itemService) ActivateItem(ctx context.Context, id uuid.UUID) (*item.Item, error) {
	return s.UpdateItemStatus(ctx, id, item.StatusActive)
}

// DeactivateItem 停用商品
func (s *itemService) DeactivateItem(ctx context.Context, id uuid.UUID) (*item.Item, error) {
	return s.UpdateItemStatus(ctx, id, item.StatusInactive)
}

// UpdateStock 更新库存（仅适用于产品）
func (s *itemService) UpdateStock(ctx context.Context, id uuid.UUID, quantity int) (*item.Item, error) {
	domainItem, err := s.itemRepo.GetByID(ctx, item.ID(id))
	if err != nil {
		return nil, fmt.Errorf("item not found: %w", err)
	}

	if domainItem.Type != item.ItemTypeProduct {
		return nil, fmt.Errorf("stock update is only applicable to products")
	}

	if quantity < 0 {
		return nil, fmt.Errorf("stock quantity cannot be negative")
	}

	domainItem.Stock = &quantity
	domainItem.UpdatedAt = time.Now()

	if err := s.itemRepo.Update(ctx, domainItem); err != nil {
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	return domainItem, nil
}

// GetLowStockItems 获取低库存商品
func (s *itemService) GetLowStockItems(ctx context.Context, threshold int) ([]*item.Item, error) {
	return s.itemRepo.GetLowStockItems(ctx, threshold)
}

// GetItemsStats 获取商品统计
func (s *itemService) GetItemsStats(ctx context.Context) (*ItemsStats, error) {
	filter := &item.StatisticsFilter{}
	domainStats, err := s.itemRepo.GetStatistics(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get items stats: %w", err)
	}

	// 转换为服务层统计格式
	stats := &ItemsStats{
		TotalItems:     domainStats.TotalCount,
		ActiveItems:    domainStats.CountByStatus[item.StatusActive],
		InactiveItems:  domainStats.CountByStatus[item.StatusInactive],
		TypeDistribution: domainStats.CountByType,
		CategoryDistribution: domainStats.CountByCategory,
		AveragePrice:    domainStats.AveragePrice,
	}

	return stats, nil
}

// GetPopularItems 获取热门商品（基于简单的逻辑）
func (s *itemService) GetPopularItems(ctx context.Context, limit int) ([]*item.Item, error) {
	// 这里可以实现更复杂的热度算法
	// 暂时返回按价格排序的商品
	pagination := &item.Pagination{
		Page:     1,
		PageSize: limit,
	}

	items, _, err := s.itemRepo.List(ctx, &item.Filter{}, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular items: %w", err)
	}

	// 过滤出活跃商品
	var activeItems []*item.Item
	for _, domainItem := range items {
		if domainItem.Status == item.StatusActive {
			activeItems = append(activeItems, domainItem)
		}
	}

	// 简单的按价格排序作为热度指标
	for i := 0; i < len(activeItems)-1; i++ {
		for j := i + 1; j < len(activeItems); j++ {
			if activeItems[i].Price < activeItems[j].Price {
				activeItems[i], activeItems[j] = activeItems[j], activeItems[i]
			}
		}
	}

	// 限制返回数量
	if len(activeItems) > limit {
		activeItems = activeItems[:limit]
	}

	return activeItems, nil
}

// validateCreateRequest 验证创建请求
func (s *itemService) validateCreateRequest(req *CreateItemRequest) error {
	// 基础验证
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}

	// 产品特定验证
	if req.Type == item.ItemTypeProduct {
		if req.Stock != nil && *req.Stock < 0 {
			return fmt.Errorf("stock cannot be negative")
		}
		if req.CostPrice != nil && *req.CostPrice < 0 {
			return fmt.Errorf("cost price cannot be negative")
		}
	}

	// 服务特定验证
	if req.Type == item.ItemTypeService {
		if req.Duration != nil && *req.Duration <= 0 {
			return fmt.Errorf("duration must be positive for services")
		}
		if req.Capacity != nil && *req.Capacity <= 0 {
			return fmt.Errorf("capacity must be positive for services")
		}
	}

	return nil
}

// isValidStatusTransition 检查状态转换是否有效
func (s *itemService) isValidStatusTransition(from, to item.ItemStatus) bool {
	// 定义允许的状态转换
	validTransitions := map[item.ItemStatus][]item.ItemStatus{
		item.StatusDraft:     {item.StatusActive, item.StatusInactive, item.StatusArchived},
		item.StatusActive:   {item.StatusInactive, item.StatusDraft, item.StatusArchived},
		item.StatusInactive: {item.StatusActive, item.StatusDraft, item.StatusArchived},
		item.StatusArchived: {item.StatusDraft, item.StatusActive, item.StatusInactive},
	}

	allowedTransitions, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowedStatus := range allowedTransitions {
		if allowedStatus == to {
			return true
		}
	}

	return false
}

