package routes

import (
	"context"
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/service"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Handlers struct {
	Auth          handler.AuthHandler
	User          handler.UserHandler
	Tenant        handler.TenantHandler
	TenantSetting handler.TenantSettingHandler
	Media         *handler.MediaHandler
	Attendance    handler.AttendanceHandler
	Org           handler.OrganizationHandler
	Leave         handler.LeaveHandler
	Overtime      handler.OvertimeHandler
	UCR           handler.UserChangeRequestHandler
	Payroll       handler.PayrollHandler
	Dashboard     handler.DashboardHandler
	Activity      handler.ActivityHandler
	TenantRole    handler.TenantRoleHandler
	Support       handler.SupportHandler
	Correction    handler.AttendanceCorrectionHandler
	Finance       handler.FinanceHandler
	Superadmin    handler.SuperadminHandler
	Subscription  handler.SubscriptionHandler
	Timesheet     handler.TimesheetHandler
	HrOps         handler.HrOpsHandler
	Performance   handler.PerformanceHandler
}

func initHandlers(db *gorm.DB, rdb *redis.Client) (*Handlers, service.CalendarCronService, service.AuthService) {
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
	payrollRepo := repository.NewPayrollRepository(db)
	supportRepo := repository.NewSupportRepository(db)
	correctionRepo := repository.NewAttendanceCorrectionRepository(db)
	performanceRepo := repository.NewPerformanceRepository(db)
	superadminRepo := repository.NewSuperadminRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)

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
	payrollService := service.NewPayrollService(payrollRepo, userRepo, tenantRepo, tenantSettingRepo, attendanceRepo, leaveRepo, userPayrollProfileRepo, overtimeRepo, hrOpsRepo)
	dashboardService := service.NewDashboardService(tenantRepo, userRepo, attendanceRepo, leaveRepo, overtimeRepo, timesheetRepo, rdb)
	tenantRoleService := service.NewTenantRoleService(roleRepo, permissionRepo, hierarchyRepo)
	supportService := service.NewSupportService(supportRepo, tenantRepo, userRepo, roleRepo, subscriptionRepo, tenantSettingRepo, userPayrollProfileRepo)
	timesheetService := service.NewTimesheetService(timesheetRepo, userRepo)
	correctionService := service.NewAttendanceCorrectionService(correctionRepo, attendanceRepo, userRepo, activityRepo)
	performanceService := service.NewPerformanceService(performanceRepo, userRepo)
	superadminService := service.NewSuperadminService(superadminRepo, userRepo, roleRepo, permissionRepo, activityRepo)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, tenantRepo)
	expenseService := service.NewExpenseService(expenseRepo, userRepo, activityRepo)
	hrOpsService := service.NewHrOpsService(hrOpsRepo, userRepo, leaveRepo, tenantSettingRepo)
	calendarCronService := service.NewCalendarCronService(hrOpsRepo, userRepo)

	// Handlers
	handlers := &Handlers{
		Auth:          handler.NewAuthHandler(authService),
		User:          handler.NewUserHandler(userService),
		Tenant:        handler.NewTenantHandler(tenantService),
		TenantSetting: handler.NewTenantSettingHandler(tenantSettingService),
		Media:         handler.NewMediaHandler(mediaService),
		Attendance:    handler.NewAttendanceHandler(attendanceService),
		Org:           handler.NewOrganizationHandler(orgService),
		Leave:         handler.NewLeaveHandler(leaveService),
		Overtime:      handler.NewOvertimeHandler(overtimeService),
		UCR:           handler.NewUserChangeRequestHandler(ucrService),
		Payroll:       handler.NewPayrollHandler(payrollService),
		Dashboard:     handler.NewDashboardHandler(dashboardService),
		Activity:      handler.NewActivityHandler(userService, leaveService, overtimeService),
		TenantRole:    handler.NewTenantRoleHandler(tenantRoleService),
		Support:       handler.NewSupportHandler(supportService),
		Correction:    handler.NewAttendanceCorrectionHandler(correctionService),
		Finance:       handler.NewFinanceHandler(expenseService),
		Superadmin:    handler.NewSuperadminHandler(superadminService),
		Subscription:  handler.NewSubscriptionHandler(subscriptionService),
		Timesheet:     handler.NewTimesheetHandler(timesheetService),
		HrOps:         handler.NewHrOpsHandler(hrOpsService),
		Performance:   handler.NewPerformanceHandler(performanceService),
	}

	return handlers, calendarCronService, authService
}
