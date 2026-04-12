package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProvisioningTicketStatus string

const (
	ProvisioningTicketStatusWaiting   ProvisioningTicketStatus = "WAITING"
	ProvisioningTicketStatusExecuting ProvisioningTicketStatus = "EXECUTING"
	ProvisioningTicketStatusCompleted ProvisioningTicketStatus = "COMPLETED"
	ProvisioningTicketStatusFailed    ProvisioningTicketStatus = "FAILED"
)

type ProvisioningTicket struct {
	ID             uuid.UUID                `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TrialRequestID uuid.UUID                `gorm:"type:uuid;not null" json:"trial_request_id"`
	Status         ProvisioningTicketStatus `gorm:"type:varchar(20);default:'WAITING'" json:"status"`
	ErrorLog       string                   `gorm:"type:text" json:"error_log"`
	ExecutedBy     *uint                    `json:"executed_by"`
	CompletedAt    *time.Time               `json:"completed_at"`
	CreatedAt      time.Time                `json:"created_at"`

	TrialRequest TrialRequest `gorm:"foreignKey:TrialRequestID" json:"trial_request,omitempty"`
	Executor     *User        `gorm:"foreignKey:ExecutedBy" json:"executor,omitempty"`
}

func (t *ProvisioningTicket) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
