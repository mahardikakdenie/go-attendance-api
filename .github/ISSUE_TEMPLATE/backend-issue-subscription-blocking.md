# Backend Task: Subscription Status Verification and Invoice Billing Endpoints

## 📌 Context
The frontend is implementing a **Subscription Blocking / Interruption Logic** on the client side.
If a tenant's subscription status is not `Active` (e.g., `Expired`, `Suspended`, `Inactive`):
1. **Admin (Owner)** users must be redirected to `/tenant-settings/billing` and have their sidebar restricted to only Invoice and Help Desk menus.
2. **Non-Admin (HR, Finance, Employee)** users must be completely blocked from the application via an overlay modal message directing them to contact their admin.

To support this logic, the frontend requires fully functional backend support for billing, subscription retrieval, and invoice operations. Currently, these endpoints are mocked or missing.

---

## 🎯 Required API Endpoints

### 1. `GET /v1/subscriptions/me` (Verification & Billing Summary)
**Description:** Fetches the subscription details of the currently authenticated tenant.
**Authentication:** Required.
**Expected Response:**
```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Subscription retrieved successfully"
  },
  "data": {
    "id": 78,
    "tenant_id": 1,
    "plan_id": 4,
    "billing_cycle": "Monthly",
    "amount": 1500000,
    "status": "Active", // "Active", "Expired", "Suspended", "Inactive"
    "next_billing_date": "2026-05-24T08:21:59Z",
    "created_at": "2026-05-10T08:21:59Z",
    "updated_at": "2026-05-10T10:20:03Z",
    "plan": {
      "id": 4,
      "name": "Enterprise",
      "max_employees": 100,
      "features": ["user", "attendance", "leave", "overtime", "payroll", "finance", "analytics"]
    }
  }
}
```

### 2. `GET /v1/billing/invoices`
**Description:** Fetches historical invoices for the authenticated tenant.
**Authentication:** Required (Tenant Admin).
**Expected Response:**
```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Invoices retrieved successfully",
    "pagination": {
      "current_page": 1,
      "last_page": 1,
      "per_page": 10,
      "total": 1
    }
  },
  "data": [
    {
      "id": 12,
      "invoice_number": "INV-2026-05-001",
      "issued_date": "2026-05-10T08:21:59Z",
      "due_date": "2026-05-17T08:21:59Z",
      "amount": 1500000,
      "status": "Paid", // "Paid", "Unpaid", "Overdue"
      "description": "Enterprise Plan - Monthly Subscription",
      "pdf_url": "https://api.attendance.app/v1/billing/invoices/12/pdf"
    }
  ]
}
```

### 3. `POST /v1/subscriptions/upgrade`
**Description:** Endpoint to handle subscription renewals, changes, or reactivation by Tenant Admins.
**Authentication:** Required (Tenant Admin).
**Payload:**
```json
{
  "plan_id": 4,
  "billing_cycle": "Monthly"
}
```
**Expected Response:**
```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Upgrade/Reactivation request initiated successfully"
  },
  "data": null
}
```

---

## 🛠️ Actions Required from Backend Team
1. Ensure `GET /v1/subscriptions/me` resolves correctly for all tenant roles.
2. Implement model structure and endpoint for `GET /v1/billing/invoices`.
3. Provide implementation for `POST /v1/subscriptions/upgrade` to allow tenants to restore active status.
4. Verify that `/user/me` endpoint correctly updates and propagates `tenant.subscription` data when status transitions to/from `Active`.
