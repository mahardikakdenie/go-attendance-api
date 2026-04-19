package seeder

import (
	"log"
	"time"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedAttendanceHistory(db *gorm.DB) {
	var employee model.User
	if err := db.Where("email = ?", "employee@friendship.com").First(&employee).Error; err != nil {
		log.Println("Seeder Attendance: Employee not found, skipping...")
		return
	}

	// Check if already seeded
	var count int64
	db.Model(&model.Attendance{}).Where("user_id = ?", employee.ID).Count(&count)
	if count > 5 {
		log.Println("Seeder Attendance: Data already exists, skipping...")
		return
	}

	// Generate 30 days of attendance
	now := time.Now()
	for i := 30; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)

		// Skip weekends
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue
		}

		// Random clock-in between 07:45 and 09:15
		// Base 08:00
		clockIn := time.Date(date.Year(), date.Month(), date.Day(), 8, 0, 0, 0, time.Local)
		
		// Add some variety
		if i%7 == 0 {
			clockIn = clockIn.Add(45 * time.Minute) // Late (08:45)
		} else if i%5 == 0 {
			clockIn = clockIn.Add(-15 * time.Minute) // Early (07:45)
		} else {
			clockIn = clockIn.Add(time.Duration(i%15) * time.Minute) // 08:00 - 08:14
		}

		// Clock-out at 17:00 + some variety
		clockOut := time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, time.Local)
		clockOut = clockOut.Add(time.Duration(i%20) * time.Minute)

		status := model.StatusWorking
		// Assuming late if after 08:30 (based on tenant_setting_seeder)
		if clockIn.Hour() > 8 || (clockIn.Hour() == 8 && clockIn.Minute() > 30) {
			status = model.StatusLate
		}

		att := model.Attendance{
			UserID:           employee.ID,
			TenantID:         employee.TenantID,
			ClockInTime:      clockIn,
			ClockOutTime:     &clockOut,
			Status:           status,
			ClockInLatitude:  -6.200000,
			ClockInLongitude: 106.816666,
			ClockInMediaUrl:  "https://i.ibb.co.com/p6119B1C/attendance-1775556680532.png",
		}

		db.Create(&att)
	}

	log.Println("Seeder: Attendance history for 30 days added")
}
