package handler

import (
	"net/http"

	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type SettingHandler interface {
	GetAllowancePresets(c *gin.Context)
}

type settingHandler struct {
	allowanceService service.AllowancePresetService
}

func NewSettingHandler(allowanceService service.AllowancePresetService) SettingHandler {
	return &settingHandler{allowanceService: allowanceService}
}

func (h *settingHandler) GetAllowancePresets(c *gin.Context) {
	res, err := h.allowanceService.GetAllPresets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch presets", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Presets retrieved successfully", 200, "success", res))
}
