package main

import (
	"testing"
	"time"

	"github.com/julesChu12/fly/appointments/internal/application/service"
	"github.com/julesChu12/fly/appointments/internal/domain/dto"
	"github.com/julesChu12/fly/appointments/internal/domain/entity"
	"github.com/julesChu12/fly/appointments/internal/infrastructure/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAppointmentIntegration(t *testing.T) {
	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&entity.Appointment{})
	require.NoError(t, err)

	// 创建仓储和服务
	appointmentRepo := database.NewAppointmentRepository(db)
	appointmentService := service.NewAppointmentService(appointmentRepo)

	
	// 测试创建预约
	now := time.Now()
	req := &dto.CreateAppointmentRequest{
		CustomerID: "550e8400-e29b-41d4-a716-446655440001",
		StaffID: "550e8400-e29b-41d4-a716-446655440002",
		ServiceID:  "550e8400-e29b-41d4-a716-446655440003",
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
		Notes:      stringPtr("Integration test appointment"),
	}

	appointment, err := appointmentService.CreateAppointment(req)
	assert.NoError(t, err)
	assert.NotNil(t, appointment)
	assert.Equal(t, entity.AppointmentStatusPending, appointment.Status)

	// 测试获取预约
	found, err := appointmentService.GetAppointmentByID(appointment.ID.String())
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, appointment.CustomerID, found.CustomerID)

	// 测试更新状态
	updateReq := &dto.UpdateStatusRequest{
		Status: "confirmed",
	}

	updated, err := appointmentService.UpdateAppointmentStatus(appointment.ID.String(), updateReq)
	assert.NoError(t, err)
	assert.Equal(t, entity.AppointmentStatusConfirmed, updated.Status)

	// 测试列表查询
	filter := &dto.AppointmentFilter{
		Page:  1,
		Limit: 10,
	}

	appointments, total, err := appointmentService.ListAppointments(filter)
	assert.NoError(t, err)
	assert.Len(t, appointments, 1)
	assert.Equal(t, int64(1), total)

	// 测试冲突检查
	conflictReq := &dto.ConflictCheckRequest{
		StaffID: "550e8400-e29b-41d4-a716-446655440002",
		StartTime:  now.Add(30 * time.Minute),
		EndTime:    now.Add(time.Hour + 30*time.Minute),
	}

	conflictInfo, err := appointmentService.CheckConflict(conflictReq)
	assert.NoError(t, err)
	assert.True(t, conflictInfo.Conflict)
	assert.Equal(t, 1, conflictInfo.ConflictCount)

	// 测试删除预约
	err = appointmentService.DeleteAppointment(appointment.ID.String())
	assert.NoError(t, err)

	// 验证已删除
	_, err = appointmentService.GetAppointmentByID(appointment.ID.String())
	assert.Error(t, err) // 应该返回错误
}

func stringPtr(s string) *string {
	return &s
}