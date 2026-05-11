package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type BillingHandler interface {
	GetInvoices(c *gin.Context)
	DownloadInvoicePDF(c *gin.Context)
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
