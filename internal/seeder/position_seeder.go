package seeder

import (
	"log"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedPositions(db *gorm.DB) {
	// Ambil tenant friendship
	var friendshipTenant model.Tenant
	if err := db.Where("code = ?", "friendship").First(&friendshipTenant).Error; err != nil {
		return
	}

	positions := []model.Position{
		{TenantID: friendshipTenant.ID, Name: "CEO", Level: 1},
		{TenantID: friendshipTenant.ID, Name: "VP", Level: 2},
		{TenantID: friendshipTenant.ID, Name: "Department Head", Level: 3},
		{TenantID: friendshipTenant.ID, Name: "Manager", Level: 4},
		{TenantID: friendshipTenant.ID, Name: "Supervisor", Level: 5},
		{TenantID: friendshipTenant.ID, Name: "Staff", Level: 6},
	}

	for _, p := range positions {
		db.FirstOrCreate(&p, model.Position{TenantID: p.TenantID, Name: p.Name})
	}

	log.Println("Seeder: Positions added")
}
