package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {

	authRepo := repository.NewAuthRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	activityRepo := repository.NewRecentActivityRepository(db)

	authService := service.NewAuthService(authRepo, activityRepo)
	authHandler := handler.NewAuthHandler(authService)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, roleRepo, activityRepo)
	userHandler := handler.NewUserHandler(userService)

	tenantRepo := repository.NewTenantRepository(db)
	tenantService := service.NewTenantService(tenantRepo)
	tenantHandler := handler.NewTenantHandler(tenantService)

	tenantSettingRepo := repository.NewTenantSettingRepository(db)
	tenantSettingService := service.NewTenantSettingService(tenantSettingRepo)
	tenantSettingHandler := handler.NewTenantSettingHandler(tenantSettingService)

	mediaRepo := repository.NewMediaRepository(db)
	mediaService := service.NewMediaService(mediaRepo)
	mediaHandler := handler.NewMediaHandler(mediaService)

	attendanceRepo := repository.NewAttendanceRepository(db)
	attendanceService := service.NewAttendanceService(attendanceRepo, userRepo, tenantSettingRepo, tenantRepo, activityRepo)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService)

	ucrRepo := repository.NewUserChangeRequestRepository(db)
	ucrService := service.NewUserChangeRequestService(ucrRepo, userRepo)
	ucrHandler := handler.NewUserChangeRequestHandler(ucrService)

	overtimeRepo := repository.NewOvertimeRepository(db)
	overtimeService := service.NewOvertimeService(overtimeRepo)
	overtimeHandler := handler.NewOvertimeHandler(overtimeService)

	leaveRepo := repository.NewLeaveRepository(db)
	leaveService := service.NewLeaveService(leaveRepo)
	leaveHandler := handler.NewLeaveHandler(leaveService)

	payrollService := service.NewPayrollService()
	payrollHandler := handler.NewPayrollHandler(payrollService)

	activityHandler := handler.NewActivityHandler(userService, leaveService, overtimeService)

	if gin.Mode() != gin.ReleaseMode {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	api.GET("/ping", attendanceHandler.HelloTest)
	api.POST("/email/test", handler.SendEmailTest)

	protected := api.Group("")
	protected.Use(middleware.SecureAuth(authService))
	{
		protected.GET("/auth/sessions", authHandler.GetSessions)

		attendance := protected.Group("/attendance")
		{
			attendance.POST("", attendanceHandler.RecordAttendance)
			attendance.GET("", attendanceHandler.GetAllAttendance)
			attendance.GET("/history", attendanceHandler.GetAttendanceHistory)
			attendance.GET("/summary", attendanceHandler.GetAttendanceSummary)
			attendance.GET("/today", attendanceHandler.GetTodayAttendance)
		}

		overtime := protected.Group("/overtime")
		{
			overtime.POST("", overtimeHandler.CreateRequest)
			overtime.GET("", overtimeHandler.GetAll)
			overtime.GET("/:id", overtimeHandler.GetByID)

			adminOnly := overtime.Group("")
			adminOnly.Use(middleware.RequireRole("superadmin", "admin", "hr"))
			{
				adminOnly.POST("/approve/:id", overtimeHandler.ApproveRequest)
				adminOnly.POST("/reject/:id", overtimeHandler.RejectRequest)
			}
		}

		tenants := protected.Group("/tenants")
		{
			// Only superadmin can list all or create tenants
			tenants.GET("", middleware.RequireRole("superadmin"), tenantHandler.GetAllTenant)
			tenants.POST("", middleware.RequireRole("superadmin"), tenantHandler.CreateTenant)
			
			// GetByID can be accessed by admin/hr but with ownership check in handler
			tenants.GET("/:id", tenantHandler.GetTenantByID)
		}

		tenantSetting := protected.Group("/tenant-setting")
		{
			tenantSetting.GET("", tenantSettingHandler.GetSetting)
			tenantSetting.PUT("", tenantSettingHandler.UpdateSetting)
		}

		users := protected.Group("/users")
		{
			users.GET("", userHandler.GetAllUsers)
			users.GET("/me", userHandler.GetMe)
			users.GET("/me/activities", userHandler.GetRecentActivities)
			users.POST("", middleware.RequireRole("superadmin", "admin"), userHandler.CreateUser)

			// ✅ NEW: update profile photo
			users.PUT("/profile-photo", userHandler.UpdateProfilePhoto)

			// ✅ NEW: Request Change System
			users.POST("/request-change", ucrHandler.CreateRequest)
			
			adminOnly := users.Group("")
			adminOnly.Use(middleware.RequireRole("admin", "hr"))
			{
				adminOnly.GET("/pending-changes", ucrHandler.GetPendingRequests)
				adminOnly.POST("/approve-change/:id", ucrHandler.ApproveRequest)
				adminOnly.POST("/reject-change/:id", ucrHandler.RejectRequest)
			}
		}

		protected.GET("/employees", userHandler.GetAllUsers)

		payroll := protected.Group("/payroll")
		{
			payroll.POST("/calculate", payrollHandler.Calculate)
		}

		protected.POST("/media/upload", mediaHandler.Upload)
		protected.POST("/auth/logout", authHandler.Logout)

		// Leave routes
		leaves := api.Group("/leaves")
		leaves.Use(middleware.SecureAuth(authService))
		{
			leaves.POST("/request", leaveHandler.RequestLeave)
			leaves.GET("", leaveHandler.GetLeaveHistory)
			leaves.GET("/balances", leaveHandler.GetLeaveBalances)
		}

		// Activity routes
		activities := api.Group("/activities")
		activities.Use(middleware.SecureAuth(authService))
		{
			activities.GET("/recent", activityHandler.GetRecentActivities)
			activities.GET("/quick-info", activityHandler.GetQuickInfo)
		}
	}
}
