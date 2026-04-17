package modelDto

import (
	"go-attendance-api/internal/model"
)

type ExpenseResponse struct {
	ID           uint                  `json:"id"`
	ClaimID      string                `json:"claimID"` // Formatted ID like EXP-001
	EmployeeName string                `json:"employeeName"`
	Avatar       string                `json:"avatar"`
	Category     model.ExpenseCategory `json:"category"`
	Amount       float64               `json:"amount"`
	Date         string                `json:"date"`
	Description  string                `json:"description"`
	Status       model.ExpenseStatus   `json:"status"`
	ReceiptUrl   string                `json:"receiptUrl"`
	AdminNotes   string                `json:"adminNotes,omitempty"`
}

type ExpenseListResponse struct {
	Data []ExpenseResponse `json:"data"`
	Meta PaginationMeta    `json:"meta"`
}

type CreateExpenseRequest struct {
	Category    model.ExpenseCategory `json:"category" binding:"required"`
	Amount      float64               `json:"amount" binding:"required"`
	Date        string                `json:"date" binding:"required"` // YYYY-MM-DD
	Description string                `json:"description" binding:"required"`
	Receipt     string                `json:"receipt"` // URL or base64
}

type RejectExpenseRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type ExpenseTopCategory struct {
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
}

type ExpenseSummaryResponse struct {
	PendingAmount           float64            `json:"pendingAmount"`
	ApprovedThisMonthAmount float64            `json:"approvedThisMonthAmount"`
	TopCategory             ExpenseTopCategory `json:"topCategory"`
}

type UpdateQuotaRequest struct {
	Quota float64 `json:"quota" binding:"required"`
}
