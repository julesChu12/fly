package database

import (
	"context"
	"testing"

	"github.com/julesChu12/fly/hermes/internal/domain/entity"
	"github.com/julesChu12/fly/hermes/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// 使用内存 SQLite 数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 启用外键约束
	db.Exec("PRAGMA foreign_keys = ON")

	// 自动迁移
	err = db.AutoMigrate(&entity.Customer{}, &entity.Contact{})
	require.NoError(t, err)

	return db
}

func TestCustomerRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	customer := &entity.Customer{
		TenantID: 1,
		Name:     "Test Customer",
		Phone:    "1234567890",
		Email:    "test@example.com",
		Tags:     "vip,premium",
	}

	err := repo.Create(ctx, customer)
	assert.NoError(t, err)
	assert.NotZero(t, customer.ID)
	assert.NotZero(t, customer.CreatedAt)
}

func TestCustomerRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	// 创建测试数据
	customer := &entity.Customer{
		TenantID: 1,
		Name:     "Test Customer",
		Email:    "test@example.com",
	}
	err := repo.Create(ctx, customer)
	require.NoError(t, err)

	// 测试获取
	found, err := repo.GetByID(ctx, customer.ID)
	assert.NoError(t, err)
	assert.Equal(t, customer.ID, found.ID)
	assert.Equal(t, customer.Name, found.Name)
	assert.Equal(t, customer.Email, found.Email)
}

func TestCustomerRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 99999)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCustomerNotFound, err)
}

func TestCustomerRepository_GetByIDAndTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	// 创建两个不同租户的客户
	customer1 := &entity.Customer{TenantID: 1, Name: "Customer 1", Email: "c1@example.com"}
	customer2 := &entity.Customer{TenantID: 2, Name: "Customer 2", Email: "c2@example.com"}

	require.NoError(t, repo.Create(ctx, customer1))
	require.NoError(t, repo.Create(ctx, customer2))

	// 测试租户隔离
	found, err := repo.GetByIDAndTenant(ctx, customer1.ID, 1)
	assert.NoError(t, err)
	assert.Equal(t, customer1.ID, found.ID)

	// 尝试用错误的租户ID访问
	_, err = repo.GetByIDAndTenant(ctx, customer1.ID, 2)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCustomerNotFound, err)
}

func TestCustomerRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	customer := &entity.Customer{
		TenantID: 1,
		Name:     "Test Customer",
		Email:    "unique@example.com",
	}
	require.NoError(t, repo.Create(ctx, customer))

	found, err := repo.GetByEmail(ctx, "unique@example.com")
	assert.NoError(t, err)
	assert.Equal(t, customer.ID, found.ID)
	assert.Equal(t, "unique@example.com", found.Email)
}

func TestCustomerRepository_GetByEmailAndTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	// 同一邮箱，不同租户（应该允许）
	customer1 := &entity.Customer{TenantID: 1, Name: "Customer 1", Email: "same@example.com"}
	customer2 := &entity.Customer{TenantID: 2, Name: "Customer 2", Email: "same@example.com"}

	require.NoError(t, repo.Create(ctx, customer1))
	require.NoError(t, repo.Create(ctx, customer2))

	// 按租户查询
	found1, err := repo.GetByEmailAndTenant(ctx, "same@example.com", 1)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), found1.TenantID)

	found2, err := repo.GetByEmailAndTenant(ctx, "same@example.com", 2)
	assert.NoError(t, err)
	assert.Equal(t, uint(2), found2.TenantID)
}

func TestCustomerRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	customer := &entity.Customer{
		TenantID: 1,
		Name:     "Original Name",
		Email:    "original@example.com",
	}
	require.NoError(t, repo.Create(ctx, customer))

	// 更新
	customer.Name = "Updated Name"
	customer.Phone = "9876543210"
	err := repo.Update(ctx, customer)
	assert.NoError(t, err)

	// 验证更新
	found, err := repo.GetByID(ctx, customer.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "9876543210", found.Phone)
}

func TestCustomerRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	customer := &entity.Customer{
		TenantID: 1,
		Name:     "To Delete",
		Email:    "delete@example.com",
	}
	require.NoError(t, repo.Create(ctx, customer))

	// 删除
	err := repo.Delete(ctx, customer.ID)
	assert.NoError(t, err)

	// 验证已删除
	_, err = repo.GetByID(ctx, customer.ID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCustomerNotFound, err)
}

func TestCustomerRepository_DeleteByTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	customer := &entity.Customer{TenantID: 1, Name: "Customer", Email: "test@example.com"}
	require.NoError(t, repo.Create(ctx, customer))

	// 尝试用错误的租户ID删除（应该失败）
	err := repo.DeleteByTenant(ctx, customer.ID, 999)
	assert.NoError(t, err) // GORM 不会报错，但不会删除

	// 验证未被删除
	found, err := repo.GetByID(ctx, customer.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)

	// 用正确的租户ID删除
	err = repo.DeleteByTenant(ctx, customer.ID, 1)
	assert.NoError(t, err)

	// 验证已删除
	_, err = repo.GetByID(ctx, customer.ID)
	assert.Error(t, err)
}

func TestCustomerRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	// 创建多个客户
	for i := 1; i <= 15; i++ {
		customer := &entity.Customer{
			TenantID: 1,
			Name:     "Customer " + string(rune(i)),
			Email:    "customer" + string(rune(i)) + "@example.com",
		}
		require.NoError(t, repo.Create(ctx, customer))
	}

	// 测试分页
	customers, total, err := repo.List(ctx, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, customers, 10)

	// 第二页
	customers, total, err = repo.List(ctx, 10, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, customers, 5)
}

func TestCustomerRepository_ListByTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	// 创建不同租户的客户
	for i := 1; i <= 5; i++ {
		c1 := &entity.Customer{TenantID: 1, Name: "T1 Customer", Email: "t1-" + string(rune(i)) + "@example.com"}
		c2 := &entity.Customer{TenantID: 2, Name: "T2 Customer", Email: "t2-" + string(rune(i)) + "@example.com"}
		require.NoError(t, repo.Create(ctx, c1))
		require.NoError(t, repo.Create(ctx, c2))
	}

	// 查询租户1的客户
	customers, total, err := repo.ListByTenant(ctx, 1, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, customers, 5)

	// 验证都是租户1的数据
	for _, c := range customers {
		assert.Equal(t, uint(1), c.TenantID)
	}
}

func TestCustomerRepository_GetWithContacts(t *testing.T) {
	db := setupTestDB(t)
	customerRepo := NewCustomerRepository(db)
	contactRepo := NewContactRepository(db)
	ctx := context.Background()

	// 创建客户
	customer := &entity.Customer{
		TenantID: 1,
		Name:     "Customer with Contacts",
		Email:    "contacts@example.com",
	}
	require.NoError(t, customerRepo.Create(ctx, customer))

	// 添加联系方式
	contacts := []*entity.Contact{
		{TenantID: 1, CustomerID: customer.ID, Type: "phone", Value: "123456", IsPrimary: true},
		{TenantID: 1, CustomerID: customer.ID, Type: "email", Value: "contact@example.com", IsPrimary: false},
	}
	for _, contact := range contacts {
		require.NoError(t, contactRepo.Create(ctx, contact))
	}

	// 获取客户及联系方式
	found, err := customerRepo.GetWithContacts(ctx, customer.ID)
	assert.NoError(t, err)
	assert.Equal(t, customer.ID, found.ID)
	assert.Len(t, found.Contacts, 2)
}

func TestCustomerRepository_GetWithContactsByTenant(t *testing.T) {
	db := setupTestDB(t)
	customerRepo := NewCustomerRepository(db)
	contactRepo := NewContactRepository(db)
	ctx := context.Background()

	// 创建客户
	customer := &entity.Customer{TenantID: 1, Name: "Customer", Email: "test@example.com"}
	require.NoError(t, customerRepo.Create(ctx, customer))

	// 添加联系方式
	contact := &entity.Contact{TenantID: 1, CustomerID: customer.ID, Type: "phone", Value: "123"}
	require.NoError(t, contactRepo.Create(ctx, contact))

	// 测试租户隔离查询
	found, err := customerRepo.GetWithContactsByTenant(ctx, customer.ID, 1)
	assert.NoError(t, err)
	assert.Len(t, found.Contacts, 1)

	// 用错误的租户ID查询
	_, err = customerRepo.GetWithContactsByTenant(ctx, customer.ID, 999)
	assert.Error(t, err)
}
