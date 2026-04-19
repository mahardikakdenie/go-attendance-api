package seeder

import (
	"log"
	"time"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

func SeedProjects(db *gorm.DB) {
	// Ambil tenant friendship
	var friendshipTenant model.Tenant
	if err := db.Where("code = ?", "friendship").First(&friendshipTenant).Error; err != nil {
		return
	}

	// 1. Seed Projects
	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	endDateActive := now.AddDate(0, 2, 0)
	endDatePast := now.AddDate(0, -1, -5)

	projects := []model.Project{
		{
			TenantID:    friendshipTenant.ID,
			Name:        "Cloud Migration 2026",
			Description: "Migrating legacy infrastructure to AWS",
			ClientName:  "Acme Corp",
			StartDate:   &startDate,
			EndDate:     &endDateActive,
			Status:      model.ProjectStatusActive,
			Budget:      150000000,
		},
		{
			TenantID:    friendshipTenant.ID,
			Name:        "E-Commerce Platform",
			Description: "Redesigning the main customer portal",
			ClientName:  "Global Retail Ltd",
			StartDate:   &startDate,
			EndDate:     &endDateActive,
			Status:      model.ProjectStatusActive,
			Budget:      250000000,
		},
		{
			TenantID:    friendshipTenant.ID,
			Name:        "Internal HR Portal",
			Description: "Building a self-service dashboard for employees",
			ClientName:  "Internal",
			StartDate:   &startDate,
			EndDate:     &endDatePast,
			Status:      model.ProjectStatusCompleted,
			Budget:      50000000,
		},
	}

	for _, p := range projects {
		db.FirstOrCreate(&p, model.Project{TenantID: p.TenantID, Name: p.Name})
	}

	// 2. Seed Project Members
	var users []model.User
	db.Where("tenant_id = ?", friendshipTenant.ID).Find(&users)

	var cloudProj model.Project
	db.Where("name = ?", "Cloud Migration 2026").First(&cloudProj)

	if cloudProj.ID != 0 && len(users) > 0 {
		for _, u := range users {
			role := "Member"
			if u.Email == "admin@friendship.com" {
				role = "Lead"
			}
			member := model.ProjectMember{
				ProjectID:      cloudProj.ID,
				UserID:         u.ID,
				Role:           role,
				AllocatedHours: 40,
			}
			db.FirstOrCreate(&member, model.ProjectMember{ProjectID: member.ProjectID, UserID: member.UserID})
		}
	}

	log.Println("Seeder: Projects and Members added")
}
