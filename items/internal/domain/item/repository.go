package item

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository 商品仓储接口
type Repository interface {
	// 基础 CRUD 操作
	Create(ctx context.Context, item *Item) error
	GetByID(ctx context.Context, id uuid.UUID) (*Item, error)
	Update(ctx context.Context, item *Item) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 批量操作
	CreateBatch(ctx context.Context, items []*Item) error
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*Item, error)
	DeleteBatch(ctx context.Context, ids []uuid.UUID) error

	// 查询操作
	List(ctx context.Context, filter *Filter, pagination *Pagination) ([]*Item, int64, error)
	Search(ctx context.Context, query *SearchQuery) ([]*Item, error)

	// 特殊查询
	GetByCategory(ctx context.Context, categoryID uuid.UUID, status ItemStatus) ([]*Item, error)
	GetBySKU(ctx context.Context, sku string) (*Item, error)
	GetActiveItems(ctx context.Context) ([]*Item, error)
	GetLowStockItems(ctx context.Context, threshold int) ([]*Item, error)

	// 统计操作
	Count(ctx context.Context, filter *Filter) (int64, error)
	GetStatistics(ctx context.Context, filter *StatisticsFilter) (*Statistics, error)
}

// Filter 查询过滤器
type Filter struct {
	Type       *ItemType  `json:"type,omitempty"`       // 商品类型
	Status     *ItemStatus `json:"status,omitempty"`     // 商品状态
	CategoryID *uuid.UUID  `json:"category_id,omitempty"` // 分类ID
	MinPrice   *float64    `json:"min_price,omitempty"`   // 最低价格
	MaxPrice   *float64    `json:"max_price,omitempty"`   // 最高价格
	HasStock   *bool       `json:"has_stock,omitempty"`   // 是否有库存
	Search     *string     `json:"search,omitempty"`     // 搜索关键词
}

// Pagination 分页参数
type Pagination struct {
	Page     int `json:"page"`      // 页码，从1开始
	PageSize int `json:"page_size"` // 每页大小
}

// GetOffset 获取偏移量
func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取限制数
func (p *Pagination) GetLimit() int {
	return p.PageSize
}

// Validate 验证分页参数
func (p *Pagination) Validate() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// SearchQuery 搜索查询
type SearchQuery struct {
	Query      string     `json:"query"`        // 搜索关键词
	Type       *ItemType  `json:"type,omitempty"` // 商品类型过滤
	CategoryID *uuid.UUID `json:"category_id,omitempty"` // 分类过滤
	MinPrice   *float64   `json:"min_price,omitempty"` // 最低价格
	MaxPrice   *float64   `json:"max_price,omitempty"` // 最高价格
	Limit      int        `json:"limit"`       // 返回结果数量限制
}

// Statistics 统计信息
type Statistics struct {
	TotalCount       int64             `json:"total_count"`        // 总数
	CountByType      map[ItemType]int64 `json:"count_by_type"`      // 按类型统计
	CountByStatus    map[ItemStatus]int64 `json:"count_by_status"`    // 按状态统计
	CountByCategory  map[string]int64   `json:"count_by_category"`  // 按分类统计
	AveragePrice     float64           `json:"average_price"`       // 平均价格
	TotalValue       float64           `json:"total_value"`         // 总价值
	LowStockCount    int64             `json:"low_stock_count"`     // 低库存数量
	OutOfStockCount  int64             `json:"out_of_stock_count"`   // 缺货数量
}

// StatisticsFilter 统计过滤器
type StatisticsFilter struct {
	StartDate *time.Time `json:"start_date,omitempty"` // 开始日期
	EndDate   *time.Time `json:"end_date,omitempty"`   // 结束日期
	Type      *ItemType  `json:"type,omitempty"`       // 商品类型
	CategoryID *uuid.UUID `json:"category_id,omitempty"` // 分类ID
}