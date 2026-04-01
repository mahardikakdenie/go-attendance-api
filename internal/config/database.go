package config

import (
	"fmt"
	"log"
	"os"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/seeder"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, dbName, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal koneksi ke database: %v", err)
	}

	if os.Getenv("RESET_DB") == "true" {
		db.Migrator().DropTable(&model.Attendance{}, &model.User{})
		log.Println("Tabel berhasil direset (Drop Table)")
	}

	if os.Getenv("RUN_MIGRATION") == "true" || os.Getenv("RESET_DB") == "true" {
		err = db.AutoMigrate(&model.User{}, &model.Attendance{})
		if err != nil {
			log.Fatalf("Gagal melakukan migrasi database: %v", err)
		}
		log.Println("Migrasi database berhasil dieksekusi")
	}

	if os.Getenv("RUN_SEEDER") == "true" {
		seeder.SeedUsers(db)
	}

	return db
}
