package handler

import (
	"strconv"

	dto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler interface {
	GetSubscriptions(c *gin.Context)
	GetMySubscription(c *gin.Context)
	UpgradeSubscription(c *gin.Context)
	RemindTenant(c *gin.Context)
	SuspendTenant(c *gin.Context)
}

type subscriptionHandler struct {
	service service.SubscriptionService
}

func NewSubscriptionHandler(service service.SubscriptionService) SubscriptionHandler {
	return &subscriptionHandler{service: service}
}

// @Summary Get Subscriptions
// @Description Get all subscriptions with statistics (Superadmin only)
// @Tags Subscription
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Limit per page"
// @Param status query string false "Filter by status"
// @Param search query string false "Search by tenant name/code"
// @Success 200 {object} utils.APIResponse{data=dto.SubscriptionsDataResponse}
// @Router /api/v1/superadmin/subscriptions [get]
func (h *subscriptionHandler) GetSubscriptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	search := c.Query("search")

	res, err := h.service.GetSubscriptions(c.Request.Context(), page, limit, status, search)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to fetch subscriptions", 500, "error", err.Error()))
		return
	}

	pagination := utils.Pagination{
		Total:       res.Total,
		PerPage:     limit,
		CurrentPage: page,
		LastPage:    int((res.Total + int64(limit) - 1) / int64(limit)),
	}

	c.JSON(200, utils.BuildResponseWithPagination("Subscriptions retrieved successfully", 200, "success", res, pagination))
}

func (h *subscriptionHandler) RemindTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid ID", 400, "error", err.Error()))
		return
	}

	err = h.service.RemindTenant(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to send reminder", 500, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Reminder sent successfully", 200, "OK", nil))
}

func (h *subscriptionHandler) SuspendTenant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(400, utils.BuildErrorResponse("Invalid ID", 400, "error", err.Error()))
		return
	}

	var req dto.SuspendRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	err = h.service.SuspendTenant(ctx.Request.Context(), uint(id), req.Reason)
	if err != nil {
		ctx.JSON(500, utils.BuildErrorResponse("Failed to suspend tenant", 500, "error", err.Error()))
		return
	}

	ctx.JSON(200, utils.BuildResponse("Tenant suspended successfully", 200, "OK", nil))
}

func (h *subscriptionHandler) GetMySubscription(ctx *gin.Context) {
	tenantID := ctx.MustGet("tenant_id").(uint)

	res, err := h.service.GetMySubscription(ctx.Request.Context(), tenantID)
	if err != nil {
		ctx.JSON(500, utils.BuildErrorResponse("Failed to fetch subscription", 500, "error", err.Error()))
		return
	}

	ctx.JSON(200, utils.BuildResponse("Success", 200, "OK", res))
}

func (h *subscriptionHandler) UpgradeSubscription(ctx *gin.Context) {
	tenantID := ctx.MustGet("tenant_id").(uint)

	var req dto.UpgradeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	err := h.service.UpgradeSubscription(ctx.Request.Context(), tenantID, req.Plan)
	if err != nil {
		ctx.JSON(400, utils.BuildErrorResponse(err.Error(), 400, "error", nil))
		return
	}

	ctx.JSON(200, utils.BuildResponse("Subscription upgraded successfully", 200, "OK", nil))
}
