package handler

import (
	"fmt"
	"net/http"
	"strconv"

	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type BillingHandler interface {
	GetInvoices(c *gin.Context)
	DownloadInvoicePDF(c *gin.Context)
	UploadTransferProof(c *gin.Context)
	VerifyInvoice(c *gin.Context)
}

type billingHandler struct {
	service service.BillingService
}

func NewBillingHandler(service service.BillingService) BillingHandler {
	return &billingHandler{service: service}
}

// @Summary Get Invoices
// @Description Get paginated list of invoices for current tenant
// @Tags Billing
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Limit per page"
// @Param status query string false "Filter by status"
// @Success 200 {object} utils.APIResponse
// @Router /api/v1/billing/invoices [get]
func (h *billingHandler) GetInvoices(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	data, total, err := h.service.GetInvoices(c.Request.Context(), tenantID, page, limit, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch invoices", 500, "error", err.Error()))
		return
	}

	pagination := utils.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: page,
		LastPage:    int((total + int64(limit) - 1) / int64(limit)),
	}

	c.JSON(http.StatusOK, utils.BuildResponseWithPagination("Invoices retrieved successfully", 200, "success", data, pagination))
}

// @Summary Download Invoice PDF
// @Description Download PDF version of a specific invoice
// @Tags Billing
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Success 200 {file} binary
// @Router /api/v1/billing/invoices/{id}/pdf [get]
func (h *billingHandler) DownloadInvoicePDF(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	invoiceID := c.Param("id")

	pdfBytes, err := h.service.GenerateInvoicePDF(c.Request.Context(), tenantID, invoiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildErrorResponse("Invoice not found", 404, "error", nil))
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=invoice-%s.pdf", invoiceID))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// @Summary Upload Transfer Proof
// @Description Upload proof of payment transfer for a specific invoice
// @Tags Billing
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Param body body modelDto.UploadProofRequest true "Transfer Proof Data"
// @Success 200 {object} utils.APIResponse
// @Router /api/v1/billing/invoices/{id}/proof [post]
// @Router /api/v1/billing/invoices/{id}/proof [put]
func (h *billingHandler) UploadTransferProof(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	userID := c.MustGet("user_id").(uint)
	invoiceID := c.Param("id")

	var req modelDto.UploadProofRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	err := h.service.UploadTransferProof(c.Request.Context(), tenantID, userID, invoiceID, req.TransferProofURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to save transfer proof", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Transfer proof uploaded successfully", 200, "success", nil))
}

// @Summary Verify Invoice Payment
// @Description Verify payment proof and activate subscription
// @Tags Superadmin Billing
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Success 200 {object} utils.APIResponse
// @Router /api/v1/superadmin/billing/invoices/{id}/verify [post]
func (h *billingHandler) VerifyInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	err := h.service.VerifyInvoice(c.Request.Context(), invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to verify payment proof", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Invoice verified successfully, subscription reactivated", 200, "success", nil))
}
