package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"go-attendance-api/internal/model"

	"github.com/gin-gonic/gin"
)

func RegisterBillingRoutes(rg *gin.RouterGroup, h handler.BillingHandler) {
	billing := rg.Group("/billing")
	billing.Use(middleware.RequireBaseRole(model.BaseRoleAdmin, model.BaseRoleFinance))
	{
		billing.GET("/invoices", h.GetInvoices)
		billing.GET("/invoices/:id/pdf", h.DownloadInvoicePDF)
		billing.POST("/invoices/:id/proof", h.UploadTransferProof)
	}

	superadmin := rg.Group("/superadmin/billing")
	superadmin.Use(middleware.RequireBaseRole(model.BaseRoleSuperAdmin))
	{
		superadmin.POST("/invoices/:id/verify", h.VerifyInvoice)
	}
}
