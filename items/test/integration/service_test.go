package integration

import (
	"context"
	"testing"

	"github.com/julesChu12/fly/items/internal/application/service"
	"github.com/julesChu12/fly/items/internal/domain/item"
	"github.com/julesChu12/fly/items/internal/infrastructure/persistence/model"
	"github.com/julesChu12/fly/items/internal/infrastructure/persistence/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TestItemService 测试Item服务层
func TestItemService(t *testing.T) {
	ctx := context.Background()
	container, err := SetupTestContainers(ctx)
	require.NoError(t, err, "Failed to setup test containers")
	defer container.Cleanup(ctx)

	// 获取数据库连接
	db := container.Database

	// 创建表结构
	err = db.AutoMigrate(&model.Item{}, &model.Category{})
	require.NoError(t, err, "Failed to auto-migrate tables")

	// 创建仓储和服务实例
	itemRepo := repository.NewItemRepository(db)
	itemService := service.NewItemService(itemRepo)

	t.Run("CreateProduct", func(t *testing.T) {
		// 首先创建分类
		categoryID := uuid.New()
		err := db.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "电子产品分类", "电子设备分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		// 创建产品商品请求
		req := &service.CreateItemRequest{
			Name:        "智能手机",
			Description: "最新款智能手机",
			Type:        item.ItemTypeProduct,
			Price:       4999.00,
			CategoryID:  categoryID,
			Stock:       toFloat64Ptr(100),
			CostPrice:   toFloat64Ptr(4000.00),
			SKU:         toStringPtr("PHONE-001"),
		}

		// 执行创建
		domainItem, err := itemService.CreateItem(ctx, req)
		require.NoError(t, err, "Failed to create product")

		// 验证创建结果
		assert.NotEqual(t, uuid.Nil, domainItem.ID)
		assert.Equal(t, "智能手机", domainItem.Name)
		assert.Equal(t, item.ItemTypeProduct, domainItem.Type)
		assert.Equal(t, 4999.00, domainItem.Price)
		assert.Equal(t, categoryID, domainItem.CategoryID)
		assert.Equal(t, item.StatusDraft, domainItem.Status) // New items start as draft
		assert.Equal(t, 100, *domainItem.Stock)
		assert.Equal(t, 4000.00, *domainItem.CostPrice)
		assert.Equal(t, "PHONE-001", *domainItem.SKU)

		// 验证服务字段为空（产品不应该有服务字段）
		assert.Nil(t, domainItem.Duration)
		assert.Nil(t, domainItem.StaffRequired)
		assert.Nil(t, domainItem.Capacity)
	})

	t.Run("CreateService", func(t *testing.T) {
		// 首先创建分类
		categoryID := uuid.New()
		err := db.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "服务分类", "各种服务", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		// 创建服务商品请求
		duration := 60
		staffRequired := true
		capacity := 1
		req := &service.CreateItemRequest{
			Name:          "理发服务",
			Description:    "专业理发服务",
			Type:          item.ItemTypeService,
			Price:         128.00,
			CategoryID:    categoryID,
			Duration:      &duration,
			StaffRequired: &staffRequired,
			Capacity:      &capacity,
		}

		// 执行创建
		domainItem, err := itemService.CreateItem(ctx, req)
		require.NoError(t, err, "Failed to create service")

		// 验证创建结果
		assert.NotEqual(t, uuid.Nil, domainItem.ID)
		assert.Equal(t, "理发服务", domainItem.Name)
		assert.Equal(t, item.ItemTypeService, domainItem.Type)
		assert.Equal(t, 128.00, domainItem.Price)
		assert.Equal(t, categoryID, domainItem.CategoryID)
		assert.Equal(t, item.StatusDraft, domainItem.Status) // New items start as draft
		assert.Equal(t, 60, *domainItem.Duration)
		assert.True(t, *domainItem.StaffRequired)
		assert.Equal(t, 1, *domainItem.Capacity)

		// 验证产品字段为空（服务不应该有产品字段）
		assert.Nil(t, domainItem.Stock)
		assert.Nil(t, domainItem.CostPrice)
		assert.Nil(t, domainItem.Weight)
		assert.Nil(t, domainItem.SKU)
	})

	t.Run("UpdateProduct", func(t *testing.T) {
		// 创建产品
		categoryID := uuid.New()
		err := db.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "测试分类2", "用于更新的测试分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		productItem := createTestProduct(ctx, t, itemService, db, categoryID)

		// 更新请求
		newPrice := 4599.00
		newStock := 80
		req := &service.UpdateItemRequest{
			Price:   &newPrice,
			Stock:   toFloat64Ptr(newStock),
		}

		// 执行更新
		updatedItem, err := itemService.UpdateItem(ctx, productItem.ID, req)
		require.NoError(t, err, "Failed to update product")

		// 验证更新结果
		assert.Equal(t, productItem.ID, updatedItem.ID)
		assert.Equal(t, productItem.Name, updatedItem.Name)
		assert.Equal(t, 4599.00, updatedItem.Price)
		assert.Equal(t, 80, *updatedItem.Stock)
		assert.True(t, updatedItem.UpdatedAt.After(productItem.UpdatedAt))
	})

	t.Run("UpdateItemStatus", func(t *testing.T) {
		// 创建草稿商品
		categoryID := uuid.New()
		err := db.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "草稿分类", "用于草稿测试的分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		draftItem := createDraftItem(ctx, t, itemService, db, categoryID)

		// 激活商品
		activatedItem, err := itemService.UpdateItemStatus(ctx, draftItem.ID, item.StatusActive)
		require.NoError(t, err, "Failed to activate item")

		// 验证状态更新
		assert.Equal(t, item.StatusActive, activatedItem.Status)

		// 停用商品
		deactivatedItem, err := itemService.UpdateItemStatus(ctx, activatedItem.ID, item.StatusInactive)
		require.NoError(t, err, "Failed to deactivate item")

		// 验证状态更新
		assert.Equal(t, item.StatusInactive, deactivatedItem.Status)
	})

	t.Run("UpdateStock", func(t *testing.T) {
		// 创建产品
		categoryID := uuid.New()
		err := db.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "库存分类", "用于库存测试的分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		productItem := createTestProduct(ctx, t, itemService, db, categoryID)

		// 更新库存
		newStock := 50
		updatedItem, err := itemService.UpdateStock(ctx, productItem.ID, newStock)
		require.NoError(t, err, "Failed to update stock")

		// 验证库存更新
		assert.Equal(t, newStock, *updatedItem.Stock)
	})

	t.Run("UpdateStock_ServiceError", func(t *testing.T) {
		// 创建服务
		categoryID := uuid.New()
		err := db.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "服务分类2", "用于服务测试的分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		serviceItem := createTestService(ctx, t, itemService, db, categoryID)

		// 尝试更新服务库存（应该失败）
		_, err = itemService.UpdateStock(ctx, serviceItem.ID, 100)
		assert.Error(t, err, "Should not be able to update stock for service")
		assert.Contains(t, err.Error(), "only applicable to products")
	})

	t.Run("ListItems", func(t *testing.T) {
		// 创建多个商品
		categoryID := uuid.New()
		err := db.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "列表分类", "用于列表测试的分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		createTestProduct(ctx, t, itemService, db, categoryID)
		createTestService(ctx, t, itemService, db, categoryID)

		// 列出商品
		req := &service.ListItemsRequest{
			Limit:  10,
			Offset: 0,
		}

		items, total, err := itemService.ListItems(ctx, req)
		require.NoError(t, err, "Failed to list items")
		assert.Greater(t, len(items), 0)
		assert.Greater(t, total, int64(0))
	})

	t.Run("SearchItems", func(t *testing.T) {
		// 创建测试商品
		categoryID := uuid.New()
		err := db.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "搜索分类", "用于搜索测试的分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		item := createTestProduct(ctx, t, itemService, db, categoryID)

		// 搜索测试
		req := &service.SearchItemsRequest{
			Query: item.Name,
			Limit: 10,
		}

		items, err := itemService.SearchItems(ctx, req)
		require.NoError(t, err, "Failed to search items")
		assert.Greater(t, len(items), 0)

		// 验证搜索结果包含创建的商品
		found := false
		for _, result := range items {
			if result.ID == item.ID {
				found = true
				break
				}
		}
		assert.True(t, found, "Search results should contain the created item")
	})

	t.Run("GetItemsStats", func(t *testing.T) {
		// 创建多个测试商品
		categoryID := uuid.New()
		err := db.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "统计分类", "用于统计测试的分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		createTestProduct(ctx, t, itemService, db, categoryID)
		createTestService(ctx, t, itemService, db, categoryID)

		// 获取统计
		stats, err := itemService.GetItemsStats(ctx)
		require.NoError(t, err, "Failed to get items stats")

		assert.Greater(t, stats.TotalItems, int64(0))
		assert.GreaterOrEqual(t, stats.ActiveItems, int64(0))
		assert.Contains(t, stats.TypeDistribution, item.ItemTypeProduct)
		assert.Contains(t, stats.TypeDistribution, item.ItemTypeService)
	})
}

// createTestProduct 创建测试产品
func createTestProduct(ctx context.Context, t *testing.T, itemService service.ItemService, db *gorm.DB, categoryID uuid.UUID) *item.Item {
	req := &service.CreateItemRequest{
		Name:        "测试产品",
		Description: "这是一个测试产品",
		Type:        item.ItemTypeProduct,
		Price:       100.00,
		CategoryID:  categoryID,
		Stock:       toFloat64Ptr(50),
		CostPrice:   toFloat64Ptr(80.00),
		SKU:         toStringPtr("TEST-PROD-001"),
	}

	domainItem, err := itemService.CreateItem(ctx, req)
	require.NoError(t, err, "Failed to create test product")
	return domainItem
}

// createTestService 创建测试服务
func createTestService(ctx context.Context, t *testing.T, itemService service.ItemService, db *gorm.DB, categoryID uuid.UUID) *item.Item {
	duration := 30
	req := &service.CreateItemRequest{
		Name:          "测试服务",
		Description:    "这是一个测试服务",
		Type:          item.ItemTypeService,
		Price:         50.00,
		CategoryID:    categoryID,
		Duration:      &duration,
		StaffRequired: toBoolPtr(true),
		Capacity:      toIntPtr(1),
	}

	domainItem, err := itemService.CreateItem(ctx, req)
	require.NoError(t, err, "Failed to create test service")
	return domainItem
}

// createDraftItem 创建草稿商品
func createDraftItem(ctx context.Context, t *testing.T, itemService service.ItemService, db *gorm.DB, categoryID uuid.UUID) *item.Item {
	req := &service.CreateItemRequest{
		Name:        "草稿商品",
		Description: "这是一个草稿商品",
		Type:        item.ItemTypeProduct,
		Price:       200.00,
		CategoryID:  categoryID,
		Stock:       toFloat64Ptr(30),
		SKU:         toStringPtr("DRAFT-001"),
	}

	domainItem, err := itemService.CreateItem(ctx, req)
	require.NoError(t, err, "Failed to create draft item")
	return domainItem
}

// 辅助函数
func toFloat64Ptr(v int) *float64 {
	f := float64(v)
	return &f
}

func toStringPtr(v string) *string {
	return &v
}

func toBoolPtr(v bool) *bool {
	return &v
}

func toIntPtr(v int) *int {
	return &v
}