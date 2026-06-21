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
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("⚠️ Warning: Gagal create extension uuid-ossp: %v\n", err)
	}

	// 🆕 CLEANUP: Drop legacy 'plan' column from subscriptions if exists
	if err := db.Exec("ALTER TABLE subscriptions DROP COLUMN IF EXISTS plan").Error; err != nil {
		log.Printf("⚠️ Warning: Gagal drop column plan: %v\n", err)
	}

	// 🆕 FIX: Ensure precision is updated for duration_hours
	// We check if table exists first to avoid error on fresh DB
	if db.Migrator().HasTable("timesheet_entries") {
		if err := db.Exec("ALTER TABLE timesheet_entries ALTER COLUMN duration_hours TYPE numeric(8,4)").Error; err != nil {
			log.Printf("⚠️ Warning: Gagal alter column duration_hours: %v\n", err)
		}
	}

	if os.Getenv("RESET_DB") == "true" {

		log.Println("⚠️ Resetting database tables...")
		// Disable FK checks for clean drop if possible or drop in strict reverse order
		db.Exec("DROP TABLE IF EXISTS support_replies, allowance_presets, notifications, subscription_features, invoices, timesheet_entries, tasks, projects, password_resets, subscriptions, appraisals, performance_cycles, performance_goals, expenses, payrolls, attendance_corrections, employee_lifecycle_tasks, lifecycle_tasks, employee_rosters, holidays, work_shifts, support_messages, provisioning_tickets, trial_requests, leaves, leave_balances, leave_types, attendances, overtimes, user_change_requests, recent_activities, tokens, media, user_payroll_profiles, users, positions, role_hierarchies, role_permissions, permissions, roles, tenant_settings, tenants CASCADE")
		log.Println("⚠️ Semua tabel berhasil di-reset (CASCADE)")
	}

	log.Println("🔄 Running migrations...")

	// Stage 1: Absolute Base (No dependencies)
	err = db.AutoMigrate(
		&model.Tenant{},
		&model.SubscriptionPlan{},
		&model.SubscriptionFeature{},
		&model.AllowancePreset{},
		&model.Menu{},
		&model.Permission{}, &model.Role{},
		&model.Position{}, &model.WorkShift{},
		&model.Holiday{},
		&model.LifecycleTask{},
	)
	if err != nil {
		log.Fatalf("❌ Gagal migrasi Stage 1: %v", err)
	}


	// Stage 2: Hierarchies and User (Depends on Stage 1)
	err = db.AutoMigrate(
		&model.RolePermission{},
		&model.RoleMenuVisibility{},
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
		&model.Invoice{},
		&model.PasswordReset{},
		&model.Project{},
		&model.Task{},
		&model.TimesheetEntry{},
		&model.Notification{},
		&model.SupportReply{},
		&model.AuditLog{},
	)
	if err != nil {
		log.Fatalf("❌ Gagal migrasi Stage 3: %v", err)
	}
	log.Println("✅ Migrasi database berhasil")

	// Seed plans early because subscriptions depend on them
	seeder.SeedPlans(db)

	// Data Backfill
	backfillSubscriptions(db)
	backfillPayrollProfiles(db)
	backfillRoleMenuVisibility(db)

	// Register Tenant Plugin
	if err := db.Use(&TenantPlugin{}); err != nil {
		log.Fatalf("❌ Gagal inisialisasi TenantPlugin: %v", err)
	}
	log.Println("✅ TenantPlugin enabled")

	if os.Getenv("RUN_SEEDER") == "true" {
		log.Println("🌱 Running seeder...")

		seeder.SeedSubscriptionFeatures(db)
		seeder.SeedAllowancePresets(db)
		seeder.SeedMenus(db)
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
	result := db.Exec(`
		INSERT INTO subscriptions (tenant_id, plan_id, billing_cycle, amount, status, next_billing_date, created_at, updated_at)
		SELECT id, 1, 'Monthly', 0, 'Trial', (NOW() + INTERVAL '14 days'), NOW(), NOW()
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

func backfillRoleMenuVisibility(db *gorm.DB) {
	log.Println("🔄 Backfilling role_menu_visibility from required_permission...")

	// This join finds which roles have the permission required by a menu
	// and inserts that mapping into the new role_menu_visibility table.
	// It handles both system roles (tenant_id IS NULL) and tenant roles.
	result := db.Exec(`
		INSERT INTO role_menu_visibility (menu_id, role_id)
		SELECT m.id, rp.role_id
		FROM menus m
		JOIN role_permissions rp ON m.required_permission = rp.permission_id
		WHERE m.required_permission IS NOT NULL AND m.required_permission != ''
		ON CONFLICT (menu_id, role_id) DO NOTHING
	`)

	if result.Error != nil {
		log.Printf("⚠️ Warning during role_menu_visibility backfill: %v\n", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("✅ Successfully backfilled %d role-menu visibility mappings\n", result.RowsAffected)
	}

	// Also handle menus that have no required_permission (public/dashboard)
	// Usually these should be visible to all roles.
	// For safety, we only backfill if the table is mostly empty or for specific important keys if needed.
	// But according to the task, we primarily want to migrate existing permission-based visibility.
}
