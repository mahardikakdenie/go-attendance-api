package handler

import (
	"net/http"
	"strconv"

	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type PerformanceHandler interface {
	GetMyGoals(c *gin.Context)
	GetUserGoals(c *gin.Context)
	CreateGoal(c *gin.Context)
	UpdateGoalProgress(c *gin.Context)

	GetAllCycles(c *gin.Context)
	GetAppraisalsByCycle(c *gin.Context)
	SubmitSelfReview(c *gin.Context)
}

type performanceHandler struct {
	service service.PerformanceService
}

func NewPerformanceHandler(service service.PerformanceService) PerformanceHandler {
	return &performanceHandler{service: service}
}

// @Summary Get My Goals
// @Tags Performance
// @Security BearerAuth
// @Router /api/v1/performance/goals/me [get]
func (h *performanceHandler) GetMyGoals(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	res, err := h.service.GetMyGoals(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch goals", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Goals fetched successfully", 200, "success", res))
}

// @Summary Get User Goals
// @Tags Performance
// @Param userId path int true "User ID"
// @Security BearerAuth
// @Router /api/v1/performance/goals/user/{userId} [get]
func (h *performanceHandler) GetUserGoals(c *gin.Context) {
	requesterID := c.MustGet("user_id").(uint)
	targetUserID, _ := strconv.Atoi(c.Param("userId"))

	res, err := h.service.GetUserGoals(c.Request.Context(), requesterID, uint(targetUserID))
	if err != nil {
		c.JSON(http.StatusForbidden, utils.BuildErrorResponse("Access denied", 403, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("User goals fetched successfully", 200, "success", res))
}

// @Summary Create Goal
// @Tags Performance
// @Accept json
// @Param body body modelDto.CreateGoalRequest true "Goal Payload"
// @Security BearerAuth
// @Router /api/v1/performance/goals [post]
func (h *performanceHandler) CreateGoal(c *gin.Context) {
	requesterID := c.MustGet("user_id").(uint)
	tenantID := c.MustGet("tenant_id").(uint)

	var req modelDto.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	res, err := h.service.CreateGoal(c.Request.Context(), tenantID, requesterID, req)
	if err != nil {
		c.JSON(http.StatusForbidden, utils.BuildErrorResponse("Failed to create goal", 403, "error", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, utils.BuildResponse("Goal created successfully", 201, "success", res))
}

// @Summary Update Goal Progress
// @Tags Performance
// @Param id path int true "Goal ID"
// @Accept json
// @Param body body modelDto.UpdateGoalProgressRequest true "Progress Payload"
// @Security BearerAuth
// @Router /api/v1/performance/goals/{id}/progress [put]
func (h *performanceHandler) UpdateGoalProgress(c *gin.Context) {
	requesterID := c.MustGet("user_id").(uint)
	goalID, _ := strconv.Atoi(c.Param("id"))

	var req modelDto.UpdateGoalProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	if err := h.service.UpdateGoalProgress(c.Request.Context(), requesterID, uint(goalID), req.CurrentProgress); err != nil {
		c.JSON(http.StatusForbidden, utils.BuildErrorResponse("Failed to update progress", 403, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Progress updated successfully", 200, "success", nil))
}

// @Summary Get All Cycles
// @Tags Performance
// @Security BearerAuth
// @Router /api/v1/performance/cycles [get]
func (h *performanceHandler) GetAllCycles(c *gin.Context) {
	res, err := h.service.GetAllCycles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch cycles", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Cycles fetched successfully", 200, "success", res))
}

// @Summary Get Appraisals by Cycle
// @Tags Performance
// @Param cycleId path int true "Cycle ID"
// @Security BearerAuth
// @Router /api/v1/performance/appraisals/cycle/{cycleId} [get]
func (h *performanceHandler) GetAppraisalsByCycle(c *gin.Context) {
	cycleID, _ := strconv.Atoi(c.Param("cycleId"))
	res, err := h.service.GetAppraisalsByCycle(c.Request.Context(), uint(cycleID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch appraisals", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Appraisals fetched successfully", 200, "success", res))
}

// @Summary Submit Self Review
// @Tags Performance
// @Param id path int true "Appraisal ID"
// @Accept json
// @Param body body modelDto.SubmitSelfReviewRequest true "Review Payload"
// @Security BearerAuth
// @Router /api/v1/performance/appraisals/{id}/self-review [put]
func (h *performanceHandler) SubmitSelfReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	appraisalID, _ := strconv.Atoi(c.Param("id"))

	var req modelDto.SubmitSelfReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	if err := h.service.SubmitSelfReview(c.Request.Context(), userID, uint(appraisalID), req); err != nil {
		c.JSON(http.StatusForbidden, utils.BuildErrorResponse("Failed to submit review", 403, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Self-review submitted successfully", 200, "success", nil))
}
