package model

import "time"

type UserChangeRequestStatus string

const (
	StatusPending  UserChangeRequestStatus = "pending"
	StatusApproved UserChangeRequestStatus = "approved"
	StatusRejected UserChangeRequestStatus = "rejected"
)

type UserChangeRequest struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	UserID    uint   `gorm:"index;not null" json:"user_id"`
	User      *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TenantID  uint   `gorm:"index;not null" json:"tenant_id"`

	// Fields to change
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Department  string `json:"department"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`

	Status     UserChangeRequestStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	AdminNotes string                  `json:"admin_notes"`

	ApprovedBy *uint      `json:"approved_by"`
	ApprovedAt *time.Time `json:"approved_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserChangeRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Department  string `json:"department"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}

type ApproveUserChangeRequest struct {
	AdminNotes string `json:"admin_notes"`
}

type UserChangeRequestResponse struct {
	ID          uint                    `json:"id"`
	UserID      uint                    `json:"user_id"`
	TenantID    uint                    `json:"tenant_id"`
	Name        string                  `json:"name"`
	Email       string                  `json:"email"`
	Department  string                  `json:"department"`
	Address     string                  `json:"address"`
	PhoneNumber string                  `json:"phone_number"`
	Status      UserChangeRequestStatus `json:"status"`
	AdminNotes  string                  `json:"admin_notes"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	
	User *UserResponse `json:"user,omitempty"`
}
