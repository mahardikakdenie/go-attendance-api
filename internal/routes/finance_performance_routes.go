package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterFinancePerformanceRoutes(rg *gin.RouterGroup, financeH handler.FinanceHandler, perfH handler.PerformanceHandler) {
	finance := rg.Group("/finance")
	{
		finance.GET("/expenses", financeH.GetAllExpenses)
		finance.GET("/expenses/summary", financeH.GetSummary)
		finance.POST("/expenses", financeH.CreateExpense)
		finance.PATCH("/expenses/:id/approve", middleware.RequireRole("superadmin", "admin", "finance"), financeH.ApproveExpense)
		finance.PATCH("/expenses/:id/reject", middleware.RequireRole("superadmin", "admin", "finance"), financeH.RejectExpense)
		finance.PATCH("/quotas/:id", middleware.RequireRole("superadmin", "admin", "finance"), financeH.UpdateQuota)
	}

	perf := rg.Group("/performance")
	{
		perf.GET("/goals/me", perfH.GetMyGoals)
		perf.GET("/goals/user/:userId", perfH.GetUserGoals)
		perf.POST("/goals", perfH.CreateGoal)
		perf.PUT("/goals/:id/progress", perfH.UpdateGoalProgress)

		perf.GET("/cycles", perfH.GetAllCycles)
		perf.GET("/appraisals/cycle/:cycleId", perfH.GetAppraisalsByCycle)
		perf.PUT("/appraisals/:id/self-review", perfH.SubmitSelfReview)
	}
}
