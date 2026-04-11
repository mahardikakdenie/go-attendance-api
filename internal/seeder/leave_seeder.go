package seeder

import (
	"log"
	"time"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedLeaves(db *gorm.DB) {
	// Ambil tenant friendship
	var friendshipTenant model.Tenant
	if err := db.Where("code = ?", "friendship").First(&friendshipTenant).Error; err != nil {
		return
	}

	// 1. Seed Leave Types
	leaveTypes := []model.LeaveType{
		{TenantID: friendshipTenant.ID, Name: "Annual Leave", Code: "ANNUAL", DefaultDays: 12},
		{TenantID: friendshipTenant.ID, Name: "Sick Leave", Code: "SICK", DefaultDays: 30},
		{TenantID: friendshipTenant.ID, Name: "Emergency Leave", Code: "EMERGENCY", DefaultDays: 3},
		{TenantID: friendshipTenant.ID, Name: "Unpaid Leave", Code: "UNPAID", DefaultDays: 0},
	}

	for _, lt := range leaveTypes {
		db.FirstOrCreate(&lt, model.LeaveType{TenantID: lt.TenantID, Code: lt.Code})
	}

	// 2. Seed Balances for existing users
	var users []model.User
	db.Where("tenant_id = ?", friendshipTenant.ID).Find(&users)

	var ltAnnual model.LeaveType
	db.Where("tenant_id = ? AND code = ?", friendshipTenant.ID, "ANNUAL").First(&ltAnnual)

	var ltSick model.LeaveType
	db.Where("tenant_id = ? AND code = ?", friendshipTenant.ID, "SICK").First(&ltSick)

	currentYear := time.Now().Year()

	for _, u := range users {
		balance := model.LeaveBalance{
			UserID:      u.ID,
			LeaveTypeID: ltAnnual.ID,
			Year:        currentYear,
			Balance:     12,
		}
		db.FirstOrCreate(&balance, model.LeaveBalance{UserID: u.ID, LeaveTypeID: ltAnnual.ID, Year: currentYear})
	}

	// 3. Seed some actual approved leave requests for Dashboard
	var employee model.User
	db.Where("email = ?", "employee@friendship.com").First(&employee)

	if employee.ID != 0 {
		leaves := []model.Leave{
			{
				TenantID:    friendshipTenant.ID,
				UserID:      employee.ID,
				LeaveTypeID: ltAnnual.ID,
				StartDate:   time.Now().AddDate(0, 0, -5),
				EndDate:     time.Now().AddDate(0, 0, -3),
				Reason:      "Vacation",
				Status:      model.LeaveStatusApproved,
			},
			{
				TenantID:    friendshipTenant.ID,
				UserID:      employee.ID,
				LeaveTypeID: ltSick.ID,
				StartDate:   time.Now().AddDate(0, -1, -2),
				EndDate:     time.Now().AddDate(0, -1, -1),
				Reason:      "Flu",
				Status:      model.LeaveStatusApproved,
			},
		}

		for _, l := range leaves {
			db.FirstOrCreate(&l, model.Leave{UserID: l.UserID, StartDate: l.StartDate})
		}
	}

	log.Println("Seeder: Leave types, balances, and requests added")
}
