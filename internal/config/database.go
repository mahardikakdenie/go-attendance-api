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

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, dbName, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Gagal koneksi ke database: %v", err)
	}

	log.Println("✅ Database connected")

	if os.Getenv("RESET_DB") == "true" {
		err := db.Migrator().DropTable(
			&model.Token{},
			&model.Attendance{},
			&model.User{},
			&model.TenantSetting{},
			&model.Tenant{},
		)
		if err != nil {
			log.Fatalf("❌ Gagal reset database: %v", err)
		}
		log.Println("⚠️ Semua tabel berhasil di-reset")
	}

	if os.Getenv("RUN_MIGRATION") == "true" || os.Getenv("RESET_DB") == "true" {

		err = db.AutoMigrate(
			&model.Tenant{},
			&model.User{},
			&model.TenantSetting{},
			&model.Attendance{},
			&model.Token{},
		)

		if err != nil {
			log.Fatalf("❌ Gagal migrasi database: %v", err)
		}

		log.Println("✅ Migrasi database berhasil")
	}

	if os.Getenv("RUN_SEEDER") == "true" {
		log.Println("🌱 Running seeder...")

		seeder.SeedTenant(db)
		seeder.SeedUsers(db)
		seeder.SeedTenantSetting(db)

		log.Println("✅ Seeder selesai")
	}

	return db
}
