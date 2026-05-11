package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"go-attendance-api/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env.local")
	godotenv.Load(".env")

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, dbName, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal koneksi ke database: %v", err)
	}

	taskID := uint(1) // dummy
	
	entry := model.TimesheetEntry{
		TenantID: 1,
		UserID: 1,
		ProjectID: 4,
		TaskID: &taskID,
		Date: time.Now(),
		DurationHours: 0.0061,
		Notes: "Test precision",
	}

	err = db.Create(&entry).Error
	if err != nil {
		log.Fatalf("Error creating entry: %v", err)
	}

	fmt.Printf("Inserted entry ID: %s with duration: %v\n", entry.ID, entry.DurationHours)

	var fetched model.TimesheetEntry
	db.First(&fetched, "id = ?", entry.ID)
	fmt.Printf("Fetched entry ID: %s with duration: %v\n", fetched.ID, fetched.DurationHours)
}
