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

	// Positions
	var ceoPos, managerPos, staffPos model.Position
	db.Where("name = ? AND tenant_id = ?", "CEO", friendshipTenant.ID).First(&ceoPos)
	db.Where("name = ? AND tenant_id = ?", "Manager", friendshipTenant.ID).First(&managerPos)
	db.Where("name = ? AND tenant_id = ?", "Staff", friendshipTenant.ID).First(&staffPos)

	mediaUrl := "http://i.ibb.co.com/p6119B1C/attendance-1775556680532.png"

	// 1. Super Admin
	sa := model.User{
		Name:        "Super Admin",
		Email:       "superadmin@yopmail.com",
		Password:    string(hashedPassword),
		TenantID:    systemTenant.ID,
		RoleID:      superAdminRole.ID,
		EmployeeID:  "SA-001",
		Department:  "SaaS Owner",
		Address:     "System HQ",
		PhoneNumber: "0000000000",
	}
	db.Create(&sa)

	// 2. Admin Tenant (CEO)
	admin := model.User{
		Name:        "Admin PT Friendship",
		Email:       "admin@friendship.com",
		Password:    string(hashedPassword),
		TenantID:    friendshipTenant.ID,
		RoleID:      adminRole.ID,
		PositionID:  &ceoPos.ID,
		EmployeeID:  "ADM-001",
		Department:  "Owner",
		Address:     "Friendship Office",
		PhoneNumber: "0811111111",
		BaseSalary:  50000000,
	}
	db.Create(&admin)

	// 3. HR Manager (Reports to Admin)
	hr := model.User{
		Name:        "HR Manager",
		Email:       "hr@friendship.com",
		Password:    string(hashedPassword),
		TenantID:    friendshipTenant.ID,
		RoleID:      hrRole.ID,
		PositionID:  &managerPos.ID,
		ManagerID:   &admin.ID,
		EmployeeID:  "HR-001",
		Department:  "HRD",
		Address:     "Friendship Office",
		PhoneNumber: "0822222222",
		MediaUrl:    mediaUrl,
		BaseSalary:  15000000,
	}
	db.Create(&hr)

	// 4. Employee User (Reports to HR)
	emp := model.User{
		Name:        "Employee User",
		Email:       "employee@friendship.com",
		Password:    string(hashedPassword),
		TenantID:    friendshipTenant.ID,
		RoleID:      employeeRole.ID,
		PositionID:  &staffPos.ID,
		ManagerID:   &hr.ID,
		EmployeeID:  "EMP-001",
		Department:  "Operations",
		Address:     "Friendship Office",
		PhoneNumber: "0833333333",
		MediaUrl:    mediaUrl,
		BaseSalary:  8000000,
	}
	db.Create(&emp)

	log.Println("Seeder: Users with Hierarchy added")
}
