# Frontend Task: Superadmin Tenant Deep-Dive Modal

## 📌 Context
We have implemented a new aggregate endpoint for the Superadmin Dashboard. Instead of the frontend making 4-5 separate calls to get a tenant's profile, subscription, employee list, and statistics, everything is now provided in a single "Deep-Dive" API.

We need to update the "See Detail" action in the Superadmin Tenant List to open a rich modal using this data.

---

## 🎯 API Integration

### `GET /v1/superadmin/tenants/{id}/full-details` (NEW)
**Description:** Fetches all necessary data for the Tenant Detail Modal.
**Authentication:** Required (Superadmin Only).

**Response Type:**
```typescript
interface TenantFullDetails {
  tenant: {
    id: number;
    name: string;
    code: string;
    created_at: string;
    is_suspended: boolean;
    suspended_reason: string;
  };
  subscription: {
    plan_name: string;
    status: string;
    amount: number;
    billing_cycle: string;
    next_billing_date: string;
  };
  usage_stats: {
    total_employees: number;
    total_attendances: number;
    total_leaves: number;
    total_payrolls: number;
    total_expenses: number;
  };
  employees: Array<{
    id: number;
    name: string;
    email: string;
    role: string;
    position: string;
    department: string;
    created_at: string;
  }>;
}
```

---

## 🎨 UI/UX Requirements

### 1. The Detail Modal Structure
Organize the modal into three or four logical sections/tabs:
- **Header:** Company Name, Badge for Account Status (Active/Suspended), and Logo.
- **Section A: Subscription:** Display the current Plan, Price, and when it expires.
- **Section B: Usage Stats:** Display the numbers in a clean "Card" or "Grid" format (e.g., "Total Payrolls: 150").
- **Section C: Employee List:** A scrollable table showing the employees belonging to this tenant.

### 2. Loading & Error States
- Show a skeleton loader or spinner while fetching the details.
- Handle the case where a tenant ID might not exist (though unlikely from the list view).

### 3. Quick Actions
- Add a button inside the modal to "Edit Profile" (pointing to the update tenant flow).
- Add a button to "Suspend/Unsuspend" the account.

---

## ✅ Success Criteria
- [ ] Clicking "See Detail" opens a modal with a loading state.
- [ ] All 4 data categories (Profile, Billing, Stats, Employees) are rendered correctly.
- [ ] Dates are formatted to a human-readable local format.
- [ ] Currency is formatted (IDR) for the subscription amount.
- [ ] No usage of `any` types in the data mapping logic.

---

## 🛠️ References
- Backend Task: `be-task-superadmin-tenant-detail.md` (Implemented)
- Swagger: `/docs/index.html#/superadmin/GetTenantDetails`
