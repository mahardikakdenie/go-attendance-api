package seeder

import (
	"log"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedRoles(db *gorm.DB) {
	// 1. Seed Permissions
	permissions := []model.Permission{
		{ID: "attendance.view", Module: "attendance", Action: "view"},
		{ID: "attendance.create", Module: "attendance", Action: "create"},
		{ID: "attendance.edit", Module: "attendance", Action: "edit"},
		{ID: "attendance.delete", Module: "attendance", Action: "delete"},
		{ID: "attendance.export", Module: "attendance", Action: "export"},

		{ID: "leave.view", Module: "leave", Action: "view"},
		{ID: "leave.create", Module: "leave", Action: "create"},
		{ID: "leave.approve", Module: "leave", Action: "approve"},
		{ID: "leave.reject", Module: "leave", Action: "reject"},

		{ID: "overtime.view", Module: "overtime", Action: "view"},
		{ID: "overtime.create", Module: "overtime", Action: "create"},
		{ID: "overtime.approve", Module: "overtime", Action: "approve"},

		{ID: "payroll.view", Module: "payroll", Action: "view"},
		{ID: "payroll.calculate", Module: "payroll", Action: "calculate"},
		{ID: "payroll.approve", Module: "payroll", Action: "approve"},

		{ID: "user.view", Module: "user", Action: "view"},
		{ID: "user.create", Module: "user", Action: "create"},
		{ID: "user.edit", Module: "user", Action: "edit"},
		{ID: "user.delete", Module: "user", Action: "delete"},

		{ID: "tenant.view", Module: "tenant", Action: "view"},
		{ID: "tenant.edit", Module: "tenant", Action: "edit"},

		{ID: "role.view", Module: "role", Action: "view"},
		{ID: "role.manage", Module: "role", Action: "manage"},
	}

	for _, p := range permissions {
		db.FirstOrCreate(&p, model.Permission{ID: p.ID})
	}

	// 2. Seed System Roles
	systemRoles := []model.Role{
		{
			Name:        "superadmin",
			Description: "Platform Owner with full access",
			BaseRole:    model.BaseRoleAdmin,
			IsSystem:    true,
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

		// Assign all permissions to superadmin and admin
		if role.BaseRole == model.BaseRoleAdmin {
			for _, p := range permissions {
				rp := model.RolePermission{RoleID: role.ID, PermissionID: p.ID}
				db.FirstOrCreate(&rp, rp)
			}
		}

		// Assign HR permissions
		if role.Name == "hr" {
			hrPerms := []string{
				"attendance.view", "attendance.export",
				"leave.view", "leave.approve", "leave.reject",
				"overtime.view", "overtime.approve",
				"user.view", "user.create", "user.edit",
			}
			for _, pID := range hrPerms {
				rp := model.RolePermission{RoleID: role.ID, PermissionID: pID}
				db.FirstOrCreate(&rp, rp)
			}
		}

		// Assign Employee permissions
		if role.Name == "employee" {
			empPerms := []string{
				"attendance.view", "attendance.create",
				"leave.view", "leave.create",
				"overtime.view", "overtime.create",
			}
			for _, pID := range empPerms {
				rp := model.RolePermission{RoleID: role.ID, PermissionID: pID}
				db.FirstOrCreate(&rp, rp)
			}
		}
	}
}
