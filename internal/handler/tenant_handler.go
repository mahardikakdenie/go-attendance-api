package handler

import (
	"strconv"

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
// @Tags Tenant
// @Accept json
// @Produce json
// @Param request body model.Tenant true "Tenant Data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
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
// @Tags Tenant
// @Produce json
// @Success 200 {object} map[string]interface{}
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
// @Tags Tenant
// @Produce json
// @Param id path int true "Tenant ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tenants/{id} [get]
func (h *tenantHandler) GetTenantByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid ID", 400, "error", err.Error()))
		return
	}

	data, err := h.service.GetTenantByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(404, utils.BuildErrorResponse("Not found", 404, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Success", 200, "success", data))
}
