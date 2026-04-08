package seeder

import (
	"log"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedRoles(db *gorm.DB) {
	roles := []string{"superadmin", "admin", "hr", "employee"}

	for _, roleName := range roles {
		var role model.Role
		err := db.Where("name = ?", roleName).First(&role).Error
		if err == gorm.ErrRecordNotFound {
			role = model.Role{Name: roleName}
			if err := db.Create(&role).Error; err != nil {
				log.Fatalf("Gagal seeder role %s: %v", roleName, err)
			}
			log.Printf("Seeder: Role %s ditambahkan", roleName)
		}
	}
}
