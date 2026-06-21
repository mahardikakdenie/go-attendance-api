package seeder

import (
	"log"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedRoles(db *gorm.DB) {
	// 1. Seed Permissions
	permissions := []model.Permission{
		// Attendance
		{ID: "attendance.view", Module: "attendance", Action: "view"},
		{ID: "attendance.create", Module: "attendance", Action: "create"},
		{ID: "attendance.edit", Module: "attendance", Action: "edit"},
		{ID: "attendance.delete", Module: "attendance", Action: "delete"},
		{ID: "attendance.export", Module: "attendance", Action: "export"},

		// Leave
		{ID: "leave.view", Module: "leave", Action: "view"},
		{ID: "leave.create", Module: "leave", Action: "create"},
		{ID: "leave.approve", Module: "leave", Action: "approve"},
		{ID: "leave.reject", Module: "leave", Action: "reject"},

		// Overtime
		{ID: "overtime.view", Module: "overtime", Action: "view"},
		{ID: "overtime.create", Module: "overtime", Action: "create"},
		{ID: "overtime.approve", Module: "overtime", Action: "approve"},

		// Payroll
		{ID: "payroll.view", Module: "payroll", Action: "view"},
		{ID: "payroll.calculate", Module: "payroll", Action: "calculate"},
		{ID: "payroll.approve", Module: "payroll", Action: "approve"},
		{ID: "payroll.edit", Module: "payroll", Action: "edit"},

		// User Management
		{ID: "user.view", Module: "user", Action: "view"},
		{ID: "user.create", Module: "user", Action: "create"},
		{ID: "user.edit", Module: "user", Action: "edit"},
		{ID: "user.delete", Module: "user", Action: "delete"},
		{ID: "user.view.detail", Module: "user", Action: "view"},

		// Tenant & SaaS
		{ID: "tenant.view", Module: "tenant", Action: "view"},
		{ID: "tenant.edit", Module: "tenant", Action: "edit"},
		{ID: "tenant.settings.view", Module: "tenant", Action: "view"},
		{ID: "subscription.manage", Module: "subscription", Action: "manage"},
		{ID: "billing.manage", Module: "tenant", Action: "manage"},
		{ID: "calendar.manage", Module: "tenant", Action: "manage"},
		{ID: "lifecycle.manage", Module: "tenant", Action: "manage"},
		{ID: "superadmin.access", Module: "superadmin", Action: "access"},
		{ID: "settings.manage", Module: "tenant", Action: "manage"},
		{ID: "billing.view", Module: "tenant", Action: "view"},
		{ID: "employee.view", Module: "user", Action: "view"},

		// RBAC
		{ID: "role.view", Module: "role", Action: "view"},
		{ID: "role.manage", Module: "role", Action: "manage"},
		{ID: "platform.roles.view", Module: "role", Action: "view"},
		{ID: "rbac.manage", Module: "role", Action: "manage"},
		{ID: "rbac.access", Module: "role", Action: "access"},

		// Support Inbox
		{ID: "support.view", Module: "support", Action: "view"},
		{ID: "support.reply", Module: "support", Action: "reply"},
		{ID: "support.assign", Module: "support", Action: "assign"},
		{ID: "support.read_state", Module: "support", Action: "read_state"},
		{ID: "support.status", Module: "support", Action: "status"},
		{ID: "support.bulk_action", Module: "support", Action: "bulk_action"},
		{ID: "support.manage", Module: "support", Action: "manage"},
		{ID: "support.access", Module: "support", Action: "access"},

		// Analytics & Reports
		{ID: "analytics.view", Module: "analytics", Action: "view"},
		{ID: "analytics.executive", Module: "analytics", Action: "view"},

		// Scheduling
		{ID: "schedule.view", Module: "schedule", Action: "view"},

		// Projects
		{ID: "project.view", Module: "project", Action: "view"},
		{ID: "project.manage", Module: "project", Action: "manage"},

		// Timesheets
		{ID: "timesheet.view", Module: "timesheet", Action: "view"},
		{ID: "timesheet.create", Module: "timesheet", Action: "create"},
		{ID: "timesheet.manage", Module: "timesheet", Action: "manage"},

		// Finance
		{ID: "finance.view", Module: "finance", Action: "view"},
		{ID: "finance.manage", Module: "finance", Action: "manage"},
		{ID: "expense.view", Module: "finance", Action: "view"},
		{ID: "loan.view", Module: "finance", Action: "view"},

		// Performance
		{ID: "performance.view", Module: "performance", Action: "view"},
		{ID: "performance.manage", Module: "performance", Action: "manage"},
	}

	for _, p := range permissions {
		db.FirstOrCreate(&p, model.Permission{ID: p.ID})
	}

	// 2. Seed System Roles
	systemRoles := []model.Role{
		{
			Name:        "superadmin",
			Description: "Platform Owner with full access",
			BaseRole:    model.BaseRoleSuperAdmin,
			IsSystem:    true,
			IsImmutable: true,
			IsEditable:  false,
		},
		{
			Name:        "admin",
			Description: "Tenant Owner / Administrator",
			BaseRole:    model.BaseRoleAdmin,
			IsSystem:    true,
			IsImmutable: true,
			IsEditable:  false,
		},
		{
			Name:        "hr",
			Description: "Human Resources Manager",
			BaseRole:    model.BaseRoleHR,
			IsSystem:    true,
			IsImmutable: true,
			IsEditable:  false,
		},
		{
			Name:        "finance",
			Description: "Finance & Payroll Manager",
			BaseRole:    model.BaseRoleFinance,
			IsSystem:    true,
			IsImmutable: true,
			IsEditable:  false,
		},
		{
			Name:        "employee",
			Description: "Regular Employee",
			BaseRole:    model.BaseRoleEmployee,
			IsSystem:    true,
			IsImmutable: false,
			IsEditable:  false,
		},
		{
			Name:        "SUPPORT SYSTEM",
			Description: "Platform Support Agent with access to all support and tenant management tools",
			BaseRole:    model.BaseRoleSuperAdmin,
			IsSystem:    true,
			IsImmutable: true,
			IsEditable:  false,
		},
	}

	for _, r := range systemRoles {
		var role model.Role
		err := db.Where("name = ?", r.Name).First(&role).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&r).Error; err != nil {
				log.Fatalf("Gagal seeder role %s: %v", r.Name, err)
			}
			role = r
			log.Printf("Seeder: Role %s ditambahkan", r.Name)
		} else {
			// Update flags and base_role for existing roles
			db.Model(&role).Updates(map[string]interface{}{
				"base_role":    r.BaseRole,
				"is_system":    r.IsSystem,
				"is_immutable": r.IsImmutable,
				"is_editable":  r.IsEditable,
			})
		}

		// Assign permissions based on role
		if role.BaseRole == model.BaseRoleSuperAdmin || role.BaseRole == model.BaseRoleAdmin {
			// Full Access
			for _, p := range permissions {
				rp := model.RolePermission{RoleID: role.ID, PermissionID: p.ID}
				db.FirstOrCreate(&rp, rp)
			}
		} else if role.Name == "hr" {
			hrPerms := []string{
				"attendance.view", "attendance.export",
				"leave.view", "leave.approve", "leave.reject",
				"overtime.view", "overtime.approve",
				"user.view", "user.create", "user.edit",
				"project.view", "timesheet.manage",
				"performance.manage", "performance.view",
				"support.access", "employee.view",
			}
			// Clear old and re-add or just FirstOrCreate
			for _, pID := range hrPerms {
				rp := model.RolePermission{RoleID: role.ID, PermissionID: pID}
				db.FirstOrCreate(&rp, rp)
			}
		} else if role.Name == "finance" {
			finPerms := []string{
				"payroll.view", "payroll.calculate", "payroll.approve",
				"finance.view", "finance.manage",
				"project.view", "timesheet.view",
				"support.access",
			}
			for _, pID := range finPerms {
				rp := model.RolePermission{RoleID: role.ID, PermissionID: pID}
				db.FirstOrCreate(&rp, rp)
			}
		} else if role.Name == "employee" {
			empPerms := []string{
				"attendance.view", "attendance.create",
				"leave.view", "leave.create",
				"overtime.view", "overtime.create",
				"project.view", "timesheet.view", "timesheet.create",
				"performance.view", "support.access",
			}
			for _, pID := range empPerms {
				rp := model.RolePermission{RoleID: role.ID, PermissionID: pID}
				db.FirstOrCreate(&rp, rp)
			}
		} else if role.Name == "SUPPORT SYSTEM" {
			supportPerms := []string{
				// Support operations
				"support.view", "support.reply", "support.assign",
				"support.read_state", "support.status", "support.bulk_action",
				"support.manage", "support.access",
				// Tenant visibility
				"tenant.view", "tenant.settings.view",
				"subscription.manage", "billing.view",
				"lifecycle.manage",
				// User visibility
				"user.view", "user.view.detail", "employee.view",
				// Platform access
				"superadmin.access", "analytics.view",
				"rbac.access", "platform.roles.view",
			}
			for _, pID := range supportPerms {
				rp := model.RolePermission{RoleID: role.ID, PermissionID: pID}
				db.FirstOrCreate(&rp, rp)
			}
		}
	}

	log.Println("Seeder: Roles and Permissions updated")

	// 3. Seed role_menu_visibility based on role permissions
	for _, r := range systemRoles {
		var role model.Role
		if err := db.Where("name = ?", r.Name).First(&role).Error; err != nil {
			continue
		}
		// Find menus visible to this role via their required_permission -> role's permissions
		insert := db.Exec(`
			INSERT INTO role_menu_visibility (menu_id, role_id)
			SELECT m.id, ?
			FROM menus m
			JOIN role_permissions rp ON m.required_permission = rp.permission_id
			WHERE rp.role_id = ?
			ON CONFLICT (menu_id, role_id) DO NOTHING
		`, role.ID, role.ID)
		if insert.Error != nil {
			log.Printf("⚠️ Gagal seeder role_menu_visibility for %s: %v", r.Name, insert.Error)
		} else if insert.RowsAffected > 0 {
			log.Printf("Seeder:  %d menu visibility mappings for role %s", insert.RowsAffected, r.Name)
		}

		// Also grant visibility to menus with no required_permission (public/dashboard/personal group)
		publicMenus := []string{
			"intelligence-group", "dashboard", "personal-group",
			"my-leaves", "my-overtime", "my-payroll", "my-timesheet", "my-support",
			"workforce-group", "performance-group", "financial-group", "governance-group",
		}
		for _, key := range publicMenus {
			var menu model.Menu
			if err := db.Where("key = ?", key).First(&menu).Error; err != nil {
				continue
			}
			rvm := model.RoleMenuVisibility{RoleID: role.ID, MenuID: menu.ID}
			db.FirstOrCreate(&rvm, rvm)
		}
	}

	log.Println("Seeder: Role-Menu visibility seeded")
}
