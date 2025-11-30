package item

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewItem(t *testing.T) {
	name := "专业理发服务"
	description := "包含洗剪吹的专业理发服务"
	itemType := ItemTypeService
	price := 88.00
	categoryID := uuid.New()

	item := NewItem(name, description, itemType, price, categoryID)

	assert.NotEqual(t, uuid.Nil, item.ID)
	assert.Equal(t, name, item.Name)
	assert.Equal(t, description, item.Description)
	assert.Equal(t, itemType, item.Type)
	assert.Equal(t, price, item.Price)
	assert.Equal(t, categoryID, item.CategoryID)
	assert.Equal(t, StatusDraft, item.Status)
	assert.False(t, item.CreatedAt.IsZero())
	assert.False(t, item.UpdatedAt.IsZero())
	assert.Nil(t, item.DeletedAt)
}

func TestItem_IsService(t *testing.T) {
	tests := []struct {
		name     string
		itemType ItemType
		want     bool
	}{
		{
			name:     "service item",
			itemType: ItemTypeService,
			want:     true,
		},
		{
			name:     "product item",
			itemType: ItemTypeProduct,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{Type: tt.itemType}
			assert.Equal(t, tt.want, item.IsService())
		})
	}
}

func TestItem_IsProduct(t *testing.T) {
	tests := []struct {
		name     string
		itemType ItemType
		want     bool
	}{
		{
			name:     "product item",
			itemType: ItemTypeProduct,
			want:     true,
		},
		{
			name:     "service item",
			itemType: ItemTypeService,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{Type: tt.itemType}
			assert.Equal(t, tt.want, item.IsProduct())
		})
	}
}

func TestItem_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status ItemStatus
		want   bool
	}{
		{
			name:   "active status",
			status: StatusActive,
			want:   true,
		},
		{
			name:   "inactive status",
			status: StatusInactive,
			want:   false,
		},
		{
			name:   "draft status",
			status: StatusDraft,
			want:   false,
		},
		{
			name:   "archived status",
			status: StatusArchived,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{Status: tt.status}
			assert.Equal(t, tt.want, item.IsActive())
		})
	}
}

func TestItem_CanDelete(t *testing.T) {
	tests := []struct {
		name   string
		status ItemStatus
		want   bool
	}{
		{
			name:   "draft status can delete",
			status: StatusDraft,
			want:   true,
		},
		{
			name:   "active status cannot delete",
			status: StatusActive,
			want:   false,
		},
		{
			name:   "inactive status cannot delete",
			status: StatusInactive,
			want:   false,
		},
		{
			name:   "archived status cannot delete",
			status: StatusArchived,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{Status: tt.status}
			assert.Equal(t, tt.want, item.CanDelete())
		})
	}
}

func TestItem_GetServiceFields(t *testing.T) {
	t.Run("service item returns service fields", func(t *testing.T) {
		duration := 60
		staffRequired := true
		capacity := 1

		item := &Item{
			Type:          ItemTypeService,
			Duration:      &duration,
			StaffRequired: &staffRequired,
			Capacity:      &capacity,
		}

		fields := item.GetServiceFields()
		require.NotNil(t, fields)
		assert.Equal(t, &duration, fields.Duration)
		assert.Equal(t, &staffRequired, fields.StaffRequired)
		assert.Equal(t, &capacity, fields.Capacity)
	})

	t.Run("product item returns nil", func(t *testing.T) {
		item := &Item{Type: ItemTypeProduct}
		assert.Nil(t, item.GetServiceFields())
	})
}

func TestItem_GetProductFields(t *testing.T) {
	t.Run("product item returns product fields", func(t *testing.T) {
		stock := 100
		costPrice := 50.00
		weight := 0.5
		sku := "PROD-001"

		item := &Item{
			Type:       ItemTypeProduct,
			Stock:      &stock,
			CostPrice:  &costPrice,
			Weight:     &weight,
			SKU:        &sku,
		}

		fields := item.GetProductFields()
		require.NotNil(t, fields)
		assert.Equal(t, &stock, fields.Stock)
		assert.Equal(t, &costPrice, fields.CostPrice)
		assert.Equal(t, &weight, fields.Weight)
		assert.Equal(t, &sku, fields.SKU)
	})

	t.Run("service item returns nil", func(t *testing.T) {
		item := &Item{Type: ItemTypeService}
		assert.Nil(t, item.GetProductFields())
	})
}

func TestItem_Update(t *testing.T) {
	item := NewItem("测试商品", "描述", ItemTypeProduct, 100.00, uuid.New())
	originalUpdatedAt := item.UpdatedAt

	// 等待一毫秒确保时间戳不同
	time.Sleep(1 * time.Millisecond)

	item.Update("更新商品", "更新描述", 150.00)

	assert.Equal(t, "更新商品", item.Name)
	assert.Equal(t, "更新描述", item.Description)
	assert.Equal(t, 150.00, item.Price)
	assert.True(t, item.UpdatedAt.After(originalUpdatedAt))
}

func TestItem_UpdateServiceFields(t *testing.T) {
	t.Run("service item can update service fields", func(t *testing.T) {
		item := NewItem("服务", "描述", ItemTypeService, 100.00, uuid.New())
		originalUpdatedAt := item.UpdatedAt

		duration := 120
		staffRequired := false
		capacity := 2

		time.Sleep(1 * time.Millisecond)
		item.UpdateServiceFields(&duration, &staffRequired, &capacity)

		assert.Equal(t, &duration, item.Duration)
		assert.Equal(t, &staffRequired, item.StaffRequired)
		assert.Equal(t, &capacity, item.Capacity)
		assert.True(t, item.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("product item cannot update service fields", func(t *testing.T) {
		duration := 60
		staffRequired := true
		capacity := 1

		item := &Item{
			Type:     ItemTypeProduct,
			Duration: &duration,
		}

		item.UpdateServiceFields(&duration, &staffRequired, &capacity)
		// 服务字段不应该被更新
		assert.Equal(t, &duration, item.Duration) // 保持原值
	})
}

func TestItem_UpdateProductFields(t *testing.T) {
	t.Run("product item can update product fields", func(t *testing.T) {
		item := NewItem("产品", "描述", ItemTypeProduct, 100.00, uuid.New())
		originalUpdatedAt := item.UpdatedAt

		stock := 200
		costPrice := 80.00
		weight := 1.5
		sku := "PROD-002"

		time.Sleep(1 * time.Millisecond)
		item.UpdateProductFields(&stock, &costPrice, &weight, &sku)

		assert.Equal(t, &stock, item.Stock)
		assert.Equal(t, &costPrice, item.CostPrice)
		assert.Equal(t, &weight, item.Weight)
		assert.Equal(t, &sku, item.SKU)
		assert.True(t, item.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("service item cannot update product fields", func(t *testing.T) {
		stock := 100

		item := &Item{
			Type:  ItemTypeService,
			Stock: &stock,
		}

		item.UpdateProductFields(&stock, nil, nil, nil)
		// 产品字段不应该被更新
		assert.Equal(t, &stock, item.Stock) // 保持原值
	})
}

func TestItem_Activate(t *testing.T) {
	item := NewItem("测试商品", "描述", ItemTypeProduct, 100.00, uuid.New())
	originalUpdatedAt := item.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	item.Activate()

	assert.Equal(t, StatusActive, item.Status)
	assert.True(t, item.UpdatedAt.After(originalUpdatedAt))
}

func TestItem_Deactivate(t *testing.T) {
	item := NewItem("测试商品", "描述", ItemTypeProduct, 100.00, uuid.New())
	item.Activate() // 先激活
	originalUpdatedAt := item.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	item.Deactivate()

	assert.Equal(t, StatusInactive, item.Status)
	assert.True(t, item.UpdatedAt.After(originalUpdatedAt))
}

func TestItem_Archive(t *testing.T) {
	item := NewItem("测试商品", "描述", ItemTypeProduct, 100.00, uuid.New())
	originalUpdatedAt := item.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	item.Archive()

	assert.Equal(t, StatusArchived, item.Status)
	assert.True(t, item.UpdatedAt.After(originalUpdatedAt))
}

func TestItem_UpdateStock(t *testing.T) {
	t.Run("product item can update stock", func(t *testing.T) {
		stock := 100
		item := &Item{
			Type: ItemTypeProduct,
			Stock: &stock,
		}

		item.UpdateStock(200)
		assert.Equal(t, 200, *item.Stock)
	})

	t.Run("service item cannot update stock", func(t *testing.T) {
		item := &Item{Type: ItemTypeService}
		originalUpdatedAt := item.UpdatedAt

		item.UpdateStock(100)
		// 服务不应该有库存更新逻辑
		assert.True(t, item.UpdatedAt.Equal(originalUpdatedAt))
	})

	t.Run("product item with nil stock cannot update", func(t *testing.T) {
		item := &Item{
			Type:  ItemTypeProduct,
			Stock: nil,
		}
		originalUpdatedAt := item.UpdatedAt

		item.UpdateStock(100)
		assert.True(t, item.UpdatedAt.Equal(originalUpdatedAt))
	})
}

func TestItem_DecreaseStock(t *testing.T) {
	t.Run("product item can decrease stock", func(t *testing.T) {
		stock := 100
		item := &Item{
			Type: ItemTypeProduct,
			Stock: &stock,
		}

		err := item.DecreaseStock(30)
		assert.NoError(t, err)
		assert.Equal(t, 70, *item.Stock)
	})

	t.Run("service item cannot decrease stock", func(t *testing.T) {
		item := &Item{Type: ItemTypeService}
		err := item.DecreaseStock(10)
		assert.Error(t, err)
		assert.Equal(t, ErrNotProductOrNoStock, err)
	})

	t.Run("product item with nil stock cannot decrease", func(t *testing.T) {
		item := &Item{
			Type:  ItemTypeProduct,
			Stock: nil,
		}
		err := item.DecreaseStock(10)
		assert.Error(t, err)
		assert.Equal(t, ErrNotProductOrNoStock, err)
	})

	t.Run("insufficient stock error", func(t *testing.T) {
		stock := 10
		item := &Item{
			Type: ItemTypeProduct,
			Stock: &stock,
		}

		err := item.DecreaseStock(20)
		assert.Error(t, err)
		assert.Equal(t, ErrInsufficientStock, err)
		assert.Equal(t, 10, *item.Stock) // 库存不变
	})
}

func TestItem_IncreaseStock(t *testing.T) {
	t.Run("product item can increase stock", func(t *testing.T) {
		stock := 100
		item := &Item{
			Type: ItemTypeProduct,
			Stock: &stock,
		}

		err := item.IncreaseStock(50)
		assert.NoError(t, err)
		assert.Equal(t, 150, *item.Stock)
	})

	t.Run("service item cannot increase stock", func(t *testing.T) {
		item := &Item{Type: ItemTypeService}
		err := item.IncreaseStock(10)
		assert.Error(t, err)
		assert.Equal(t, ErrNotProductOrNoStock, err)
	})

	t.Run("product item with nil stock cannot increase", func(t *testing.T) {
		item := &Item{
			Type:  ItemTypeProduct,
			Stock: nil,
		}
		err := item.IncreaseStock(10)
		assert.Error(t, err)
		assert.Equal(t, ErrNotProductOrNoStock, err)
	})
}

func TestItem_TableName(t *testing.T) {
	item := &Item{}
	assert.Equal(t, "items", item.TableName())
}

func TestItem_Constants(t *testing.T) {
	assert.Equal(t, ItemType("SERVICE"), ItemTypeService)
	assert.Equal(t, ItemType("PRODUCT"), ItemTypeProduct)

	assert.Equal(t, ItemStatus("ACTIVE"), StatusActive)
	assert.Equal(t, ItemStatus("INACTIVE"), StatusInactive)
	assert.Equal(t, ItemStatus("DRAFT"), StatusDraft)
	assert.Equal(t, ItemStatus("ARCHIVED"), StatusArchived)
}