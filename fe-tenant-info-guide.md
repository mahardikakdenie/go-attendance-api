# Frontend Integration Guide: Tenant Information & Settings

## 🌟 Overview
This guide details how to integrate the Tenant Information and Organizational Settings within the dashboard. These endpoints are primarily used in the **Organization Control** (Governance) section.

## 📡 API Endpoints

### 1. Get Specific Tenant Details
Used to fetch general information about a tenant (Company Name, Code, etc.).

*   **Endpoint**: `GET /api/v1/tenants/:id`
*   **Security**: Requires authenticated session.
*   **Response Structure**:
    ```json
    {
      "success": true,
      "message": "Tenant fetched successfully",
      "data": {
        "id": 2,
        "name": "PT Friendship Logistics",
        "code": "friendship",
        "is_active": true,
        "created_at": "2026-05-15T00:00:00Z"
      }
    }
    ```

### 2. Get Current Tenant Settings
Fetches organizational policies, branding (logo), and specific tenant configurations.

*   **Endpoint**: `GET /api/v1/tenant-setting`
*   **Security**: Requires `tenant.settings.view` permission.
*   **Response Structure**:
    ```json
    {
      "success": true,
      "message": "Tenant settings retrieved",
      "data": {
        "tenant_id": 2,
        "tenant_logo": "https://cdn.example.com/logos/friendship.png",
        "late_threshold": 15,
        "overtime_auto_approve": false,
        "office_start_time": "08:00",
        "office_end_time": "17:00",
        "timezone": "Asia/Jakarta"
      }
    }
    ```

### 3. Update Tenant Settings
Update organizational policies and branding.

*   **Endpoint**: `PUT /api/v1/tenant-setting`
*   **Security**: Requires `tenant.edit` or `tenant.settings.manage` permission.
*   **Payload**:
    ```json
    {
      "late_threshold": 10,
      "office_start_time": "08:30"
    }
    ```

## 🛠️ Implementation Strategy

### 1. UI Placement
- **Route**: `/tenant-settings/info`
- **Component**: `TenantInfoCard` and `OrganizationSettingsForm`.

### 2. Branding (Logo) Integration
The `tenant_logo` from the settings API should be used globally in the sidebar or navbar to reflect the tenant's brand.

```tsx
// Example Global State Hook
const useTenantBranding = () => {
  const { data: settings } = useTenantSettings();
  return {
    logo: settings?.tenant_logo || DEFAULT_LOGO,
    companyName: settings?.company_name
  };
};
```

### 3. Handling Permissions
Always wrap the "Edit" button in a permission check:

```tsx
<PermissionGuard permission="tenant.edit">
  <Button onClick={handleUpdate}>Save Changes</Button>
</PermissionGuard>
```

---
**Status**: Ready for implementation.
**Base URL**: `http://localhost:3000/api/v1`
