package handler

import (
	"net/http"
	"strconv"

	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type FinanceHandler interface {
	GetAllExpenses(c *gin.Context)
	GetSummary(c *gin.Context)
	CreateExpense(c *gin.Context)
	ApproveExpense(c *gin.Context)
	RejectExpense(c *gin.Context)
	UpdateQuota(c *gin.Context)
}

type financeHandler struct {
	service service.ExpenseService
}

func NewFinanceHandler(service service.ExpenseService) FinanceHandler {
	return &financeHandler{service: service}
}

// ... rest of the methods ...

// @Summary Update User Expense Quota
// @Tags Finance
// @Param id path int true "User ID"
// @Param body body modelDto.UpdateQuotaRequest true "Quota Payload"
// @Security BearerAuth
// @Router /api/v1/finance/quotas/{id} [patch]
func (h *financeHandler) UpdateQuota(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("id"))
	adminID := c.MustGet("user_id").(uint)

	var req modelDto.UpdateQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	if err := h.service.UpdateQuota(c.Request.Context(), uint(userID), req.Quota, adminID); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to update quota", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("User quota updated successfully", 200, "success", nil))
}

// @Summary Get All Expenses
// @Tags Finance
// @Produce json
// @Param status query string false "Pending, Approved, Rejected"
// @Param search query string false "ID or Name"
// @Param page query int false "Page number"
// @Param limit query int false "Limit per page"
// @Security BearerAuth
// @Router /api/v1/finance/expenses [get]
func (h *financeHandler) GetAllExpenses(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	status := c.Query("status")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	filter := model.ExpenseFilter{
		TenantID: tenantID,
		Status:   model.ExpenseStatus(status),
		Search:   search,
		Limit:    limit,
		Offset:   offset,
	}

	// Hierarchical Scoping if needed could be added here
	
	expenses, total, err := h.service.GetAllExpenses(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch expenses", 500, "error", err.Error()))
		return
	}

	res := modelDto.ExpenseListResponse{
		Data: expenses,
		Meta: modelDto.PaginationMeta{
			Total:       total,
			CurrentPage: page,
			LastPage:    (int(total) + limit - 1) / limit,
		},
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Expenses fetched successfully", 200, "success", res))
}

// @Summary Get Expense Summary
// @Tags Finance
// @Produce json
// @Security BearerAuth
// @Router /api/v1/finance/expenses/summary [get]
func (h *financeHandler) GetSummary(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	res, err := h.service.GetSummary(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch summary", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Summary fetched successfully", 200, "success", res))
}

// @Summary Create Expense Claim
// @Tags Finance
// @Accept json
// @Produce json
// @Param body body modelDto.CreateExpenseRequest true "Expense Payload"
// @Security BearerAuth
// @Router /api/v1/finance/expenses [post]
func (h *financeHandler) CreateExpense(c *gin.Context) {
	var req modelDto.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	userID := c.MustGet("user_id").(uint)
	tenantID := c.MustGet("tenant_id").(uint)

	res, err := h.service.SubmitExpense(c.Request.Context(), userID, tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to submit expense", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Expense submitted successfully", 201, "success", res))
}

// @Summary Approve Expense Claim
// @Tags Finance
// @Param id path int true "Expense ID"
// @Security BearerAuth
// @Router /api/v1/finance/expenses/{id}/approve [patch]
func (h *financeHandler) ApproveExpense(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	adminID := c.MustGet("user_id").(uint)

	if err := h.service.ApproveExpense(c.Request.Context(), uint(id), adminID); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to approve expense", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Expense approved successfully", 200, "success", nil))
}

// @Summary Reject Expense Claim
// @Tags Finance
// @Param id path int true "Expense ID"
// @Param body body modelDto.RejectExpenseRequest true "Reject Payload"
// @Security BearerAuth
// @Router /api/v1/finance/expenses/{id}/reject [patch]
func (h *financeHandler) RejectExpense(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	adminID := c.MustGet("user_id").(uint)

	var req modelDto.RejectExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	if err := h.service.RejectExpense(c.Request.Context(), uint(id), adminID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to reject expense", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Expense rejected successfully", 200, "success", nil))
}
