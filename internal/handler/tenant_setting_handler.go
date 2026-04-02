package handler

import (
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type TenantSettingHandler interface {
	GetSetting(c *gin.Context)
	UpdateSetting(c *gin.Context)
}

type tenantSettingHandler struct {
	service service.TenantSettingService
}

func NewTenantSettingHandler(service service.TenantSettingService) TenantSettingHandler {
	return &tenantSettingHandler{service: service}
}

// @Summary Get Tenant Setting
// @Tags Tenant Setting
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tenant-setting [get]
func (h *tenantSettingHandler) GetSetting(c *gin.Context) {
	tenantIDVal, _ := c.Get("tenant_id")
	tenantID := tenantIDVal.(uint)

	data, err := h.service.GetByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(404, utils.BuildErrorResponse("Not found", 404, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Success", 200, "success", data))
}

// @Summary Update Tenant Setting
// @Tags Tenant Setting
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body model.TenantSetting true "Tenant Setting"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tenant-setting [put]
func (h *tenantSettingHandler) UpdateSetting(c *gin.Context) {
	var req model.TenantSetting

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantIDVal, _ := c.Get("tenant_id")
	tenantID := tenantIDVal.(uint)

	data, err := h.service.UpdateSetting(c.Request.Context(), tenantID, req)
	if err != nil {
		c.JSON(400, utils.BuildErrorResponse("Failed", 400, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Updated", 200, "success", data))
}
