package model

import "time"

type AttendanceRequest struct {
	KaryawanID int `json:"karyawan_id" binding:"required"`
}

type AttendanceResponse struct {
	ID         int       `json:"id"`
	KaryawanID int       `json:"karyawan_id"`
	WaktuMasuk time.Time `json:"waktu_masuk"`
	Status     string    `json:"status"`
}

type Attendance struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	User       User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	WaktuMasuk time.Time `gorm:"not null" json:"waktu_masuk"`
	Status     string    `gorm:"type:varchar(50)" json:"status"`
}
