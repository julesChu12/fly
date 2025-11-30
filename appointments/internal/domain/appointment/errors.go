package appointment

import "errors"

// Common domain errors
var (
	ErrAppointmentNotFound     = errors.New("预约不存在")
	ErrInvalidTimeRange        = errors.New("无效的时间范围")
	ErrTimeConflict            = errors.New("时间段冲突")
	ErrInvalidStatusTransition = errors.New("无效的状态转换")
	ErrPastAppointment         = errors.New("不能创建过去的预约")
	ErrInvalidCustomerID       = errors.New("无效的客户ID")
	ErrInvalidStaffID          = errors.New("无效的员工ID")
	ErrInvalidServiceID        = errors.New("无效的服务ID")
)
