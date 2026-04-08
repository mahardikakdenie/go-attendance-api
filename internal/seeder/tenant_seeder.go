package seeder

import (
	"log"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedTenant(db *gorm.DB) {
	var count int64
	db.Model(&model.Tenant{}).Count(&count)

	if count > 0 {
		return
	}

	tenants := []model.Tenant{
		{
			Name: "SaaS System",
			Code: "system",
		},
		{
			Name: "PT Friendship Logistics",
			Code: "friendship",
		},
		{
			Name: "Remote Company Inc",
			Code: "remote-co",
		},
		{
			Name: "Hybrid Corp",
			Code: "hybrid",
		},
	}

	if err := db.Create(&tenants).Error; err != nil {
		log.Fatalf("Gagal menjalankan seeder tenant: %v", err)
	}

	log.Println("Seeder: Tenant berhasil ditambahkan")
}
