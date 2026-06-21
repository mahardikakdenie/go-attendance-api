package handler

import (
	"fmt"
	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MenuHandler interface {
	GetMyMenus(c *gin.Context)
	GetAllMenus(c *gin.Context)
	GetRoleMenuOverview(c *gin.Context)
	CreateMenu(c *gin.Context)
	UpdateMenu(c *gin.Context)
}

type menuHandler struct {
	service service.MenuService
}

func NewMenuHandler(service service.MenuService) MenuHandler {
	return &menuHandler{service: service}
}

func (h *menuHandler) CreateMenu(c *gin.Context) {
	var req modelDto.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	res, err := h.service.CreateMenu(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to create menu", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Menu created successfully", 201, "success", res))
}

func (h *menuHandler) GetMyMenus(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	roleID := c.MustGet("role_id").(uint)
	baseRole := fmt.Sprintf("%v", c.MustGet("base_role"))
	planFeatures := c.MustGet("plan_features").([]string)

	isRestricted := false
	if val, ok := c.Get("is_restricted"); ok {
		isRestricted = val.(bool)
	}

	res, err := h.service.GetMyMenus(c.Request.Context(), userID, roleID, baseRole, planFeatures, isRestricted)

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch menus", 500, "error", err.Error()))
		return
	}

	fmt.Printf("[DEBUG] GetMyMenus for roleID=%d, restricted: %v, menuCount: %d\n", roleID, isRestricted, len(res))
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

func (h *menuHandler) GetRoleMenuOverview(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)

	res, err := h.service.GetRolesMenuOverview(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch menu overview", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Menu overview retrieved successfully", 200, "success", res))
}

func (h *menuHandler) UpdateMenu(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req modelDto.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	res, err := h.service.UpdateMenu(c.Request.Context(), uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to update menu", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Menu updated successfully", 200, "success", res))
}
