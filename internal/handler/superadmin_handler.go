package handler

import (
	"strconv"

	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type SuperadminHandler interface {
	GetOwnersWithStats(c *gin.Context)
}

type superadminHandler struct {
	service service.SuperadminService
}

func NewSuperadminHandler(service service.SuperadminService) SuperadminHandler {
	return &superadminHandler{service: service}
}

// @Summary Get Owners with Stats
// @Description Get list of owners (Admins) with their tenant statistics (SuperAdmin only)
// @Tags SuperAdmin
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=[]modelDto.OwnerWithStatsResponse}
// @Failure 401 {object} utils.APIResponse
// @Failure 403 {object} utils.APIResponse
// @Router /api/v1/superadmin/owners-stats [get]
func (h *superadminHandler) GetOwnersWithStats(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	results, total, err := h.service.GetOwnersWithStats(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to fetch owners statistics", 500, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Success", 200, "success", gin.H{
		"items": results,
		"total": total,
	}))
}
