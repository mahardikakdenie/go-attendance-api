# Frontend Task: Subscription Plan Model Updates (Price & Duration)

## 📌 Context
The backend `SubscriptionPlan` model has been enhanced to support dynamic pricing and expiration logic. We have added two new critical fields: `price` and `days`. These fields are now required when creating or updating plans, and they drive the automated billing and status calculation logic.

---

## 🎯 API Contract Updates

### 1. New Fields in Plan Object
All Plan-related endpoints (`GET /v1/superadmin/plans`, `GET /v1/subscriptions/plans`, etc.) now include:
```typescript
interface SubscriptionPlan {
  id: number;
  name: string;
  price: number;        // The cost of the plan (e.g., 150000)
  days: number;         // Active duration in days (e.g., 30 for monthly, 365 for yearly)
  max_employees: number;
  features: string[];
  is_active: boolean;
}
```

### 2. Update/Create Plan Payload
**Endpoints:** `POST /api/v1/superadmin/plans` and `PUT /api/v1/superadmin/plans/:id`
**New Payload Structure:**
```json
{
  "name": "Business Pro",
  "price": 500000,
  "days": 30,
  "max_employees": 100,
  "features": ["attendance", "payroll", "leave"]
}
```

---

## 🎨 UI/UX Requirements

### 1. Plan Management (Superadmin Dashboard)
- **Create/Edit Form:** Add two new input fields:
    - **Price (IDR):** Use a formatted currency input if possible.
    - **Duration (Days):** A numeric input to define how long the subscription lasts before it expires.
- **Plan List Table:** Add columns for **Price** and **Duration** to the main plans table.

### 2. Billing & Upgrade UI (Tenant Admin)
- Update the pricing cards/table to display the actual `price` fetched from the API instead of hardcoded labels.
- Display the duration clearly (e.g., "per 30 days" or "per month" derived from the `days` field).

---

## ✅ Success Criteria
- [ ] Superadmin can successfully save a plan with a custom price and day count.
- [ ] Tenant Upgrade flow uses the dynamic price from the selected plan card.
- [ ] No `any` types used for the new fields in the Plan interfaces.
- [ ] Form validation ensures `days` is at least 1 and `price` is 0 or greater.

---

## 🛠️ References
- Backend PR: Implemented dynamic pricing and expiration logic.
- Swagger: `/docs/index.html#/superadmin/UpdatePlan`
