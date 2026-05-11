# Backend Task: Superadmin Tenant Detail API

## 📌 Context
To support the "See Detail" action in the Superadmin Dashboard, we need a single endpoint that aggregates all information about a tenant. This avoids the frontend having to make multiple calls to different services.

## 🎯 Requirements
Implement a new endpoint `GET /api/v1/superadmin/tenants/{id}/full-details`.

### 🚀 Data Points to Include:
1. **Tenant Profile:** Name, Code, Created Date, Status.
2. **Membership Info:** Active Plan name, Billing cycle, Amount, Next billing date.
3. **Usage Statistics:** 
   - Total Employees
   - Total Attendance records
   - Total Leave requests
   - Total Payroll records
   - Total Expense claims
4. **Employee List:** A snapshot list of current employees (Name, Email, Role, Department).

### 🛠️ API Contract
**Endpoint:** `GET /api/v1/superadmin/tenants/{id}/full-details`
**Akses:** BaseRole `SUPERADMIN` only.

**Response Structure:**
```json
{
  "meta": { "code": 200, "status": "success" },
  "data": {
    "tenant": { ... },
    "subscription": {
        "plan_name": "Enterprise",
        "status": "Active",
        "amount": 1500000,
        "next_billing_date": "..."
    },
    "usage_stats": {
        "total_employees": 45,
        "total_attendances": 1250,
        ...
    },
    "employees": [
        { "id": 1, "name": "...", "role": "admin", ... }
    ]
  }
}
```

---

## ✅ Implementation Status (Backend)
- [x] DTO `TenantFullDetailsResponse` defined in `internal/dto/superadmin.go`.
- [x] Repository method `GetTenantFullDetails` implemented with optimized counts.
- [x] Service method and Handler implemented.
- [x] Route registered under `/superadmin` group.
- [x] Compilation verified.
