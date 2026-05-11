package handler

import (
	"net/http"
	"strconv"

	// modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserChangeRequestHandler interface {
	CreateRequest(c *gin.Context)
	GetMyRequests(c *gin.Context)
	GetRequests(c *gin.Context)
	ApproveRequest(c *gin.Context)
	RejectRequest(c *gin.Context)
	CancelRequest(c *gin.Context)
}

type userChangeRequestHandler struct {
	service service.UserChangeRequestService
}

func NewUserChangeRequestHandler(service service.UserChangeRequestService) UserChangeRequestHandler {
	return &userChangeRequestHandler{
		service: service,
	}
}

// @Summary Create User Change Request
// @Description Create a request to change user data (needs approval)
// @Tags UserChangeRequests
// @Accept json
// @Produce json
// @Param body body model.CreateUserChangeRequest true "Request Body"
// @Security BearerAuth
// @Security CookieAuth
// @Success 201 {object} utils.APIResponse{data=model.UserChangeRequestResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/users/request-change [post]
func (h *userChangeRequestHandler) CreateRequest(c *gin.Context) {
	var req model.CreateUserChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request body", 400, "error", err.Error()))
		return
	}

	userID := c.MustGet("user_id").(uint)
	tenantID := c.MustGet("tenant_id").(uint)

	res, err := h.service.CreateRequest(c.Request.Context(), userID, tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to create request", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Request created successfully", 201, "success", res))
}

// @Summary Get My Change Requests
// @Description Get all change requests for the current user
// @Tags UserChangeRequests
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=[]model.UserChangeRequestResponse}
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/users/me/change-requests [get]
func (h *userChangeRequestHandler) GetMyRequests(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	res, err := h.service.GetMyRequests(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch requests", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Requests fetched successfully", 200, "success", res))
}

// @Summary Get Change Requests
// @Description Get all change requests for the tenant (filtered by status)
// @Tags UserChangeRequests
// @Produce json
// @Param status query string false "Filter by status"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=[]model.UserChangeRequestResponse}
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/users/change-requests [get]
func (h *userChangeRequestHandler) GetRequests(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	status := c.Query("status")

	res, err := h.service.GetAllRequests(c.Request.Context(), tenantID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch requests", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Requests fetched successfully", 200, "success", res))
}

// @Summary Approve Change Request
// @Description Approve a pending change request
// @Tags UserChangeRequests
// @Produce json
// @Param id path int true "Request ID"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/users/approve-change/{id} [post]
func (h *userChangeRequestHandler) ApproveRequest(c *gin.Context) {
	idParam := c.Param("id")
	id, _ := strconv.Atoi(idParam)
	adminID := c.MustGet("user_id").(uint)

	err := h.service.ApproveRequest(c.Request.Context(), uint(id), adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to approve request", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Request approved successfully", 200, "success", nil))
}

// @Summary Reject Change Request
// @Description Reject a pending change request
// @Tags UserChangeRequests
// @Accept json
// @Produce json
// @Param id path int true "Request ID"
// @Param body body model.ApproveUserChangeRequest true "Notes"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/users/reject-change/{id} [post]
func (h *userChangeRequestHandler) RejectRequest(c *gin.Context) {
	idParam := c.Param("id")
	id, _ := strconv.Atoi(idParam)
	adminID := c.MustGet("user_id").(uint)

	var req model.ApproveUserChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request body", 400, "error", err.Error()))
		return
	}

	if req.AdminNotes == "" {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Admin notes are required for rejection", 400, "error", nil))
		return
	}

	err := h.service.RejectRequest(c.Request.Context(), uint(id), adminID, req.AdminNotes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to reject request", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Request rejected successfully", 200, "success", nil))
}

// @Summary Cancel Change Request
// @Description Cancel a pending or draft change request
// @Tags UserChangeRequests
// @Produce json
// @Param id path int true "Request ID"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/users/me/change-requests/{id}/cancel [patch]
func (h *userChangeRequestHandler) CancelRequest(c *gin.Context) {
	idParam := c.Param("id")
	id, _ := strconv.Atoi(idParam)
	userID := c.MustGet("user_id").(uint)

	err := h.service.CancelRequest(c.Request.Context(), uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to cancel request", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Request cancelled successfully", 200, "success", nil))
}
