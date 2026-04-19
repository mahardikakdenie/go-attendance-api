package handler

import (
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
	CreateSystemRole(c *gin.Context)
	UpdateSystemRole(c *gin.Context)
	DeleteSystemRole(c *gin.Context)
}

type superadminHandler struct {
	service service.SuperadminService
}

func NewSuperadminHandler(service service.SuperadminService) SuperadminHandler {
	return &superadminHandler{service: service}
}

// @Summary Get Owners with Stats
func (h *superadminHandler) GetOwnersWithStats(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	results, total, err := h.service.GetOwnersWithStats(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to fetch owners statistics", 500, "error", err.Error()))
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	results, total, err := h.service.GetPlatformAccounts(c.Request.Context(), search, limit, offset)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to fetch platform accounts", 500, "error", err.Error()))
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
		c.JSON(500, utils.BuildErrorResponse("Failed to create account", 500, "error", err.Error()))
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
		c.JSON(500, utils.BuildErrorResponse("Failed to update account", 500, "error", err.Error()))
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
		c.JSON(500, utils.BuildErrorResponse("Failed to update status", 500, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Status updated successfully", 200, "success", nil))
}

// @Summary List System Roles
func (h *superadminHandler) ListSystemRoles(c *gin.Context) {
	roles, err := h.service.ListSystemRoles(c.Request.Context())
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to list system roles", 500, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Success", 200, "success", roles))
}

// @Summary List All Available Permissions
func (h *superadminHandler) ListAllPermissions(c *gin.Context) {
	permissions, err := h.service.ListAllPermissions(c.Request.Context())
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to list permissions", 500, "error", err.Error()))
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
		c.JSON(500, utils.BuildErrorResponse("Failed to create system role", 500, "error", err.Error()))
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
		c.JSON(500, utils.BuildErrorResponse("Failed to update system role", 500, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("System role updated successfully", 200, "success", role))
}

// @Summary Delete System Role
func (h *superadminHandler) DeleteSystemRole(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, _ := strconv.Atoi(c.Param("id"))
	err := h.service.DeleteSystemRole(c.Request.Context(), uint(id), userID)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to delete system role", 500, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("System role deleted successfully", 200, "success", nil))
}
