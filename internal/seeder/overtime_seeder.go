package seeder

import (
	"log"
	"time"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedOvertimes(db *gorm.DB) {
	var employee model.User
	if err := db.Where("email = ?", "employee@friendship.com").First(&employee).Error; err != nil {
		return
	}

	overtimes := []model.Overtime{
		{
			UserID:    employee.ID,
			TenantID:  employee.TenantID,
			Date:      time.Now().AddDate(0, 0, -1),
			StartTime: "17:30",
			EndTime:   "19:30",
			Reason:    "Finalizing project reports",
			Status:    model.OvertimeStatusApproved,
		},
		{
			UserID:    employee.ID,
			TenantID:  employee.TenantID,
			Date:      time.Now().AddDate(0, 0, -3),
			StartTime: "18:00",
			EndTime:   "20:00",
			Reason:    "Deployment support",
			Status:    model.OvertimeStatusApproved,
		},
	}

	for _, o := range overtimes {
		db.FirstOrCreate(&o, model.Overtime{UserID: o.UserID, Date: o.Date})
	}

	log.Println("Seeder: Overtime requests added")
}
