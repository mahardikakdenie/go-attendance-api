package seeder

import (
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
	db.FirstOrCreate(&trial, model.TrialRequest{Email: trial.Email})

	// 2. Seed Support Message (Tenant 1, User 1 - assuming superadmin)
	msg := model.SupportMessage{
		ID:        uuid.New(),
		TenantID:  1,
		UserID:    1,
		Subject:   "Technical Issue",
		Message:   "Cannot clock in using mobile app.",
		Category:  model.SupportCategoryTechnical,
		Status:    model.SupportStatusPending,
		CreatedAt: time.Now(),
	}
	db.FirstOrCreate(&msg, model.SupportMessage{Subject: msg.Subject})

	log.Println("Seeder: Support data added")
}
