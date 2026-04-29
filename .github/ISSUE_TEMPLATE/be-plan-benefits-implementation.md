# Backend Task: Implementing Plan-Based Benefits & Trial Automation

**Title:** Implementation of Subscription Plans, Benefit Enforcement, and 14-Day Trial Automation
**Priority:** High
**Component:** Subscription & Tenant Management
**Status:** Open

## 1. Description
We need to evolve our current subscription string-based model into a robust Benefit Management System. This involves creating a master `plans` table that defines limits (e.g., max employees) and allowed features (e.g., modules) per tenant. The system must also automate the 14-day trial period during the provisioning process.

## 2. Requirements

### A. Database Model: `SubscriptionPlan`
Create a new model `Plan` (or `SubscriptionPlan`):
- `id` (PK)
- `name` (String: 'Trial', 'Basic', 'Pro', 'Enterprise')
- `max_employees` (Integer: Limit of active users)
- `features` (JSON/Array: List of allowed modules like 'attendance', 'payroll', 'finance')
- `is_active` (Boolean)

### B. Automation: 14-Day Trial Provisioning
Update `ExecuteProvisioning` logic:
1. Lookup the `PlanID` for the name 'Trial'.
2. When creating a `Subscription` for the new tenant:
   - Set `PlanID` to the Trial plan.
   - Set `Status` to 'Trial'.
   - Set `NextBillingDate` to `Now + 14 days`.

### C. Enforcement: Employee Limit
Update `internal/service/user_service.go` (`CreateUser`):
1. Fetch the tenant's current plan and employee limit.
2. Count the current number of active users in the tenant.
3. Return an error `400 Bad Request` if the limit is reached (e.g., "Employee limit reached for your Trial plan (Max: 3)").

### D. Enforcement: Feature Blocking
Create a Middleware or update `HasPermission`:
1. Check if the module of the requested permission (e.g., `payroll.edit` belongs to `payroll` module) is present in the `Tenant.Subscription.Plan.Features` list.
2. If not, return `403 Forbidden` with details on the required plan.

## 3. Data Seeder (`internal/seeder/plan_seeder.go`)
Initialize the following plans:
- **Trial**: `max_employees: 3`, `features: ["user", "attendance"]`
- **Starter**: `max_employees: 50`, `features: ["user", "attendance", "leave", "overtime"]`
- **Business**: `max_employees: 200`, `features: ["user", "attendance", "leave", "overtime", "payroll", "finance"]`
- **Enterprise**: `max_employees: 0` (unlimited), `features: ["*"]` (all)

## 4. Acceptance Criteria
- [ ] Trial tenants are automatically restricted to 3 employees.
- [ ] Users cannot access modules not included in their plan (e.g., Trial user cannot access Payroll).
- [ ] Provisioning automatically sets the 14-day expiration.
- [ ] Superadmin can successfully change a tenant's plan to 'Pro' to lift restrictions.

---
**Documented by:** CEO (Gemini CLI Agent)
**Target:** Backend Engineering Team
