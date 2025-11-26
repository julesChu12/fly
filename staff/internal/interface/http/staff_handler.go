package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/staff/internal/application/service"
	"github.com/julesChu12/fly/staff/internal/domain/dto"
	"github.com/julesChu12/fly/staff/internal/domain/entity"
)

// StaffHandler 员工HTTP处理器
type StaffHandler struct {
	staffService service.StaffService
}

// NewStaffHandler 创建员工HTTP处理器
func NewStaffHandler(staffService service.StaffService) *StaffHandler {
	return &StaffHandler{
		staffService: staffService,
	}
}

// RegisterRoutes 注册路由
func (h *StaffHandler) RegisterRoutes(router *gin.RouterGroup) {
	staff := router.Group("/staff")
	{
		staff.GET("", h.ListStaff)                    // 获取员工列表
		staff.POST("", h.CreateStaff)                  // 创建员工
		staff.GET("/:id", h.GetStaff)                    // 获取员工详情
		staff.PUT("/:id", h.UpdateStaff)                 // 更新员工
		staff.DELETE("/:id", h.DeleteStaff)              // 删除员工
		staff.PUT("/:id/status", h.UpdateStatus)          // 更新员工状态
		staff.GET("/available", h.GetAvailableStaff)     // 获取可用员工
		staff.GET("/stats", h.GetStats)                  // 获取统计数据

		// 角色管理
		roles := staff.Group("/roles")
		{
			roles.GET("", h.ListRoles)                    // 获取角色列表
			roles.POST("", h.CreateRole)                  // 创建角色
			roles.GET("/:id", h.GetRole)                     // 获取角色详情
			roles.PUT("/:id", h.UpdateRole)                 // 更新角色
			roles.DELETE("/:id", h.DeleteRole)              // 删除角色
		}

		// 可用性管理
		availability := staff.Group("/availability")
		{
			availability.GET("/:staff_id", h.GetAvailability)              // 获取员工可用性
			availability.PUT("/:staff_id", h.SetAvailability)              // 设置员工可用性
			availability.POST("/search", h.GetAvailableStaffForTime)       // 查询可用员工
		}
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
// @Param sort query string false "排序字段" default("created_at")
// @Param order query string false "排序方向" default("desc")
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /staff [get]
func (h *StaffHandler) ListStaff(c *gin.Context) {
	filter := &dto.StaffFilter{}

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
		statusEntity := entity.StaffStatus(status)
		filter.Status = &statusEntity
	}
	if isAvailable := c.Query("is_available"); isAvailable != "" {
		if available, err := strconv.ParseBool(isAvailable); err == nil {
			filter.IsAvailable = &available
		}
	}
	if sort := c.Query("sort"); sort != "" {
		filter.Sort = sort
	}
	if order := c.Query("order"); order != "" {
		filter.Order = order
	}

	staff, total, err := h.staffService.ListStaff(filter)
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
			"staff": staff,
			"total": total,
			"page":  filter.Page,
			"limit": filter.Limit,
		},
	})
}

// CreateStaff godoc
// @Summary 创建员工
// @Description 创建新的员工
// @Tags 员工管理
// @Accept json
// @Produce json
// @Param staff body dto.CreateStaffRequest true "员工信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /staff [post]
func (h *StaffHandler) CreateStaff(c *gin.Context) {
	var req dto.CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	staff, err := h.staffService.CreateStaff(&req)
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
// @Router /staff/{id} [get]
func (h *StaffHandler) GetStaff(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	staff, err := h.staffService.GetStaffByID(id)
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
// @Param staff body dto.UpdateStaffRequest true "员工信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /staff/{id} [put]
func (h *StaffHandler) UpdateStaff(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	var req dto.UpdateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	staff, err := h.staffService.UpdateStaff(id, &req)
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
// @Router /staff/{id} [delete]
func (h *StaffHandler) DeleteStaff(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	if err := h.staffService.DeleteStaff(id); err != nil {
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
// @Param status body dto.UpdateStatusRequest true "状态信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /staff/{id}/status [put]
func (h *StaffHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
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

	staff, err := h.staffService.UpdateStaffStatus(id, &req)
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
// @Router /staff/available [get]
func (h *StaffHandler) GetAvailableStaff(c *gin.Context) {
	filter := &dto.StaffFilter{}

	if department := c.Query("department"); department != "" {
		filter.Department = &department
	}

	if skills := c.QueryArray("skills"); len(skills) > 0 {
		filter.Skills = skills
	}

	available := true
	filter.IsAvailable = &available

	staff, _, err := h.staffService.ListStaff(filter)
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

// GetStats godoc
// @Summary 获取统计数据
// @Description 获取员工统计数据
// @Tags 员工管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /staff/stats [get]
func (h *StaffHandler) GetStats(c *gin.Context) {
	stats, err := h.staffService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取统计数据失败",
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

// ListRoles godoc
// @Summary 获取角色列表
// @Description 获取角色列表，支持分页和过滤
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Param status query string false "角色状态"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /staff/roles [get]
func (h *StaffHandler) ListRoles(c *gin.Context) {
	filter := &dto.RoleFilter{}

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
	if status := c.Query("status"); status != "" {
		statusEntity := entity.StaffStatus(status)
		filter.Status = &statusEntity
	}

	roles, err := h.staffService.ListRoles(filter)
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
		"data": map[string]interface{}{
			"roles": roles,
			"page":  filter.Page,
			"limit": filter.Limit,
		},
	})
}

// CreateRole godoc
// @Summary 创建角色
// @Description 创建新的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param role body dto.CreateRoleRequest true "角色信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /staff/roles [post]
func (h *StaffHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	role, err := h.staffService.CreateRole(&req)
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

// GetRole godoc
// @Summary 获取角色详情
// @Description 根据ID获取角色详情
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path string true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /staff/roles/{id} [get]
func (h *StaffHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "角色ID不能为空",
		})
		return
	}

	role, err := h.staffService.GetRoleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "角色不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    role,
	})
}

// UpdateRole godoc
// @Summary 更新角色
// @Description 更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path string true "角色ID"
// @Param role body dto.UpdateRoleRequest true "角色信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /staff/roles/{id} [put]
func (h *StaffHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "角色ID不能为空",
		})
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	role, err := h.staffService.UpdateRole(id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "更新角色失败",
			"error":   err.Error(),
	})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    role,
	})
}

// DeleteRole godoc
// @Summary 删除角色
// @Description 删除角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path string true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /staff/roles/{id} [delete]
func (h *StaffHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "角色ID不能为空",
		})
		return
	}

	if err := h.staffService.DeleteRole(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "删除角色失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
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
// @Router /staff/availability/{staff_id} [get]
func (h *StaffHandler) GetAvailability(c *gin.Context) {
	staffID := c.Param("staff_id")
	if staffID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	availability, err := h.staffService.GetAvailability(staffID)
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
// @Param availability body dto.AvailabilityRequest true "可用性信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /staff/availability/{staff_id} [put]
func (h *StaffHandler) SetAvailability(c *gin.Context) {
	staffID := c.Param("staff_id")
	if staffID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工ID不能为空",
		})
		return
	}

	var req dto.AvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	err := h.staffService.SetAvailability(staffID, &req)
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

// GetAvailableStaffForTime godoc
// @Summary 查询可用员工
// @Description 根据时间和条件查询可用员工
// @Tags 可用性管理
// @Accept json
// @Produce json
// @Param request body dto.AvailableStaffRequest true "查询条件"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /staff/availability/search [post]
func (h *StaffHandler) GetAvailableStaffForTime(c *gin.Context) {
	var req dto.AvailableStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	result, err := h.staffService.GetAvailableStaffForTime(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "查询可用员工失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "查询成功",
		"data":    result,
	})
}