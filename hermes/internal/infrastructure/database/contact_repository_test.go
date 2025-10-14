package database

import (
	"context"
	"testing"

	"github.com/julesChu12/fly/hermes/internal/domain/entity"
	"github.com/julesChu12/fly/hermes/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContactRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContactRepository(db)
	ctx := context.Background()

	// 先创建客户
	customer := &entity.Customer{TenantID: 1, Name: "Customer", Email: "test@example.com"}
	require.NoError(t, db.Create(customer).Error)

	contact := &entity.Contact{
		TenantID:   1,
		CustomerID: customer.ID,
		Type:       "phone",
		Value:      "1234567890",
		IsPrimary:  true,
	}

	err := repo.Create(ctx, contact)
	assert.NoError(t, err)
	assert.NotZero(t, contact.ID)
	assert.NotZero(t, contact.CreatedAt)
}

func TestContactRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContactRepository(db)
	ctx := context.Background()

	// 创建测试数据
	customer := &entity.Customer{TenantID: 1, Name: "Customer", Email: "test@example.com"}
	require.NoError(t, db.Create(customer).Error)

	contact := &entity.Contact{
		TenantID:   1,
		CustomerID: customer.ID,
		Type:       "email",
		Value:      "contact@example.com",
		IsPrimary:  true,
	}
	require.NoError(t, repo.Create(ctx, contact))

	// 测试获取
	found, err := repo.GetByID(ctx, contact.ID)
	assert.NoError(t, err)
	assert.Equal(t, contact.ID, found.ID)
	assert.Equal(t, contact.Type, found.Type)
	assert.Equal(t, contact.Value, found.Value)
	assert.True(t, found.IsPrimary)
}

func TestContactRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContactRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 99999)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrContactNotFound, err)
}

func TestContactRepository_GetByIDAndTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContactRepository(db)
	ctx := context.Background()

	// 创建两个不同租户的联系方式
	customer1 := &entity.Customer{TenantID: 1, Name: "C1", Email: "c1@example.com"}
	customer2 := &entity.Customer{TenantID: 2, Name: "C2", Email: "c2@example.com"}
	require.NoError(t, db.Create(customer1).Error)
	require.NoError(t, db.Create(customer2).Error)

	contact1 := &entity.Contact{TenantID: 1, CustomerID: customer1.ID, Type: "phone", Value: "111"}
	contact2 := &entity.Contact{TenantID: 2, CustomerID: customer2.ID, Type: "phone", Value: "222"}
	require.NoError(t, repo.Create(ctx, contact1))
	require.NoError(t, repo.Create(ctx, contact2))

	// 测试租户隔离
	found, err := repo.GetByIDAndTenant(ctx, contact1.ID, 1)
	assert.NoError(t, err)
	assert.Equal(t, contact1.ID, found.ID)

	// 尝试用错误的租户ID访问
	_, err = repo.GetByIDAndTenant(ctx, contact1.ID, 2)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrContactNotFound, err)
}

func TestContactRepository_GetByCustomerID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContactRepository(db)
	ctx := context.Background()

	// 创建客户
	customer := &entity.Customer{TenantID: 1, Name: "Customer", Email: "test@example.com"}
	require.NoError(t, db.Create(customer).Error)

	// 创建多个联系方式
	contacts := []*entity.Contact{
		{TenantID: 1, CustomerID: customer.ID, Type: "phone", Value: "111"},
		{TenantID: 1, CustomerID: customer.ID, Type: "email", Value: "test@example.com"},
		{TenantID: 1, CustomerID: customer.ID, Type: "address", Value: "123 Main St"},
	}
	for _, c := range contacts {
		require.NoError(t, repo.Create(ctx, c))
	}

	// 获取客户的所有联系方式
	found, err := repo.GetByCustomerID(ctx, customer.ID)
	assert.NoError(t, err)
	assert.Len(t, found, 3)
}

func TestContactRepository_GetByCustomerIDAndTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContactRepository(db)
	ctx := context.Background()

	// 创建两个租户的客户
	customer1 := &entity.Customer{TenantID: 1, Name: "C1", Email: "c1@example.com"}
	customer2 := &entity.Customer{TenantID: 2, Name: "C2", Email: "c2@example.com"}
	require.NoError(t, db.Create(customer1).Error)
	require.NoError(t, db.Create(customer2).Error)

	// 为每个客户创建联系方式
	contact1 := &entity.Contact{TenantID: 1, CustomerID: customer1.ID, Type: "phone", Value: "111"}
	contact2 := &entity.Contact{TenantID: 2, CustomerID: customer2.ID, Type: "phone", Value: "222"}
	require.NoError(t, repo.Create(ctx, contact1))
	require.NoError(t, repo.Create(ctx, contact2))

	// 测试租户隔离
	found, err := repo.GetByCustomerIDAndTenant(ctx, customer1.ID, 1)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, uint(1), found[0].TenantID)

	// 用错误的租户ID查询
	found, err = repo.GetByCustomerIDAndTenant(ctx, customer1.ID, 2)
	assert.NoError(t, err)
	assert.Len(t, found, 0) // 应该返回空列表
}

func TestContactRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContactRepository(db)
	ctx := context.Background()

	customer := &entity.Customer{TenantID: 1, Name: "Customer", Email: "test@example.com"}
	require.NoError(t, db.Create(customer).Error)

	contact := &entity.Contact{
		TenantID:   1,
		CustomerID: customer.ID,
		Type:       "phone",
		Value:      "111",
		IsPrimary:  false,
	}
	require.NoError(t, repo.Create(ctx, contact))

	// 更新
	contact.Value = "999"
	contact.IsPrimary = true
	err := repo.Update(ctx, contact)
	assert.NoError(t, err)

	// 验证更新
	found, err := repo.GetByID(ctx, contact.ID)
	assert.NoError(t, err)
	assert.Equal(t, "999", found.Value)
	assert.True(t, found.IsPrimary)
}

func TestContactRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContactRepository(db)
	ctx := context.Background()

	customer := &entity.Customer{TenantID: 1, Name: "Customer", Email: "test@example.com"}
	require.NoError(t, db.Create(customer).Error)

	contact := &entity.Contact{TenantID: 1, CustomerID: customer.ID, Type: "phone", Value: "111"}
	require.NoError(t, repo.Create(ctx, contact))

	// 删除
	err := repo.Delete(ctx, contact.ID)
	assert.NoError(t, err)

	// 验证已删除
	_, err = repo.GetByID(ctx, contact.ID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrContactNotFound, err)
}

func TestContactRepository_DeleteByTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContactRepository(db)
	ctx := context.Background()

	customer := &entity.Customer{TenantID: 1, Name: "Customer", Email: "test@example.com"}
	require.NoError(t, db.Create(customer).Error)

	contact := &entity.Contact{TenantID: 1, CustomerID: customer.ID, Type: "phone", Value: "111"}
	require.NoError(t, repo.Create(ctx, contact))

	// 尝试用错误的租户ID删除
	err := repo.DeleteByTenant(ctx, contact.ID, 999)
	assert.NoError(t, err) // GORM 不会报错

	// 验证未被删除
	found, err := repo.GetByID(ctx, contact.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)

	// 用正确的租户ID删除
	err = repo.DeleteByTenant(ctx, contact.ID, 1)
	assert.NoError(t, err)

	// 验证已删除
	_, err = repo.GetByID(ctx, contact.ID)
	assert.Error(t, err)
}

func TestContactRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContactRepository(db)
	ctx := context.Background()

	customer := &entity.Customer{TenantID: 1, Name: "Customer", Email: "test@example.com"}
	require.NoError(t, db.Create(customer).Error)

	// 创建多个联系方式
	for i := 1; i <= 15; i++ {
		contact := &entity.Contact{
			TenantID:   1,
			CustomerID: customer.ID,
			Type:       "phone",
			Value:      string(rune(i)),
		}
		require.NoError(t, repo.Create(ctx, contact))
	}

	// 测试分页
	contacts, err := repo.List(ctx, 0, 10)
	assert.NoError(t, err)
	assert.Len(t, contacts, 10)

	// 第二页
	contacts, err = repo.List(ctx, 10, 10)
	assert.NoError(t, err)
	assert.Len(t, contacts, 5)
}

func TestContactRepository_ListByTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewContactRepository(db)
	ctx := context.Background()

	// 创建两个租户的数据
	customer1 := &entity.Customer{TenantID: 1, Name: "C1", Email: "c1@example.com"}
	customer2 := &entity.Customer{TenantID: 2, Name: "C2", Email: "c2@example.com"}
	require.NoError(t, db.Create(customer1).Error)
	require.NoError(t, db.Create(customer2).Error)

	// 为每个租户创建联系方式
	for i := 1; i <= 5; i++ {
		c1 := &entity.Contact{TenantID: 1, CustomerID: customer1.ID, Type: "phone", Value: "1-" + string(rune(i))}
		c2 := &entity.Contact{TenantID: 2, CustomerID: customer2.ID, Type: "phone", Value: "2-" + string(rune(i))}
		require.NoError(t, repo.Create(ctx, c1))
		require.NoError(t, repo.Create(ctx, c2))
	}

	// 查询租户1的联系方式
	contacts, err := repo.ListByTenant(ctx, 1, 0, 10)
	assert.NoError(t, err)
	assert.Len(t, contacts, 5)

	// 验证都是租户1的数据
	for _, c := range contacts {
		assert.Equal(t, uint(1), c.TenantID)
	}
}

func TestContactRepository_CascadeDelete(t *testing.T) {
	db := setupTestDB(t)
	customerRepo := NewCustomerRepository(db)
	contactRepo := NewContactRepository(db)
	ctx := context.Background()

	// 创建客户和联系方式
	customer := &entity.Customer{TenantID: 1, Name: "Customer", Email: "test@example.com"}
	require.NoError(t, customerRepo.Create(ctx, customer))

	contact := &entity.Contact{TenantID: 1, CustomerID: customer.ID, Type: "phone", Value: "111"}
	require.NoError(t, contactRepo.Create(ctx, contact))

	// 删除客户
	err := customerRepo.Delete(ctx, customer.ID)
	assert.NoError(t, err)

	// 验证联系方式也被删除（级联删除）
	_, err = contactRepo.GetByID(ctx, contact.ID)
	assert.Error(t, err)
}
