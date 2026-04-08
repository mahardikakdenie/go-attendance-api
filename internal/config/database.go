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
			&model.Media{},
			&model.Attendance{},
			&model.Overtime{},
			&model.UserChangeRequest{},
			&model.RecentActivity{},
			&model.User{},
			&model.Role{},
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
			&model.Role{},
			&model.User{},
			&model.RecentActivity{},
			&model.UserChangeRequest{},
			&model.Overtime{},
			&model.TenantSetting{},
			&model.Attendance{},
			&model.Token{},
			&model.Media{},
		)

		if err != nil {
			log.Fatalf("❌ Gagal migrasi database: %v", err)
		}

		log.Println("✅ Migrasi database berhasil")
	}

	if os.Getenv("RUN_SEEDER") == "true" {
		log.Println("🌱 Running seeder...")

		seeder.SeedTenant(db)
		seeder.SeedRoles(db)
		seeder.SeedUsers(db)
		seeder.SeedTenantSetting(db)
		seeder.SeedRecentActivities(db)

		log.Println("✅ Seeder selesai")
	}

	return db
}
