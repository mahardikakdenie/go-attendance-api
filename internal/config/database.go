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

	// Enable UUID extension for PostgreSQL
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	if os.Getenv("RESET_DB") == "true" {
		log.Println("⚠️ Resetting database tables...")
		// Disable FK checks for clean drop if possible or drop in strict reverse order
		db.Exec("DROP TABLE IF EXISTS attendance_corrections, employee_lifecycle_tasks, lifecycle_tasks, employee_rosters, holidays, work_shifts, support_messages, provisioning_tickets, trial_requests, leaves, leave_balances, leave_types, attendances, overtimes, user_change_requests, recent_activities, tokens, media, users, role_hierarchies, role_permissions, permissions, roles, tenant_settings, tenants CASCADE")
		log.Println("⚠️ Semua tabel berhasil di-reset (CASCADE)")
	}

	if os.Getenv("RUN_MIGRATION") == "true" || os.Getenv("RESET_DB") == "true" {
		log.Println("🔄 Running migrations...")

		// Stage 1: Absolute Base (No dependencies)
		err = db.AutoMigrate(
			&model.Tenant{},
			&model.Permission{},
			&model.Role{},
			&model.Position{},
			&model.WorkShift{},
			&model.Holiday{},
			&model.LifecycleTask{},
		)
		if err != nil {
			log.Fatalf("❌ Gagal migrasi Stage 1: %v", err)
		}

		// Stage 2: Hierarchies and User (Depends on Stage 1)
		err = db.AutoMigrate(
			&model.RolePermission{},
			&model.RoleHierarchy{},
			&model.User{},
		)
		if err != nil {
			log.Fatalf("❌ Gagal migrasi Stage 2: %v", err)
		}

		// Stage 3: Business Logic (Depends on User)
		err = db.AutoMigrate(
			&model.RecentActivity{},
			&model.UserChangeRequest{},
			&model.Overtime{},
			&model.TenantSetting{},
			&model.Attendance{},
			&model.Token{},
			&model.Media{},
			&model.LeaveType{},
			&model.LeaveBalance{},
			&model.Leave{},
			&model.TrialRequest{},
			&model.ProvisioningTicket{},
			&model.SupportMessage{},
			&model.EmployeeRoster{},
			&model.EmployeeLifecycleTask{},
			&model.AttendanceCorrection{},
			&model.Payroll{},
			&model.Expense{},
			&model.PerformanceGoal{},
			&model.PerformanceCycle{},
			&model.Appraisal{},
		)
		if err != nil {
			log.Fatalf("❌ Gagal migrasi Stage 3: %v", err)
		}

		log.Println("✅ Migrasi database berhasil")
	}

	// Register Tenant Plugin AFTER migration to avoid interference with schema changes
	if err := db.Use(&TenantPlugin{}); err != nil {
		log.Fatalf("❌ Gagal inisialisasi TenantPlugin: %v", err)
	}
	log.Println("✅ TenantPlugin enabled")

	if os.Getenv("RUN_SEEDER") == "true" {
		log.Println("🌱 Running seeder...")

		seeder.SeedTenant(db)
		seeder.SeedRoles(db)
		seeder.SeedRoleHierarchy(db)
		seeder.SeedPositions(db)
		seeder.SeedUsers(db)
		seeder.SeedTenantSetting(db)
		seeder.SeedRecentActivities(db)
		seeder.SeedLeaves(db)
		seeder.SeedAttendanceHistory(db)
		seeder.SeedOvertimes(db)
		seeder.SeedSupport(db)

		log.Println("✅ Seeder selesai")
	}

	return db
}
