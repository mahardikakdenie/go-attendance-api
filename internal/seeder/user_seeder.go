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
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Gagal hash password seeder: %v", err)
	}

	user := model.User{
		Name:     "admin",
		Email:    "admin@yopmail.com",
		Password: string(hashedPassword),
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("Gagal menjalankan seeder user: %v", err)
	}

	log.Println("Seeder: Berhasil menambahkan Karyawan Teladan")
}
