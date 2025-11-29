package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/items/internal/domain/item"
)

// Item GORM 数据库模型
type Item struct {
	ID          string    `gorm:"column:id;primaryKey;type:char(36)" json:"id"`
	Name        string    `gorm:"column:name;not null;size:255;index" json:"name"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	Type        string    `gorm:"column:type;not null;size:20;index" json:"type"`
	Price       float64   `gorm:"column:price;not null;type:decimal(10,2)" json:"price"`

	// 服务特有字段
	Duration      *int   `gorm:"column:duration;comment:服务时长(分钟)" json:"duration,omitempty"`
	StaffRequired *bool  `gorm:"column:staff_required;comment:是否需要员工" json:"staff_required,omitempty"`
	Capacity      *int   `gorm:"column:capacity;comment:服务容量" json:"capacity,omitempty"`

	// 产品特有字段
	Stock     *float64 `gorm:"column:stock;comment:库存数量" json:"stock,omitempty"`
	CostPrice *float64 `gorm:"column:cost_price;type:decimal(10,2);comment:成本价" json:"cost_price,omitempty"`
	Weight    *float64 `gorm:"column:weight;type:decimal(8,2);comment:重量(kg)" json:"weight,omitempty"`
	SKU       *string  `gorm:"column:sku;size:100;comment:商品编码" json:"sku,omitempty"`

	// 通用字段
	CategoryID string  `gorm:"column:category_id;not null;type:char(36);index" json:"category_id"`
	Status     string  `gorm:"column:status;not null;size:20;default:'DRAFT';index" json:"status"`
	ImageURL   *string `gorm:"column:image_url;size:500" json:"image_url,omitempty"`
	Tags       *string `gorm:"column:tags;type:text;comment:逗号分隔的标签" json:"tags,omitempty"`

	// 时间戳
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`

	// 关联
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// TableName 设置表名
func (Item) TableName() string {
	return "items"
}

// ToDomain 转换为领域模型
func (m *Item) ToDomain() *item.Item {
	// 解析 UUID
	id, _ := uuid.Parse(m.ID)
	categoryID, _ := uuid.Parse(m.CategoryID)

	domain := &item.Item{
		ID:          id,
		Name:        m.Name,
		Description: m.Description,
		Type:        item.ItemType(m.Type),
		Price:       m.Price,
		CategoryID:  categoryID,
		Status:      item.ItemStatus(m.Status),
		ImageURL:    m.ImageURL,
		Tags:        m.Tags,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
	}

	// 设置服务字段
	domain.Duration = m.Duration
	domain.StaffRequired = m.StaffRequired
	domain.Capacity = m.Capacity

	// 设置产品字段
	if m.Stock != nil {
		stock := int(*m.Stock)
		domain.Stock = &stock
	}
	domain.CostPrice = m.CostPrice
	domain.Weight = m.Weight
	domain.SKU = m.SKU

	return domain
}

// FromDomain 从领域模型转换
func FromDomain(domain *item.Item) *Item {
	if domain == nil {
		return nil
	}

	model := &Item{
		ID:          domain.ID.String(),
		Name:        domain.Name,
		Description: domain.Description,
		Type:        string(domain.Type),
		Price:       domain.Price,
		CategoryID:  domain.CategoryID.String(),
		Status:      string(domain.Status),
		ImageURL:    domain.ImageURL,
		Tags:        domain.Tags,
		CreatedAt:   domain.CreatedAt,
		UpdatedAt:   domain.UpdatedAt,
	}

	// 设置服务字段
	model.Duration = domain.Duration
	model.StaffRequired = domain.StaffRequired
	model.Capacity = domain.Capacity

	// 设置产品字段
	if domain.Stock != nil {
		stock := float64(*domain.Stock)
		model.Stock = &stock
	}
	model.CostPrice = domain.CostPrice
	model.Weight = domain.Weight
	model.SKU = domain.SKU

	// 设置删除时间
	if domain.DeletedAt != nil {
		model.DeletedAt = domain.DeletedAt
	}

	return model
}

// BatchToDomain 批量转换为领域模型
func BatchToDomain(models []*Item) []*item.Item {
	domains := make([]*item.Item, len(models))
	for i, model := range models {
		domains[i] = model.ToDomain()
	}
	return domains
}

// BatchFromDomain 批量从领域模型转换
func BatchFromDomain(domains []*item.Item) []*Item {
	models := make([]*Item, len(domains))
	for i, domain := range domains {
		models[i] = FromDomain(domain)
	}
	return models
}