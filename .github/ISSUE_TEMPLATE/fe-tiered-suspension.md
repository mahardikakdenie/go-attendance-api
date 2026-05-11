# Frontend Task: Tiered Tenant Suspension Handling

## 📌 Context
The backend suspension logic has been refined. Instead of blocking the entire application with a hard error, the system now allows users to log in, but restricts their access based on their **Role**. This ensures that administrators can still access billing pages to resolve the suspension while regular employees are properly informed.

---

## 🎯 Backend Behavior Update
- **Superadmin (Tenant 1):** Bypasses all suspension blocks. Full access is maintained.
- **Tenant Admin / HR:** Allowed to access `/users/me`, `/billing`, `/subscriptions`, and `/notifications`. All other endpoints (Attendance, Payroll, etc.) will return a `403 Forbidden` with an `is_suspended: true` flag.
- **Employees:** Only allowed to access `/users/me` and `/notifications`. All other actions are blocked.

---

## 🎨 UI/UX Requirements

### 1. Global Redirection (Admin/HR)
If `user.tenant.is_suspended` is true AND the user has an **Admin/HR** role:
- **Redirect:** If the user attempts to access functional modules (e.g., Dashboard, Payroll), redirect them to the **Invoice/Billing Page** (`/settings/billing`).
- **Alert Banner:** Display a persistent, prominent banner at the top of the screen:
    - **Wording:** *"Akun Anda sedang di-suspend karena: **{user.tenant.suspended_reason}**. Silakan selesaikan tagihan Anda atau hubungi Admin SaaS untuk bantuan."*

### 2. Blocking Modal (Regular Employees)
If `user.tenant.is_suspended` is true AND the user is a **Regular Employee**:
- **Overlay Modal:** Show a non-dismissible modal overlay that blocks the entire dashboard.
- **Wording:** *"Akses ditangguhkan. Akun perusahaan Anda sedang dalam masa suspend. Silakan hubungi Admin Perusahaan Anda untuk informasi lebih lanjut."*
- **Allowed Actions:** Only the "Logout" button should be clickable.

### 3. Sidebar Filtering
- If a tenant is suspended, the Sidebar menu should be filtered:
    - **For Admins:** Only show "Billing" and "Settings" menus.
    - **For Employees:** Hide all functional menus.

---

## ✅ Success Criteria
- [ ] Superadmins can still manage the system regardless of suspension.
- [ ] Tenant Admins are automatically forced to the Billing page with a clear reason.
- [ ] Employees see a polite blocking modal directing them to their own company's admin.
- [ ] All functional API calls (Attendance, Payroll) are correctly blocked at the middleware level with a 403 status.

---

## 🛠️ API Reference
- Check `user.tenant.is_suspended` and `user.tenant.suspended_reason` in the `/me` response.
- Functional APIs will return: `{"message": "...", "is_suspended": true, "reason": "..."}` when blocked.
