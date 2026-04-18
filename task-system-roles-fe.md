# [FE] Implement Platform System Roles & Permissions Management

## 📝 Overview
Replace mock data in the Superadmin "System Roles" management page with live API integration. This feature allows Root Admins to manage global role blueprints and their associated permissions.

## 🛠 Technical Requirements

### 1. API Integration
- **Fetch System Roles:** `GET /api/v1/superadmin/system-roles`
- **Fetch All Permissions:** `GET /api/v1/superadmin/permissions`
- **Create Role:** `POST /api/v1/superadmin/system-roles`
    - Payload: `{ "name", "description", "base_role", "permission_ids": [] }`
- **Update Role:** `PUT /api/v1/superadmin/system-roles/:id`
- **Delete Role:** `DELETE /api/v1/superadmin/system-roles/:id`

### 2. UI Components
- **Role Table:** Displays Name, Description, Base Role, and a list of Permissions (as tags or tooltips).
- **Create/Edit Modal:** 
    - Input for Name and Description.
    - Select for Base Role (SUPERADMIN, SUPPORT, ENGINEER, ADMIN, etc.).
    - Multi-select or Checkbox Group for Permissions (organized by Module).
- **Immutable Status:** Disable "Delete" and "Edit Permissions" buttons for roles marked as `is_immutable: true` (e.g., SUPERADMIN).

### 3. Business Rules (Frontend)
- **Base Role Mapping:** Ensure the `base_role` selection matches the `BaseRole` enum in the backend.
- **Permission Grouping:** Group permissions by their module (e.g., `user.*`, `attendance.*`) for a better user experience in the selection list.

## ✅ Acceptance Criteria
- [ ] Superadmin can view and manage all global role templates.
- [ ] Users assigned to a modified system role immediately receive the updated permissions (verified by refreshing their session).
- [ ] Deletion is blocked if the role is currently in use (API returns 500/400 with a message).
- [ ] Audit logs reflect all modifications to system roles.
