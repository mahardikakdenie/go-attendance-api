package seeder

import (
	"log"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedTenantSetting(db *gorm.DB) {
	var tenants []model.Tenant

	if err := db.Find(&tenants).Error; err != nil {
		log.Fatalf("Gagal ambil tenant: %v", err)
	}

	for _, tenant := range tenants {

		var count int64
		db.Model(&model.TenantSetting{}).
			Where("tenant_id = ?", tenant.ID).
			Count(&count)

		if count > 0 {
			continue
		}

		setting := model.TenantSetting{
			TenantID: tenant.ID,

			// Default (bisa di-custom per tenant)
			OfficeLatitude:     -6.1339179,
			OfficeLongitude:    106.8329504,
			MaxRadiusMeter:     100,
			AllowRemote:        false,
			RequireLocation:    true,
			ClockInStartTime:   "07:00",
			ClockInEndTime:     "09:00",
			LateAfterMinute:    480,
			ClockOutStartTime:  "16:00",
			ClockOutEndTime:    "23:00",
			RequireSelfie:      true,
			AllowMultipleCheck: false,
		}

		// Custom per tenant
		switch tenant.Code {

		case "remote-co":
			setting.AllowRemote = true
			setting.RequireLocation = false
			setting.RequireSelfie = false
			setting.MaxRadiusMeter = 0

		case "hybrid":
			setting.AllowRemote = true
			setting.MaxRadiusMeter = 200
		}

		if err := db.Create(&setting).Error; err != nil {
			log.Fatalf("Gagal seeder tenant setting: %v", err)
		}
	}

	log.Println("Seeder: TenantSetting berhasil ditambahkan")
}
