package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterBillingRoutes(rg *gin.RouterGroup, h handler.BillingHandler) {
	billing := rg.Group("/billing")
	billing.Use(middleware.RequireRole("superadmin", "admin"))
	{
		billing.GET("/invoices", h.GetInvoices)
		billing.GET("/invoices/:id/pdf", h.DownloadInvoicePDF)
	}
}
