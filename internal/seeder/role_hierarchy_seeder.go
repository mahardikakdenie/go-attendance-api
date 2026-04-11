package seeder

import (
	"log"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedRoleHierarchy(db *gorm.DB) {
	// Ambil tenant friendship
	var friendshipTenant model.Tenant
	if err := db.Where("code = ?", "friendship").First(&friendshipTenant).Error; err != nil {
		return
	}

	// Roles
	var adminRole, hrRole, employeeRole model.Role
	db.Where("name = ?", "admin").First(&adminRole)
	db.Where("name = ?", "hr").First(&hrRole)
	db.Where("name = ?", "employee").First(&employeeRole)

	hierarchies := []model.RoleHierarchy{
		// Admin supervises HR and Employee
		{TenantID: friendshipTenant.ID, ParentRoleID: adminRole.ID, ChildRoleID: hrRole.ID},
		{TenantID: friendshipTenant.ID, ParentRoleID: adminRole.ID, ChildRoleID: employeeRole.ID},
		// HR supervises Employee
		{TenantID: friendshipTenant.ID, ParentRoleID: hrRole.ID, ChildRoleID: employeeRole.ID},
	}

	for _, h := range hierarchies {
		db.FirstOrCreate(&h, model.RoleHierarchy{
			TenantID:     h.TenantID,
			ParentRoleID: h.ParentRoleID,
			ChildRoleID:  h.ChildRoleID,
		})
	}

	log.Println("Seeder: Role hierarchy added")
}
