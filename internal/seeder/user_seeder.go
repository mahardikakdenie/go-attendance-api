package seeder

import (
	"log"

	"go-attendance-api/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Count(&count)

	if count > 0 {
		log.Println("Seeder: User sudah ada, skip...")
		return
	}

	// Ambil tenant system
	var systemTenant model.Tenant
	if err := db.Where("code = ?", "system").First(&systemTenant).Error; err != nil {
		log.Fatalf("Seeder: Tenant system tidak ditemukan: %v", err)
	}

	// Ambil tenant PT Friendship
	var friendshipTenant model.Tenant
	if err := db.Where("code = ?", "friendship").First(&friendshipTenant).Error; err != nil {
		log.Fatalf("Seeder: Tenant friendship tidak ditemukan: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Gagal hash password seeder: %v", err)
	}

	// Roles
	var superAdminRole, adminRole, hrRole, employeeRole model.Role
	db.Where("name = ?", "superadmin").First(&superAdminRole)
	db.Where("name = ?", "admin").First(&adminRole)
	db.Where("name = ?", "hr").First(&hrRole)
	db.Where("name = ?", "employee").First(&employeeRole)

	mediaUrl := "http://i.ibb.co.com/p6119B1C/attendance-1775556680532.png"

	users := []model.User{
		{
			Name:        "Super Admin",
			Email:       "superadmin@yopmail.com",
			Password:    string(hashedPassword),
			TenantID:    systemTenant.ID,
			RoleID:      superAdminRole.ID,
			EmployeeID:  "SA-001",
			Department:  "SaaS Owner",
			Address:     "System HQ",
			PhoneNumber: "0000000000",
		},
		{
			Name:        "Admin PT Friendship",
			Email:       "admin@friendship.com",
			Password:    string(hashedPassword),
			TenantID:    friendshipTenant.ID,
			RoleID:      adminRole.ID,
			EmployeeID:  "ADM-001",
			Department:  "Owner",
			Address:     "Friendship Office",
			PhoneNumber: "0811111111",
		},
		{
			Name:        "HR Manager",
			Email:       "hr@friendship.com",
			Password:    string(hashedPassword),
			TenantID:    friendshipTenant.ID,
			RoleID:      hrRole.ID,
			EmployeeID:  "HR-001",
			Department:  "HRD",
			Address:     "Friendship Office",
			PhoneNumber: "0822222222",
			MediaUrl:    mediaUrl,
		},
		{
			Name:        "Employee User",
			Email:       "employee@friendship.com",
			Password:    string(hashedPassword),
			TenantID:    friendshipTenant.ID,
			RoleID:      employeeRole.ID,
			EmployeeID:  "EMP-001",
			Department:  "Operations",
			Address:     "Friendship Office",
			PhoneNumber: "0833333333",
			MediaUrl:    mediaUrl,
		},
	}

	if err := db.Create(&users).Error; err != nil {
		log.Fatalf("Gagal menjalankan seeder user: %v", err)
	}

	log.Println("Seeder: Berhasil menambahkan SuperAdmin, Admin, HR, dan Employee")
}
