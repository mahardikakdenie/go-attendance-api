package seeder

import (
	"log"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedAllowancePresets(db *gorm.DB) {
	presets := []model.AllowancePreset{
		{Name: "Uang Makan", Type: model.AllowanceTypeVariable},
		{Name: "Uang Transport", Type: model.AllowanceTypeVariable},
		{Name: "Tunjangan Internet / Komunikasi", Type: model.AllowanceTypeVariable},
		{Name: "Tunjangan Jabatan", Type: model.AllowanceTypeFixed},
	}

	for _, p := range presets {
		var existing model.AllowancePreset
		err := db.Where("name = ?", p.Name).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&p).Error; err != nil {
				log.Printf("⚠️ Gagal seeder allowance preset %s: %v", p.Name, err)
			}
		}
	}

	log.Println("Seeder: Allowance Presets seeded successfully")
}
