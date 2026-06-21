package model

import (
	"time"

	"gorm.io/gorm"
)

type InvoiceStatus string

const (
	InvoiceStatusPaid     InvoiceStatus = "Paid"
	InvoiceStatusUnpaid   InvoiceStatus = "Unpaid"
	InvoiceStatusOverdue  InvoiceStatus = "Overdue"
	InvoiceStatusCanceled  InvoiceStatus = "Canceled"
	InvoiceStatusVerifying InvoiceStatus = "Verifying"
)

type Invoice struct {
	ID               string         `gorm:"primaryKey;type:varchar(50)" json:"id"`
	TenantID         uint           `gorm:"not null;index" json:"tenant_id"`
	Tenant           *Tenant        `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	InvoiceNumber    string         `gorm:"type:varchar(50);not null;unique" json:"invoice_number"`
	IssuedDate       time.Time      `gorm:"not null" json:"issued_date"`
	DueDate          time.Time      `gorm:"not null" json:"due_date"`
	Amount           float64        `gorm:"type:decimal(15,2);not null" json:"amount"`
	Currency         string         `gorm:"type:varchar(10);default:'IDR'" json:"currency"`
	Status           InvoiceStatus  `gorm:"type:varchar(20);not null;default:'Unpaid'" json:"status"`
	Description      string         `gorm:"type:text" json:"description"`
	PdfUrl           string         `gorm:"type:varchar(255)" json:"pdf_url"`
	TransferProofURL string         `gorm:"type:varchar(255)" json:"transfer_proof_url,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}
