package model

import "time"

type RecentActivity struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Action    string    `gorm:"type:varchar(100);not null" json:"action"`
	Status    string    `gorm:"type:varchar(100)" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type RecentActivityResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Action    string    `json:"action"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
