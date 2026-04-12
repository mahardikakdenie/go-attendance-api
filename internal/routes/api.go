package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client) {
	// Repositories
	authRepo := repository.NewAuthRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	tenantSettingRepo := repository.NewTenantSettingRepository(db)
	mediaRepo := repository.NewMediaRepository(db)
	activityRepo := repository.NewRecentActivityRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	leaveRepo := repository.NewLeaveRepository(db)
	overtimeRepo := repository.NewOvertimeRepository(db)
	ucrRepo := repository.NewUserChangeRequestRepository(db)
	positionRepo := repository.NewPositionRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	hierarchyRepo := repository.NewRoleHierarchyRepository(db)
	hrOpsRepo := repository.NewHrOpsRepository(db)

	// Services
	authService := service.NewAuthService(authRepo, activityRepo)
	userService := service.NewUserService(userRepo, roleRepo, activityRepo, hierarchyRepo)
	tenantService := service.NewTenantService(tenantRepo)
	tenantSettingService := service.NewTenantSettingService(tenantSettingRepo)
	mediaService := service.NewMediaService(mediaRepo)
	attendanceService := service.NewAttendanceService(attendanceRepo, userRepo, tenantSettingRepo, tenantRepo, activityRepo, hrOpsRepo, userService, rdb)
	orgService := service.NewOrganizationService(userRepo, leaveRepo, positionRepo)
	leaveService := service.NewLeaveService(leaveRepo, activityRepo, userRepo, orgService, userService, rdb)
	overtimeService := service.NewOvertimeService(overtimeRepo, userService)
	ucrService := service.NewUserChangeRequestService(ucrRepo, userRepo)
	payrollService := service.NewPayrollService()
	dashboardService := service.NewDashboardService(tenantRepo, userRepo, attendanceRepo, leaveRepo, overtimeRepo, rdb)
	tenantRoleService := service.NewTenantRoleService(roleRepo, permissionRepo, hierarchyRepo)
	supportRepo := repository.NewSupportRepository(db)
	supportService := service.NewSupportService(supportRepo, tenantRepo, userRepo, roleRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	tenantHandler := handler.NewTenantHandler(tenantService)
	tenantSettingHandler := handler.NewTenantSettingHandler(tenantSettingService)
	mediaHandler := handler.NewMediaHandler(mediaService)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService)
	orgHandler := handler.NewOrganizationHandler(orgService)
	leaveHandler := handler.NewLeaveHandler(leaveService)
	overtimeHandler := handler.NewOvertimeHandler(overtimeService)
	ucrHandler := handler.NewUserChangeRequestHandler(ucrService)
	payrollHandler := handler.NewPayrollHandler(payrollService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	activityHandler := handler.NewActivityHandler(userService, leaveService, overtimeService)
	tenantRoleHandler := handler.NewTenantRoleHandler(tenantRoleService)
	supportHandler := handler.NewSupportHandler(supportService)
	hrOpsService := service.NewHrOpsService(hrOpsRepo, userRepo)
	hrOpsHandler := handler.NewHrOpsHandler(hrOpsService)

	if gin.Mode() != gin.ReleaseMode {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := r.Group("/api/v1")

	// Public routes
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	api.POST("/public/trial-request", supportHandler.CreateTrialRequest)

	api.GET("/ping", attendanceHandler.HelloTest)
	api.POST("/email/test", handler.SendEmailTest)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.SecureAuth(authService))
	{
		protected.GET("/auth/sessions", authHandler.GetSessions)
		protected.POST("/auth/logout", authHandler.Logout)

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

			overtime.POST("/approve/:id", middleware.RequireRole("superadmin", "admin", "hr"), overtimeHandler.ApproveRequest)
			overtime.POST("/reject/:id", middleware.RequireRole("superadmin", "admin", "hr"), overtimeHandler.RejectRequest)
		}

		tenants := protected.Group("/tenants")
		{
			tenants.GET("", middleware.RequireRole("superadmin"), tenantHandler.GetAllTenant)
			tenants.POST("", middleware.RequireRole("superadmin"), tenantHandler.CreateTenant)
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
			users.PUT("/profile-photo", userHandler.UpdateProfilePhoto)
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

		dashboards := protected.Group("/dashboards")
		{
			dashboards.GET("/admin", middleware.RequireBaseRole(model.BaseRoleSuperAdmin), dashboardHandler.GetAdminDashboard)
			dashboards.GET("/hr", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), dashboardHandler.GetHrDashboard)
			dashboards.GET("/hr/daily-pulse", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), dashboardHandler.GetDailyPulse)
			dashboards.GET("/finance", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleFinance), dashboardHandler.GetFinanceDashboard)
			dashboards.GET("/heatmap", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), dashboardHandler.GetHeatmap)
		}


		protected.POST("/media/upload", mediaHandler.Upload)

		leaves := protected.Group("/leaves")
		{
			leaves.POST("/request", leaveHandler.RequestLeave)
			leaves.GET("", leaveHandler.GetLeaveHistory)
			leaves.GET("/balances", leaveHandler.GetLeaveBalances)
			leaves.POST("/approve/:id", middleware.RequireRole("superadmin", "admin", "hr"), leaveHandler.ApproveLeave)
			leaves.POST("/reject/:id", middleware.RequireRole("superadmin", "admin", "hr"), leaveHandler.RejectLeave)
		}

		activities := protected.Group("/activities")
		{
			activities.GET("/recent", activityHandler.GetRecentActivities)
			activities.GET("/quick-info", activityHandler.GetQuickInfo)
		}

		org := protected.Group("/organization")
		{
			org.GET("/chart", orgHandler.GetOrgTree)
			org.GET("/positions", orgHandler.GetPositions)
			org.POST("/positions", middleware.RequireRole("superadmin", "admin"), orgHandler.CreatePosition)
		}

		// Custom Tenant Roles
		tenantRoles := protected.Group("/tenant-roles")
		{
			tenantRoles.GET("", middleware.RequireRole("superadmin", "admin"), tenantRoleHandler.ListRoles)
			tenantRoles.POST("", middleware.RequireRole("superadmin", "admin"), tenantRoleHandler.CreateRole)
			tenantRoles.PATCH("/:id", middleware.RequireRole("superadmin", "admin"), tenantRoleHandler.UpdateRole)
			tenantRoles.DELETE("/:id", middleware.RequireRole("superadmin", "admin"), tenantRoleHandler.DeleteRole)
			tenantRoles.GET("/:id/hierarchy", middleware.RequireRole("superadmin", "admin"), tenantRoleHandler.GetHierarchy)
			tenantRoles.POST("/hierarchy", middleware.RequireRole("superadmin", "admin"), tenantRoleHandler.SaveHierarchy)
		}

		// HR Advanced Operations
		hrOps := protected.Group("/hr")
		{
			hrOps.GET("/shifts", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), hrOpsHandler.GetAllShifts)
			hrOps.POST("/shifts", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), hrOpsHandler.CreateShift)
			hrOps.GET("/roster", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), hrOpsHandler.GetWeeklyRoster)
			hrOps.POST("/roster/save", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), hrOpsHandler.SaveRoster)
			hrOps.GET("/calendar", hrOpsHandler.GetHolidays)
			hrOps.POST("/calendar", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), hrOpsHandler.CreateHoliday)
			hrOps.PUT("/calendar/:id", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), hrOpsHandler.UpdateHoliday)
			hrOps.DELETE("/calendar/:id", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), hrOpsHandler.DeleteHoliday)
			hrOps.GET("/employees/:id/lifecycle", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), hrOpsHandler.GetEmployeeLifecycle)
			hrOps.PATCH("/employees/:id/lifecycle/tasks/:task_id", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), hrOpsHandler.UpdateLifecycleTask)
		}

		// Support & Provisioning
		support := protected.Group("/admin/support")
		support.Use(middleware.RequireTenant(1)) // HQ only
		{
			// Superadmin or any role with support.manage permission in Tenant 1
			support.GET("/inbox", middleware.HasPermission("support.manage"), supportHandler.GetAllSupportMessages)
			support.PATCH("/inbox/:id", middleware.HasPermission("support.manage"), supportHandler.UpdateSupportStatus)
			support.GET("/trials", middleware.HasPermission("support.manage"), supportHandler.GetAllTrialRequests)
			support.PATCH("/trials/:id", middleware.HasPermission("support.manage"), supportHandler.UpdateTrialStatus)

			// Provisioning (Superadmin only)
			support.GET("/provisioning", middleware.RequireBaseRole(model.BaseRoleSuperAdmin), supportHandler.GetAllProvisioningTickets)
			support.POST("/provisioning/:id/execute", middleware.RequireBaseRole(model.BaseRoleSuperAdmin), supportHandler.ExecuteProvisioning)
		}

		// User side support
		protected.POST("/support/message", supportHandler.CreateSupportMessage)
	}
}
