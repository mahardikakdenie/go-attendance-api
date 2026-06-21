package repository

import (
	"context"
	"errors"
	"strings"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTrialRequestNotFound       = errors.New("trial request not found")
	ErrProvisioningTicketNotFound = errors.New("provisioning ticket not found")
	ErrSupportMessageNotFound     = errors.New("support message not found")
)

type SupportRepository interface {
	// Trial Request
	CreateTrialRequest(ctx context.Context, trial *model.TrialRequest) error
	FindAllTrialRequests(ctx context.Context) ([]model.TrialRequest, error)
	FindTrialRequestByID(ctx context.Context, id uuid.UUID) (*model.TrialRequest, error)
	UpdateTrialRequest(ctx context.Context, trial *model.TrialRequest) error

	// Provisioning Ticket
	CreateProvisioningTicket(ctx context.Context, ticket *model.ProvisioningTicket) error
	FindAllProvisioningTickets(ctx context.Context, includes []string) ([]model.ProvisioningTicket, error)
	FindProvisioningTicketByID(ctx context.Context, id uuid.UUID, includes []string) (*model.ProvisioningTicket, error)
	UpdateProvisioningTicket(ctx context.Context, ticket *model.ProvisioningTicket) error

	// Support Message
	CreateSupportMessage(ctx context.Context, message *model.SupportMessage) error
	FindAllSupportMessages(ctx context.Context, filter model.SupportMessageFilter, includes []string) ([]model.SupportMessage, int64, error)
	FindAllUserSupportMessages(ctx context.Context, userID uint, filter model.SupportMessageFilter, includes []string) ([]model.SupportMessage, int64, error)
	FindSupportMessageByID(ctx context.Context, id uuid.UUID, includes []string) (*model.SupportMessage, error)
	UpdateSupportMessage(ctx context.Context, message *model.SupportMessage) error
	BulkUpdateSupportMessages(ctx context.Context, ids []uuid.UUID, updates map[string]interface{}) error

	// Support Reply
	CreateReply(ctx context.Context, reply *model.SupportReply) error
	FindRepliesByMessageID(ctx context.Context, messageID uuid.UUID) ([]model.SupportReply, error)

	// Transaction
	Transaction(ctx context.Context, fn func(repo SupportRepository) error) error
}

type supportRepository struct {
	db *gorm.DB
}

func NewSupportRepository(db *gorm.DB) SupportRepository {
	return &supportRepository{db: db}
}

func (r *supportRepository) Transaction(ctx context.Context, fn func(repo SupportRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewSupportRepository(tx))
	})
}

var supportPreloadMap = map[string]string{
	"trial_request": "TrialRequest",
	"executor":      "Executor",
	"tenant":        "Tenant",
	"user":          "User",
	"assigned_to":   "AssignedTo",
}

// Trial Request
func (r *supportRepository) CreateTrialRequest(ctx context.Context, trial *model.TrialRequest) error {
	return r.db.WithContext(ctx).Create(trial).Error
}

func (r *supportRepository) FindAllTrialRequests(ctx context.Context) ([]model.TrialRequest, error) {
	var trials []model.TrialRequest
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&trials).Error
	return trials, err
}

func (r *supportRepository) FindTrialRequestByID(ctx context.Context, id uuid.UUID) (*model.TrialRequest, error) {
	var trial model.TrialRequest
	err := r.db.WithContext(ctx).First(&trial, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTrialRequestNotFound
		}
		return nil, err
	}
	return &trial, nil
}

func (r *supportRepository) UpdateTrialRequest(ctx context.Context, trial *model.TrialRequest) error {
	return r.db.WithContext(ctx).Save(trial).Error
}

// Provisioning Ticket
func (r *supportRepository) CreateProvisioningTicket(ctx context.Context, ticket *model.ProvisioningTicket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

func (r *supportRepository) FindAllProvisioningTickets(ctx context.Context, includes []string) ([]model.ProvisioningTicket, error) {
	var tickets []model.ProvisioningTicket
	query := r.db.WithContext(ctx).Model(&model.ProvisioningTicket{})
	query = utils.ApplyPreloads(query, includes, supportPreloadMap)
	err := query.Order("created_at DESC").Find(&tickets).Error
	return tickets, err
}

func (r *supportRepository) FindProvisioningTicketByID(ctx context.Context, id uuid.UUID, includes []string) (*model.ProvisioningTicket, error) {
	var ticket model.ProvisioningTicket
	query := r.db.WithContext(ctx).Model(&model.ProvisioningTicket{})
	query = utils.ApplyPreloads(query, includes, supportPreloadMap)
	err := query.First(&ticket, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProvisioningTicketNotFound
		}
		return nil, err
	}
	return &ticket, nil
}

func (r *supportRepository) UpdateProvisioningTicket(ctx context.Context, ticket *model.ProvisioningTicket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

// Support Message
func (r *supportRepository) CreateSupportMessage(ctx context.Context, message *model.SupportMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *supportRepository) FindAllSupportMessages(ctx context.Context, filter model.SupportMessageFilter, includes []string) ([]model.SupportMessage, int64, error) {
	var messages []model.SupportMessage
	var total int64

	query := r.db.WithContext(ctx).Model(&model.SupportMessage{}).
		Joins("LEFT JOIN tenants ON tenants.id = support_messages.tenant_id").
		Joins("LEFT JOIN users sender_users ON sender_users.id = support_messages.user_id")

	if filter.Search != "" {
		pattern := "%" + strings.TrimSpace(filter.Search) + "%"
		query = query.Where(
			"support_messages.subject ILIKE ? OR support_messages.message ILIKE ? OR sender_users.name ILIKE ? OR tenants.name ILIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}

	if filter.Category != "" {
		query = query.Where("support_messages.category = ?", filter.Category)
	}

	if filter.Status != "" {
		query = query.Where("support_messages.status = ?", filter.Status)
	}

	if filter.Priority != "" {
		query = query.Where("support_messages.priority = ?", filter.Priority)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = utils.ApplyPreloads(query, includes, supportPreloadMap)

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit).Offset(filter.Offset)
	}

	err := query.Order("support_messages.created_at DESC").Find(&messages).Error
	return messages, total, err
}

func (r *supportRepository) FindSupportMessageByID(ctx context.Context, id uuid.UUID, includes []string) (*model.SupportMessage, error) {
	var message model.SupportMessage
	query := r.db.WithContext(ctx).Model(&model.SupportMessage{})
	query = utils.ApplyPreloads(query, includes, supportPreloadMap)
	err := query.First(&message, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupportMessageNotFound
		}
		return nil, err
	}
	return &message, nil
}

func (r *supportRepository) FindAllUserSupportMessages(ctx context.Context, userID uint, filter model.SupportMessageFilter, includes []string) ([]model.SupportMessage, int64, error) {
	var messages []model.SupportMessage
	var total int64

	query := r.db.WithContext(ctx).Model(&model.SupportMessage{}).
		Where("support_messages.user_id = ?", userID)

	if filter.Search != "" {
		pattern := "%" + strings.TrimSpace(filter.Search) + "%"
		query = query.Where(
			"support_messages.subject ILIKE ? OR support_messages.message ILIKE ?",
			pattern, pattern,
		)
	}

	if filter.Status != "" {
		query = query.Where("support_messages.status = ?", filter.Status)
	}

	if filter.Priority != "" {
		query = query.Where("support_messages.priority = ?", filter.Priority)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = utils.ApplyPreloads(query, includes, supportPreloadMap)

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit).Offset(filter.Offset)
	}

	// Dynamic deep preloads for replies: Preload replies, replies' user, and replies' user's role
	// GORM preloads them sequentially
	query = query.Preload("Replies").
		Preload("Replies.User").
		Preload("Replies.User.Role")

	err := query.Order("support_messages.created_at DESC").Find(&messages).Error
	return messages, total, err
}

func (r *supportRepository) UpdateSupportMessage(ctx context.Context, message *model.SupportMessage) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *supportRepository) BulkUpdateSupportMessages(ctx context.Context, ids []uuid.UUID, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.SupportMessage{}).
		Where("id IN ?", ids).
		Updates(updates).Error
}

func (r *supportRepository) CreateReply(ctx context.Context, reply *model.SupportReply) error {
	return r.db.WithContext(ctx).Create(reply).Error
}

func (r *supportRepository) FindRepliesByMessageID(ctx context.Context, messageID uuid.UUID) ([]model.SupportReply, error) {
	var replies []model.SupportReply
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("message_id = ?", messageID).
		Order("created_at ASC").
		Find(&replies).Error
	return replies, err
}
