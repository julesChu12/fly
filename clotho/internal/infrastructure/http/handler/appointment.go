package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/application/usecase"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
)

// AppointmentHandler handles HTTP requests for appointment management
type AppointmentHandler struct {
	appointmentProxy *usecase.AppointmentProxy
}

// NewAppointmentHandler creates a new AppointmentHandler instance
func NewAppointmentHandler(appointmentProxy *usecase.AppointmentProxy) *AppointmentHandler {
	return &AppointmentHandler{
		appointmentProxy: appointmentProxy,
	}
}

// RegisterRoutes registers appointment-related routes
func (h *AppointmentHandler) RegisterRoutes(router *gin.RouterGroup) {
	appointments := router.Group("/appointments")
	{
		appointments.GET("", h.ListAppointments)                    // 获取预约列表
		appointments.POST("", h.CreateAppointment)                  // 创建预约
		appointments.GET("/:id", h.GetAppointment)                   // 获取预约详情
		appointments.PUT("/:id", h.UpdateAppointment)                // 更新预约
		appointments.DELETE("/:id", h.DeleteAppointment)             // 删除预约
		appointments.DELETE("/batch", h.BatchDeleteAppointments)     // 批量删除预约
		appointments.PUT("/:id/status", h.UpdateStatus)              // 更新预约状态

		// 可用性和冲突检查
		appointments.GET("/availability", h.CheckAvailability)       // 检查可用时间
		appointments.POST("/conflict-check", h.CheckConflict)       // 检查时间冲突

		// 客户和员工相关
		appointments.GET("/customer/:customerId", h.GetAppointmentsByCustomer)
		appointments.GET("/employee/:employeeId", h.GetAppointmentsByEmployee)
	}
}

// ListAppointments godoc
// @Summary 获取预约列表
// @Description 获取预约列表，支持分页和过滤
// @Tags appointments
// @Accept json
// @Produce json
// @Param customer_id query string false "客户ID"
// @Param staff_id query string false "员工ID"
// @Param service_id query string false "服务ID"
// @Param status query string false "预约状态"
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param sort query string false "排序字段" default(start_time)
// @Param order query string false "排序方向" default(asc)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments [get]
func (h *AppointmentHandler) ListAppointments(c *gin.Context) {
	// 构建过滤条件
	filter := &client.AppointmentFilter{
		CustomerID: stringPtr(c.Query("customer_id")),
		StaffID: stringPtr(c.Query("staff_id")),
		ServiceID:  stringPtr(c.Query("service_id")),
		Page:       parseIntQueryParam(c.Query("page"), 1),
		Limit:      parseIntQueryParam(c.Query("limit"), 20),
		Sort:       c.DefaultQuery("sort", "start_time"),
		Order:      c.DefaultQuery("order", "asc"),
	}

	// 解析状态参数
	if statusStr := c.Query("status"); statusStr != "" {
		status := client.AppointmentStatus(statusStr)
		filter.Status = &status
	}

	// 解析日期参数
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	appointments, total, err := h.appointmentProxy.ListAppointments(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取预约列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"appointments": appointments,
			"total":        total,
			"page":         filter.Page,
			"limit":        filter.Limit,
		},
	})
}

// CreateAppointment godoc
// @Summary 创建预约
// @Description 创建新的预约
// @Tags appointments
// @Accept json
// @Produce json
// @Param appointment body client.CreateAppointmentRequestHTTP true "预约信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments [post]
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	var req client.CreateAppointmentRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	appointment, err := h.appointmentProxy.CreateAppointment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "创建预约失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data":    appointment,
	})
}

// GetAppointment godoc
// @Summary 获取预约详情
// @Description 根据ID获取预约详情
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path string true "预约ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments/{id} [get]
func (h *AppointmentHandler) GetAppointment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "预约ID不能为空",
		})
		return
	}

	appointment, err := h.appointmentProxy.GetAppointment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "预约不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    appointment,
	})
}

// UpdateAppointment godoc
// @Summary 更新预约
// @Description 更新现有预约信息
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path string true "预约ID"
// @Param appointment body client.UpdateAppointmentRequestHTTP true "预约信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments/{id} [put]
func (h *AppointmentHandler) UpdateAppointment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "预约ID不能为空",
		})
		return
	}

	var req client.UpdateAppointmentRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	appointment, err := h.appointmentProxy.UpdateAppointment(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "更新预约失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    appointment,
	})
}

// DeleteAppointment godoc
// @Summary 删除预约
// @Description 删除指定预约
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path string true "预约ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments/{id} [delete]
func (h *AppointmentHandler) DeleteAppointment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "预约ID不能为空",
		})
		return
	}

	err := h.appointmentProxy.DeleteAppointment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "删除预约失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// UpdateStatus godoc
// @Summary 更新预约状态
// @Description 更新预约状态
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path string true "预约ID"
// @Param status body client.UpdateStatusRequestHTTP true "状态信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments/{id}/status [put]
func (h *AppointmentHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "预约ID不能为空",
		})
		return
	}

	var req client.UpdateAppointmentStatusRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	appointment, err := h.appointmentProxy.UpdateAppointmentStatus(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "更新状态失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新状态成功",
		"data":    appointment,
	})
}

// CheckAvailability godoc
// @Summary 检查可用时间
// @Description 检查指定员工在指定日期的可用时间
// @Tags appointments
// @Accept json
// @Produce json
// @Param staff_id query string true "员工ID"
// @Param date query string true "日期"
// @Param service_duration query int true "服务时长(秒)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments/availability [get]
func (h *AppointmentHandler) CheckAvailability(c *gin.Context) {
	employeeID := c.Query("staff_id")
	dateStr := c.Query("date")
	serviceDurationStr := c.Query("service_duration")

	if employeeID == "" || dateStr == "" || serviceDurationStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID、日期和服务时长不能为空",
		})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "日期格式错误，请使用 YYYY-MM-DD 格式",
			"error":   err.Error(),
		})
		return
	}

	serviceDuration, err := strconv.ParseInt(serviceDurationStr, 10, 64)
	if err != nil || serviceDuration <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "服务时长必须是正整数",
			"error":   err.Error(),
		})
		return
	}

	req := &client.AppointmentAvailabilityRequest{
		StaffID:      employeeID,
		Date:            date,
		ServiceDuration: time.Duration(serviceDuration) * time.Second,
	}

	availability, err := h.appointmentProxy.CheckAvailability(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "检查可用时间失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "检查成功",
		"data":    availability,
	})
}

// CheckConflict godoc
// @Summary 检查时间冲突
// @Description 检查预约时间是否存在冲突
// @Tags appointments
// @Accept json
// @Produce json
// @Param conflict body client.ConflictCheckRequest true "冲突检查信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments/conflict-check [post]
func (h *AppointmentHandler) CheckConflict(c *gin.Context) {
	var req client.ConflictCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	conflictInfo, err := h.appointmentProxy.CheckConflict(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "检查冲突失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "检查成功",
		"data":    conflictInfo,
	})
}

// GetAppointmentsByCustomer godoc
// @Summary 获取客户预约列表
// @Description 根据客户ID获取预约列表
// @Tags appointments
// @Accept json
// @Produce json
// @Param customerId path string true "客户ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param status query string false "预约状态"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments/customer/{customerId} [get]
func (h *AppointmentHandler) GetAppointmentsByCustomer(c *gin.Context) {
	customerID := c.Param("customerId")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "客户ID不能为空",
		})
		return
	}

	// 构建过滤条件
	filter := &client.AppointmentFilter{
		Page:  parseIntQueryParam(c.Query("page"), 1),
		Limit: parseIntQueryParam(c.Query("limit"), 20),
	}

	// 解析状态参数
	if statusStr := c.Query("status"); statusStr != "" {
		status := client.AppointmentStatus(statusStr)
		filter.Status = &status
	}

	appointments, err := h.appointmentProxy.GetAppointmentsByCustomer(c.Request.Context(), customerID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取客户预约失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    appointments,
	})
}

// GetAppointmentsByEmployee godoc
// @Summary 获取员工预约列表
// @Description 根据员工ID获取预约列表
// @Tags appointments
// @Accept json
// @Produce json
// @Param employeeId path string true "员工ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param status query string false "预约状态"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments/employee/{employeeId} [get]
func (h *AppointmentHandler) GetAppointmentsByEmployee(c *gin.Context) {
	employeeID := c.Param("employeeId")
	if employeeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	// 构建过滤条件
	filter := &client.AppointmentFilter{
		Page:  parseIntQueryParam(c.Query("page"), 1),
		Limit: parseIntQueryParam(c.Query("limit"), 20),
	}

	// 解析状态参数
	if statusStr := c.Query("status"); statusStr != "" {
		status := client.AppointmentStatus(statusStr)
		filter.Status = &status
	}

	appointments, err := h.appointmentProxy.GetAppointmentsByEmployee(c.Request.Context(), employeeID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取员工预约失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    appointments,
	})
}

// BatchDeleteAppointments godoc
// @Summary 批量删除预约
// @Description 批量删除指定的预约
// @Tags appointments
// @Accept json
// @Produce json
// @Param request body map[string][]int true "预约ID列表"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments/batch [delete]
func (h *AppointmentHandler) BatchDeleteAppointments(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "预约ID列表不能为空",
		})
		return
	}

	// 批量删除预约
	err := h.appointmentProxy.BatchDeleteAppointments(c.Request.Context(), req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "批量删除预约失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "批量删除成功",
		"data": gin.H{
			"deleted_count": len(req.IDs),
		},
	})
}

// Helper functions

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseIntQueryParam(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultValue
}
// GetAppointmentStats godoc
// @Summary 获取预约统计信息
// @Description 获取预约相关的统计数据，包括总数、待确认、已完成等
// @Tags appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/appointments/appointments/stats [get]
func (h *AppointmentHandler) GetAppointmentStats(c *gin.Context) {
	var stats interface{} = map[string]interface{}{}
	err := fmt.Errorf("not implemented yet")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取预约统计失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}
