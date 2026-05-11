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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Gagal hash password seeder: %v", err)
	}

	// Roles
	var superAdminRole, adminRole, hrRole, financeRole, employeeRole model.Role
	db.Where("name = ?", "superadmin").First(&superAdminRole)
	db.Where("name = ?", "admin").First(&adminRole)
	db.Where("name = ?", "hr").First(&hrRole)
	db.Where("name = ?", "finance").First(&financeRole)
	db.Where("name = ?", "employee").First(&employeeRole)

	// Fetch all tenants
	var tenants []model.Tenant
	db.Find(&tenants)

	mediaUrl := "http://i.ibb.co.com/p6119B1C/attendance-1775556680532.png"

	for _, t := range tenants {
		if t.Code == "system" {
			// 1. Super Admin for System Tenant
			sa := model.User{
				Name:        "Super Admin",
				Email:       "superadmin@yopmail.com",
				Password:    string(hashedPassword),
				TenantID:    t.ID,
				RoleID:      superAdminRole.ID,
				EmployeeID:  "SA-001",
				Department:  "SaaS Owner",
				Address:     "System HQ",
				PhoneNumber: "0000000000",
			}
			db.Create(&sa)
			continue
		}

		// Create Admin for each tenant
		adminEmail := "admin@" + t.Code + ".com"
		if t.Code == "friendship" {
			adminEmail = "admin@friendship.com"
		}

		// Positions for this tenant
		var ceoPos model.Position
		db.Where("name = ? AND tenant_id = ?", "CEO", t.ID).First(&ceoPos)
		if ceoPos.ID == 0 {
			// Create a default CEO position if not exists
			ceoPos = model.Position{TenantID: t.ID, Name: "CEO", Level: 1}
			db.Create(&ceoPos)
		}

		admin := model.User{
			Name:        "Admin " + t.Name,
			Email:       adminEmail,
			Password:    string(hashedPassword),
			TenantID:    t.ID,
			RoleID:      adminRole.ID,
			PositionID:  &ceoPos.ID,
			EmployeeID:  "ADM-001",
			Department:  "Owner",
			Address:     t.Name + " Office",
			PhoneNumber: "0811111111",
			BaseSalary:  50000000,
		}
		db.Create(&admin)

		// Specific additional users for PT Friendship
		if t.Code == "friendship" {
			var managerPos, staffPos model.Position
			db.Where("name = ? AND tenant_id = ?", "Manager", t.ID).First(&managerPos)
			db.Where("name = ? AND tenant_id = ?", "Staff", t.ID).First(&staffPos)

			// HR Manager
			hr := model.User{
				Name:        "HR Manager",
				Email:       "hr@friendship.com",
				Password:    string(hashedPassword),
				TenantID:    t.ID,
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

			// Finance Manager
			financeUser := model.User{
				Name:        "Finance Manager",
				Email:       "finance@friendship.com",
				Password:    string(hashedPassword),
				TenantID:    t.ID,
				RoleID:      financeRole.ID,
				PositionID:  &managerPos.ID,
				ManagerID:   &admin.ID,
				EmployeeID:  "FIN-001",
				Department:  "Finance",
				Address:     "Friendship Office",
				PhoneNumber: "0844444444",
				MediaUrl:    mediaUrl,
				BaseSalary:  14000000,
			}
			db.Create(&financeUser)

			// Employee
			emp := model.User{
				Name:        "Employee User",
				Email:       "employee@friendship.com",
				Password:    string(hashedPassword),
				TenantID:    t.ID,
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
		}

		// Variety in Subscription Statuses
		if t.Code == "remote-co" {
			db.Model(&model.Subscription{}).Where("tenant_id = ?", t.ID).Update("status", model.SubscriptionStatusPastDue)
		}
		if t.Code == "hybrid" {
			db.Model(&model.Subscription{}).Where("tenant_id = ?", t.ID).Update("status", model.SubscriptionStatusCanceled)
		}
	}

	log.Println("Seeder: Users for all tenants and status variety added")
}
