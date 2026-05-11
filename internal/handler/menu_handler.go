package handler

import (
	"fmt"
	"net/http"

	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type MenuHandler interface {
	GetMyMenus(c *gin.Context)
	GetAllMenus(c *gin.Context)
}

type menuHandler struct {
	service service.MenuService
}

func NewMenuHandler(service service.MenuService) MenuHandler {
	return &menuHandler{service: service}
}

func (h *menuHandler) GetMyMenus(c *gin.Context) {
	baseRole := fmt.Sprintf("%v", c.MustGet("base_role"))
	permissions := c.MustGet("permissions").([]string)
	planFeatures := c.MustGet("plan_features").([]string)

	isRestricted := false
	if val, ok := c.Get("is_restricted"); ok {
		isRestricted = val.(bool)
	}

	res, err := h.service.GetMyMenus(c.Request.Context(), baseRole, permissions, planFeatures, isRestricted)

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch menus", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Menus retrieved successfully", 200, "success", res))
}

func (h *menuHandler) GetAllMenus(c *gin.Context) {
	res, err := h.service.GetAllMenus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch all menus", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("All menus retrieved successfully", 200, "success", res))
}
