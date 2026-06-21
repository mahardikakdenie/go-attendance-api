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
	Subject       string                `json:"subject" binding:"required"`
	Message       string                `json:"message" binding:"required"`
	Category      model.SupportCategory `json:"category" binding:"required"`
	Priority      model.SupportPriority `json:"priority"`
	AttachmentURL string                `json:"attachment_url"`
}

type SupportMessageResponse struct {
	ID            uuid.UUID               `json:"id"`
	TenantID      uint                    `json:"tenant_id"`
	UserID        uint                    `json:"user_id"`
	Subject       string                  `json:"subject"`
	Message       string                  `json:"message"`
	Category      model.SupportCategory   `json:"category"`
	Priority      model.SupportPriority   `json:"priority"`
	Status        model.SupportStatus     `json:"status"`
	IsRead        bool                    `json:"is_read"`
	AttachmentURL string                  `json:"attachment_url,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
	Tenant        *model.TenantResponse   `json:"tenant,omitempty"`
	User          *model.UserResponse     `json:"user,omitempty"`
	AssignedTo    *SupportAssigneeResponse `json:"assigned_to,omitempty"`
}

type SupportAssigneeResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type SupportInboxFilterRequest struct {
	Search   string               `form:"search"`
	Category model.SupportCategory `form:"category"`
	Status   model.SupportStatus   `form:"status"`
	Limit    int                  `form:"limit"`
	Offset   int                  `form:"offset"`
}

type BulkSupportInboxRequest struct {
	IDs        []uuid.UUID          `json:"ids" binding:"required,min=1"`
	Action     string               `json:"action" binding:"required,oneof=MARK_READ MARK_UNREAD RESOLVE ASSIGN"`
	AssignToID *uint                `json:"assign_to_id,omitempty"`
}

type UpdateSupportReadStateRequest struct {
	IsRead bool `json:"is_read" binding:"required"`
}

type AssignSupportAgentRequest struct {
	AgentID uint `json:"agent_id" binding:"required"`
}

type BulkAssignSupportRequest struct {
	IDs     []uuid.UUID `json:"ids" binding:"required,min=1"`
	AgentID uint        `json:"agent_id" binding:"required"`
}

type BulkAssignResponse struct {
	Updated int `json:"updated"`
	Failed  int `json:"failed"`
}

type SupportAgentResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

type UpdateSupportMessageStatusRequest struct {
	Status model.SupportStatus `json:"status" binding:"required"`
}

type CreateSupportReplyRequest struct {
	Message string `json:"message" binding:"required"`
}

type SupportReplyResponse struct {
	ID        uuid.UUID           `json:"id"`
	MessageID uuid.UUID           `json:"message_id"`
	UserID    uint                `json:"user_id"`
	Message   string              `json:"message"`
	CreatedAt time.Time           `json:"created_at"`
	User      *model.UserResponse `json:"user,omitempty"`
}

// User Support History (BE-006) DTOs
type UserSupportReplyItemResponse struct {
	ID         uuid.UUID `json:"id"`
	SenderType string    `json:"sender_type"` // "ADMIN" or "USER"
	SenderName string    `json:"sender_name"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserSupportHistoryResponse struct {
	ID        uuid.UUID                       `json:"id"`
	Subject   string                          `json:"subject"`
	Category  model.SupportCategory           `json:"category"`
	Priority  model.SupportPriority           `json:"priority"`
	Status    model.SupportStatus             `json:"status"`
	Message   string                          `json:"message"`
	CreatedAt time.Time                       `json:"created_at"`
	Replies   []UserSupportReplyItemResponse  `json:"replies"`
}

type UserSupportHistoryFilterRequest struct {
	Search   string              `form:"search"`
	Status   model.SupportStatus `form:"status"`
	Priority model.SupportPriority `form:"priority"`
	Limit    int                 `form:"limit"`
	Offset   int                 `form:"offset"`
}

type UserReplySupportRequest struct {
	Message string `json:"message" binding:"required"`
}

type SupportCategoryInfo struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type SupportPriorityInfo struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Color string `json:"color"`
}
