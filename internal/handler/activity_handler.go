package handler

import (
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ActivityHandler interface {
	GetRecentActivities(c *gin.Context)
}

type activityHandler struct {
	userService service.UserService
}

func NewActivityHandler(userService service.UserService) ActivityHandler {
	return &activityHandler{userService: userService}
}

// @Summary Get Recent Activities
// @Description Get combined recent activities for the logged-in user
// @Tags Activity
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=[]model.RecentActivityResponse}
// @Router /api/v1/activities/recent [get]
func (h *activityHandler) GetRecentActivities(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	activities, err := h.userService.GetRecentActivities(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch activities", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Activities fetched successfully", 200, "success", activities))
}
