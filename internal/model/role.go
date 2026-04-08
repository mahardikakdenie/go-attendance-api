package model

import "time"

type Role struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(50);unique;not null" json:"name" example:"admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RoleResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
