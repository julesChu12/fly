package category

import (
	"context"

	"github.com/google/uuid"
)

// Repository 分类仓储接口
type Repository interface {
	// 基础 CRUD 操作
	Create(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 查询操作
	List(ctx context.Context, filter *Filter) ([]*Category, error)
	GetTree(ctx context.Context, filter *Filter) ([]*CategoryTree, error)
	GetRoots(ctx context.Context) ([]*Category, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*Category, error)
	GetPath(ctx context.Context, id uuid.UUID) ([]*CategoryPath, error)

	// 特殊查询
	GetByName(ctx context.Context, name string) (*Category, error)
	FindByParentID(ctx context.Context, parentID *uuid.UUID) ([]*Category, error)
	GetActiveCategories(ctx context.Context) ([]*Category, error)
	HasChildren(ctx context.Context, id uuid.UUID) (bool, error)
	GetDescendantIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error)

	// 统计操作
	UpdateItemCount(ctx context.Context, id uuid.UUID, count int) error
	GetItemCount(ctx context.Context, id uuid.UUID) (int, error)
}

// Filter 查询过滤器
type Filter struct {
	ParentID *uuid.UUID      `json:"parent_id,omitempty"` // 父分类ID
	Status   *CategoryStatus `json:"status,omitempty"`    // 状态
	Level    *int            `json:"level,omitempty"`     // 层级
	Search   *string         `json:"search,omitempty"`    // 搜索关键词
}

