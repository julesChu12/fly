package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/application/usecase"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
)

// StaffHandler handles HTTP requests for staff management
type StaffHandler struct {
	staffProxy *usecase.StaffProxy
}

// NewStaffHandler creates a new StaffHandler instance
func NewStaffHandler(staffProxy *usecase.StaffProxy) *StaffHandler {
	return &StaffHandler{
		staffProxy: staffProxy,
	}
}

// RegisterRoutes registers staff-related routes
func (h *StaffHandler) RegisterRoutes(router *gin.RouterGroup) {
	staff := router.Group("/employees")
	{
		staff.GET("", h.ListStaff)                    // 获取员工列表
		staff.POST("", h.CreateStaff)                  // 创建员工
		staff.GET("/:id", h.GetStaff)                   // 获取员工详情
		staff.PUT("/:id", h.UpdateStaff)                // 更新员工
		staff.DELETE("/:id", h.DeleteStaff)             // 删除员工
		staff.PUT("/:id/status", h.UpdateStatus)         // 更新员工状态
		staff.GET("/available", h.GetAvailableStaff)    // 获取可用员工
	}

	// 角色管理
	roles := staff.Group("/roles")
	{
		roles.GET("", h.ListRoles)                      // 获取角色列表
		roles.POST("", h.CreateRole)                    // 创建角色
	}

	// 可用性管理
	availability := staff.Group("/availability")
	{
		availability.GET("/:staff_id", h.GetAvailability)          // 获取员工可用性
		availability.PUT("/:staff_id", h.SetAvailability)          // 设置员工可用性
	}
}

// ListStaff godoc
// @Summary 获取员工列表
// @Description 获取员工列表，支持分页和过滤
// @Tags 员工管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Param department query string false "部门"
// @Param role_id query string false "角色ID"
// @Param status query string false "员工状态"
// @Param is_available query bool false "是否可用"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /employees [get]
func (h *StaffHandler) ListStaff(c *gin.Context) {
	filter := &client.StaffFilter{}

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		filter.Page = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
		filter.Limit = limit
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}
	if department := c.Query("department"); department != "" {
		filter.Department = &department
	}
	if roleID := c.Query("role_id"); roleID != "" {
		filter.RoleID = &roleID
	}
	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}
	if isAvailable := c.Query("is_available"); isAvailable != "" {
		if available, err := strconv.ParseBool(isAvailable); err == nil {
			filter.IsAvailable = &available
		}
	}

	staff, total, err := h.staffProxy.ListStaff(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取员工列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": map[string]interface{}{
			"employees": staff,
			"total":     total,
			"page":      filter.Page,
			"limit":     filter.Limit,
		},
	})
}

// CreateStaff godoc
// @Summary 创建员工
// @Description 创建新的员工
// @Tags 员工管理
// @Accept json
// @Produce json
// @Param employee body client.CreateStaffRequest true "员工信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /employees [post]
func (h *StaffHandler) CreateStaff(c *gin.Context) {
	var req client.CreateStaffRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	staff, err := h.staffProxy.CreateStaff(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "创建员工失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data":    staff,
	})
}

// GetStaff godoc
// @Summary 获取员工详情
// @Description 根据ID获取员工详情
// @Tags 员工管理
// @Accept json
// @Produce json
// @Param id path string true "员工ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /employees/{id} [get]
func (h *StaffHandler) GetStaff(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	staff, err := h.staffProxy.GetStaff(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "员工不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    staff,
	})
}

// UpdateStaff godoc
// @Summary 更新员工
// @Description 更新员工信息
// @Tags 员工管理
// @Accept json
// @Produce json
// @Param id path string true "员工ID"
// @Param employee body client.UpdateStaffRequest true "员工信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /employees/{id} [put]
func (h *StaffHandler) UpdateStaff(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	var req client.UpdateStaffRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	staff, err := h.staffProxy.UpdateStaff(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "更新员工失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    staff,
	})
}

// DeleteStaff godoc
// @Summary 删除员工
// @Description 删除员工
// @Tags 员工管理
// @Accept json
// @Produce json
// @Param id path string true "员工ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /employees/{id} [delete]
func (h *StaffHandler) DeleteStaff(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	if err := h.staffProxy.DeleteStaff(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "删除员工失败",
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
// @Summary 更新员工状态
// @Description 更新员工状态
// @Tags 员工管理
// @Accept json
// @Produce json
// @Param id path string true "员工ID"
// @Param status body client.UpdateStatusRequest true "状态信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /employees/{id}/status [put]
func (h *StaffHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	var req client.UpdateStatusRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	staff, err := h.staffProxy.UpdateStaffStatus(c.Request.Context(), id, &req)
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
		"data":    staff,
	})
}

// GetAvailableStaff godoc
// @Summary 获取可用员工
// @Description 获取当前可用的员工列表
// @Tags 员工管理
// @Accept json
// @Produce json
// @Param department query string false "部门"
// @Param skills query array false "技能列表"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /employees/available [get]
func (h *StaffHandler) GetAvailableStaff(c *gin.Context) {
	filter := &client.StaffFilter{}

	if department := c.Query("department"); department != "" {
		filter.Department = &department
	}

	if skills := c.QueryArray("skills"); len(skills) > 0 {
		filter.Skills = skills
	}

	staff, err := h.staffProxy.GetAvailableStaff(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取可用员工失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    staff,
	})
}

// ListRoles godoc
// @Summary 获取角色列表
// @Description 获取角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /employees/roles [get]
func (h *StaffHandler) ListRoles(c *gin.Context) {
	roles, err := h.staffProxy.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取角色列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    roles,
	})
}

// CreateRole godoc
// @Summary 创建角色
// @Description 创建新的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param role body map[string]interface{} true "角色信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /employees/roles [post]
func (h *StaffHandler) CreateRole(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	role, err := h.staffProxy.CreateRole(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "创建角色失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data":    role,
	})
}

// GetAvailability godoc
// @Summary 获取员工可用性
// @Description 获取员工的可用性时间安排
// @Tags 可用性管理
// @Accept json
// @Produce json
// @Param staff_id path string true "员工ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /employees/availability/{staff_id} [get]
func (h *StaffHandler) GetAvailability(c *gin.Context) {
	staffID := c.Param("staff_id")
	if staffID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	availability, err := h.staffProxy.GetAvailability(c.Request.Context(), staffID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "获取可用性失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    availability,
	})
}

// SetAvailability godoc
// @Summary 设置员工可用性
// @Description 设置员工的可用性时间安排
// @Tags 可用性管理
// @Accept json
// @Produce json
// @Param staff_id path string true "员工ID"
// @Param availability body client.AvailabilityRequest true "可用性信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /employees/availability/{staff_id} [put]
func (h *StaffHandler) SetAvailability(c *gin.Context) {
	staffID := c.Param("staff_id")
	if staffID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	var req client.AvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	err := h.staffProxy.SetAvailability(c.Request.Context(), staffID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "设置可用性失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "设置成功",
	})
}