# Frontend Task: Support Inbox Role-Based Permissions

## 📌 Context
Following the recent enhancements to the Support Inbox (BE-004 & BE-005), the backend has introduced granular permissions to replace the single `support.manage` permission. The Support Desk UI must now adapt to these specific permissions to dynamically show, hide, or disable specific actions based on the logged-in user's capabilities.

---

## 🎯 New Permissions Mapping

The following new permissions have been seeded in the backend and will be returned in the user's `permissions` array upon login:

| Permission ID | UI Component / Action |
|--------------|------------------------|
| `support.view` | Access to `/admin/support` route and rendering the table. |
| `support.reply` | Ability to open a ticket and send a reply message. |
| `support.assign` | The Assignee dropdown selector for individual tickets. |
| `support.status` | Ability to mark a ticket as Resolved, In Progress, etc. |
| `support.read_state` | The Envelope icon button (Mark Read/Unread). |
| `support.bulk_action` | The master checkbox, row checkboxes, and the bulk action menu. |
| `support.manage` | Full access fallback (Superadmin bypass). |

---

## 🚀 Implementation Guide

### 1. Route Protection
Ensure the `SupportDesk` view in your router is protected. Users must have at least `support.view` or `support.manage` to access the page.

### 2. UI Component Adapters
In `src/views/admin/SupportDesk.tsx` (or your table component), wrap the action elements with your permission checker logic (e.g., `HasPermission` component or `usePermission` hook).

#### Example Handling:
- **Assign Dropdown**: Disable the `Select` component or hide it entirely if `!hasPermission('support.assign')`.
- **Envelope Icon**: Hide the read/unread toggle button if `!hasPermission('support.read_state')`.
- **Checkboxes**: Hide the first column (checkboxes) and the bulk action toolbar if `!hasPermission('support.bulk_action')`.
- **Action Menu (...)**: Filter out options like "Resolve" if `!hasPermission('support.status')`.

```tsx
{/* Example: Hiding the Assign Dropdown */}
{hasPermission('support.assign') ? (
  <AssignDropdown ticketId={row.id} currentAssignee={row.assigned_to} />
) : (
  <span className="text-gray-500">{row.assigned_to?.name || 'Unassigned'}</span>
)}
```

---

## ✅ Success Criteria
- [ ] Users without `support.assign` cannot change ticket assignments.
- [ ] Users without `support.bulk_action` cannot see or use checkboxes.
- [ ] Users without `support.read_state` cannot toggle the read/unread envelope icon.
- [ ] The UI gracefully degrades (disables/hides elements) instead of throwing API errors.
- [ ] Superadmins (or users with `support.manage`) retain full access to all features.
