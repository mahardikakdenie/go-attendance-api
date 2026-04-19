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
		db.Exec("DROP TABLE IF EXISTS timesheet_entries, tasks, projects, password_resets, subscriptions, appraisals, performance_cycles, performance_goals, expenses, payrolls, attendance_corrections, employee_lifecycle_tasks, lifecycle_tasks, employee_rosters, holidays, work_shifts, support_messages, provisioning_tickets, trial_requests, leaves, leave_balances, leave_types, attendances, overtimes, user_change_requests, recent_activities, tokens, media, user_payroll_profiles, users, positions, role_hierarchies, role_permissions, permissions, roles, tenant_settings, tenants CASCADE")
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
			&model.UserPayrollProfile{},
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
			&model.Subscription{},
			&model.PasswordReset{},
			&model.Project{},
			&model.Task{},
			&model.TimesheetEntry{},
		)
		if err != nil {
			log.Fatalf("❌ Gagal migrasi Stage 3: %v", err)
		}
		log.Println("✅ Migrasi database berhasil")

		// 🔄 Data Backfill: Ensure existing tenants have a subscription record
		backfillSubscriptions(db)
		backfillPayrollProfiles(db)
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
		seeder.SeedProjects(db)
		seeder.SeedLeaves(db)
		seeder.SeedAttendanceHistory(db)
		seeder.SeedOvertimes(db)
		seeder.SeedSupport(db)
		seeder.SeedUserPayrollProfiles(db)

		log.Println("✅ Seeder selesai")
	}

	return db
}

func backfillSubscriptions(db *gorm.DB) {
	log.Println("🔄 Checking for tenants without subscriptions...")

	// Query: Insert default subscription for tenants that don't have one
	// Kita ambil status 'Active' untuk tenant lama agar tidak langsung tersuspensi
	result := db.Exec(`
		INSERT INTO subscriptions (tenant_id, plan, billing_cycle, amount, status, next_billing_date, created_at, updated_at)
		SELECT id, plan, 'Monthly', 0, 'Active', (NOW() + INTERVAL '30 days'), NOW(), NOW()
		FROM tenants
		WHERE id NOT IN (SELECT tenant_id FROM subscriptions)
	`)

	if result.Error != nil {
		log.Printf("⚠️ Warning during subscription backfill: %v\n", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("✅ Successfully backfilled %d tenant subscriptions\n", result.RowsAffected)
	}
}

func backfillPayrollProfiles(db *gorm.DB) {
	log.Println("🔄 Checking for users without payroll profiles...")

	result := db.Exec(`
		INSERT INTO user_payroll_profiles (user_id, ptkp_status, basic_salary, fixed_allowance, created_at, updated_at)
		SELECT id, 'TK/0', base_salary, 0, NOW(), NOW()
		FROM users
		WHERE id NOT IN (SELECT user_id FROM user_payroll_profiles)
	`)

	if result.Error != nil {
		log.Printf("⚠️ Warning during payroll profile backfill: %v\n", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("✅ Successfully backfilled %d user payroll profiles\n", result.RowsAffected)
	}
}
