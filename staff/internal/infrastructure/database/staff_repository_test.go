package database

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/staff/internal/domain/entity"
	"github.com/julesChu12/fly/staff/internal/domain/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&entity.Staff{}, &entity.StaffRole{}, &entity.StaffAvailability{})
	require.NoError(t, err)

	// 清理表数据确保测试隔离
	db.Exec("DELETE FROM staff_availability")
	db.Exec("DELETE FROM staff_roles")
	db.Exec("DELETE FROM staff")

	return db
}

func TestStaffRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewStaffRepository(db)

	roleID := uuid.New()
	staff := &entity.Staff{
		ID:       uuid.New(),
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Phone:    "1234567890",
		Position: "Developer",
		RoleID:   roleID,
		Status:   entity.StaffStatusActive,
	}

	err := repo.Create(staff)
	assert.NoError(t, err)

	// 验证创建的记录
	var saved entity.Staff
	err = db.First(&saved, "id = ?", staff.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, staff.Name, saved.Name)
	assert.Equal(t, staff.Email, saved.Email)
}

func TestStaffRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewStaffRepository(db)

	// 创建测试数据
	roleID := uuid.New()
	staff := &entity.Staff{
		ID:     uuid.New(),
		Name:   "John Doe",
		Email:  "john.doe@example.com",
		RoleID: roleID,
		Status: entity.StaffStatusActive,
	}
	_ = repo.Create(staff)

	// 测试获取
	found, err := repo.GetByID(staff.ID.String())
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, staff.Name, found.Name)
}

func TestStaffRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewStaffRepository(db)

	roleID := uuid.New()
	staff := &entity.Staff{
		ID:     uuid.New(),
		Name:   "John Doe",
		Email:  "john.doe@example.com",
		RoleID: roleID,
		Status: entity.StaffStatusActive,
	}
	_ = repo.Create(staff)

	// 更新状态
	staff.Status = entity.StaffStatusInactive
	err := repo.Update(staff)
	assert.NoError(t, err)

	// 验证更新
	found, _ := repo.GetByID(staff.ID.String())
	assert.Equal(t, entity.StaffStatusInactive, found.Status)
}

func TestStaffRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewStaffRepository(db)

	// 创建多个测试数据
	roleID := uuid.New()
	for i := 0; i < 5; i++ {
		staff := &entity.Staff{
			ID:     uuid.New(),
			Name:   "John Doe",
			Email:  fmt.Sprintf("john.doe%d@example.com", i),
			RoleID: roleID,
			Status: entity.StaffStatusActive,
		}
		_ = repo.Create(staff)
	}

	// 测试列表查询
	filter := &dto.StaffFilter{
		Page:  1,
		Limit: 10,
	}
	staffs, err := repo.List(filter)
	assert.NoError(t, err)
	assert.Len(t, staffs, 5)
}