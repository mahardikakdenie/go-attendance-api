package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	godotenv.Load(".env.local")
	godotenv.Load(".env")

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbName, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal koneksi ke database: %v", err)
	}

	var results []struct {
		ID            string
		DurationHours float64
	}

	db.Raw("SELECT id, duration_hours FROM timesheet_entries ORDER BY created_at DESC LIMIT 5").Scan(&results)

	fmt.Println("Recent Timesheet Entries from DB:")
	for _, r := range results {
		fmt.Printf("ID: %s | Duration: %v\n", r.ID, r.DurationHours)
	}

	// Check column type
	var colType []struct{
		DataType string
		NumericPrecision int
		NumericScale int
	}
	db.Raw("SELECT data_type, numeric_precision, numeric_scale FROM information_schema.columns WHERE table_name = 'timesheet_entries' AND column_name = 'duration_hours'").Scan(&colType)
	fmt.Printf("Column Info: %+v\n", colType)
}
