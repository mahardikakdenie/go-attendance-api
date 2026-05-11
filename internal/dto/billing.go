package modelDto

import "time"

type InvoiceResponse struct {
	ID            string    `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	IssuedDate    time.Time `json:"issued_date"`
	DueDate       time.Time `json:"due_date"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	Description   string    `json:"description"`
	PdfUrl        string    `json:"pdf_url"`
}

type InvoicesListData struct {
	Invoices []InvoiceResponse `json:"invoices"`
}
