# Backend Issue: Missing Tenant Roles Endpoints

**Title:** Implement Custom Tenant Roles CRUD & Hierarchy Endpoints
**Priority:** High
**Component:** User & Role Management (RBAC)

## Description
The frontend is currently implementing the Tenant Access Control (Role Management) interface at `src/views/tenant-settings/TenantRoles.tsx`. However, the API endpoints for managing these custom roles per tenant are either missing (returning 404, specifically noted for `PUT/DELETE /api/v1/tenant-roles/{id}`) or incomplete.

We need the backend team to implement a full suite of CRUD operations for Tenant Roles, along with an endpoint to manage role hierarchies (reporting lines).

---

## 📋 API Contract Specifications

All requests are expected to have the standard security headers (`X-Timestamp`, `X-Request-ID`, `X-Signature`) and authenticate the user's session to determine the `tenant_id`. Custom roles created/updated must belong to the user's tenant.

### 1. Get All Tenant Roles
*   **Endpoint:** `GET /v1/tenant-roles`
*   **Description:** Fetch all roles available for the current tenant, including system roles (`tenant_id: null`) and custom roles specific to the tenant.
*   **Response Structure:**
    ```json
    {
      "success": true,
      "meta": {
        "message": "Roles fetched successfully",
        "code": 200,
        "status": "success"
      },
      "data": [
        {
          "id": 1,
          "tenant_id": null,
          "name": "HR",
          "description": "System Default HR Role",
          "base_role": "HR",
          "is_system": true,
          "permissions": [
            { "id": "user.view", "module": "user", "action": "view" },
            { "id": "user.create", "module": "user", "action": "create" }
          ]
        },
        {
          "id": 5,
          "tenant_id": 2,
          "name": "Senior HR Specialist",
          "description": "Custom HR role with elevated access",
          "base_role": "HR",
          "is_system": false,
          "permissions": [
            { "id": "user.view", "module": "user", "action": "view" }
          ]
        }
      ]
    }
    ```

### 2. Create Custom Role
*   **Endpoint:** `POST /v1/tenant-roles`
*   **Description:** Create a new custom role for the tenant. `is_system` should automatically be `false`.
*   **Payload:**
    ```json
    {
      "name": "Senior HR Specialist",
      "description": "Handles advanced HR tasks",
      "base_role": "HR",
      "department": "Human Resources",
      "permissions": ["user.view", "attendance.view"] 
    }
    ```
    *Note: `permissions` is an array of permission ID strings.*
*   **Response:** `201 Created` with the created role object in `data`.

### 3. Update Custom Role (Permissions) - **CURRENTLY 404**
*   **Endpoint:** `PUT /v1/tenant-roles/{id}`
*   **Description:** Update the name, description, and assigned permissions of a custom role. **System roles (`is_system: true`) must not be editable.**
*   **Payload:**
    ```json
    {
      "name": "Senior HR Specialist",
      "description": "Updated description",
      "base_role": "HR",
      "permissions": ["user.view", "attendance.view", "attendance.edit"]
    }
    ```
*   **Response:** `200 OK` with the updated role object.

### 4. Delete Custom Role - **CURRENTLY 404**
*   **Endpoint:** `DELETE /v1/tenant-roles/{id}`
*   **Description:** Delete a custom role. **System roles cannot be deleted.** Ensure validation to prevent deletion if the role is currently assigned to active users.
*   **Response:**
    ```json
    {
      "success": true,
      "meta": {
        "message": "Role deleted successfully",
        "code": 200,
        "status": "success"
      },
      "data": null
    }
    ```

### 5. Save Role Hierarchy
*   **Endpoint:** `POST /v1/tenant-roles/hierarchy`
*   **Description:** Defines which roles report to a specific parent role. This establishes the approval chain (e.g., who can approve leave requests).
*   **Payload:**
    ```json
    {
      "parent_role_id": 5,
      "child_role_ids": [6, 7, 8]
    }
    ```
*   **Response:** `200 OK` indicating the hierarchy was saved.

---

## 🔑 Data Models & Enums

**`base_role` Enum:**
Must be one of: `SUPERADMIN`, `ADMIN`, `HR`, `FINANCE`, `EMPLOYEE`.

**Permissions Array mapping:**
The frontend sends an array of strings (e.g., `["attendance.view", "payroll.edit"]`). The backend should map these to the respective junction tables linking the `Role` and `Permission` entities.

## Acceptance Criteria
- [ ] `GET /v1/tenant-roles` returns a mix of system and tenant-specific roles.
- [ ] `POST /v1/tenant-roles` successfully creates a role tied to the user's tenant.
- [ ] `PUT /v1/tenant-roles/{id}` properly syncs the permission junction table (replaces old permissions with new ones).
- [ ] `DELETE /v1/tenant-roles/{id}` works for custom roles and throws an error for system roles.
- [ ] Role hierarchy endpoint correctly maps parent-child relationships for roles within the same tenant.