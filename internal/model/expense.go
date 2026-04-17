package model

import (
	"time"

	"gorm.io/gorm"
)

type ExpenseStatus string

const (
	ExpenseStatusPending  ExpenseStatus = "Pending"
	ExpenseStatusApproved ExpenseStatus = "Approved"
	ExpenseStatusRejected ExpenseStatus = "Rejected"
)

type ExpenseCategory string

const (
	ExpenseCategoryTravel    ExpenseCategory = "Travel"
	ExpenseCategoryMedical   ExpenseCategory = "Medical"
	ExpenseCategorySupplies  ExpenseCategory = "Supplies"
	ExpenseCategoryEquipment ExpenseCategory = "Equipment"
	ExpenseCategoryOther     ExpenseCategory = "Other"
)

type Expense struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"index;not null" json:"tenant_id"`
	UserID      uint           `gorm:"index;not null" json:"user_id"`
	Category    ExpenseCategory `gorm:"type:varchar(50);not null" json:"category"`
	Amount      float64        `gorm:"type:decimal(15,2);not null" json:"amount"`
	Date        time.Time      `gorm:"type:date;not null" json:"date"`
	Description string         `gorm:"type:text" json:"description"`
	Status      ExpenseStatus  `gorm:"type:varchar(20);default:'Pending'" json:"status"`
	ReceiptUrl  string         `gorm:"type:varchar(255)" json:"receipt_url"`
	AdminNotes  string         `gorm:"type:text" json:"admin_notes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

type ExpenseFilter struct {
	TenantID       uint
	UserID         uint
	Status         ExpenseStatus
	Search         string
	AllowedRoleIDs []uint
	Limit          int
	Offset         int
}
