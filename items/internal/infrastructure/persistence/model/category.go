package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/items/internal/domain/category"
)

// Category GORM 数据库模型
type Category struct {
	ID          string             `gorm:"column:id;primaryKey;type:char(36)" json:"id"`
	Name        string             `gorm:"column:name;not null;size:255;index" json:"name"`
	Description string             `gorm:"column:description;type:text" json:"description"`
	ParentID    *string            `gorm:"column:parent_id;type:char(36);index" json:"parent_id,omitempty"`
	Icon        *string            `gorm:"column:icon;size:500" json:"icon,omitempty"`
	SortOrder   int                `gorm:"column:sort_order;default:0" json:"sort_order"`
	Status      string             `gorm:"column:status;not null;size:20;default:'ACTIVE';index" json:"status"`
	ItemCount   int                `gorm:"column:item_count;default:0" json:"item_count"`

	// 时间戳
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`

	// 关联
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []*Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Items    []*Item     `gorm:"foreignKey:CategoryID" json:"items,omitempty"`
}

// TableName 设置表名
func (Category) TableName() string {
	return "categories"
}

// ToDomain 转换为领域模型
func (m *Category) ToDomain() *category.Category {
	// 解析 UUID
	id, _ := uuid.Parse(m.ID)
	var parentID *uuid.UUID
	if m.ParentID != nil {
		pid, _ := uuid.Parse(*m.ParentID)
		parentID = &pid
	}

	domain := &category.Category{
		ID:          id,
		Name:        m.Name,
		Description: m.Description,
		ParentID:    parentID,
		Icon:        m.Icon,
		SortOrder:   m.SortOrder,
		Status:      category.CategoryStatus(m.Status),
		ItemCount:   m.ItemCount,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
	}

	return domain
}

// FromDomain 从领域模型转换
func FromCategoryDomain(domain *category.Category) *Category {
	if domain == nil {
		return nil
	}

	model := &Category{
		ID:          domain.ID.String(),
		Name:        domain.Name,
		Description: domain.Description,
		Icon:        domain.Icon,
		SortOrder:   domain.SortOrder,
		Status:      string(domain.Status),
		ItemCount:   domain.ItemCount,
		CreatedAt:   domain.CreatedAt,
		UpdatedAt:   domain.UpdatedAt,
	}

	// 设置父分类ID
	if domain.ParentID != nil {
		parentID := domain.ParentID.String()
		model.ParentID = &parentID
	}

	// 设置删除时间
	if domain.DeletedAt != nil {
		model.DeletedAt = domain.DeletedAt
	}

	return model
}

// BatchToDomain 批量转换为领域模型
func BatchToCategoryDomain(models []*Category) []*category.Category {
	domains := make([]*category.Category, len(models))
	for i, model := range models {
		domains[i] = model.ToDomain()
	}
	return domains
}

// BatchFromDomain 批量从领域模型转换
func BatchFromCategoryDomain(domains []*category.Category) []*Category {
	models := make([]*Category, len(domains))
	for i, domain := range domains {
		models[i] = FromCategoryDomain(domain)
	}
	return models
}