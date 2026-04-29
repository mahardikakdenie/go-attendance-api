package handler

import (
	"strconv"

	modelDto "go-attendance-api/internal/dto"
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

	// Superadmin Plan Management
	GetAllPlans(c *gin.Context)
	GetPlanByID(c *gin.Context)
	CreatePlan(c *gin.Context)
	UpdatePlan(c *gin.Context)
	DeletePlan(c *gin.Context)

	// Superadmin Subscription Management
	UpdateTenantSubscription(c *gin.Context)
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
// @Success 200 {object} utils.APIResponse{data=modelDto.SubscriptionsDataResponse}
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

// @Summary Send Reminder
// @Description Send payment reminder to tenant owner
// @Tags Subscription
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Success 200 {object} utils.APIResponse
// @Router /api/v1/superadmin/subscriptions/{id}/remind [post]
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

// @Summary Suspend Tenant
// @Description Suspend tenant access
// @Tags Subscription
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param body body modelDto.SuspendRequest true "Suspension Reason"
// @Success 200 {object} utils.APIResponse
// @Router /api/v1/superadmin/subscriptions/{id}/suspend [post]
func (h *subscriptionHandler) SuspendTenant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(400, utils.BuildErrorResponse("Invalid ID", 400, "error", err.Error()))
		return
	}

	var req modelDto.SuspendRequest
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

// @Summary Get My Subscription
// @Description Get current tenant subscription details
// @Tags Subscription
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=model.Subscription}
// @Router /api/v1/subscriptions/me [get]
func (h *subscriptionHandler) GetMySubscription(ctx *gin.Context) {
	tenantID := ctx.MustGet("tenant_id").(uint)

	res, err := h.service.GetMySubscription(ctx.Request.Context(), tenantID)
	if err != nil {
		ctx.JSON(500, utils.BuildErrorResponse("Failed to fetch subscription", 500, "error", err.Error()))
		return
	}

	ctx.JSON(200, utils.BuildResponse("Success", 200, "OK", res))
}

// @Summary Upgrade Subscription
// @Description Upgrade current tenant plan
// @Tags Subscription
// @Security BearerAuth
// @Param body body modelDto.UpgradeRequest true "Target Plan"
// @Success 200 {object} utils.APIResponse
// @Router /api/v1/subscriptions/upgrade [post]
func (h *subscriptionHandler) UpgradeSubscription(ctx *gin.Context) {
	tenantID := ctx.MustGet("tenant_id").(uint)

	var req modelDto.UpgradeRequest
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

// @Summary List Plans
// @Description Get all global subscription plans (Superadmin only)
// @Tags Plan
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]model.SubscriptionPlan}
// @Router /api/v1/superadmin/plans [get]
func (h *subscriptionHandler) GetAllPlans(c *gin.Context) {
	res, err := h.service.GetAllPlans(c.Request.Context())
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to fetch plans", 500, "error", err.Error()))
		return
	}
	c.JSON(200, utils.BuildResponse("Plans retrieved successfully", 200, "success", res))
}

// @Summary Get Plan by ID
// @Description Get details of a subscription plan
// @Tags Plan
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Success 200 {object} utils.APIResponse{data=model.SubscriptionPlan}
// @Router /api/v1/superadmin/plans/{id} [get]
func (h *subscriptionHandler) GetPlanByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	res, err := h.service.GetPlanByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to fetch plan", 500, "error", err.Error()))
		return
	}
	c.JSON(200, utils.BuildResponse("Plan retrieved successfully", 200, "success", res))
}

// @Summary Create Plan
// @Description Create a new global subscription plan
// @Tags Plan
// @Security BearerAuth
// @Param body body modelDto.CreatePlanRequest true "Plan Details"
// @Success 201 {object} utils.APIResponse{data=model.SubscriptionPlan}
// @Router /api/v1/superadmin/plans [post]
func (h *subscriptionHandler) CreatePlan(c *gin.Context) {
	var req modelDto.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	res, err := h.service.CreatePlan(c.Request.Context(), req)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to create plan", 500, "error", err.Error()))
		return
	}
	c.JSON(201, utils.BuildResponse("Plan created successfully", 201, "success", res))
}

// @Summary Update Plan
// @Description Update an existing subscription plan
// @Tags Plan
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Param body body modelDto.UpdatePlanRequest true "Updated Plan Details"
// @Success 200 {object} utils.APIResponse{data=model.SubscriptionPlan}
// @Router /api/v1/superadmin/plans/{id} [put]
func (h *subscriptionHandler) UpdatePlan(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req modelDto.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	res, err := h.service.UpdatePlan(c.Request.Context(), uint(id), req)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to update plan", 500, "error", err.Error()))
		return
	}
	c.JSON(200, utils.BuildResponse("Plan updated successfully", 200, "success", res))
}

// @Summary Delete Plan
// @Description Delete a subscription plan
// @Tags Plan
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Success 200 {object} utils.APIResponse
// @Router /api/v1/superadmin/plans/{id} [delete]
func (h *subscriptionHandler) DeletePlan(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	err := h.service.DeletePlan(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to delete plan", 500, "error", err.Error()))
		return
	}
	c.JSON(200, utils.BuildResponse("Plan deleted successfully", 200, "success", nil))
}

// @Summary Update Tenant Subscription
// @Description Manually update a tenant's subscription (Superadmin override)
// @Tags Subscription
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param body body modelDto.UpdateTenantSubscriptionRequest true "Update Details"
// @Success 200 {object} utils.APIResponse{data=model.Subscription}
// @Router /api/v1/superadmin/subscriptions/{id} [put]
func (h *subscriptionHandler) UpdateTenantSubscription(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req modelDto.UpdateTenantSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	res, err := h.service.UpdateTenantSubscription(c.Request.Context(), uint(id), req)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to update subscription", 500, "error", err.Error()))
		return
	}
	c.JSON(200, utils.BuildResponse("Subscription updated successfully", 200, "success", res))
}
