package seeder

import (
	"log"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedSubscriptionFeatures(db *gorm.DB) {
	features := []model.SubscriptionFeature{
		{FeatureKey: "user", Label: "Employee Management", IsActive: true},
		{FeatureKey: "attendance", Label: "Advanced Attendance", IsActive: true},
		{FeatureKey: "leave", Label: "Leave Requests", IsActive: true},
		{FeatureKey: "overtime", Label: "Overtime Tracking", IsActive: true},
		{FeatureKey: "payroll", Label: "Payroll & Slips", IsActive: true},
		{FeatureKey: "finance", Label: "Finance & Claims", IsActive: true},
		{FeatureKey: "analytics", Label: "Advanced Analytics", IsActive: true},
		{FeatureKey: "timesheet", Label: "Project Timesheet", IsActive: true},
		{FeatureKey: "schedule", Label: "Work Schedules & Rosters", IsActive: true},
		{FeatureKey: "performance", Label: "Performance Management", IsActive: true},
		{FeatureKey: "project", Label: "Project Management", IsActive: true},
	}

	for _, f := range features {
		var existing model.SubscriptionFeature
		err := db.Where("feature_key = ?", f.FeatureKey).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&f).Error; err != nil {
				log.Printf("⚠️ Gagal seeder feature %s: %v", f.FeatureKey, err)
			}
		}
	}

	log.Println("Seeder: Subscription Features seeded successfully")
}
