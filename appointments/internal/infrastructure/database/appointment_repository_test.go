package database

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/appointments/internal/domain/entity"
	"github.com/julesChu12/fly/appointments/internal/domain/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&entity.Appointment{})
	require.NoError(t, err)

	// 清理表数据确保测试隔离
	db.Exec("DELETE FROM appointments")

	return db
}

func TestAppointmentRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)

	now := time.Now()
	appointment := &entity.Appointment{
		ID:         uuid.New(),
		CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		StaffID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		ServiceID:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
		Status:     entity.AppointmentStatusPending,
		Notes:      stringPtr("Test appointment"),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err := repo.Create(appointment)
	assert.NoError(t, err)

	// 验证创建的记录
	var saved entity.Appointment
	err = db.First(&saved, "id = ?", appointment.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, appointment.CustomerID, saved.CustomerID)
	assert.Equal(t, appointment.Status, saved.Status)
}

func TestAppointmentRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)

	// 创建测试数据
	appointment := &entity.Appointment{
		ID:         uuid.New(),
		CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		StaffID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		ServiceID:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		Status:     entity.AppointmentStatusConfirmed,
	}
	_ = repo.Create(appointment)

	// 测试获取
	found, err := repo.GetByID(appointment.ID.String())
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, appointment.CustomerID, found.CustomerID)
}

func TestAppointmentRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)

	appointment := &entity.Appointment{
		ID:         uuid.New(),
		CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		StaffID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		ServiceID:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		Status:     entity.AppointmentStatusPending,
	}
	_ = repo.Create(appointment)

	// 更新状态
	appointment.Status = entity.AppointmentStatusConfirmed
	err := repo.Update(appointment)
	assert.NoError(t, err)

	// 验证更新
	found, _ := repo.GetByID(appointment.ID.String())
	assert.Equal(t, entity.AppointmentStatusConfirmed, found.Status)
}

func TestAppointmentRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)

	appointment := &entity.Appointment{
		ID:         uuid.New(),
		CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		StaffID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		ServiceID:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
	}
	_ = repo.Create(appointment)

	err := repo.Delete(appointment.ID.String())
	assert.NoError(t, err)

	// 验证删除
	found, _ := repo.GetByID(appointment.ID.String())
	assert.Nil(t, found) // 应该返回nil
}

func TestAppointmentRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)

	// 创建多个测试数据
	now := time.Now()
	for i := 0; i < 5; i++ {
		appointment := &entity.Appointment{
			ID:         uuid.New(),
			CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			StaffID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			ServiceID:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
			Status:     entity.AppointmentStatusConfirmed,
			StartTime:  now.Add(time.Duration(i) * time.Hour),
			EndTime:    now.Add(time.Duration(i) * time.Hour).Add(time.Hour),
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		_ = repo.Create(appointment)
	}

	// 测试列表查询
	customerID := "550e8400-e29b-41d4-a716-446655440001"
	filter := &dto.AppointmentFilter{
		CustomerID: &customerID,
		Status:     &[]entity.AppointmentStatus{entity.AppointmentStatusConfirmed}[0],
		Page:       1,
		Limit:      3,
		Sort:       "start_time",
		Order:      "asc",
	}

	appointments, err := repo.List(filter)
	assert.NoError(t, err)
	assert.Len(t, appointments, 3) // 应该返回3条记录
}

func TestAppointmentRepository_CheckConflict(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)

	now := time.Now().Truncate(time.Minute) // Truncate to avoid second-level precision issues

	// 创建一个现有预约 9:00-10:00
	existing := &entity.Appointment{
		ID:         uuid.New(),
		CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		StaffID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		ServiceID:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
		Status:     entity.AppointmentStatusConfirmed,
	}
	_ = repo.Create(existing)

	employeeID := "550e8400-e29b-41d4-a716-446655440002"
	excludeID := existing.ID.String()

	// 测试时间重叠检测 - 9:30-10:30 应该检测到冲突
	conflicts, err := repo.CheckConflict(employeeID, now.Add(30*time.Minute), now.Add(time.Hour+30*time.Minute), nil)
	assert.NoError(t, err)
	assert.Len(t, conflicts, 1) // 应该检测到冲突

	// 测试排除当前预���的时间 - 9:30-10:30 使用排除ID应该没有冲突
	conflicts, err = repo.CheckConflict(employeeID, now.Add(30*time.Minute), now.Add(time.Hour+30*time.Minute), &excludeID)
	assert.NoError(t, err)
	assert.Empty(t, conflicts) // 排除后应该没有冲突

	// 测试无冲突的时间段 - 11:00-12:00 应该没有冲突
	noConflicts, err := repo.CheckConflict(employeeID, now.Add(2*time.Hour), now.Add(3*time.Hour), nil)
	assert.NoError(t, err)
	assert.Empty(t, noConflicts) // 应该没有冲突
}

func TestAppointmentRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)

	appointment := &entity.Appointment{
		ID:         uuid.New(),
		CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		StaffID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		ServiceID:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		Status:     entity.AppointmentStatusPending,
	}
	_ = repo.Create(appointment)

	err := repo.UpdateStatus(appointment.ID.String(), entity.AppointmentStatusConfirmed)
	assert.NoError(t, err)

	updated, _ := repo.GetByID(appointment.ID.String())
	assert.Equal(t, entity.AppointmentStatusConfirmed, updated.Status)
}

func stringPtr(s string) *string {
	return &s
}