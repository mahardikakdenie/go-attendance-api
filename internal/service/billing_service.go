package service

import (
	"context"
	"fmt"
	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"
)

type BillingService interface {
	GetInvoices(ctx context.Context, tenantID uint, page, limit int, status string) ([]modelDto.InvoiceResponse, int64, error)
	GenerateInvoicePDF(ctx context.Context, tenantID uint, invoiceID string) ([]byte, error)
	GenerateExpiringInvoices(ctx context.Context) error
}

type billingService struct {
	invoiceRepo repository.InvoiceRepository
	subRepo     repository.SubscriptionRepository
}

func NewBillingService(invoiceRepo repository.InvoiceRepository, subRepo repository.SubscriptionRepository) BillingService {
	return &billingService{
		invoiceRepo: invoiceRepo,
		subRepo:     subRepo,
	}
}

func (s *billingService) GetInvoices(ctx context.Context, tenantID uint, page, limit int, status string) ([]modelDto.InvoiceResponse, int64, error) {
	invoices, total, err := s.invoiceRepo.FindAllByTenant(ctx, tenantID, page, limit, status)
	if err != nil {
		return nil, 0, err
	}

	var res []modelDto.InvoiceResponse
	for _, inv := range invoices {
		res = append(res, modelDto.InvoiceResponse{
			ID:            inv.ID,
			InvoiceNumber: inv.InvoiceNumber,
			IssuedDate:    inv.IssuedDate,
			DueDate:       inv.DueDate,
			Amount:        inv.Amount,
			Currency:      inv.Currency,
			Status:        string(inv.Status),
			Description:   inv.Description,
			PdfUrl:        inv.PdfUrl,
		})
	}

	return res, total, nil
}

func (s *billingService) GenerateInvoicePDF(ctx context.Context, tenantID uint, invoiceID string) ([]byte, error) {
	_, err := s.invoiceRepo.FindByIDAndTenant(ctx, invoiceID, tenantID)
	if err != nil {
		return nil, err
	}

	// Placeholder for real PDF generation
	dummyPDF := []byte("%PDF-1.4\n1 0 obj\n<< /Title (Invoice) >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")
	return dummyPDF, nil
}

func (s *billingService) GenerateExpiringInvoices(ctx context.Context) error {
	// Find subscriptions expiring in 7 days
	subs, err := s.subRepo.FindExpiringSubscriptions(ctx, 7)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		// Calculate the next billing period range (roughly)
		// Usually an invoice is for the UPCOMING period
		issuedDate := utils.Now()
		dueDate := sub.NextBillingDate

		// Generate a unique ID and Invoice Number
		// e.g. INV-{TenantID}-{NextBillingDate}
		invoiceNum := fmt.Sprintf("INV-%d-%s", sub.TenantID, sub.NextBillingDate.Format("20060102"))

		invoice := &model.Invoice{
			ID:            utils.GenerateRandomString(12),
			TenantID:      sub.TenantID,
			InvoiceNumber: invoiceNum,
			IssuedDate:    issuedDate,
			DueDate:       dueDate,
			Amount:        sub.Amount,
			Currency:      "IDR",
			Status:        model.InvoiceStatusUnpaid,
			Description:   fmt.Sprintf("Renewal for %s Plan", sub.Plan.Name),
		}

		// Save to DB (Create will fail if invoice_number exists due to unique constraint,
		// which acts as a natural de-duplication)
		_ = s.invoiceRepo.Create(ctx, invoice)
	}

	return nil
}
