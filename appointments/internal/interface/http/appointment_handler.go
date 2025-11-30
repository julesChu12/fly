package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/appointments/internal/application/service"
	"github.com/julesChu12/fly/appointments/internal/domain/appointment"
	"github.com/julesChu12/fly/appointments/internal/domain/dto"
)

// AppointmentHandler 预约HTTP处理器
type AppointmentHandler struct {
	appointmentService service.AppointmentService
}

// NewAppointmentHandler 创建预约HTTP处理器
func NewAppointmentHandler(appointmentService service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{
		appointmentService: appointmentService,
	}
}

// RegisterRoutes 注册路由
func (h *AppointmentHandler) RegisterRoutes(router *gin.RouterGroup) {
	appointments := router.Group("/appointments")
	{
		appointments.GET("", h.ListAppointments)         // 获取预约列表
		appointments.POST("", h.CreateAppointment)       // 创建预约
		appointments.GET("/:id", h.GetAppointment)       // 获取预约详情
		appointments.PUT("/:id", h.UpdateAppointment)    // 更新预约
		appointments.DELETE("/:id", h.DeleteAppointment) // 删除预约
		appointments.PUT("/:id/status", h.UpdateStatus)  // 更新预约状态

		// 日历和可用时间相关
		appointments.GET("/calendar", h.GetCalendarView)       // 获取日历视图
		appointments.GET("/availability", h.CheckAvailability) // 检查可用时间
		appointments.POST("/conflict-check", h.CheckConflict)  // 检查时间冲突

		// 客户和员工相关
		appointments.GET("/customer/:customerId", h.GetAppointmentsByCustomer)
		appointments.GET("/staff/:staffId", h.GetAppointmentsByEmployee)
		appointments.GET("/staff/:staffId/upcoming", h.GetUpcomingAppointments)
	}
}

// ListAppointments godoc
// @Summary 获取预约列表
// @Description 获取预约列表，支持分页和过滤
// @Tags 预约管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param customer_id query string false "客户ID"
// @Param staff_id query string false "员工ID"
// @Param service_id query string false "服务ID"
// @Param status query string false "��约状态"
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Param sort query string false "排序字段" default("start_time")
// @Param order query string false "排序方向" default("asc")
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments [get]
func (h *AppointmentHandler) ListAppointments(c *gin.Context) {
	filter := &dto.AppointmentFilter{}

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		filter.Page = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
		filter.Limit = limit
	}

	if customerID := c.Query("customer_id"); customerID != "" {
		filter.CustomerID = &customerID
	}
	if staffID := c.Query("staff_id"); staffID != "" {
		filter.StaffID = &staffID
	}
	if serviceID := c.Query("service_id"); serviceID != "" {
		filter.ServiceID = &serviceID
	}
	if status := c.Query("status"); status != "" {
		statusEntity := appointment.AppointmentStatus(status)
		filter.Status = &statusEntity
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}
	if sort := c.Query("sort"); sort != "" {
		filter.Sort = sort
	}
	if order := c.Query("order"); order != "" {
		filter.Order = order
	}

	appointments, total, err := h.appointmentService.ListAppointments(filter)
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
		"data": map[string]interface{}{
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
// @Tags 预约管理
// @Accept json
// @Produce json
// @Param appointment body dto.CreateAppointmentRequest true "预约信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /appointments [post]
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	var req dto.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	appointment, err := h.appointmentService.CreateAppointment(&req)
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
// @Tags 预约管理
// @Accept json
// @Produce json
// @Param id path string true "预约ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
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

	appointment, err := h.appointmentService.GetAppointmentByID(id)
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
// @Description 更新预约信息
// @Tags 预约管理
// @Accept json
// @Produce json
// @Param id path string true "预约ID"
// @Param appointment body dto.UpdateAppointmentRequest true "预约信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
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

	var req dto.UpdateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	appointment, err := h.appointmentService.UpdateAppointment(id, &req)
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
// @Description 删除预约
// @Tags 预约管理
// @Accept json
// @Produce json
// @Param id path string true "预约ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
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

	if err := h.appointmentService.DeleteAppointment(id); err != nil {
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
// @Tags 预约管理
// @Accept json
// @Produce json
// @Param id path string true "预约ID"
// @Param status body dto.UpdateStatusRequest true "状态信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
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

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	appointment, err := h.appointmentService.UpdateAppointmentStatus(id, &req)
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
		"message": "更新成功",
		"data":    appointment,
	})
}

// GetCalendarView godoc
// @Summary 获取日历视图
// @Description 获取日历视图
// @Tags 预约管理
// @Accept json
// @Produce json
// @Param staff_id query string false "员工ID"
// @Param start_date query string true "开始日期"
// @Param end_date query string true "结束日期"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /appointments/calendar [get]
func (h *AppointmentHandler) GetCalendarView(c *gin.Context) {
	var req dto.CalendarViewRequest

	// 解析查询参数
	if staffID := c.Query("staff_id"); staffID != "" {
		req.StaffID = &staffID
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = startDate
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "开始日期格式错误",
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始日期不能为空",
		})
		return
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = endDate
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "结束日期格式错误",
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束日期不能为空",
		})
		return
	}

	events, err := h.appointmentService.GetCalendarView(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取日历视图失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    events,
	})
}

// CheckAvailability godoc
// @Summary 检查可用时间
// @Description 检查员工在指定日期的可用时间
// @Tags 预约管理
// @Accept json
// @Produce json
// @Param staff_id query string true "员工ID"
// @Param date query string true "日期"
// @Param service_duration query int true "服务时长(分钟)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /appointments/availability [get]
func (h *AppointmentHandler) CheckAvailability(c *gin.Context) {
	staffID := c.Query("staff_id")
	if staffID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "日期不能为空",
		})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "日期格式错误",
		})
		return
	}

	durationStr := c.Query("service_duration")
	if durationStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "服务时长不能为空",
		})
		return
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "服务时长必须为正整数",
		})
		return
	}

	req := &dto.AvailabilityRequest{
		StaffID:         staffID,
		Date:            date,
		ServiceDuration: time.Duration(duration) * time.Minute,
	}

	response, err := h.appointmentService.CheckAvailability(req)
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
		"message": "获取成功",
		"data":    response,
	})
}

// CheckConflict godoc
// @Summary 检查时间冲突
// @Description 检查预约时间冲突
// @Tags 预约管理
// @Accept json
// @Produce json
// @Param conflict body dto.ConflictCheckRequest true "冲突检查信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /appointments/conflict-check [post]
func (h *AppointmentHandler) CheckConflict(c *gin.Context) {
	var req dto.ConflictCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	conflictInfo, err := h.appointmentService.CheckConflict(&req)
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
// @Summary 获取客户预约
// @Description 根据客户ID获取预约列表
// @Tags 预约管理
// @Accept json
// @Produce json
// @Param customerId path string true "客户ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
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

	filter := &dto.AppointmentFilter{}
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		filter.Page = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
		filter.Limit = limit
	}

	appointments, err := h.appointmentService.GetAppointmentsByCustomerID(customerID, filter)
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
// @Summary 获取员工预约
// @Description 根据员工ID获取预约列表
// @Tags 预约管理
// @Accept json
// @Produce json
// @Param staffId path string true "员工ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /appointments/staff/{staffId} [get]
func (h *AppointmentHandler) GetAppointmentsByEmployee(c *gin.Context) {
	staffID := c.Param("staffId")
	if staffID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	filter := &dto.AppointmentFilter{}
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		filter.Page = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
		filter.Limit = limit
	}

	appointments, err := h.appointmentService.GetAppointmentsByStaffID(staffID, filter)
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

// GetUpcomingAppointments godoc
// @Summary 获取即将到来的预约
// @Description 获取员工即将到来的预约
// @Tags 预约管理
// @Accept json
// @Produce json
// @Param staffId path string true "员工ID"
// @Param limit query int false "数量限制" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /appointments/staff/{staffId}/upcoming [get]
func (h *AppointmentHandler) GetUpcomingAppointments(c *gin.Context) {
	staffID := c.Param("staffId")
	if staffID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	appointments, err := h.appointmentService.GetUpcomingAppointments(staffID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取即将到来的预约失败",
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
