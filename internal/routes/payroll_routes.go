package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterPayrollRoutes(rg *gin.RouterGroup, h handler.PayrollHandler) {
	payroll := rg.Group("/payroll")
	{
		payroll.POST("/calculate", h.Calculate)
		payroll.POST("/generate", h.Generate)
		payroll.GET("", h.GetList)
		payroll.GET("/summary", h.GetSummary)
		payroll.PATCH("/:id/publish", h.Publish)

		payroll.GET("/employee/:user_id/baseline", h.GetBaseline)
		payroll.GET("/employee/:user_id/attendance-sync", h.SyncAttendance)
		payroll.POST("/employee/:user_id/save", h.SaveIndividual)
	}

	myPayroll := rg.Group("/my-payroll")
	{
		myPayroll.GET("/profile", h.GetMyPayrollProfile)
		myPayroll.GET("/slips", h.GetMySlip)
		myPayroll.GET("/history", h.GetMyPayrolls)
	}

	adminUsers := rg.Group("/admin/users")
	adminUsers.Use(middleware.RequireRole("superadmin", "admin", "hr"))
	{
		adminUsers.GET("/:id/payroll-profile", h.GetPayrollProfile)
		adminUsers.PUT("/:id/payroll-profile", h.UpdatePayrollProfile)
	}
}
