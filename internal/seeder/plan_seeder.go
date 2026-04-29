package seeder

import (
	"go-attendance-api/internal/model"
	"gorm.io/gorm"
)

func SeedPlans(db *gorm.DB) {
	plans := []model.SubscriptionPlan{
		{
			ID:           1,
			Name:         "Trial",
			MaxEmployees: 3,
			Features:     []string{"user", "attendance"},
			IsActive:     true,
		},
		{
			ID:           2,
			Name:         "Starter",
			MaxEmployees: 50,
			Features:     []string{"user", "attendance", "leave", "overtime"},
			IsActive:     true,
		},
		{
			ID:           3,
			Name:         "Business",
			MaxEmployees: 200,
			Features:     []string{"user", "attendance", "leave", "overtime", "payroll", "finance"},
			IsActive:     true,
		},
		{
			ID:           4,
			Name:         "Enterprise",
			MaxEmployees: 0, // unlimited
			Features:     []string{"*"},
			IsActive:     true,
		},
	}

	for _, plan := range plans {
		db.FirstOrCreate(&plan, model.SubscriptionPlan{Name: plan.Name})
	}

	// 🆕 Reset sequence agar auto-increment sinkron setelah manual insert ID
	db.Exec("SELECT setval('subscription_plans_id_seq', (SELECT MAX(id) FROM subscription_plans))")
}
