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
		m.ID = existing.ID
		if err := db.Save(m).Error; err != nil {
			log.Printf("⚠️ Gagal update menu %s: %v", m.Key, err)
		}
	}
}

func perm(id string) *string {
	if id == "" {
		return nil
	}
	return &id
}

func SeedMenus(db *gorm.DB) {
	// 1) Platform Control
	platformGroup := model.Menu{
		Key:                "platform-group",
		Label:              "Platform Control",
		Icon:               "ShieldCheck",
		IsSystem:           true,
		SortOrder:          1,
		RequiredPermission: nil, // group container, visible if children pass filter
	}
	upsertMenu(db, &platformGroup)

	platformChildren := []model.Menu{
		{ParentID: &platformGroup.ID, Key: "manage-tenants", Label: "Tenant Directory", Icon: "Building2", Path: "/admin/tenants", SortOrder: 1, RequiredPermission: perm("superadmin.access")},
		{ParentID: &platformGroup.ID, Key: "subscriptions", Label: "Global Billing", Icon: "CreditCard", Path: "/admin/subscriptions", SortOrder: 2, RequiredPermission: perm("superadmin.access")},
		{ParentID: &platformGroup.ID, Key: "accounts", Label: "Platform Admins", Icon: "UserCheck", Path: "/admin/accounts", SortOrder: 3, RequiredPermission: perm("superadmin.access")},
		{ParentID: &platformGroup.ID, Key: "platform-menus", Label: "Menu Management", Icon: "LayoutGrid", Path: "/admin/menus", IsSystem: true, SortOrder: 4, RequiredPermission: perm("superadmin.access")},
		{ParentID: &platformGroup.ID, Key: "platform-roles", Label: "System Governance", Icon: "ShieldAlert", Path: "/admin/roles", SortOrder: 5, RequiredPermission: perm("superadmin.access")},
		{ParentID: &platformGroup.ID, Key: "support-desk", Label: "Support Desk", Icon: "MessageSquare", Path: "/admin/support", SortOrder: 6, RequiredPermission: perm("superadmin.access")},
	}
	for _, c := range platformChildren {
		upsertMenu(db, &c)
	}

	// 2) Intelligence Hub
	intelGroup := model.Menu{Key: "intelligence-group", Label: "Intelligence Hub", Icon: "LayoutDashboard", SortOrder: 2}
	upsertMenu(db, &intelGroup)

	intelChildren := []model.Menu{
		{ParentID: &intelGroup.ID, Key: "dashboard", Label: "Main Dashboard", Icon: "LayoutDashboard", Path: "/", Module: "attendance", SortOrder: 1, RequiredPermission: nil},
		{ParentID: &intelGroup.ID, Key: "analytics", Label: "Workforce Intel", Icon: "TrendingUp", Path: "/analytics", Module: "analytics", SortOrder: 2, RequiredPermission: perm("analytics.executive")},
	}
	for _, c := range intelChildren {
		upsertMenu(db, &c)
	}

	// 3) Workforce Management
	workforceGroup := model.Menu{Key: "workforce-group", Label: "Workforce Management", Icon: "Users", SortOrder: 3}
	upsertMenu(db, &workforceGroup)

	workforceChildren := []model.Menu{
		{ParentID: &workforceGroup.ID, Key: "all-employees", Label: "Staff Directory", Icon: "Users", Path: "/employees", Module: "user", SortOrder: 1, RequiredPermission: perm("employee.view")},
		{ParentID: &workforceGroup.ID, Key: "all-attendance", Label: "Attendance Logs", Icon: "CalendarDays", Path: "/attendances", Module: "attendance", SortOrder: 2, RequiredPermission: perm("attendance.view")},
		{ParentID: &workforceGroup.ID, Key: "work-schedules", Label: "Shift Rosters", Icon: "Clock", Path: "/schedules", Module: "schedule", SortOrder: 3, RequiredPermission: perm("schedule.view")},
		{ParentID: &workforceGroup.ID, Key: "manage-leaves", Label: "Leave Approvals", Icon: "CalendarX", Path: "/leaves", Module: "leave", SortOrder: 4, RequiredPermission: perm("leave.view")},
		{ParentID: &workforceGroup.ID, Key: "manage-overtime", Label: "Overtime Desk", Icon: "Clock", Path: "/overtime", Module: "overtime", SortOrder: 5, RequiredPermission: perm("overtime.view")},
	}
	for _, c := range workforceChildren {
		upsertMenu(db, &c)
	}

	// 4) Performance & Ops
	perfGroup := model.Menu{Key: "performance-group", Label: "Performance & Ops", Icon: "Target", SortOrder: 4}
	upsertMenu(db, &perfGroup)

	perfChildren := []model.Menu{
		{ParentID: &perfGroup.ID, Key: "performance-goals", Label: "Strategic Goals", Icon: "Target", Path: "/performance/goals", Module: "performance", SortOrder: 1, RequiredPermission: perm("performance.manage")},
		{ParentID: &perfGroup.ID, Key: "performance-appraisals", Label: "Staff Appraisals", Icon: "Star", Path: "/performance/appraisals", Module: "performance", SortOrder: 2, RequiredPermission: perm("performance.view")},
		{ParentID: &perfGroup.ID, Key: "projects", Label: "Project Tracker", Icon: "Briefcase", Path: "/projects", Module: "project", SortOrder: 3, RequiredPermission: perm("project.manage")},
		{ParentID: &perfGroup.ID, Key: "timesheet-monitoring", Label: "Timesheet Audit", Icon: "ActivityIcon", Path: "/timesheet/monitoring", Module: "project", SortOrder: 4, RequiredPermission: perm("project.manage")},
	}
	for _, c := range perfChildren {
		upsertMenu(db, &c)
	}

	// 5) Financial Center
	financeGroup := model.Menu{Key: "financial-group", Label: "Financial Center", Icon: "Coins", SortOrder: 5}
	upsertMenu(db, &financeGroup)

	financeChildren := []model.Menu{
		{ParentID: &financeGroup.ID, Key: "payroll-list", Label: "Payroll Ledger", Icon: "FileText", Path: "/payroll", Module: "payroll", SortOrder: 1, RequiredPermission: perm("payroll.view")},
		{ParentID: &financeGroup.ID, Key: "payroll-calc", Label: "Salary Engine", Icon: "Calculator", Path: "/payroll/calculator", Module: "payroll", SortOrder: 2, RequiredPermission: perm("payroll.calculate")},
		{ParentID: &financeGroup.ID, Key: "expenses", Label: "Claims & Expenses", Icon: "Receipt", Path: "/finance/expenses", Module: "finance", SortOrder: 3, RequiredPermission: perm("expense.view")},
		{ParentID: &financeGroup.ID, Key: "loans", Label: "Employee Loans", Icon: "Landmark", Path: "/finance/loans", Module: "finance", SortOrder: 4, RequiredPermission: perm("loan.view")},
	}
	for _, c := range financeChildren {
		upsertMenu(db, &c)
	}

	// 6) Organization Governance
	govGroup := model.Menu{Key: "governance-group", Label: "Organization Control", Icon: "Settings", SortOrder: 6}
	upsertMenu(db, &govGroup)

	govChildren := []model.Menu{
		{ParentID: &govGroup.ID, Key: "tenant-info", Label: "Tenant Info", Icon: "Info", Path: "/tenant-settings/info", SortOrder: 1, RequiredPermission: perm("settings.manage")},
		{ParentID: &govGroup.ID, Key: "tenant-settings-general", Label: "General Policies", Icon: "Building2", Path: "/tenant-settings", SortOrder: 2, RequiredPermission: perm("settings.manage")},
		{ParentID: &govGroup.ID, Key: "tenant-settings-billing", Label: "Plans & Billing", Icon: "CreditCard", Path: "/tenant-settings/billing", SortOrder: 3, RequiredPermission: perm("billing.view")},
		{ParentID: &govGroup.ID, Key: "company-calendar", Label: "Holiday Calendar", Icon: "Calendar", Path: "/tenant-settings/calendar", SortOrder: 4, RequiredPermission: perm("settings.manage")},
		{ParentID: &govGroup.ID, Key: "employee-lifecycle", Label: "Lifecycle Master", Icon: "ListChecks", Path: "/tenant-settings/lifecycle", SortOrder: 5, RequiredPermission: perm("lifecycle.manage")},
		{ParentID: &govGroup.ID, Key: "tenant-roles", Label: "Roles & Access", Icon: "ShieldAlert", Path: "/tenant-settings/roles", SortOrder: 6, RequiredPermission: perm("rbac.access")},
	}
	for _, c := range govChildren {
		upsertMenu(db, &c)
	}

	// 7) My Personal Hub
	personalGroup := model.Menu{Key: "personal-group", Label: "My Personal Hub", Icon: "UserCog", SortOrder: 7}
	upsertMenu(db, &personalGroup)

	personalChildren := []model.Menu{
		{ParentID: &personalGroup.ID, Key: "my-leaves", Label: "Leave Request", Icon: "CalendarX", Path: "/leaves", Module: "leave", SortOrder: 1, RequiredPermission: perm("leave.view")},
		{ParentID: &personalGroup.ID, Key: "my-overtime", Label: "Overtime Desk", Icon: "Clock", Path: "/overtime", Module: "overtime", SortOrder: 2, RequiredPermission: perm("overtime.view")},
		{ParentID: &personalGroup.ID, Key: "my-payroll", Label: "My Salary & Slips", Icon: "Wallet", Path: "/payroll", Module: "payroll", SortOrder: 3, RequiredPermission: perm("payroll.view")},
		{ParentID: &personalGroup.ID, Key: "my-timesheet", Label: "My Timesheet", Icon: "ActivityIcon", Path: "/timesheet", Module: "project", SortOrder: 4, RequiredPermission: perm("timesheet.view")},
		{ParentID: &personalGroup.ID, Key: "my-support", Label: "Helpdesk", Icon: "LifeBuoy", Path: "/support", Module: "support", SortOrder: 5, RequiredPermission: perm("support.access")},
	}
	for _, c := range personalChildren {
		upsertMenu(db, &c)
	}

	log.Println("Seeder: Menus seeded successfully (required_permission mode)")
}
