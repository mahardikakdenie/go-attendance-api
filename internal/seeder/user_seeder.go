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

	// Ambil tenant pertama (WAJIB kalau ada FK)
	var tenant model.Tenant
	if err := db.First(&tenant).Error; err != nil {
		log.Fatalf("Seeder: Tenant tidak ditemukan, jalankan seeder tenant dulu: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Gagal hash password seeder: %v", err)
	}

	user := model.User{
		Name:        "Admin",
		Email:       "admin@yopmail.com",
		Password:    string(hashedPassword),
		TenantID:    tenant.ID, // 🔥 FIX penting (foreign key)
		Role:        model.RoleAdmin,
		EmployeeID:  "FT-001",
		Department:  "Management",
		Address:     "Head Office",
		PhoneNumber: "08123456789",
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("Gagal menjalankan seeder user: %v", err)
	}

	log.Println("Seeder: Berhasil menambahkan Admin")
}
