# Backend Task: API Integration for Tenant Billing Dashboard

## 📌 Context
The frontend team is redesigning the **Tenant Billing Dashboard** (`/tenant-settings/billing`). The goal is to shift from a pricing-grid focus to a comprehensive billing management view that shows **Current Subscription Status**, **Upcoming Invoices**, and a **Billing History (List of Invoices)**.

To support this UI refactor, we need new API endpoints and updates to existing ones to provide invoice data.

---

## 🎯 Required API Endpoints

### 1. `GET /v1/billing/invoices` (NEW)
**Description:** Fetches a paginated list of billing invoices for the currently authenticated tenant.
**Authentication:** Required (Tenant Admin / Billing Manager).

**Query Parameters:**
- `page` (optional, default: 1)
- `limit` (optional, default: 10)
- `status` (optional) - Filter by status (e.g., `Paid`, `Unpaid`, `Overdue`).

**API Contract (Success Response):**
```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Invoices retrieved successfully",
    "pagination": {
      "current_page": 1,
      "last_page": 5,
      "per_page": 10,
      "total": 45
    }
  },
  "data": [
    {
      "id": "inv_123456789",
      "invoice_number": "INV-2026-05-001",
      "issued_date": "2026-05-01T00:00:00Z",
      "due_date": "2026-05-08T00:00:00Z",
      "amount": 1500000,
      "currency": "IDR",
      "status": "Paid", 
      "description": "Enterprise Plan - Monthly Subscription",
      "pdf_url": "https://api.attendance.app/v1/billing/invoices/inv_123456789/pdf"
    },
    {
      "id": "inv_123456788",
      "invoice_number": "INV-2026-04-001",
      "issued_date": "2026-04-01T00:00:00Z",
      "due_date": "2026-04-08T00:00:00Z",
      "amount": 1500000,
      "currency": "IDR",
      "status": "Paid",
      "description": "Enterprise Plan - Monthly Subscription",
      "pdf_url": "https://api.attendance.app/v1/billing/invoices/inv_123456788/pdf"
    }
  ]
}
```

### 2. `GET /v1/billing/invoices/{id}/pdf` (NEW)
**Description:** Generates or returns a downloadable PDF version of a specific invoice.
**Authentication:** Required (Tenant Admin / Billing Manager).

**Response:**
- `Content-Type: application/pdf`
- Binary PDF stream.

### 3. Update `GET /v1/subscriptions/me` (ENHANCEMENT)
**Description:** Ensure the current subscription endpoint returns enough data for the "Upcoming Billing" summary.

**Expected Output Verification:**
```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Subscription retrieved successfully"
  },
  "data": {
    "id": 1,
    "status": "Active",
    "amount": 1500000,
    "next_billing_date": "2026-06-01T00:00:00Z",
    "billing_cycle": "Monthly",
    "plan": {
      "id": 3,
      "name": "Enterprise",
      "max_employees": 0,
      "features": ["Feature A", "Feature B"]
    }
  }
}
```

---

## 🛠️ Actions Required from Backend Team
1. Create the **Invoice Model** (if not already existing) tied to the Tenant Subscription.
2. Implement the `GET /v1/billing/invoices` endpoint with pagination and basic filtering.
3. Implement the PDF generation endpoint for invoices.
4. Provide Postman collection updates / Swagger docs for these new endpoints once deployed to staging.
