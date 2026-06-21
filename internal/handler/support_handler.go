package handler

import (
	"net/http"
	"strconv"

	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
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
	GetSupportAgents(c *gin.Context)
	UpdateSupportStatus(c *gin.Context)
	BulkUpdateSupportMessages(c *gin.Context)
	BulkAssignSupport(c *gin.Context)
	UpdateSupportReadState(c *gin.Context)
	AssignSupportAgent(c *gin.Context)

	// Superadmin Only
	GetAllProvisioningTickets(c *gin.Context)
	ExecuteProvisioning(c *gin.Context)

	// Tenant User
	CreateSupportMessage(c *gin.Context)
	GetSupportCategories(c *gin.Context)
	GetSupportPriorities(c *gin.Context)
	GetUserSupportHistory(c *gin.Context)
	ReplyToTicketByUser(c *gin.Context)
	CreateReply(c *gin.Context)
	GetReplies(c *gin.Context)
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
// @Description List all inbound support messages with server-side filters
// @Tags Support
// @Produce json
// @Param search query string false "Search subject/message/sender/tenant"
// @Param category query string false "TECHNICAL|BILLING|FEATURE|OTHER"
// @Param status query string false "PENDING|IN_PROGRESS|RESOLVED|CLOSED"
// @Param limit query int false "Pagination limit"
// @Param offset query int false "Pagination offset"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]modelDto.SupportMessageResponse}
// @Router /v1/admin/support/inbox [get]
func (h *supportHandler) GetAllSupportMessages(c *gin.Context) {
	var req modelDto.SupportInboxFilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid query params", 400, "error", err.Error()))
		return
	}

	limit := req.Limit
	offset := req.Offset
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	filter := model.SupportMessageFilter{
		Search:   req.Search,
		Category: req.Category,
		Status:   req.Status,
		Limit:    limit,
		Offset:   offset,
	}

	res, total, err := h.service.GetAllSupportMessages(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch support messages", 500, "error", err.Error()))
		return
	}

	currentPage := (offset / limit) + 1
	lastPage := int((total + int64(limit) - 1) / int64(limit))
	if lastPage == 0 {
		lastPage = 1
	}

	pagination := utils.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: currentPage,
		LastPage:    lastPage,
	}

	c.JSON(http.StatusOK, utils.BuildResponseWithPagination("Support messages fetched successfully", 200, "success", res, pagination))
}

// @Summary Bulk Update Support Inbox
// @Description Bulk actions for support inbox (mark read/unread, resolve, assign)
// @Tags Support
// @Accept json
// @Produce json
// @Param body body modelDto.BulkSupportInboxRequest true "Bulk Action Payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Router /v1/admin/support/inbox/bulk [patch]
func (h *supportHandler) BulkUpdateSupportMessages(c *gin.Context) {
	var req modelDto.BulkSupportInboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	if err := h.service.BulkUpdateSupportMessages(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Failed to apply bulk action", 400, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Bulk action applied successfully", 200, "success", nil))
}

// @Summary Update Support Read State
// @Description Mark ticket read/unread
// @Tags Support
// @Accept json
// @Produce json
// @Param id path string true "Message ID (UUID)"
// @Param body body modelDto.UpdateSupportReadStateRequest true "Read state payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Router /v1/admin/support/inbox/{id}/read-state [patch]
func (h *supportHandler) UpdateSupportReadState(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid UUID", 400, "error", err.Error()))
		return
	}

	var req modelDto.UpdateSupportReadStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	if err := h.service.UpdateSupportReadState(c.Request.Context(), id, req.IsRead); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to update read state", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Read state updated successfully", 200, "success", nil))
}

// @Summary Assign Support Agent
// @Description Assign agent to support ticket
// @Tags Support
// @Accept json
// @Produce json
// @Param id path string true "Message ID (UUID)"
// @Param body body modelDto.AssignSupportAgentRequest true "Assign payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=modelDto.SupportMessageResponse}
// @Router /v1/admin/support/inbox/{id}/assign [patch]
func (h *supportHandler) AssignSupportAgent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid UUID", 400, "error", err.Error()))
		return
	}

	var req modelDto.AssignSupportAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	if req.AgentID == 0 {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid agent_id", 400, "error", "agent_id must be > 0"))
		return
	}

	updatedMsg, err := h.service.AssignSupportAgent(c.Request.Context(), id, req.AgentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Failed to assign agent", 400, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Support ticket assigned successfully", 200, "success", updatedMsg))
}

// @Summary Get Assignable Support Agents
// @Description List all assignable support agents (active HQ users)
// @Tags Support
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]model.UserResponse}
// @Router /v1/admin/support/agents [get]
func (h *supportHandler) GetSupportAgents(c *gin.Context) {
	search := c.Query("search")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	results, total, err := h.service.GetSupportAgents(c.Request.Context(), search, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch support agents", 500, "error", err.Error()))
		return
	}

	pagination := utils.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: (offset / limit) + 1,
		LastPage:    int((total + int64(limit) - 1) / int64(limit)),
	}

	c.JSON(http.StatusOK, utils.BuildResponseWithPagination("Support agents fetched successfully", 200, "success", results, pagination))
}

// @Summary Bulk Assign Support Tickets
// @Description Assign multiple support tickets to a specific agent
// @Tags Support
// @Accept json
// @Produce json
// @Param body body modelDto.BulkAssignSupportRequest true "Bulk Assign Payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=modelDto.BulkAssignResponse}
// @Router /v1/admin/support/inbox/bulk-assign [patch]
func (h *supportHandler) BulkAssignSupport(c *gin.Context) {
	var req modelDto.BulkAssignSupportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	res, err := h.service.BulkAssignSupport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Failed to bulk assign tickets", 400, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Bulk assign completed", 200, "success", res))
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

// @Summary Get User Support History
// @Description List all support tickets sent by the logged-in user
// @Tags Support
// @Produce json
// @Param search query string false "Search subject/message"
// @Param status query string false "PENDING|IN_PROGRESS|RESOLVED|CLOSED"
// @Param limit query int false "Pagination limit"
// @Param offset query int false "Pagination offset"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]modelDto.UserSupportHistoryResponse}
// @Router /v1/support/history [get]
func (h *supportHandler) GetUserSupportHistory(c *gin.Context) {
	var req modelDto.UserSupportHistoryFilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid query params", 400, "error", err.Error()))
		return
	}

	userID := c.MustGet("user_id").(uint)

	limit := req.Limit
	offset := req.Offset
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	filter := model.SupportMessageFilter{
		Search: req.Search,
		Status: req.Status,
		Limit:  limit,
		Offset: offset,
	}

	res, total, err := h.service.GetUserSupportHistory(c.Request.Context(), userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch history", 500, "error", err.Error()))
		return
	}

	pagination := utils.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: (offset / limit) + 1,
		LastPage:    int((total + int64(limit) - 1) / int64(limit)),
	}

	c.JSON(http.StatusOK, utils.BuildResponseWithPagination("Support history fetched successfully", 200, "success", res, pagination))
}

// @Summary Reply to Ticket (User)
// @Description Allow user to reply to their own active support ticket
// @Tags Support
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID (UUID)"
// @Param body body modelDto.UserReplySupportRequest true "Reply Payload"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=modelDto.SupportReplyResponse}
// @Router /v1/support/tickets/{id}/reply [post]
func (h *supportHandler) ReplyToTicketByUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid UUID", 400, "error", err.Error()))
		return
	}

	var req modelDto.UserReplySupportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	userID := c.MustGet("user_id").(uint)

	res, err := h.service.ReplyToTicketByUser(c.Request.Context(), userID, id, req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Failed to send reply", 400, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Reply sent successfully", 201, "success", res))
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

// @Summary Get Support Ticket Categories
// @Description List of available categories and priorities for support tickets
// @Tags Support
// @Produce json
// @Success 200 {object} utils.APIResponse{data=[]modelDto.SupportCategoryInfo}
// @Router /v1/tickets/categories [get]
func (h *supportHandler) GetSupportCategories(c *gin.Context) {
	res := h.service.GetSupportCategories(c.Request.Context())
	c.JSON(http.StatusOK, utils.BuildResponse("Support categories fetched", 200, "success", res))
}

// @Summary Get Support Ticket Priorities
// @Description List of available priority levels for support tickets
// @Tags Support
// @Produce json
// @Success 200 {object} utils.APIResponse{data=[]modelDto.SupportPriorityInfo}
// @Router /v1/tickets/priorities [get]
func (h *supportHandler) GetSupportPriorities(c *gin.Context) {
	res := h.service.GetSupportPriorities(c.Request.Context())
	c.JSON(http.StatusOK, utils.BuildResponse("Support priorities fetched", 200, "success", res))
}

// @Summary Reply to Support Message
// @Description Send a reply to a support ticket
// @Tags Support
// @Accept json
// @Produce json
// @Param id path string true "Message ID (UUID)"
// @Param body body modelDto.CreateSupportReplyRequest true "Reply Payload"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=modelDto.SupportReplyResponse}
// @Router /v1/support/message/{id}/reply [post]
func (h *supportHandler) CreateReply(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid UUID", 400, "error", err.Error()))
		return
	}

	var req modelDto.CreateSupportReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	userID := c.MustGet("user_id").(uint)

	res, err := h.service.CreateReply(c.Request.Context(), userID, id, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to send reply", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Reply sent successfully", 201, "success", res))
}

// @Summary Get Replies for Message
// @Description List all replies for a support ticket
// @Tags Support
// @Produce json
// @Param id path string true "Message ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]modelDto.SupportReplyResponse}
// @Router /v1/support/message/{id}/replies [get]
func (h *supportHandler) GetReplies(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid UUID", 400, "error", err.Error()))
		return
	}

	res, err := h.service.GetReplies(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch replies", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Replies fetched successfully", 200, "success", res))
}
