package category

import (
	"time"

	"github.com/google/uuid"
)

// CategoryStatus 分类状态
type CategoryStatus string

const (
	CategoryStatusActive   CategoryStatus = "ACTIVE"   // 激活
	CategoryStatusInactive CategoryStatus = "INACTIVE" // 停用
)

// Category 商品分类实体
type Category struct {
	AutoID      uint64         `json:"auto_id" gorm:"column:auto_id;primaryKey;autoIncrement"`     // 自增主键
	ID          uuid.UUID      `json:"uuid" gorm:"column:id;uniqueIndex;type:char(36)"`                // 业务UUID
	Name        string         `json:"name" gorm:"column:name;not null;size:255;index"`
	Description string         `json:"description" gorm:"column:description;type:text"`
	ParentID    *uuid.UUID     `json:"parent_id,omitempty" gorm:"column:parent_id;type:char(36);index"`
	Icon        *string        `json:"icon,omitempty" gorm:"column:icon;size:500"`
	SortOrder   int            `json:"sort_order" gorm:"column:sort_order;default:0"`
	Status      CategoryStatus `json:"status" gorm:"column:status;not null;size:20;default:'ACTIVE';index"`

	// 关联
	Children  []*Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Parent    *Category   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	ItemCount int         `json:"item_count" gorm:"column:item_count;default:0"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at;index"`
}

// TableName 设置表名
func (Category) TableName() string {
	return "categories"
}

// IsRoot 判断是否为根分类
func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}

// IsActive 判断是否为激活状态
func (c *Category) IsActive() bool {
	return c.Status == CategoryStatusActive
}

// GetPath 获取分类路径
func (c *Category) GetPath() string {
	if c.IsRoot() {
		return c.Name
	}
	if c.Parent != nil {
		return c.Parent.GetPath() + " > " + c.Name
	}
	return c.Name
}

// GetLevel 获取分类层级
func (c *Category) GetLevel() int {
	if c.IsRoot() {
		return 0
	}
	if c.Parent != nil {
		return c.Parent.GetLevel() + 1
	}
	return 0
}

// CanDelete 判断是否可以删除
func (c *Category) CanDelete() bool {
	return c.ItemCount == 0 && len(c.Children) == 0
}

// NewCategory 创建新分类
func NewCategory(name, description string, parentID *uuid.UUID) *Category {
	return &Category{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		ParentID:    parentID,
		Status:      CategoryStatusActive,
		SortOrder:   0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Update 更新分类信息
func (c *Category) Update(name, description string, icon *string, sortOrder int) {
	c.Name = name
	c.Description = description
	c.Icon = icon
	c.SortOrder = sortOrder
	c.UpdatedAt = time.Now()
}

// Activate 激活分类
func (c *Category) Activate() {
	c.Status = CategoryStatusActive
	c.UpdatedAt = time.Now()
}

// Deactivate 停用分类
func (c *Category) Deactivate() {
	c.Status = CategoryStatusInactive
	c.UpdatedAt = time.Now()
}

// UpdateItemCount 更新商品数量
func (c *Category) UpdateItemCount(count int) {
	c.ItemCount = count
	c.UpdatedAt = time.Now()
}

// AddChild 添加子分类
func (c *Category) AddChild(child *Category) {
	if c.Children == nil {
		c.Children = make([]*Category, 0)
	}
	c.Children = append(c.Children, child)
	child.ParentID = &c.ID
	child.Parent = c
}

// ToTree 转换为树结构
func (c *Category) ToTree() *CategoryTree {
	node := &CategoryTree{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		ParentID:    c.ParentID,
		Icon:        c.Icon,
		SortOrder:   c.SortOrder,
		Status:      c.Status,
		ItemCount:   c.ItemCount,
		Children:    make([]*CategoryTree, 0),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	if c.Children != nil {
		for _, child := range c.Children {
			node.Children = append(node.Children, child.ToTree())
		}
	}

	return node
}

// CategoryTree 分类树结构
type CategoryTree struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ParentID    *uuid.UUID     `json:"parent_id,omitempty"`
	Icon        *string        `json:"icon,omitempty"`
	SortOrder   int            `json:"sort_order"`
	Status      CategoryStatus `json:"status"`
	ItemCount   int            `json:"item_count"`
	Level       int            `json:"level"`
	Path        string         `json:"path"`
	Children    []*CategoryTree `json:"children,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// CategoryPath 分类路径
type CategoryPath struct {
	CategoryID uuid.UUID `json:"category_id"`
	Name       string    `json:"name"`
	Level      int       `json:"level"`
	Path       string    `json:"path"`
}