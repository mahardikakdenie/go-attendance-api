package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"go-attendance-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client) service.CalendarCronService {
	// Initialize handlers (keep the same as original)
	handlers, cronService, authService := initHandlers(db, rdb)

	if gin.Mode() != gin.ReleaseMode {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := r.Group("/api/v1")

	// Public routes
	RegisterAuthRoutes(api, handlers.Auth)
	api.POST("/public/trial-request", handlers.Support.CreateTrialRequest)
	api.GET("/ping", handlers.Attendance.HelloTest)
	api.POST("/email/test", handler.SendEmailTest)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.SecureAuth(authService))
	{
		RegisterAuthenticatedAuthRoutes(protected, handlers.Auth)
		RegisterAttendanceRoutes(protected, handlers.Attendance, handlers.Correction)
		RegisterOvertimeRoutes(protected, handlers.Overtime)
		RegisterTenantRoutes(protected, handlers.Tenant, handlers.TenantSetting)
		RegisterUserRoutes(protected, handlers.User, handlers.UCR, handlers.Media)
		RegisterPayrollRoutes(protected, handlers.Payroll)
		RegisterDashboardRoutes(protected, handlers.Dashboard)
		RegisterLeaveRoutes(protected, handlers.Leave)
		RegisterMiscRoutes(protected, handlers.Activity)
		RegisterOrgRoutes(protected, handlers.Org, handlers.TenantRole, handlers.Superadmin)
		RegisterHrOpsRoutes(protected, handlers.HrOps)
		RegisterSupportRoutes(protected, handlers.Support)
		RegisterSubscriptionRoutes(protected, handlers.Subscription)
		RegisterFinancePerformanceRoutes(protected, handlers.Finance, handlers.Performance)
		RegisterTimesheetRoutes(protected, handlers.Timesheet)
		RegisterSuperadminRoutes(protected, handlers.Superadmin, handlers.Tenant, handlers.Subscription)
	}

	return cronService
}
