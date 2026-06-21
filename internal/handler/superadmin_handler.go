package handler

import (
	"errors"
	"strconv"

	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type SuperadminHandler interface {
	GetOwnersWithStats(c *gin.Context)
	GetPlatformAccounts(c *gin.Context)
	CreatePlatformAccount(c *gin.Context)
	UpdatePlatformAccount(c *gin.Context)
	TogglePlatformAccountStatus(c *gin.Context)

	// System Role Management
	ListSystemRoles(c *gin.Context)
	ListAllPermissions(c *gin.Context)
	ListTenantModules(c *gin.Context)
	CreateSystemRole(c *gin.Context)
	UpdateSystemRole(c *gin.Context)
	PatchSystemRole(c *gin.Context)
	DeleteSystemRole(c *gin.Context)

	GetAnalyticsDashboard(c *gin.Context)
	GetTenantDetails(c *gin.Context)
}

type superadminHandler struct {
	service service.SuperadminService
}

func NewSuperadminHandler(service service.SuperadminService) SuperadminHandler {
	return &superadminHandler{service: service}
}

func (h *superadminHandler) handleError(c *gin.Context, message string, err error) {
	var appErr *utils.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Code, utils.BuildErrorResponse(appErr.Message, appErr.Code, "error", appErr.Details))
		return
	}
	c.JSON(500, utils.BuildErrorResponse(message, 500, "error", err.Error()))
}

// @Summary Get Owners with Stats
func (h *superadminHandler) GetOwnersWithStats(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	search := c.Query("search")
	status := c.Query("status")
	plan := c.Query("plan")

	// Validate status param if provided
	if status != "" && status != "Active" && status != "Suspended" {
		c.JSON(400, utils.BuildErrorResponse("Invalid status filter. Allowed values: Active, Suspended", 400, "error", nil))
		return
	}

	results, total, err := h.service.GetOwnersWithStats(c.Request.Context(), limit, offset, search, status, plan)
	if err != nil {
		h.handleError(c, "Failed to fetch owners statistics", err)
		return
	}

	pagination := utils.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: (offset / limit) + 1,
		LastPage:    int((total + int64(limit) - 1) / int64(limit)),
	}

	c.JSON(200, utils.BuildResponseWithPagination("Success", 200, "success", results, pagination))
}

// @Summary Get Platform Accounts
func (h *superadminHandler) GetPlatformAccounts(c *gin.Context) {
	search := c.Query("search")
	role := c.Query("role")
	if role == "" {
		role = c.Query("base_role")
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	results, total, err := h.service.GetPlatformAccounts(c.Request.Context(), search, role, limit, offset)
	if err != nil {
		h.handleError(c, "Failed to fetch platform accounts", err)
		return
	}

	pagination := utils.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: (offset / limit) + 1,
		LastPage:    int((total + int64(limit) - 1) / int64(limit)),
	}

	c.JSON(200, utils.BuildResponseWithPagination("Success", 200, "success", results, pagination))
}

// @Summary Create Platform Account
func (h *superadminHandler) CreatePlatformAccount(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	res, err := h.service.CreatePlatformAccount(c.Request.Context(), req, userID)
	if err != nil {
		h.handleError(c, "Failed to create account", err)
		return
	}

	c.JSON(201, utils.BuildResponse("Account created successfully", 201, "success", res))
}

// @Summary Update Platform Account
func (h *superadminHandler) UpdatePlatformAccount(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	res, err := h.service.UpdatePlatformAccount(c.Request.Context(), uint(id), req, userID)
	if err != nil {
		h.handleError(c, "Failed to update account", err)
		return
	}

	c.JSON(200, utils.BuildResponse("Account updated successfully", 200, "success", res))
}

// @Summary Toggle Platform Account Status
func (h *superadminHandler) TogglePlatformAccountStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	isActive := req.Status == "active"
	err := h.service.TogglePlatformAccountStatus(c.Request.Context(), uint(id), isActive)
	if err != nil {
		h.handleError(c, "Failed to update status", err)
		return
	}

	c.JSON(200, utils.BuildResponse("Status updated successfully", 200, "success", nil))
}

// @Summary List System Roles
func (h *superadminHandler) ListSystemRoles(c *gin.Context) {
	roles, err := h.service.ListSystemRoles(c.Request.Context())
	if err != nil {
		h.handleError(c, "Failed to list system roles", err)
		return
	}

	c.JSON(200, utils.BuildResponse("Success", 200, "success", roles))
}

// @Summary List All Available Permissions
func (h *superadminHandler) ListAllPermissions(c *gin.Context) {
	scope := c.DefaultQuery("scope", "system") // Default to system for superadmin
	permissions, err := h.service.ListAllPermissions(c.Request.Context(), scope)
	if err != nil {
		h.handleError(c, "Failed to list permissions", err)
		return
	}

	c.JSON(200, utils.BuildResponse("Success", 200, "success", permissions))
}

// @Summary List Tenant Modules
func (h *superadminHandler) ListTenantModules(c *gin.Context) {
	permissions, err := h.service.ListAllPermissions(c.Request.Context(), "tenant")
	if err != nil {
		h.handleError(c, "Failed to list tenant modules", err)
		return
	}

	c.JSON(200, utils.BuildResponse("Success", 200, "success", permissions))
}

// @Summary Create System Role
func (h *superadminHandler) CreateSystemRole(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req modelDto.CreateSystemRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	role, err := h.service.CreateSystemRole(c.Request.Context(), req, userID)
	if err != nil {
		h.handleError(c, "Failed to create system role", err)
		return
	}

	c.JSON(201, utils.BuildResponse("System role created successfully", 201, "success", role))
}

// @Summary Update System Role
func (h *superadminHandler) UpdateSystemRole(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, _ := strconv.Atoi(c.Param("id"))
	var req modelDto.CreateSystemRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	role, err := h.service.UpdateSystemRole(c.Request.Context(), uint(id), req, userID)
	if err != nil {
		h.handleError(c, "Failed to update system role", err)
		return
	}

	c.JSON(200, utils.BuildResponse("System role updated successfully", 200, "success", role))
}

func (h *superadminHandler) PatchSystemRole(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, _ := strconv.Atoi(c.Param("id"))
	var req modelDto.UpdateSystemRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	role, err := h.service.PatchSystemRole(c.Request.Context(), uint(id), req, userID)
	if err != nil {
		h.handleError(c, "Failed to patch system role", err)
		return
	}

	c.JSON(200, utils.BuildResponse("System role patched successfully", 200, "success", role))
}

// @Summary Delete System Role
func (h *superadminHandler) DeleteSystemRole(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, _ := strconv.Atoi(c.Param("id"))
	err := h.service.DeleteSystemRole(c.Request.Context(), uint(id), userID)
	if err != nil {
		h.handleError(c, "Failed to delete system role", err)
		return
	}

	c.JSON(200, utils.BuildResponse("System role deleted successfully", 200, "success", nil))
}

// @Summary Get Superadmin Analytics Dashboard
func (h *superadminHandler) GetAnalyticsDashboard(c *gin.Context) {
	period := c.DefaultQuery("period", "this_year")
	res, err := h.service.GetAnalyticsDashboard(c.Request.Context(), period)
	if err != nil {
		h.handleError(c, "Failed to fetch analytics dashboard", err)
		return
	}

	c.JSON(200, utils.BuildResponse("Analytics retrieved successfully", 200, "success", res))
}

func (h *superadminHandler) GetTenantDetails(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid tenant ID", 400, "error", nil))
		return
	}

	res, err := h.service.GetTenantFullDetails(c.Request.Context(), uint(id))
	if err != nil {
		h.handleError(c, "Failed to fetch tenant details", err)
		return
	}

	c.JSON(200, utils.BuildResponse("Tenant details retrieved successfully", 200, "success", res))
}
