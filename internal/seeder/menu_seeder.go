package seeder

import (
	"log"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func upsertMenu(db *gorm.DB, m *model.Menu) {
	var existing model.Menu
	err := db.Where("key = ?", m.Key).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		if err := db.Create(m).Error; err != nil {
			log.Printf("⚠️ Gagal create menu %s: %v", m.Key, err)
		}
	} else {
		// Update existing record with new values
		m.ID = existing.ID
		if err := db.Save(m).Error; err != nil {
			log.Printf("⚠️ Gagal update menu %s: %v", m.Key, err)
		}
	}
}

func SeedMenus(db *gorm.DB) {
	// 1. Group: Platform Control (SaaS Level)
	platformGroup := model.Menu{
		Key:          "platform-group",
		Label:        "Platform Control",
		Icon:         "ShieldCheck",
		AllowedRoles: []string{"SUPERADMIN"},
		IsSystem:     true,
		SortOrder:    1,
	}
	upsertMenu(db, &platformGroup)

	platformChildren := []model.Menu{
		{ParentID: &platformGroup.ID, Key: "manage-tenants", Label: "Tenant Directory", Icon: "Building2", Path: "/admin/tenants", AllowedRoles: []string{"SUPERADMIN"}, SortOrder: 1},
		{ParentID: &platformGroup.ID, Key: "subscriptions", Label: "Global Billing", Icon: "CreditCard", Path: "/admin/subscriptions", AllowedRoles: []string{"SUPERADMIN"}, SortOrder: 2},
		{ParentID: &platformGroup.ID, Key: "accounts", Label: "Platform Admins", Icon: "UserCheck", Path: "/admin/accounts", AllowedRoles: []string{"SUPERADMIN"}, SortOrder: 3},
		{ParentID: &platformGroup.ID, Key: "platform-roles", Label: "System Governance", Icon: "ShieldAlert", Path: "/admin/roles", AllowedRoles: []string{"SUPERADMIN"}, Permission: "platform.roles.view", SortOrder: 4},
		{ParentID: &platformGroup.ID, Key: "support-desk", Label: "Support Desk", Icon: "MessageSquare", Path: "/admin/support", AllowedRoles: []string{"SUPERADMIN"}, Permission: "support.manage", SortOrder: 5},
	}
	for _, c := range platformChildren {
		upsertMenu(db, &c)
	}

	// 2. Group: Intelligence Hub
	intelGroup := model.Menu{
		Key:          "intelligence-group",
		Label:        "Intelligence Hub",
		Icon:         "LayoutDashboard",
		AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR", "FINANCE", "EMPLOYEE"},
		SortOrder:    2,
	}
	upsertMenu(db, &intelGroup)

	intelChildren := []model.Menu{
		{ParentID: &intelGroup.ID, Key: "dashboard", Label: "Main Dashboard", Icon: "LayoutDashboard", Path: "/", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR", "FINANCE", "EMPLOYEE"}, Module: "attendance", SortOrder: 1},
		{ParentID: &intelGroup.ID, Key: "analytics", Label: "Workforce Intel", Icon: "TrendingUp", Path: "/analytics", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR", "FINANCE"}, Permission: "analytics.view", Module: "analytics", SortOrder: 2},
	}
	for _, c := range intelChildren {
		upsertMenu(db, &c)
	}

	// 3. Group: Workforce Management
	workforceGroup := model.Menu{
		Key:          "workforce-group",
		Label:        "Workforce Management",
		Icon:         "Users",
		AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR"},
		SortOrder:    3,
	}
	upsertMenu(db, &workforceGroup)

	workforceChildren := []model.Menu{
		{ParentID: &workforceGroup.ID, Key: "all-employees", Label: "Staff Directory", Icon: "Users", Path: "/employees", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR"}, Permission: "user.view", Module: "user", SortOrder: 1},
		{ParentID: &workforceGroup.ID, Key: "all-attendance", Label: "Attendance Logs", Icon: "CalendarDays", Path: "/attendances", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR"}, Permission: "attendance.view", Module: "attendance", SortOrder: 2},
		{ParentID: &workforceGroup.ID, Key: "work-schedules", Label: "Shift Rosters", Icon: "Clock", Path: "/schedules", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR"}, Permission: "schedule.view", Module: "schedule", SortOrder: 3},
		{ParentID: &workforceGroup.ID, Key: "manage-leaves", Label: "Leave Approvals", Icon: "CalendarX", Path: "/leaves", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR"}, Permission: "leave.view", Module: "leave", SortOrder: 4},
		{ParentID: &workforceGroup.ID, Key: "manage-overtime", Label: "Overtime Desk", Icon: "Clock", Path: "/overtime", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR"}, Permission: "overtime.view", Module: "overtime", SortOrder: 5},
	}
	for _, c := range workforceChildren {
		upsertMenu(db, &c)
	}

	// 4. Group: Performance & Ops
	perfGroup := model.Menu{
		Key:          "performance-group",
		Label:        "Performance & Ops",
		Icon:         "Target",
		AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR"},
		SortOrder:    4,
	}
	upsertMenu(db, &perfGroup)

	perfChildren := []model.Menu{
		{ParentID: &perfGroup.ID, Key: "performance-goals", Label: "Strategic Goals", Icon: "Target", Path: "/performance/goals", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR"}, Permission: "performance.manage", Module: "performance", SortOrder: 1},
		{ParentID: &perfGroup.ID, Key: "performance-appraisals", Label: "Staff Appraisals", Icon: "Star", Path: "/performance/appraisals", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR", "EMPLOYEE"}, Permission: "performance.view", Module: "performance", SortOrder: 2},
		{ParentID: &perfGroup.ID, Key: "projects", Label: "Project Tracker", Icon: "Briefcase", Path: "/projects", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR"}, Permission: "project.manage", Module: "project", SortOrder: 3},
		{ParentID: &perfGroup.ID, Key: "timesheet-monitoring", Label: "Timesheet Audit", Icon: "ActivityIcon", Path: "/timesheet/monitoring", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR"}, Permission: "project.manage", Module: "project", SortOrder: 4},
	}
	for _, c := range perfChildren {
		upsertMenu(db, &c)
	}

	// 5. Group: Financial Center
	financeGroup := model.Menu{
		Key:          "financial-group",
		Label:        "Financial Center",
		Icon:         "Coins",
		AllowedRoles: []string{"SUPERADMIN", "ADMIN", "FINANCE"},
		SortOrder:    5,
	}
	upsertMenu(db, &financeGroup)

	financeChildren := []model.Menu{
		{ParentID: &financeGroup.ID, Key: "payroll-list", Label: "Payroll Ledger", Icon: "FileText", Path: "/payroll", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "FINANCE"}, Permission: "payroll.view", Module: "payroll", SortOrder: 1},
		{ParentID: &financeGroup.ID, Key: "payroll-calc", Label: "Salary Engine", Icon: "Calculator", Path: "/payroll/calculator", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "FINANCE"}, Permission: "payroll.calculate", Module: "payroll", SortOrder: 2},
		{ParentID: &financeGroup.ID, Key: "expenses", Label: "Claims & Expenses", Icon: "Receipt", Path: "/finance/expenses", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "FINANCE"}, Permission: "expense.view", Module: "finance", SortOrder: 3},
		{ParentID: &financeGroup.ID, Key: "loans", Label: "Employee Loans", Icon: "Landmark", Path: "/finance/loans", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "FINANCE"}, Permission: "loan.view", Module: "finance", SortOrder: 4},
	}
	for _, c := range financeChildren {
		upsertMenu(db, &c)
	}

	// 6. Group: Organization Governance
	govGroup := model.Menu{
		Key:          "governance-group",
		Label:        "Organization Control",
		Icon:         "Settings",
		AllowedRoles: []string{"SUPERADMIN", "ADMIN"},
		SortOrder:    6,
	}
	upsertMenu(db, &govGroup)

	govChildren := []model.Menu{
		{ParentID: &govGroup.ID, Key: "tenant-info", Label: "Tenant Info", Icon: "Info", Path: "/tenant-settings/info", AllowedRoles: []string{"SUPERADMIN", "ADMIN"}, Permission: "tenant.view", SortOrder: 1},
		{ParentID: &govGroup.ID, Key: "tenant-settings-general", Label: "General Policies", Icon: "Building2", Path: "/tenant-settings", AllowedRoles: []string{"SUPERADMIN", "ADMIN"}, Permission: "tenant.settings.view", SortOrder: 2},
		{ParentID: &govGroup.ID, Key: "tenant-settings-billing", Label: "Plans & Billing", Icon: "CreditCard", Path: "/tenant-settings/billing", AllowedRoles: []string{"SUPERADMIN", "ADMIN"}, Permission: "billing.manage", SortOrder: 3},
		{ParentID: &govGroup.ID, Key: "company-calendar", Label: "Holiday Calendar", Icon: "Calendar", Path: "/tenant-settings/calendar", AllowedRoles: []string{"SUPERADMIN", "ADMIN"}, Permission: "calendar.manage", SortOrder: 4},
		{ParentID: &govGroup.ID, Key: "employee-lifecycle", Label: "Lifecycle Master", Icon: "ListChecks", Path: "/tenant-settings/lifecycle", AllowedRoles: []string{"SUPERADMIN", "ADMIN"}, Permission: "lifecycle.manage", SortOrder: 5},
		{ParentID: &govGroup.ID, Key: "tenant-roles", Label: "Roles & Access", Icon: "ShieldAlert", Path: "/tenant-settings/roles", AllowedRoles: []string{"SUPERADMIN", "ADMIN"}, Permission: "role.view", SortOrder: 6},
	}
	for _, c := range govChildren {
		upsertMenu(db, &c)
	}

	// 7. Group: My Personal Hub
	personalGroup := model.Menu{
		Key:          "personal-group",
		Label:        "My Personal Hub",
		Icon:         "UserCog",
		AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR", "FINANCE", "EMPLOYEE", "SUPPORT", "ENGINEER"},
		SortOrder:    7,
	}
	upsertMenu(db, &personalGroup)

	personalChildren := []model.Menu{
		{ParentID: &personalGroup.ID, Key: "my-leaves", Label: "Leave Request", Icon: "CalendarX", Path: "/leaves", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR", "FINANCE", "EMPLOYEE", "SUPPORT", "ENGINEER"}, Module: "leave", SortOrder: 1},
		{ParentID: &personalGroup.ID, Key: "my-overtime", Label: "Overtime Desk", Icon: "Clock", Path: "/overtime", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR", "FINANCE", "EMPLOYEE", "SUPPORT", "ENGINEER"}, Module: "overtime", SortOrder: 2},
		{ParentID: &personalGroup.ID, Key: "my-payroll", Label: "My Salary & Slips", Icon: "Wallet", Path: "/payroll", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR", "FINANCE", "EMPLOYEE", "SUPPORT", "ENGINEER"}, Module: "payroll", SortOrder: 3},
		{ParentID: &personalGroup.ID, Key: "my-timesheet", Label: "My Timesheet", Icon: "ActivityIcon", Path: "/timesheet", AllowedRoles: []string{"SUPERADMIN", "ADMIN", "HR", "FINANCE", "EMPLOYEE", "SUPPORT", "ENGINEER"}, Module: "project", SortOrder: 4},
	}
	for _, c := range personalChildren {
		upsertMenu(db, &c)
	}

	log.Println("Seeder: Menus seeded successfully")
}
