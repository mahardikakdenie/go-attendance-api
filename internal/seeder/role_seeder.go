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

		// User Management
		{ID: "user.view", Module: "user", Action: "view"},
		{ID: "user.create", Module: "user", Action: "create"},
		{ID: "user.edit", Module: "user", Action: "edit"},
		{ID: "user.delete", Module: "user", Action: "delete"},

		// Tenant & SaaS
		{ID: "tenant.view", Module: "tenant", Action: "view"},
		{ID: "tenant.edit", Module: "tenant", Action: "edit"},
		{ID: "subscription.manage", Module: "subscription", Action: "manage"},

		// RBAC
		{ID: "role.view", Module: "role", Action: "view"},
		{ID: "role.manage", Module: "role", Action: "manage"},

		// Support & Provisioning
		{ID: "support.manage", Module: "support", Action: "manage"},

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
		},
		{
			Name:        "admin",
			Description: "Tenant Owner / Administrator",
			BaseRole:    model.BaseRoleAdmin,
			IsSystem:    true,
		},
		{
			Name:        "hr",
			Description: "Human Resources Manager",
			BaseRole:    model.BaseRoleHR,
			IsSystem:    true,
		},
		{
			Name:        "finance",
			Description: "Finance & Payroll Manager",
			BaseRole:    model.BaseRoleFinance,
			IsSystem:    true,
		},
		{
			Name:        "employee",
			Description: "Regular Employee",
			BaseRole:    model.BaseRoleEmployee,
			IsSystem:    true,
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
		}

		// Update fields if role exists (optional, but good for keeping seeder up to date)
		db.Model(&role).Updates(model.Role{
			Description: r.Description,
			BaseRole:    r.BaseRole,
			IsSystem:    r.IsSystem,
			IsImmutable: r.IsImmutable,
		})

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
				"performance.view",
			}
			for _, pID := range empPerms {
				rp := model.RolePermission{RoleID: role.ID, PermissionID: pID}
				db.FirstOrCreate(&rp, rp)
			}
		}
	}

	log.Println("Seeder: Roles and Permissions updated")
}
