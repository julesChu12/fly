package integration

import (
	"context"
	"testing"

	"github.com/julesChu12/fly/items/internal/infrastructure/persistence/model"
	"github.com/julesChu12/fly/items/internal/infrastructure/persistence/repository"
	"github.com/julesChu12/fly/items/internal/domain/item"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/google/uuid"
)

// TestMinimalIntegration 最小化集成测试 - 验证测试容器和基本连接
func TestMinimalIntegration(t *testing.T) {
	ctx := context.Background()
	container, err := SetupTestContainers(ctx)
	require.NoError(t, err, "Failed to setup test containers")
	defer container.Cleanup(ctx)

	// 获取数据库连接
	db := container.Database

	// 不需要重置数据库，因为migrations已经创建了表结构
	// err = container.ResetDatabase(db)
	// require.NoError(t, err, "Failed to reset database")

	// 创建仓储实例
	itemRepo := repository.NewItemRepository(db)

	t.Run("BasicConnection", func(t *testing.T) {
		// 验证数据库连接
		sqlDB, err := db.DB()
		require.NoError(t, err)
		assert.NotNil(t, sqlDB)

		// 执行简单查询测试
		var result int64
		err = db.Raw("SELECT 1").Scan(&result).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), result)
	})

	t.Run("TableCreation", func(t *testing.T) {
		// 使用GORM自动创建表
		err := db.AutoMigrate(&model.Item{}, &model.Category{})
		require.NoError(t, err, "Failed to auto-migrate tables")

		// 验证表是否创建成功 - 修复scan类型问题
		var tableName string
		err = db.Raw("SHOW TABLES LIKE 'items'").Scan(&tableName).Error
		require.NoError(t, err)
		assert.Equal(t, "items", tableName, "items table should exist")
	})

	t.Run("SimpleCreate", func(t *testing.T) {
		// 首先手动创建一个分类记录
		categoryID := uuid.New()
		err := db.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "测试分类", "用于集成测试的分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		// 创建一个简单的商品，使用刚创建的分类ID
		itemID := uuid.New()
		sku := "SIMPLE-001"
		price := 99.99

		domainItem := &item.Item{
			ID:          itemID,
			Name:        "简单测试商品",
			Description: "这是一个简单的测试商品",
			Type:        item.ItemTypeProduct,
			Price:       price,
			CategoryID:  categoryID,
			Status:      item.StatusActive,
			SKU:         &sku,
		}

		// 执行创建
		err = itemRepo.Create(ctx, domainItem)
		require.NoError(t, err, "Failed to create item")

		// 验证创建结果 - 通过仓储查询
		foundItem, err := itemRepo.GetByID(ctx, itemID)
		require.NoError(t, err, "Failed to get created item")
		require.NotNil(t, foundItem)

		// 验证创建结果
		assert.Equal(t, itemID, foundItem.ID)
		assert.Equal(t, "简单测试商品", foundItem.Name)
		assert.Equal(t, item.ItemTypeProduct, foundItem.Type)
		assert.Equal(t, 99.99, foundItem.Price)
	})

	t.Run("SimpleQuery", func(t *testing.T) {
		// 查询所有商品
		var count int64
		err := db.Model(&model.Item{}).Count(&count).Error
		require.NoError(t, err)
		assert.Greater(t, count, int64(0), "应该有测试数据")
	})
}