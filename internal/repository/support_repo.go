package repository

import (
	"context"
	"errors"

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
	FindAllSupportMessages(ctx context.Context, includes []string) ([]model.SupportMessage, error)
	FindSupportMessageByID(ctx context.Context, id uuid.UUID, includes []string) (*model.SupportMessage, error)
	UpdateSupportMessage(ctx context.Context, message *model.SupportMessage) error

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

func (r *supportRepository) FindAllSupportMessages(ctx context.Context, includes []string) ([]model.SupportMessage, error) {
	var messages []model.SupportMessage
	query := r.db.WithContext(ctx).Model(&model.SupportMessage{})
	query = utils.ApplyPreloads(query, includes, supportPreloadMap)
	err := query.Order("created_at DESC").Find(&messages).Error
	return messages, err
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

func (r *supportRepository) UpdateSupportMessage(ctx context.Context, message *model.SupportMessage) error {
	return r.db.WithContext(ctx).Save(message).Error
}
