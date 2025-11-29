package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/julesChu12/fly/items/internal/domain/item"
	"github.com/julesChu12/fly/items/internal/infrastructure/persistence/model"
)

// itemRepository 商品仓储实现
type itemRepository struct {
	db *gorm.DB
}

// NewItemRepository 创建商品仓储
func NewItemRepository(db *gorm.DB) item.Repository {
	return &itemRepository{
		db: db,
	}
}

// Create 创建商品
func (r *itemRepository) Create(ctx context.Context, item *item.Item) error {
	model := model.FromDomain(item)
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID 根据ID获取商品
func (r *itemRepository) GetByID(ctx context.Context, id item.ID) (*item.Item, error) {
	var model model.Item
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, item.ErrItemNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

// Update 更新商品
func (r *itemRepository) Update(ctx context.Context, item *item.Item) error {
	model := model.FromDomain(item)
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete 删除商品
func (r *itemRepository) Delete(ctx context.Context, id item.ID) error {
	return r.db.WithContext(ctx).Delete(&model.Item{}, "id = ?", id.String()).Error
}

// CreateBatch 批量创建商品
func (r *itemRepository) CreateBatch(ctx context.Context, items []*item.Item) error {
	models := model.BatchFromDomain(items)
	return r.db.WithContext(ctx).CreateInBatches(models, 100).Error
}

// GetByIDs 根据ID列表获取商品
func (r *itemRepository) GetByIDs(ctx context.Context, ids []item.ID) ([]*item.Item, error) {
	var models []*model.Item
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.String()
	}

	err := r.db.WithContext(ctx).Where("id IN ?", idStrings).Find(&models).Error
	if err != nil {
		return nil, err
	}
	return model.BatchToDomain(models), nil
}

// DeleteBatch 批量删除商品
func (r *itemRepository) DeleteBatch(ctx context.Context, ids []item.ID) error {
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.String()
	}
	return r.db.WithContext(ctx).Delete(&model.Item{}, "id IN ?", idStrings).Error
}

// List 获取商品列表
func (r *itemRepository) List(ctx context.Context, filter *item.Filter, pagination *item.Pagination) ([]*item.Item, int64, error) {
	var models []*model.Item
	var total int64

	query := r.buildFilterQuery(filter)

	// 查询总数
	if err := query.WithContext(ctx).Model(&model.Item{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	offset := pagination.GetOffset()
	limit := pagination.GetLimit()
	err := query.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, 0, err
	}

	return model.BatchToDomain(models), total, nil
}

// Search 搜索商品
func (r *itemRepository) Search(ctx context.Context, query *item.SearchQuery) ([]*item.Item, error) {
	var models []*model.Item

	db := r.db.WithContext(ctx).
		Where("name LIKE ? OR description LIKE ?", "%"+query.Query+"%", "%"+query.Query+"%")

	if query.Type != nil {
		db = db.Where("type = ?", *query.Type)
	}
	if query.CategoryID != nil {
		db = db.Where("category_id = ?", query.CategoryID.String())
	}
	if query.MinPrice != nil {
		db = db.Where("price >= ?", *query.MinPrice)
	}
	if query.MaxPrice != nil {
		db = db.Where("price <= ?", *query.MaxPrice)
	}

	err := db.Limit(query.Limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	return model.BatchToDomain(models), nil
}

// GetByCategory 根据分类获取商品
func (r *itemRepository) GetByCategory(ctx context.Context, categoryID item.ID, status item.ItemStatus) ([]*item.Item, error) {
	var models []*model.Item

	err := r.db.WithContext(ctx).
		Where("category_id = ? AND status = ?", categoryID.String(), string(status)).
		Order("sort_order ASC, created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	return model.BatchToDomain(models), nil
}

// GetBySKU 根据SKU获取商品
func (r *itemRepository) GetBySKU(ctx context.Context, sku string) (*item.Item, error) {
	var model model.Item
	err := r.db.WithContext(ctx).Where("sku = ?", sku).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, item.ErrItemNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

// GetActiveItems 获取活跃商品
func (r *itemRepository) GetActiveItems(ctx context.Context) ([]*item.Item, error) {
	var models []*model.Item

	err := r.db.WithContext(ctx).
		Where("status = ?", string(item.StatusActive)).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	return model.BatchToDomain(models), nil
}

// GetLowStockItems 获取低库存商品
func (r *itemRepository) GetLowStockItems(ctx context.Context, threshold int) ([]*item.Item, error) {
	var models []*model.Item

	err := r.db.WithContext(ctx).
		Where("type = ? AND stock IS NOT NULL AND stock <= ?", string(item.ItemTypeProduct), threshold).
		Order("stock ASC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	return model.BatchToDomain(models), nil
}

// Count 统计商品数量
func (r *itemRepository) Count(ctx context.Context, filter *item.Filter) (int64, error) {
	var count int64
	query := r.buildFilterQuery(filter)
	err := query.WithContext(ctx).Model(&model.Item{}).Count(&count).Error
	return count, err
}

// GetStatistics 获取统计信息
func (r *itemRepository) GetStatistics(ctx context.Context, filter *item.StatisticsFilter) (*item.Statistics, error) {
	stats := &item.Statistics{
		CountByType:     make(map[item.ItemType]int64),
		CountByStatus:   make(map[item.ItemStatus]int64),
		CountByCategory: make(map[string]int64),
	}

	query := r.db.WithContext(ctx).Model(&model.Item{})

	// 应用过滤器
	if filter.StartDate != nil || filter.EndDate != nil {
		if filter.StartDate != nil {
			query = query.Where("created_at >= ?", *filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("created_at <= ?", *filter.EndDate)
		}
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", filter.CategoryID.String())
	}

	// 总数
	query.Count(&stats.TotalCount)

	// 按类型统计
	var typeStats []struct {
		Type  string
		Count int64
	}
	query.Select("type, COUNT(*) as count").Group("type").Scan(&typeStats)
	for _, stat := range typeStats {
		stats.CountByType[item.ItemType(stat.Type)] = stat.Count
	}

	// 按状态统计
	var statusStats []struct {
		Status string
		Count  int64
	}
	query.Select("status, COUNT(*) as count").Group("status").Scan(&statusStats)
	for _, stat := range statusStats {
		stats.CountByStatus[item.ItemStatus(stat.Status)] = stat.Count
	}

	// 按分类统计
	var categoryStats []struct {
		CategoryID string
		Count      int64
	}
	query.Select("category_id, COUNT(*) as count").Group("category_id").Scan(&categoryStats)
	for _, stat := range categoryStats {
		stats.CountByCategory[stat.CategoryID] = stat.Count
	}

	// 价格统计
	var priceStats struct {
		AvgPrice float64
		SumPrice float64
	}
	query.Select("AVG(price) as avg_price, SUM(price) as sum_price").Scan(&priceStats)
	stats.AveragePrice = priceStats.AvgPrice
	stats.TotalValue = priceStats.SumPrice

	// 库存统计
	var stockStats struct {
		LowStock   int64
		OutOfStock int64
	}
	r.db.WithContext(ctx).
		Model(&model.Item{}).
		Where("type = ? AND stock IS NOT NULL", string(item.ItemTypeProduct)).
		Select("COUNT(CASE WHEN stock <= 10 THEN 1 END) as low_stock, COUNT(CASE WHEN stock = 0 THEN 1 END) as out_of_stock").
		Scan(&stockStats)
	stats.LowStockCount = stockStats.LowStock
	stats.OutOfStockCount = stockStats.OutOfStock

	return stats, nil
}

// buildFilterQuery 构建过滤查询
func (r *itemRepository) buildFilterQuery(filter *item.Filter) *gorm.DB {
	query := r.db.Model(&model.Item{})

	if filter == nil {
		return query
	}

	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", filter.CategoryID.String())
	}
	if filter.MinPrice != nil {
		query = query.Where("price >= ?", *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		query = query.Where("price <= ?", *filter.MaxPrice)
	}
	if filter.HasStock != nil {
		if *filter.HasStock {
			query = query.Where("stock IS NOT NULL AND stock > 0")
		} else {
			query = query.Where("stock IS NULL OR stock = 0")
		}
	}
	if filter.Search != nil && *filter.Search != "" {
		searchTerm := "%" + *filter.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchTerm, searchTerm)
	}

	return query
}