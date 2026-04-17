package seeder

import (
	"log"
	// "time"

	// "go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedAttendanceHistory(db *gorm.DB) {
	// var employee model.User
	// if err := db.Where("email = ?", "employee@friendship.com").First(&employee).Error; err != nil {
	// 	return
	// }

	// // Generate 30 days of attendance
	// now := time.Now()
	// for i := 30; i >= 0; i-- {
	// 	date := now.AddDate(0, 0, -i)

	// 	// Skip weekends
	// 	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
	// 		continue
	// 	}

	// 	// Random clock-in between 08:00 and 09:30
	// 	clockIn := time.Date(date.Year(), date.Month(), date.Day(), 8, 0, 0, 0, time.Local)
	// 	if i%3 == 0 {
	// 		clockIn = clockIn.Add(time.Duration(i*2) * time.Minute) // Some late ones
	// 	} else {
	// 		clockIn = clockIn.Add(time.Duration(i) * time.Minute)
	// 	}

	// 	// Clock-out at 17:00
	// 	clockOut := time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, time.Local)

	// 	status := model.StatusWorking
	// 	if clockIn.Hour() > 8 || (clockIn.Hour() == 8 && clockIn.Minute() > 30) {
	// 		status = model.StatusLate
	// 	}

	// 	att := model.Attendance{
	// 		UserID: employee.ID,
	// 		TenantID: employee.TenantID,
	// 		ClockInTime: clockIn,
	// 		ClockOutTime: &clockOut,
	// 		Status: status,
	// 		ClockInLatitude: -6.200000,
	// 		ClockInLongitude: 106.816666,
	// 	}

	// 	db.Create(&att)
	// }

	log.Println("Seeder: Attendance history for 30 days added")
}
