package seeder

import (
	"errors"
	"log"
	"time"

	"go-attendance-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedSupport(db *gorm.DB) {
	// 1. Seed Trial Request
	trial := model.TrialRequest{
		ID:                 uuid.New(),
		CompanyName:        "Acme Corp",
		ContactName:        "John Doe",
		Email:              "john@acme.com",
		PhoneNumber:        "08123456789",
		EmployeeCountRange: model.EmployeeCountRange11To50,
		Industry:           "Technology",
		Status:             model.TrialRequestStatusNew,
		CreatedAt:          time.Now(),
	}
	if err := db.FirstOrCreate(&trial, model.TrialRequest{Email: trial.Email}).Error; err != nil {
		log.Printf("Seeder: gagal create trial request: %v", err)
	}

	// 2. Seed Support Message (from a real tenant user, not hardcoded ID)
	var tenant model.Tenant
	if err := db.Where("code = ?", "friendship").First(&tenant).Error; err != nil {
		log.Printf("Seeder: tenant friendship tidak ditemukan, skip support message: %v", err)
		return
	}

	var sender model.User
	err := db.Where("email = ?", "admin@friendship.com").First(&sender).Error
	if err != nil {
		// fallback: any active user in that tenant
		err = db.Where("tenant_id = ?", tenant.ID).Order("id ASC").First(&sender).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Seeder: tidak ada user tenant %d untuk support sender, skip", tenant.ID)
				return
			}
			log.Printf("Seeder: gagal mencari support sender: %v", err)
			return
		}
	}

	msg := model.SupportMessage{
		ID:        uuid.New(),
		TenantID:  sender.TenantID,
		UserID:    sender.ID,
		Subject:   "Technical Issue",
		Message:   "Cannot clock in using mobile app.",
		Category:  model.SupportCategoryTechnical,
		Status:    model.SupportStatusPending,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if err := db.FirstOrCreate(&msg, model.SupportMessage{
		TenantID: msg.TenantID,
		UserID:   msg.UserID,
		Subject:  msg.Subject,
	}).Error; err != nil {
		log.Printf("Seeder: gagal create support message: %v", err)
		return
	}

	log.Printf("Seeder: Support data added (sender=%s, tenant=%s)", sender.Email, tenant.Code)
}
