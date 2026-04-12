# Backend Task: Customer Success, Trial Management & Automated Provisioning

## 1. Background
To support SaaS growth, we need a backend infrastructure that manages inbound leads (Trials), customer support communication (Inbox), and the technical execution of new tenant creation (Provisioning). This system centers around **Tenant ID 1** as the "SaaS Headquarters."

---

## 2. Updated Constants & RBAC
Ensure the `BASE_ROLE` enum/constant in the backend matches the following:
```go
const (
    BaseRoleSuperAdmin BaseRole = "SUPERADMIN"
    BaseRoleAdmin      BaseRole = "ADMIN"
    BaseRoleHR         BaseRole = "HR"
    BaseRoleFinance    BaseRole = "FINANCE"
    BaseRoleEmployee   BaseRole = "EMPLOYEE"
)
```
*Note: Any role created by Superadmin in Tenant ID 1 with a custom name (e.g., "Customer Service") must inherit from these base behaviors.*

---

## 3. Database Schema Requirements

### A. Table: `trial_requests` (Leads)
Stores data from prospective clients.
- `id` (UUID/Primary Key)
- `company_name` (String)
- `contact_name` (String)
- `email` (String, Unique)
- `phone_number` (String)
- `employee_count_range` (Enum: '1-10', '11-50', '51-200', '201+')
- `industry` (String)
- `status` (Enum: 'NEW', 'QUALIFIED', 'APPROVED', 'REJECTED')
- `created_at` (Timestamp)

### B. Table: `provisioning_tickets` (Activation Queue)
Triggered automatically when `trial_requests.status` becomes 'APPROVED'.
- `id` (UUID/Primary Key)
- `trial_request_id` (FK to trial_requests)
- `status` (Enum: 'WAITING', 'EXECUTING', 'COMPLETED', 'FAILED')
- `error_log` (Text, Nullable) - To capture failures during DB creation.
- `executed_by` (FK to users/superadmin)
- `completed_at` (Timestamp)

### C. Table: `support_messages` (Inbox)
Messages sent from active Tenant Admins to the platform.
- `id` (UUID)
- `tenant_id` (FK to tenants)
- `user_id` (FK to users - the sender)
- `subject` (String)
- `message` (Text)
- `category` (Enum: 'TECHNICAL', 'BILLING', 'FEATURE', 'OTHER')
- `status` (Enum: 'PENDING', 'IN_PROGRESS', 'RESOLVED')

---

## 4. API Endpoint Requirements

### A. Lead Capture (Public)
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/v1/public/trial-request` | Captures leads from the landing page. No auth required. |

### B. Support & Success (Auth Required: Superadmin or CS Role in Tenant 1)
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/v1/admin/support/inbox` | List all inbound support messages. |
| `PATCH` | `/v1/admin/support/inbox/{id}` | Update message status. |
| `GET` | `/v1/admin/support/trials` | List all trial requests. |
| `PATCH` | `/v1/admin/support/trials/{id}` | Update status (e.g., to 'APPROVED'). |

### C. Provisioning (High Privilege: Superadmin Only)
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/v1/admin/support/provisioning` | List all activation tickets. |
| `POST` | `/v1/admin/support/provisioning/{id}/execute` | **CRITICAL:** Triggers the automated tenant setup. |

---

## 5. Logic: The Provisioning Pipeline
When `execute` is called on a Provisioning Ticket, the backend must perform the following in a transaction:
1. **Create Tenant**: Create a new record in the `tenants` table.
2. **Setup Admin Account**: Create a new user record associated with the new `tenant_id`.
3. **Assign Base Role**: Assign the `ADMIN` base role to this user.
4. **Generate Credentials**: Generate a temporary password.
5. **Dispatch Email**: Trigger an asynchronous email worker to send the "Welcome & Credentials" email to the user.
6. **Mark Ticket**: Update `provisioning_tickets.status` to 'COMPLETED'.

---

## 6. Security Middleware
Ensure that any user with a role in **Tenant ID 1** whose `base_role` is NOT `SUPERADMIN` is restricted to only the `/v1/admin/support` endpoints if they have the specific permission (e.g., `support.manage`). They must NOT be able to access global platform settings like Subscription Pricing or Master Tenant Deletion.

---
**Documented by:** CEO (Gemini CLI Agent)
**Target:** Backend Engineering Team
