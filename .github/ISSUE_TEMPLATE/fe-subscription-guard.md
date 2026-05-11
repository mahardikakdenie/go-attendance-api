# Frontend Task: Global Subscription Guard & Redirect

## 📌 Context
To ensure system integrity, every tenant (except Superadmin) must have an active subscription plan. If a tenant has no active plan record in the database, they should not be able to access the main dashboard features and must be redirected to the Billing/Tagihan page to select a plan.

The backend has been updated to return a `200 OK` with an empty object `{}` instead of an error when no subscription is found.

---

## 🎯 Required Logic Implementation

### 1. Subscription Check (App Initialization)
On application load (e.g., in a `useAuth` hook, `App.tsx`, or a global middleware), call:
**Endpoint:** `GET /api/v1/subscriptions/me`

### 2. Guard Logic
Check the `data` field in the response:
- **Case A: `data` is empty `{}`**
  - **Action:** Redirect the user to `/settings/billing` or a dedicated `/subscription-required` page.
  - **Action:** Disable/Hide the main sidebar navigation to prevent the user from navigating away until a plan is chosen.
- **Case B: `data` has content (e.g., `status: "Active"`)**
  - **Action:** Allow normal access to the dashboard.
- **Case C: User is Superadmin**
  - **Identifier:** `user.tenant.id === 1` OR `user.tenant.plan === "Unlimited (System)"`.
  - **Action:** Always bypass this guard.

### 3. Navigation Interceptor
- Implement a global router guard (if using React Router, Next.js, etc.).
- If a user attempts to access any route other than `/billing` while having no subscription, force redirect back to `/billing`.

---

## 🎨 UI/UX Requirements
1. **Blocking Overlay:** If the subscription is missing, show a clear message: *"Your account does not have an active plan. Please select a plan to continue using the platform."*
2. **Billing Redirect:** Provide a prominent button to "View Plans".
3. **Logout Option:** Ensure the "Logout" button remains accessible even when access is blocked, so the user can switch accounts if needed.

---

## ✅ Success Criteria
- [ ] New tenants with no data in `subscriptions` table are successfully redirected to the billing page.
- [ ] Users cannot manually type URLs in the browser to "skip" the billing page.
- [ ] Superadmin (Tenant 1) is never blocked by this logic.
- [ ] Existing active tenants can access the app normally without any interruptions.

---

## 🛠️ API Reference
- **Check Endpoint:** `GET /api/v1/subscriptions/me`
- **Upgrade/Select Plan:** `POST /api/v1/subscriptions/upgrade`
