package item

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ID 商品ID类型
type ID = uuid.UUID

// ItemType 商品类型
type ItemType string

const (
	ItemTypeService ItemType = "SERVICE" // 服务类商品
	ItemTypeProduct ItemType = "PRODUCT" // 产品类商品
)

// ItemStatus 商品状态
type ItemStatus string

const (
	StatusActive   ItemStatus = "ACTIVE"   // 活跃状态
	StatusInactive ItemStatus = "INACTIVE" // 停用状态
	StatusDraft    ItemStatus = "DRAFT"    // 草稿状态
	StatusArchived ItemStatus = "ARCHIVED" // 归档状态
)

// Item 商品实体
type Item struct {
	AutoID      uint64     `json:"auto_id" gorm:"column:auto_id;primaryKey;autoIncrement"`     // 自增主键
	ID          uuid.UUID  `json:"uuid" gorm:"column:id;uniqueIndex;type:char(36)"`                // 业务UUID
	Name        string     `json:"name" gorm:"column:name;not null;size:255"`
	Description string     `json:"description" gorm:"column:description;type:text"`
	Type        ItemType  `json:"type" gorm:"column:type;not null;size:20;index"`
	Price       float64    `json:"price" gorm:"column:price;not null;type:decimal(10,2)"`

	// 服务特有字段
	Duration       *int `json:"duration,omitempty" gorm:"column:duration;comment:服务时长(分钟)"`
	StaffRequired  *bool `json:"staff_required,omitempty" gorm:"column:staff_required;comment:是否需要员工"`
	Capacity       *int `json:"capacity,omitempty" gorm:"column:capacity;comment:服务容量"`

	// 产品特有字段
	Stock       *int   `json:"stock,omitempty" gorm:"column:stock;comment:库存数量"`
	CostPrice   *float64 `json:"cost_price,omitempty" gorm:"column:cost_price;type:decimal(10,2);comment:成本价"`
	Weight      *float64 `json:"weight,omitempty" gorm:"column:weight;type:decimal(8,2);comment:重量(kg)"`
	SKU         *string `json:"sku,omitempty" gorm:"column:sku;size:100;comment:商品编码"`

	// 通用字段
	CategoryID   uuid.UUID `json:"category_id" gorm:"column:category_id;not null;type:char(36);index"`
	Status       ItemStatus `json:"status" gorm:"column:status;not null;size:20;default:'DRAFT';index"`
	ImageURL     *string    `json:"image_url,omitempty" gorm:"column:image_url;size:500"`
	Tags         *string    `json:"tags,omitempty" gorm:"column:tags;type:text;comment:逗号分隔的标签"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at;index"`
}

// TableName 设置表名
func (Item) TableName() string {
	return "items"
}

// BeforeCreate GORM钩子 - 在创建前生成UUID
func (i *Item) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// IsService 判断是否为服务类型
func (i *Item) IsService() bool {
	return i.Type == ItemTypeService
}

// IsProduct 判断是否为产品类型
func (i *Item) IsProduct() bool {
	return i.Type == ItemTypeProduct
}

// IsActive 判断是否为活跃状态
func (i *Item) IsActive() bool {
	return i.Status == StatusActive
}

// CanDelete 判断是否可以删除
func (i *Item) CanDelete() bool {
	return i.Status == StatusDraft
}

// GetServiceFields 获取服务相关字段
func (i *Item) GetServiceFields() *ServiceFields {
	if !i.IsService() {
		return nil
	}
	return &ServiceFields{
		Duration:      i.Duration,
		StaffRequired: i.StaffRequired,
		Capacity:      i.Capacity,
	}
}

// GetProductFields 获取产品相关字段
func (i *Item) GetProductFields() *ProductFields {
	if !i.IsProduct() {
		return nil
	}
	return &ProductFields{
		Stock:     i.Stock,
		CostPrice: i.CostPrice,
		Weight:    i.Weight,
		SKU:       i.SKU,
	}
}

// ServiceFields 服务特有字段
type ServiceFields struct {
	Duration      *int  `json:"duration,omitempty"`      // 服务时长(分钟)
	StaffRequired *bool `json:"staff_required,omitempty"` // 是否需要员工
	Capacity      *int  `json:"capacity,omitempty"`      // 服务容量
}

// ProductFields 产品特有字段
type ProductFields struct {
	Stock     *int     `json:"stock,omitempty"`     // 库存数量
	CostPrice *float64 `json:"cost_price,omitempty"` // 成本价
	Weight    *float64 `json:"weight,omitempty"`    // 重量(kg)
	SKU       *string  `json:"sku,omitempty"`       // 商品编码
}

// NewItem 创建新商品
func NewItem(name, description string, itemType ItemType, price float64, categoryID uuid.UUID) *Item {
	return &Item{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Type:        itemType,
		Price:       price,
		CategoryID:  categoryID,
		Status:      StatusDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Update 更新商品信息
func (i *Item) Update(name, description string, price float64) {
	i.Name = name
	i.Description = description
	i.Price = price
	i.UpdatedAt = time.Now()
}

// UpdateServiceFields 更新服务字段
func (i *Item) UpdateServiceFields(duration *int, staffRequired *bool, capacity *int) {
	if i.IsService() {
		i.Duration = duration
		i.StaffRequired = staffRequired
		i.Capacity = capacity
		i.UpdatedAt = time.Now()
	}
}

// UpdateProductFields 更新产品字段
func (i *Item) UpdateProductFields(stock *int, costPrice *float64, weight *float64, sku *string) {
	if i.IsProduct() {
		i.Stock = stock
		i.CostPrice = costPrice
		i.Weight = weight
		i.SKU = sku
		i.UpdatedAt = time.Now()
	}
}

// Activate 激活商品
func (i *Item) Activate() {
	i.Status = StatusActive
	i.UpdatedAt = time.Now()
}

// Deactivate 停用商品
func (i *Item) Deactivate() {
	i.Status = StatusInactive
	i.UpdatedAt = time.Now()
}

// Archive 归档商品
func (i *Item) Archive() {
	i.Status = StatusArchived
	i.UpdatedAt = time.Now()
}

// UpdateStock 更新库存（仅产品）
func (i *Item) UpdateStock(stock int) {
	if i.IsProduct() && i.Stock != nil {
		*i.Stock = stock
		i.UpdatedAt = time.Now()
	}
}

// DecreaseStock 减少库存（仅产品）
func (i *Item) DecreaseStock(quantity int) error {
	if !i.IsProduct() || i.Stock == nil {
		return ErrNotProductOrNoStock
	}

	if *i.Stock < quantity {
		return ErrInsufficientStock
	}

	*i.Stock -= quantity
	i.UpdatedAt = time.Now()
	return nil
}

// IncreaseStock 增加库存（仅产品）
func (i *Item) IncreaseStock(quantity int) error {
	if !i.IsProduct() || i.Stock == nil {
		return ErrNotProductOrNoStock
	}

	*i.Stock += quantity
	i.UpdatedAt = time.Now()
	return nil
}