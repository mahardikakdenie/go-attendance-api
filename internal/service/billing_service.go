package service

import (
	"bytes"
	"context"
	"fmt"
	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"
	"strings"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
)

type BillingService interface {
	GetInvoices(ctx context.Context, tenantID uint, page, limit int, status string) ([]modelDto.InvoiceResponse, int64, error)
	GenerateInvoicePDF(ctx context.Context, tenantID uint, invoiceID string) ([]byte, error)
	GenerateExpiringInvoices(ctx context.Context) error
	UploadTransferProof(ctx context.Context, tenantID uint, userID uint, invoiceID string, proofURL string) error
	VerifyInvoice(ctx context.Context, invoiceID string) error
}

type billingService struct {
	invoiceRepo  repository.InvoiceRepository
	subRepo      repository.SubscriptionRepository
	supportRepo  repository.SupportRepository
	tenantRepo   repository.TenantRepository
	userRepo     repository.UserRepository
	notifService NotificationService
}

func NewBillingService(
	invoiceRepo repository.InvoiceRepository,
	subRepo repository.SubscriptionRepository,
	supportRepo repository.SupportRepository,
	tenantRepo repository.TenantRepository,
	userRepo repository.UserRepository,
	notifService NotificationService,
) BillingService {
	return &billingService{
		invoiceRepo:  invoiceRepo,
		subRepo:      subRepo,
		supportRepo:  supportRepo,
		tenantRepo:   tenantRepo,
		userRepo:     userRepo,
		notifService: notifService,
	}
}

func (s *billingService) GetInvoices(ctx context.Context, tenantID uint, page, limit int, status string) ([]modelDto.InvoiceResponse, int64, error) {
	// Check subscription status
	sub, err := s.subRepo.FindByTenantID(ctx, tenantID)
	if err == nil && sub != nil {
		if sub.Status == model.SubscriptionStatusNonActive || sub.Status == model.SubscriptionStatusPastDue || sub.Status == model.SubscriptionStatusCanceled {
			// Check if there is an existing Unpaid invoice
			unpaidInvoices, _, err := s.invoiceRepo.FindAllByTenant(ctx, tenantID, 1, 1, string(model.InvoiceStatusUnpaid))
			if err == nil && len(unpaidInvoices) == 0 {
				// Generate new invoice based on the previous plan
				issuedDate := utils.Now()
				dueDate := issuedDate.AddDate(0, 0, 7) // Due in 7 days
				invoiceNum := fmt.Sprintf("INV-%d-%s", sub.TenantID, issuedDate.Format("20060102150405"))

				var planName string
				if sub.Plan != nil {
					planName = sub.Plan.Name
				} else {
					planName = "Previous"
				}

				invoice := &model.Invoice{
					ID:            utils.GenerateRandomString(12),
					TenantID:      sub.TenantID,
					InvoiceNumber: invoiceNum,
					IssuedDate:    issuedDate,
					DueDate:       dueDate,
					Amount:        sub.Amount,
					Currency:      "IDR",
					Status:        model.InvoiceStatusUnpaid,
					Description:   fmt.Sprintf("Renewal for %s Plan", planName),
				}
				_ = s.invoiceRepo.Create(ctx, invoice)
			}
		}
	}

	invoices, total, err := s.invoiceRepo.FindAllByTenant(ctx, tenantID, page, limit, status)
	if err != nil {
		return nil, 0, err
	}

	var res []modelDto.InvoiceResponse
	for _, inv := range invoices {
		res = append(res, modelDto.InvoiceResponse{
			ID:               inv.ID,
			InvoiceNumber:    inv.InvoiceNumber,
			IssuedDate:       inv.IssuedDate,
			DueDate:          inv.DueDate,
			Amount:           inv.Amount,
			Currency:         inv.Currency,
			Status:           string(inv.Status),
			Description:      inv.Description,
			PdfUrl:           inv.PdfUrl,
			TransferProofURL: inv.TransferProofURL,
		})
	}

	return res, total, nil
}

func (s *billingService) UploadTransferProof(ctx context.Context, tenantID uint, userID uint, invoiceID string, proofURL string) error {
	invoice, err := s.invoiceRepo.FindByIDAndTenant(ctx, invoiceID, tenantID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	invoice.TransferProofURL = proofURL
	invoice.Status = model.InvoiceStatusVerifying

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return fmt.Errorf("failed to save invoice transfer proof: %w", err)
	}

	// Automatically create support ticket
	supportMsg := &model.SupportMessage{
		ID:            uuid.New(),
		TenantID:      tenantID,
		UserID:        userID,
		Subject:       fmt.Sprintf("Billing Payment Proof - Invoice #%s", invoice.InvoiceNumber),
		Message:       fmt.Sprintf("User has uploaded payment proof for Invoice #%s (Amount: IDR %s). Please verify the attachment.\nInvoice ID: %s", invoice.InvoiceNumber, formatIDR(invoice.Amount), invoice.ID),
		Category:      model.SupportCategoryBilling,
		Priority:      model.SupportPriorityHigh,
		Status:        model.SupportStatusPending,
		AttachmentURL: proofURL,
	}

	if err := s.supportRepo.CreateSupportMessage(ctx, supportMsg); err != nil {
		fmt.Printf("Failed to create billing support ticket: %v\n", err)
	} else {
		// Notify HQ Admins (Tenant 1)
		hqAdmins, _, _ := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: 1}, nil)
		for _, admin := range hqAdmins {
			if admin.Role != nil && admin.Role.Name == "admin" {
				s.notifService.SendNotification(ctx, 1, admin.ID, "New Support Ticket", fmt.Sprintf("New ticket from Tenant %d: %s", tenantID, supportMsg.Subject), model.NotificationTypeSupport)
			}
		}

		// Notify Superadmins
		superadmins, _, _ := s.userRepo.FindAll(ctx, model.UserFilter{BaseRole: model.BaseRoleSuperAdmin}, nil)
		for _, sa := range superadmins {
			s.notifService.SendNotification(ctx, sa.TenantID, sa.ID, "Bukti Pembayaran Baru", fmt.Sprintf("Tenant %d mengunggah bukti transfer untuk Invoice #%s", tenantID, invoice.InvoiceNumber), model.NotificationTypeSubscription)
		}
	}

	return nil
}

func (s *billingService) VerifyInvoice(ctx context.Context, invoiceID string) error {
	invoice, err := s.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	if invoice.Status == model.InvoiceStatusPaid {
		return fmt.Errorf("invoice is already paid")
	}

	// 1. Mark Invoice as Paid
	invoice.Status = model.InvoiceStatusPaid
	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return err
	}

	// 2. Reactivate Subscription
	sub, err := s.subRepo.FindByTenantID(ctx, invoice.TenantID)
	if err != nil {
		return fmt.Errorf("subscription not found for tenant: %w", err)
	}

	sub.Status = model.SubscriptionStatusActive

	// Lift suspension
	if invoice.Tenant != nil {
		tenant := invoice.Tenant
		tenant.IsSuspended = false
		tenant.SuspendedReason = ""
		_ = s.tenantRepo.Update(ctx, tenant)
	} else {
		tenant, err := s.tenantRepo.FindByID(ctx, invoice.TenantID)
		if err == nil && tenant != nil {
			tenant.IsSuspended = false
			tenant.SuspendedReason = ""
			_ = s.tenantRepo.Update(ctx, tenant)
		}
	}

	// Extend NextBillingDate
	now := utils.Now()
	duration := 30 // Default 30 days
	if sub.Plan != nil && sub.Plan.Days > 0 {
		duration = sub.Plan.Days
	}
	sub.NextBillingDate = now.AddDate(0, 0, duration)

	if err := s.subRepo.Update(ctx, sub); err != nil {
		return err
	}

	// 3. Find and automatically Resolve the support ticket associated with this invoice
	tickets, _, err := s.supportRepo.FindAllSupportMessages(ctx, model.SupportMessageFilter{
		Category: model.SupportCategoryBilling,
		Status:   model.SupportStatusPending,
	}, []string{})
	if err == nil {
		for _, ticket := range tickets {
			if strings.Contains(ticket.Message, invoice.ID) {
				ticket.Status = model.SupportStatusResolved
				_ = s.supportRepo.UpdateSupportMessage(ctx, &ticket)
			}
		}
	}

	// 4. Notify Tenant Owner/Admins
	admins, _, _ := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: invoice.TenantID}, nil)
	for _, admin := range admins {
		if admin.Role != nil && (admin.Role.Name == "admin" || admin.Role.Name == "hr") {
			s.notifService.SendNotification(ctx, invoice.TenantID, admin.ID, "Subscription Restored", "Your organization subscription has been reactivated. Full access restored.", model.NotificationTypeSubscription)
		}
	}

	return nil
}

func (s *billingService) GenerateInvoicePDF(ctx context.Context, tenantID uint, invoiceID string) ([]byte, error) {
	invoice, err := s.invoiceRepo.FindByIDAndTenant(ctx, invoiceID, tenantID)
	if err != nil {
		return nil, err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetMargins(15, 15, 15)

	// Colors
	primaryColor := []int{79, 70, 229} // Indigo: #4F46E5

	// Header - Company Info & Title
	pdf.SetFont("Helvetica", "B", 24)
	pdf.SetTextColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.CellFormat(0, 10, "INVOICE", "", 0, "L", false, 0, "")
	pdf.Ln(12)

	// Company Info (Left) & Invoice Info (Right)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(75, 85, 99) // Gray-600
	
	// Left Column: Issuer Info
	pdf.CellFormat(100, 5, "AttendancePro Inc.", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Invoice Number: %s", invoice.InvoiceNumber), "", 1, "R", false, 0, "")
	
	pdf.CellFormat(100, 5, "Jakarta, Indonesia", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Issued Date: %s", invoice.IssuedDate.Format("02 Jan 2006")), "", 1, "R", false, 0, "")
	
	pdf.CellFormat(100, 5, "support@attendancepro.com", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Due Date: %s", invoice.DueDate.Format("02 Jan 2006")), "", 1, "R", false, 0, "")

	pdf.CellFormat(100, 5, "", "", 0, "L", false, 0, "")
	statusStr := string(invoice.Status)
	pdf.CellFormat(0, 5, fmt.Sprintf("Status: %s", statusStr), "", 1, "R", false, 0, "")
	
	pdf.Ln(15)

	// Bill To Info
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(17, 24, 39) // Gray-900
	pdf.CellFormat(0, 6, "BILL TO:", "", 1, "L", false, 0, "")
	
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(75, 85, 99)
	tenantName := "Tenant Organization"
	tenantCode := "N/A"
	if invoice.Tenant != nil {
		tenantName = invoice.Tenant.Name
		tenantCode = invoice.Tenant.Code
	}
	pdf.CellFormat(0, 5, tenantName, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Tenant Code: %s", tenantCode), "", 1, "L", false, 0, "")
	pdf.Ln(10)

	// Table Header
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFillColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.CellFormat(120, 8, "Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 8, "Amount", "1", 1, "R", true, 0, "")

	// Table Body
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(17, 24, 39)
	pdf.SetFillColor(249, 250, 251) // Gray-50 alternate background
	
	desc := invoice.Description
	if desc == "" {
		desc = "Subscription Plan Renewal"
	}
	
	amountStr := fmt.Sprintf("%s %.2f", invoice.Currency, invoice.Amount)
	if invoice.Currency == "IDR" {
		amountStr = fmt.Sprintf("IDR %s", formatIDR(invoice.Amount))
	}
	
	pdf.CellFormat(120, 10, desc, "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 10, amountStr, "1", 1, "R", false, 0, "")

	// Total Row
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(120, 10, "Total Amount Due", "1", 0, "R", false, 0, "")
	pdf.CellFormat(60, 10, amountStr, "1", 1, "R", false, 0, "")

	pdf.Ln(25)

	// Footer / Payment Instructions
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(0, 5, "Payment Instructions:", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(107, 114, 128)
	pdf.CellFormat(0, 5, "Please transfer the total amount to Bank Mandiri Account: 123-456-7890 a.n AttendancePro", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, "After transfer, please send proof of payment to finance@attendancepro.com", "", 1, "L", false, 0, "")

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func formatIDR(val float64) string {
	s := fmt.Sprintf("%.0f", val)
	var result []byte
	n := len(s)
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, s[i])
	}
	return string(result)
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
