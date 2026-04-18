# Backend Task: Superadmin B2B Subscriptions & Billing API

**Title:** Implement Superadmin Subscriptions Monitoring and Actions
**Priority:** High
**Component:** Financial Operations (Superadmin Only)

## 📝 Overview
The Superadmin dashboard requires a comprehensive view of all SaaS tenants' billing statuses, Monthly Recurring Revenue (MRR), and the ability to take administrative actions on overdue accounts (e.g., sending reminders, suspending).

## 🛠 Required Endpoints

### 1. `GET /v1/superadmin/subscriptions`
Retrieves a paginated list of all tenant subscriptions along with aggregate financial statistics.

- **Access Level:** Superadmin Only (`billing.manage` or `analytics.executive` permissions)
- **Query Parameters:**
  - `page` (int, default: 1)
  - `limit` (int, default: 10)
  - `status` (string, optional: "Active" | "Past Due" | "Canceled" | "Trial")
  - `search` (string, optional: searches tenant name or code)

- **Response Format (`APIResponse<SubscriptionsResponse>`):**
```json
{
  "success": true,
  "meta": {
    "message": "Subscriptions retrieved successfully",
    "code": 200,
    "status": "OK",
    "pagination": {
      "total": 150,
      "per_page": 10,
      "current_page": 1,
      "last_page": 15
    }
  },
  "data": {
    "stats": {
      "mrr": 450000000,
      "mrr_growth": "+12.5%",
      "active_tenants": 142,
      "active_tenants_growth": "+5",
      "past_due_amount": 25000000,
      "past_due_growth": "-2.1%"
    },
    "items": [
      {
        "id": 1,
        "tenant_id": 101,
        "tenant_name": "TechCorp Indonesia",
        "tenant_code": "TCI",
        "tenant_logo": "https://url.com/logo.png",
        "plan": "Enterprise",
        "billing_cycle": "Monthly",
        "amount": 15000000,
        "status": "Active",
        "next_billing_date": "2026-05-12T00:00:00Z",
        "active_employees": 124,
        "created_at": "2024-01-10T00:00:00Z"
      }
    ],
    "total": 150
  }
}
```

### 2. `POST /v1/superadmin/subscriptions/{id}/remind`
Triggers an automated email reminder to the Tenant Owner regarding an upcoming or overdue payment.
- **Path Param:** `id` (Subscription ID)
- **Response:** 200 OK (Success message)

### 3. `POST /v1/superadmin/subscriptions/{id}/suspend`
Suspends a tenant's access to the platform (usually due to prolonged non-payment). Users of that tenant should not be able to log in or use the apps until the suspension is lifted.
- **Path Param:** `id` (Subscription ID)
- **Body:**
```json
{
  "reason": "Overdue payment for 60 days"
}
```
- **Response:** 200 OK (Success message, subscription status updated to "Canceled" or "Suspended")

## ✅ Acceptance Criteria (AC)
- [ ] Only users with Superadmin roles can access these endpoints.
- [ ] The `stats` object must accurately aggregate MRR (sum of all Active/Trial active amounts) and Past Due amounts across the entire platform.
- [ ] The suspension endpoint must properly revoke or flag the tenant in the database to intercept login attempts.
- [ ] Pagination and search filters must work efficiently.
- [ ] Standardized `APIResponse` structure is strictly followed.
