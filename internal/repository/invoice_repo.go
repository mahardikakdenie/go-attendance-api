package repository

import (
	"context"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type InvoiceRepository interface {
	FindAllByTenant(ctx context.Context, tenantID uint, page, limit int, status string) ([]model.Invoice, int64, error)
	FindByIDAndTenant(ctx context.Context, id string, tenantID uint) (*model.Invoice, error)
	FindByID(ctx context.Context, id string) (*model.Invoice, error)
	Create(ctx context.Context, invoice *model.Invoice) error
	Update(ctx context.Context, invoice *model.Invoice) error
}

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) FindAllByTenant(ctx context.Context, tenantID uint, page, limit int, status string) ([]model.Invoice, int64, error) {
	var invoices []model.Invoice
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Invoice{}).Where("tenant_id = ?", tenantID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Offset(offset).Limit(limit).Order("issued_date DESC").Find(&invoices).Error

	return invoices, total, err
}

func (r *invoiceRepository) FindByIDAndTenant(ctx context.Context, id string, tenantID uint) (*model.Invoice, error) {
	var invoice model.Invoice
	err := r.db.WithContext(ctx).Preload("Tenant").Where("id = ? AND tenant_id = ?", id, tenantID).First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) FindByID(ctx context.Context, id string) (*model.Invoice, error) {
	var invoice model.Invoice
	err := r.db.WithContext(ctx).Preload("Tenant").Where("id = ?", id).First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) Create(ctx context.Context, invoice *model.Invoice) error {
	return r.db.WithContext(ctx).Create(invoice).Error
}

func (r *invoiceRepository) Update(ctx context.Context, invoice *model.Invoice) error {
	return r.db.WithContext(ctx).Save(invoice).Error
}
