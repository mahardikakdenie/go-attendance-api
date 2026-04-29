# Frontend Task: Plan-Based Feature Locking & Capacity UI

**Title:** Implement Plan-Based Feature Locking, Employee Caps, and Trial Countdown
**Priority:** High
**Component:** Global UI / Access Control
**Status:** Open

## 1. Description
The backend now enforces subscription plan limits (Trial, Starter, Business, Enterprise). We need to update the UI to reflect these restrictions, provide feedback to users when they hit limits, and show a countdown for Trial periods.

## 2. Requirements

### A. Dynamic Sidebar & Module Access
- Update the Sidebar navigation logic to check both `permissions` AND `plan_features` (now available in the `/me` response).
- If a module (e.g., `payroll`) is not in `plan_features`, the menu item should either be **hidden** or **disabled with a "Lock" icon**.
- Modules to check: `user`, `attendance`, `leave`, `overtime`, `payroll`, `finance`.
- Special Case: If `plan_features` contains `*`, allow all modules.

### B. Employee Capacity Indicator
- In the **Employee Directory** page, display a capacity indicator (e.g., a progress bar or text: "3 of 50 employees used").
- Use the `max_employees` info from the tenant's plan (available via `/api/v1/superadmin/subscriptions/me` or updated `/me`).
- If the user tries to create a new employee and the limit is reached, show a clear "Upgrade Required" modal instead of just a raw API error.

### C. Trial Banner & Countdown
- If the tenant status is `Trial`, display a non-intrusive banner at the top of the dashboard.
- Text: "Your Trial expires in **X days**. [Upgrade Now]"
- Calculate days remaining using `next_billing_date` from the subscription data.

### D. Global 403 Error Handling
- Update the API interceptor to catch 403 errors with the specific message: `"Feature not available in your current plan"`.
- Instead of a generic "Forbidden" toast, show a branded "Feature Locked" popup with a Call-to-Action to contact the administrator or upgrade.

## 3. Technical Specifications

### Data Source: `GET /api/v1/users/me`
```json
{
  "data": {
    "plan_features": ["user", "attendance"],
    "tenant": {
      "plan": "Trial"
    },
    ...
  }
}
```

### Implementation Logic (Pseudo-code)
```typescript
const hasModuleAccess = (moduleName) => {
  if (user.plan_features.includes('*')) return true;
  return user.plan_features.includes(moduleName);
}
```

## 4. Acceptance Criteria
- [ ] Users on **Trial** plan cannot see/access Payroll or Finance menus.
- [ ] Employee Directory shows "X / 3" for Trial users.
- [ ] Trial countdown banner appears correctly.
- [ ] Clicking a locked feature or hitting the employee cap triggers an "Upgrade" CTA.

---
**Prepared by:** Gemini CLI Agent (Backend Implementation Team)
**Target:** Frontend Engineering Team
