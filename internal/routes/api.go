package routes

import (
	"context"
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client) service.CalendarCronService {
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
	positionRepo.FindAll(context.TODO(), 0) // dummy to avoid unused
	permissionRepo := repository.NewPermissionRepository(db)
	hierarchyRepo := repository.NewRoleHierarchyRepository(db)
	hrOpsRepo := repository.NewHrOpsRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	userPayrollProfileRepo := repository.NewUserPayrollProfileRepository(db)
	timesheetRepo := repository.NewTimesheetRepository(db)

	// Services
	authService := service.NewAuthService(authRepo, activityRepo)
	userService := service.NewUserService(userRepo, roleRepo, activityRepo, hierarchyRepo, hrOpsRepo, leaveRepo, userPayrollProfileRepo)
	tenantService := service.NewTenantService(tenantRepo, subscriptionRepo)
	tenantSettingService := service.NewTenantSettingService(tenantSettingRepo)
	mediaService := service.NewMediaService(mediaRepo)
	attendanceService := service.NewAttendanceService(attendanceRepo, userRepo, tenantSettingRepo, tenantRepo, activityRepo, hrOpsRepo, leaveRepo, userService, rdb)
	orgService := service.NewOrganizationService(userRepo, leaveRepo, positionRepo)
	leaveService := service.NewLeaveService(leaveRepo, activityRepo, userRepo, orgService, userService, rdb)
	overtimeService := service.NewOvertimeService(overtimeRepo, userService)
	ucrService := service.NewUserChangeRequestService(ucrRepo, userRepo)
	payrollRepo := repository.NewPayrollRepository(db)
	payrollService := service.NewPayrollService(payrollRepo, userRepo, tenantRepo, tenantSettingRepo, attendanceRepo, leaveRepo, userPayrollProfileRepo)
	dashboardService := service.NewDashboardService(tenantRepo, userRepo, attendanceRepo, leaveRepo, overtimeRepo, timesheetRepo, rdb)
	tenantRoleService := service.NewTenantRoleService(roleRepo, permissionRepo, hierarchyRepo)
	supportRepo := repository.NewSupportRepository(db)
	supportService := service.NewSupportService(supportRepo, tenantRepo, userRepo, roleRepo, subscriptionRepo, tenantSettingRepo, userPayrollProfileRepo)
	correctionRepo := repository.NewAttendanceCorrectionRepository(db)
	timesheetService := service.NewTimesheetService(timesheetRepo, userRepo)

	correctionService := service.NewAttendanceCorrectionService(correctionRepo, attendanceRepo, userRepo, activityRepo)
	performanceRepo := repository.NewPerformanceRepository(db)
	performanceService := service.NewPerformanceService(performanceRepo, userRepo)

	superadminRepo := repository.NewSuperadminRepository(db)
	superadminService := service.NewSuperadminService(superadminRepo, userRepo, roleRepo, permissionRepo, activityRepo)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, tenantRepo)

	expenseRepo := repository.NewExpenseRepository(db)
	expenseService := service.NewExpenseService(expenseRepo, userRepo, activityRepo)

	calendarCronService := service.NewCalendarCronService(hrOpsRepo, userRepo)

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
	correctionHandler := handler.NewAttendanceCorrectionHandler(correctionService)
	financeHandler := handler.NewFinanceHandler(expenseService)
	superadminHandler := handler.NewSuperadminHandler(superadminService)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)
	timesheetHandler := handler.NewTimesheetHandler(timesheetService)
	hrOpsService := service.NewHrOpsService(hrOpsRepo, userRepo, leaveRepo, tenantSettingRepo)
	hrOpsHandler := handler.NewHrOpsHandler(hrOpsService)
	performanceHandler := handler.NewPerformanceHandler(performanceService)
	if gin.Mode() != gin.ReleaseMode {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := r.Group("/api/v1")

	// Public routes
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/reset-password", authHandler.ResetPassword)
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
		protected.POST("/auth/change-password", authHandler.ChangePassword)

		attendance := protected.Group("/attendance")
		{
			attendance.POST("", attendanceHandler.RecordAttendance)
			attendance.GET("", attendanceHandler.GetAllAttendance)
			attendance.GET("/history", attendanceHandler.GetAttendanceHistory)
			attendance.GET("/summary", attendanceHandler.GetAttendanceSummary)
			attendance.GET("/today", attendanceHandler.GetTodayAttendance)

			// Correction requests
			corrections := attendance.Group("/corrections")
			{
				corrections.POST("", correctionHandler.RequestCorrection)
				corrections.GET("", correctionHandler.GetCorrections)
				corrections.POST("/:id/approve", middleware.RequireRole("superadmin", "admin", "hr"), correctionHandler.ApproveCorrection)
				corrections.POST("/:id/reject", middleware.RequireRole("superadmin", "admin", "hr"), correctionHandler.RejectCorrection)
			}
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
			payroll.POST("/generate", payrollHandler.Generate)
			payroll.GET("", payrollHandler.GetList)
			payroll.GET("/summary", payrollHandler.GetSummary)
			payroll.PATCH("/:id/publish", payrollHandler.Publish)

			// Individual Extensions
			payroll.GET("/employee/:user_id/baseline", payrollHandler.GetBaseline)
			payroll.GET("/employee/:user_id/attendance-sync", payrollHandler.SyncAttendance)
			payroll.POST("/employee/:user_id/save", payrollHandler.SaveIndividual)
		}

		// Self Service Payroll
		myPayroll := protected.Group("/my-payroll")
		{
			myPayroll.GET("/profile", payrollHandler.GetMyPayrollProfile)
			myPayroll.GET("/slips", payrollHandler.GetMySlip)
			myPayroll.GET("/history", payrollHandler.GetMyPayrolls)
		}

		// Admin side user profile management
		adminUsers := protected.Group("/admin/users")
		adminUsers.Use(middleware.RequireRole("superadmin", "admin", "hr"))
		{
			adminUsers.GET("/:id/payroll-profile", payrollHandler.GetPayrollProfile)
			adminUsers.PUT("/:id/payroll-profile", payrollHandler.UpdatePayrollProfile)
		}

		dashboards := protected.Group("/dashboards")
		{
			dashboards.GET("/admin", middleware.RequireBaseRole(model.BaseRoleSuperAdmin), dashboardHandler.GetAdminDashboard)
			dashboards.GET("/hr", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), dashboardHandler.GetHrDashboard)
			dashboards.GET("/hr/daily-pulse", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), dashboardHandler.GetDailyPulse)
			dashboards.GET("/hr/employee-dna/:user_id", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), dashboardHandler.GetEmployeeDNA)
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
			hrOps.GET("/lifecycle-templates",
				middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin,
					model.BaseRoleHR), hrOpsHandler.GetLifecycleTemplates)
			hrOps.POST("/lifecycle-templates",
				middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin,
					model.BaseRoleHR), hrOpsHandler.CreateLifecycleTemplate)
			hrOps.DELETE("/lifecycle-templates/:id",
				middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin,
					model.BaseRoleHR), hrOpsHandler.DeleteLifecycleTemplate)
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

		// Subscription (Admin/Owner side)
		subs := protected.Group("/subscriptions")
		subs.Use(middleware.RequireRole("superadmin", "admin"))
		{
			subs.GET("/me", subscriptionHandler.GetMySubscription)
			subs.POST("/upgrade", subscriptionHandler.UpgradeSubscription)
		}

		// Finance Module
		finance := protected.Group("/finance")
		{
			finance.GET("/expenses", financeHandler.GetAllExpenses)
			finance.GET("/expenses/summary", financeHandler.GetSummary)
			finance.POST("/expenses", financeHandler.CreateExpense)
			finance.PATCH("/expenses/:id/approve", middleware.RequireRole("superadmin", "admin", "finance"), financeHandler.ApproveExpense)
			finance.PATCH("/expenses/:id/reject", middleware.RequireRole("superadmin", "admin", "finance"), financeHandler.RejectExpense)
			finance.PATCH("/quotas/:id", middleware.RequireRole("superadmin", "admin", "finance"), financeHandler.UpdateQuota)
		}

		// Performance Module
		perf := protected.Group("/performance")
		{
			perf.GET("/goals/me", performanceHandler.GetMyGoals)
			perf.GET("/goals/user/:userId", performanceHandler.GetUserGoals)
			perf.POST("/goals", performanceHandler.CreateGoal)
			perf.PUT("/goals/:id/progress", performanceHandler.UpdateGoalProgress)

			perf.GET("/cycles", performanceHandler.GetAllCycles)
			perf.GET("/appraisals/cycle/:cycleId", performanceHandler.GetAppraisalsByCycle)
			perf.PUT("/appraisals/:id/self-review", performanceHandler.SubmitSelfReview)
		}

		// Timesheet & Project Module
		timesheet := protected.Group("/timesheet")
		{
			// Employee endpoints
			timesheet.POST("/entries", timesheetHandler.CreateEntry)
			timesheet.GET("/me/report", timesheetHandler.GetMyReport)
			timesheet.POST("/tasks", timesheetHandler.CreateTask)
			timesheet.GET("/tasks", timesheetHandler.GetTasks)
			
			// Project endpoints (Accessible by all for selection)
			timesheet.GET("/projects", timesheetHandler.GetProjects)
			timesheet.GET("/projects/:id/members", timesheetHandler.GetMembers)

			// HR/Admin endpoints
			adminTimesheet := timesheet.Group("/admin")
			adminTimesheet.Use(middleware.RequireRole("superadmin", "admin", "hr"))
			{
				// Project Management
				adminTimesheet.POST("/projects", timesheetHandler.CreateProject)
				adminTimesheet.PUT("/projects/:id", timesheetHandler.UpdateProject)
				adminTimesheet.DELETE("/projects/:id", timesheetHandler.DeleteProject)
				
				// Member Management
				adminTimesheet.POST("/projects/:id/members", timesheetHandler.AssignMembers)
				adminTimesheet.DELETE("/projects/:id/members/:user_id", timesheetHandler.RemoveMember)

				// Report Management
				adminTimesheet.GET("/report/employee/:user_id", timesheetHandler.GetEmployeeReport)
			}
		}

		// Backward Compatibility / Explicit Project Endpoints as per Issue
		projects := protected.Group("/projects")
		{
			projects.GET("", middleware.HasPermission("user.view"), timesheetHandler.GetProjects)
			projects.POST("", middleware.HasPermission("project.manage"), timesheetHandler.CreateProject)
			projects.PUT("/:id", middleware.HasPermission("project.manage"), timesheetHandler.UpdateProject)
			projects.DELETE("/:id", middleware.HasPermission("project.manage"), timesheetHandler.DeleteProject)
			
			projects.POST("/:id/members", middleware.HasPermission("project.manage"), timesheetHandler.AssignMembers)
			projects.DELETE("/:id/members/:user_id", middleware.HasPermission("project.manage"), timesheetHandler.RemoveMember)
			projects.GET("/:id/members", middleware.HasPermission("user.view"), timesheetHandler.GetMembers)
		}

		// Superadmin specialized routes
		superadmin := protected.Group("/superadmin")
		superadmin.Use(middleware.RequireBaseRole(model.BaseRoleSuperAdmin))
		{
			superadmin.GET("/owners-stats", superadminHandler.GetOwnersWithStats)
			superadmin.PUT("/tenants/:id", tenantHandler.UpdateTenant)

			// Platform Accounts Management
			platform := superadmin.Group("/platform-accounts")
			{
				platform.GET("", superadminHandler.GetPlatformAccounts)
				platform.POST("", superadminHandler.CreatePlatformAccount)
				platform.PUT("/:id", superadminHandler.UpdatePlatformAccount)
				platform.PATCH("/:id/status", superadminHandler.TogglePlatformAccountStatus)
			}

			// System Roles & Permissions
			roles := superadmin.Group("/system-roles")
			{
				roles.GET("", superadminHandler.ListSystemRoles)
				roles.POST("", superadminHandler.CreateSystemRole)
				roles.PUT("/:id", superadminHandler.UpdateSystemRole)
				roles.DELETE("/:id", superadminHandler.DeleteSystemRole)
			}
			superadmin.GET("/permissions", superadminHandler.ListAllPermissions)

			// Subscription & Billing
			subscriptions := superadmin.Group("/subscriptions")
			{
				subscriptions.GET("", subscriptionHandler.GetSubscriptions)
				subscriptions.POST("/:id/remind", subscriptionHandler.RemindTenant)
				subscriptions.POST("/:id/suspend", subscriptionHandler.SuspendTenant)
			}
		}
	}

	return calendarCronService
}
