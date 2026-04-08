package seeder

import (
	"log"
	"time"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedRecentActivities(db *gorm.DB) {
	var count int64
	db.Model(&model.RecentActivity{}).Count(&count)

	if count > 0 {
		log.Println("Seeder: Recent activities sudah ada, skip...")
		return
	}

	var users []model.User
	db.Find(&users)

	if len(users) == 0 {
		log.Println("Seeder: Tidak ada user untuk ditambahkan activity, skip...")
		return
	}

	activities := []model.RecentActivity{
		{
			UserID:    users[0].ID,
			Title:     "Logged in",
			Action:    "Login",
			Status:    "Success",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			UserID:    users[0].ID,
			Title:     "Clocked in",
			Action:    "Attendance",
			Status:    "Success",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			UserID:    users[0].ID,
			Title:     "Updated profile",
			Action:    "Profile",
			Status:    "Success",
			CreatedAt: time.Now().Add(-3 * time.Hour),
		},
	}

	// Add some activities for other users if available
	if len(users) > 1 {
		activities = append(activities, model.RecentActivity{
			UserID:    users[1].ID,
			Title:     "Clocked in",
			Action:    "Attendance",
			Status:    "Late",
			CreatedAt: time.Now().Add(-30 * time.Minute),
		})
	}

	if err := db.Create(&activities).Error; err != nil {
		log.Fatalf("Gagal menjalankan seeder activity: %v", err)
	}

	log.Println("Seeder: Berhasil menambahkan Recent Activities")
}
