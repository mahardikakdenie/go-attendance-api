package handler

import (
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TenantRoleHandler interface {
	ListRoles(c *gin.Context)
	CreateRole(c *gin.Context)
	UpdateRole(c *gin.Context)
	DeleteRole(c *gin.Context)
	GetHierarchy(c *gin.Context)
	SaveHierarchy(c *gin.Context)
}

type tenantRoleHandler struct {
	service service.TenantRoleService
}

func NewTenantRoleHandler(service service.TenantRoleService) TenantRoleHandler {
	return &tenantRoleHandler{
		service: service,
	}
}

// @Summary List all roles for tenant
// @Tags Tenant Roles
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/tenant-roles [get]
func (h *tenantRoleHandler) ListRoles(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	roles, err := h.service.ListRoles(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch roles", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Roles fetched successfully", 200, "success", roles))
}

// @Summary Create custom role
// @Tags Tenant Roles
// @Accept json
// @Produce json
// @Param body body service.CreateRoleRequest true "Role Data"
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/tenant-roles [post]
func (h *tenantRoleHandler) CreateRole(c *gin.Context) {
	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	role, err := h.service.CreateRole(c.Request.Context(), tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to create role", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, utils.BuildResponse("Role created successfully", 201, "success", role))
}

// @Summary Update role
// @Tags Tenant Roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param body body service.UpdateRoleRequest true "Update Data"
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/tenant-roles/{id} [patch]
func (h *tenantRoleHandler) UpdateRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	role, err := h.service.UpdateRole(c.Request.Context(), tenantID, uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to update role", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Role updated successfully", 200, "success", role))
}

// @Summary Delete role
// @Tags Tenant Roles
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/tenant-roles/{id} [delete]
func (h *tenantRoleHandler) DeleteRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	tenantID := c.MustGet("tenant_id").(uint)

	if err := h.service.DeleteRole(c.Request.Context(), tenantID, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to delete role", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Role deleted successfully", 200, "success", nil))
}

// @Summary Get role hierarchy
// @Tags Tenant Roles
// @Param id path int true "Role ID"
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/tenant-roles/{id}/hierarchy [get]
func (h *tenantRoleHandler) GetHierarchy(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	hierarchy, err := h.service.GetHierarchy(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch hierarchy", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Hierarchy fetched successfully", 200, "success", hierarchy))
}

type SaveHierarchyRequest struct {
	ParentRoleID uint   `json:"parent_role_id" binding:"required"`
	ChildRoleIDs []uint `json:"child_role_ids" binding:"required"`
}

// @Summary Save role hierarchy
// @Tags Tenant Roles
// @Accept json
// @Produce json
// @Param body body SaveHierarchyRequest true "Hierarchy Data"
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/tenant-roles/hierarchy [post]
func (h *tenantRoleHandler) SaveHierarchy(c *gin.Context) {
	var req SaveHierarchyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	if err := h.service.SaveHierarchy(c.Request.Context(), tenantID, req.ParentRoleID, req.ChildRoleIDs); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to save hierarchy", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Hierarchy saved successfully", 200, "success", nil))
}
