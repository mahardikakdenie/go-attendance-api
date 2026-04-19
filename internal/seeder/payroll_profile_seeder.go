package seeder

import (
	"log"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedUserPayrollProfiles(db *gorm.DB) {
	var users []model.User
	db.Find(&users)

	for _, user := range users {
		var count int64
		db.Model(&model.UserPayrollProfile{}).Where("user_id = ?", user.ID).Count(&count)

		if count == 0 {
			profile := model.UserPayrollProfile{
				UserID:            user.ID,
				BankName:          "BCA",
				BankAccountNumber: "1234567890",
				BankAccountHolder: user.Name,
				PtkpStatus:        model.PtkpTK0,
				BasicSalary:       user.BaseSalary,
				FixedAllowance:    0,
			}
			
			// Custom values for specific seed users
			if user.Email == "admin@friendship.com" {
				profile.PtkpStatus = model.PtkpK1
				profile.FixedAllowance = 5000000
			} else if user.Email == "hr@friendship.com" {
				profile.PtkpStatus = model.PtkpK0
				profile.FixedAllowance = 2000000
			} else if user.Email == "finance@friendship.com" {
				profile.PtkpStatus = model.PtkpK0
				profile.FixedAllowance = 1500000
			}

			db.Create(&profile)
		}
	}

	log.Println("Seeder: User Payroll Profiles added")
}
