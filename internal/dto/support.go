package modelDto

import (
	"time"

	"go-attendance-api/internal/model"

	"github.com/google/uuid"
)

// Trial Request DTOs
type CreateTrialRequestRequest struct {
	CompanyName        string                   `json:"company_name" binding:"required"`
	ContactName        string                   `json:"contact_name" binding:"required"`
	Email              string                   `json:"email" binding:"required,email"`
	PhoneNumber        string                   `json:"phone_number"`
	EmployeeCountRange model.EmployeeCountRange `json:"employee_count_range" binding:"required"`
	Industry           string                   `json:"industry"`
}

type TrialRequestResponse struct {
	ID                 uuid.UUID                `json:"id"`
	CompanyName        string                   `json:"company_name"`
	ContactName        string                   `json:"contact_name"`
	Email              string                   `json:"email"`
	PhoneNumber        string                   `json:"phone_number"`
	EmployeeCountRange model.EmployeeCountRange `json:"employee_count_range"`
	Industry           string                   `json:"industry"`
	Status             model.TrialRequestStatus `json:"status"`
	CreatedAt          time.Time                `json:"created_at"`
}

type UpdateTrialRequestStatusRequest struct {
	Status model.TrialRequestStatus `json:"status" binding:"required"`
}

// Provisioning Ticket DTOs
type ProvisioningTicketResponse struct {
	ID             uuid.UUID                      `json:"id"`
	TrialRequestID uuid.UUID                      `json:"trial_request_id"`
	Status         model.ProvisioningTicketStatus `json:"status"`
	ErrorLog       string                         `json:"error_log"`
	ExecutedBy     *uint                          `json:"executed_by"`
	CompletedAt    *time.Time                     `json:"completed_at"`
	CreatedAt      time.Time                      `json:"created_at"`
	TrialRequest   *TrialRequestResponse          `json:"trial_request,omitempty"`
}

// Support Message DTOs
type CreateSupportMessageRequest struct {
	Subject  string                `json:"subject" binding:"required"`
	Message  string                `json:"message" binding:"required"`
	Category model.SupportCategory `json:"category" binding:"required"`
}

type SupportMessageResponse struct {
	ID        uuid.UUID             `json:"id"`
	TenantID  uint                  `json:"tenant_id"`
	UserID    uint                  `json:"user_id"`
	Subject   string                `json:"subject"`
	Message   string                `json:"message"`
	Category  model.SupportCategory `json:"category"`
	Status    model.SupportStatus   `json:"status"`
	CreatedAt time.Time             `json:"created_at"`
	Tenant    *model.TenantResponse `json:"tenant,omitempty"`
	User      *model.UserResponse   `json:"user,omitempty"`
}

type UpdateSupportMessageStatusRequest struct {
	Status model.SupportStatus `json:"status" binding:"required"`
}
