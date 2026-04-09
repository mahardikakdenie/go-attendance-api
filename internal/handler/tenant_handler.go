package handler

import (
	"strconv"

	// modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type TenantHandler interface {
	CreateTenant(c *gin.Context)
	GetAllTenant(c *gin.Context)
	GetTenantByID(c *gin.Context)
}

type tenantHandler struct {
	service service.TenantService
}

func NewTenantHandler(service service.TenantService) TenantHandler {
	return &tenantHandler{service: service}
}

// @Summary Create Tenant
// @Description Create a new tenant (SuperAdmin only)
// @Tags Tenant
// @Accept json
// @Produce json
// @Param request body model.Tenant true "Tenant Data"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=model.Tenant}
// @Failure 400 {object} utils.APIResponse
// @Router /api/v1/tenants [post]
func (h *tenantHandler) CreateTenant(c *gin.Context) {
	var req model.Tenant

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	res, err := h.service.CreateTenant(c.Request.Context(), req)
	if err != nil {
		c.JSON(400, utils.BuildErrorResponse("Failed", 400, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Tenant created", 200, "success", res))
}

// @Summary Get All Tenants
// @Description Get list of all tenants (SuperAdmin only)
// @Tags Tenant
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=[]model.Tenant}
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/tenants [get]
func (h *tenantHandler) GetAllTenant(c *gin.Context) {
	data, err := h.service.GetAllTenants(c.Request.Context())
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed", 500, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Success", 200, "success", data))
}

// @Summary Get Tenant By ID
// @Description Get detailed information about a specific tenant
// @Tags Tenant
// @Produce json
// @Param id path int true "Tenant ID"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=model.Tenant}
// @Failure 400 {object} utils.APIResponse
// @Failure 403 {object} utils.APIResponse
// @Failure 404 {object} utils.APIResponse
// @Router /api/v1/tenants/{id} [get]
func (h *tenantHandler) GetTenantByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid ID", 400, "error", err.Error()))
		return
	}

	// 🛡️ SECURITY CHECK: Non-superadmin can only see their own tenant
	role := c.MustGet("role").(string)
	userTenantID := c.MustGet("tenant_id").(uint)

	if role != "superadmin" && userTenantID != uint(id) {
		c.JSON(403, utils.BuildErrorResponse("Forbidden: You can only access your own tenant data", 403, "error", nil))
		return
	}

	data, err := h.service.GetTenantByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(404, utils.BuildErrorResponse("Not found", 404, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Success", 200, "success", data))
}
