package handler

import (
	"net/http"

	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SupportHandler interface {
	// Public
	CreateTrialRequest(c *gin.Context)

	// Admin (Tenant 1)
	GetAllTrialRequests(c *gin.Context)
	UpdateTrialStatus(c *gin.Context)

	GetAllSupportMessages(c *gin.Context)
	UpdateSupportStatus(c *gin.Context)

	// Superadmin Only
	GetAllProvisioningTickets(c *gin.Context)
	ExecuteProvisioning(c *gin.Context)

	// Tenant User
	CreateSupportMessage(c *gin.Context)
}

type supportHandler struct {
	service service.SupportService
}

func NewSupportHandler(service service.SupportService) SupportHandler {
	return &supportHandler{service: service}
}

// @Summary Create Trial Request
// @Description Submit a new trial request from landing page
// @Tags Public
// @Accept json
// @Produce json
// @Param body body modelDto.CreateTrialRequestRequest true "Trial Request Payload"
// @Success 201 {object} utils.APIResponse{data=modelDto.TrialRequestResponse}
// @Router /v1/public/trial-request [post]
func (h *supportHandler) CreateTrialRequest(c *gin.Context) {
	var req modelDto.CreateTrialRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	res, err := h.service.CreateTrialRequest(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to create trial request", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Trial request submitted successfully", 201, "success", res))
}

// @Summary Get All Trial Requests
// @Description List all trial requests (Superadmin or CS Role in Tenant 1)
// @Tags Support
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]modelDto.TrialRequestResponse}
// @Router /v1/admin/support/trials [get]
func (h *supportHandler) GetAllTrialRequests(c *gin.Context) {
	res, err := h.service.GetAllTrialRequests(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch trial requests", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Trial requests fetched successfully", 200, "success", res))
}

// @Summary Update Trial Status
// @Description Update trial request status (e.g., to APPROVED)
// @Tags Support
// @Accept json
// @Produce json
// @Param id path string true "Trial Request ID (UUID)"
// @Param body body modelDto.UpdateTrialRequestStatusRequest true "Status Payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Router /v1/admin/support/trials/{id} [patch]
func (h *supportHandler) UpdateTrialStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid UUID", 400, "error", err.Error()))
		return
	}

	var req modelDto.UpdateTrialRequestStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	if err := h.service.UpdateTrialStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to update status", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Trial status updated successfully", 200, "success", nil))
}

// @Summary Get All Support Messages
// @Description List all inbound support messages
// @Tags Support
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]modelDto.SupportMessageResponse}
// @Router /v1/admin/support/inbox [get]
func (h *supportHandler) GetAllSupportMessages(c *gin.Context) {
	res, err := h.service.GetAllSupportMessages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch support messages", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Support messages fetched successfully", 200, "success", res))
}

// @Summary Update Support Status
// @Description Update support message status
// @Tags Support
// @Accept json
// @Produce json
// @Param id path string true "Message ID (UUID)"
// @Param body body modelDto.UpdateSupportMessageStatusRequest true "Status Payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Router /v1/admin/support/inbox/{id} [patch]
func (h *supportHandler) UpdateSupportStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid UUID", 400, "error", err.Error()))
		return
	}

	var req modelDto.UpdateSupportMessageStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	if err := h.service.UpdateSupportStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to update status", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Support status updated successfully", 200, "success", nil))
}

// @Summary Get All Provisioning Tickets
// @Description List all activation tickets (Superadmin only)
// @Tags Provisioning
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]modelDto.ProvisioningTicketResponse}
// @Router /v1/admin/support/provisioning [get]
func (h *supportHandler) GetAllProvisioningTickets(c *gin.Context) {
	res, err := h.service.GetAllProvisioningTickets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch provisioning tickets", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Provisioning tickets fetched successfully", 200, "success", res))
}

// @Summary Execute Provisioning
// @Description Trigger the automated tenant setup (Superadmin only)
// @Tags Provisioning
// @Produce json
// @Param id path string true "Ticket ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Router /v1/admin/support/provisioning/{id}/execute [post]
func (h *supportHandler) ExecuteProvisioning(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid UUID", 400, "error", err.Error()))
		return
	}

	adminID := c.MustGet("user_id").(uint)

	if err := h.service.ExecuteProvisioning(c.Request.Context(), id, adminID); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Provisioning failed", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Provisioning completed successfully", 200, "success", nil))
}

// @Summary Create Support Message
// @Description Send a support message (Auth required: Any tenant user)
// @Tags Support
// @Accept json
// @Produce json
// @Param body body modelDto.CreateSupportMessageRequest true "Support Message Payload"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=modelDto.SupportMessageResponse}
// @Router /v1/support/message [post]
func (h *supportHandler) CreateSupportMessage(c *gin.Context) {
	var req modelDto.CreateSupportMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	userID := c.MustGet("user_id").(uint)

	res, err := h.service.CreateSupportMessage(c.Request.Context(), tenantID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to send support message", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Support message sent successfully", 201, "success", res))
}
