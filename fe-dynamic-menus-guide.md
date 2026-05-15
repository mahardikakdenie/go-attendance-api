# Frontend Implementation Guide: Dynamic Menu System

## 🌟 Overview
The application has transitioned from a static, hardcoded menu system to a dynamic, server-driven menu system. The backend now provides a hierarchical menu structure tailored to the user's role, permissions, and subscription plan.

## 🚀 Key Changes
1.  **Remove Static MENUS Constant**: Delete the hardcoded `MENUS` array in your frontend configuration files (e.g., `src/constants/menus.ts` or similar).
2.  **API Integration**: Fetch menus from the new endpoint after a successful login or during application initialization.
3.  **Global State**: Store the fetched menus in your global state (Redux, Pinia, or Context API).

## 📡 API Reference

### Get User Menus
Fetches the hierarchical menu structure for the currently authenticated user.

*   **Endpoint**: `GET /api/v1/menus/me`
*   **Security**: Requires valid session cookie (`access_token`).
*   **Response Structure**:
    ```json
    {
      "success": true,
      "message": "Menus retrieved successfully",
      "data": [
        {
          "id": 1,
          "key": "platform-group",
          "label": "Platform Control",
          "icon": "ShieldCheck",
          "children": [
            {
              "id": 2,
              "key": "manage-tenants",
              "label": "Tenant Directory",
              "icon": "Building2",
              "path": "/admin/tenants"
            }
          ]
        }
      ]
    }
    ```

## 🛠️ Implementation Steps

### 1. Create Menu Service/Hook
Create a hook or service to fetch menus and handle loading/error states.

```typescript
// Example using React Query
export const useMenus = () => {
  return useQuery({
    queryKey: ['menus'],
    queryFn: async () => {
      const response = await api.get('/api/v1/menus/me');
      return response.data.data;
    },
    staleTime: 24 * 60 * 60 * 1000, // Menus change rarely, cache for 24h
  });
};
```

### 2. Update Sidebar Component
Refactor your Sidebar or Navigation component to map through the data received from the API instead of the static constant.

```tsx
// Simplified Example
const Sidebar = () => {
  const { data: menus, isLoading } = useMenus();

  if (isLoading) return <SidebarSkeleton />;

  return (
    <nav>
      {menus.map((group) => (
        <MenuGroup key={group.key} data={group} />
      ))}
    </nav>
  );
};
```

### 3. Handle Icons Dynamically
The `icon` field now returns a string matching the icon name (e.g., "ShieldCheck", "Users"). Use a mapping utility or a dynamic icon component (like `lucide-react`'s dynamic import or a switch statement).

```tsx
import * as Icons from 'lucide-react';

const DynamicIcon = ({ name, ...props }) => {
  const IconComponent = Icons[name];
  return IconComponent ? <IconComponent {...props} /> : <Icons.HelpCircle {...props} />;
};
```

## ⚠️ Important Considerations
*   **Permissions**: The API already filters menus based on permissions. You no longer need to check `user.permissions.includes(...)` for the primary navigation.
*   **Subscription Plans**: Menus are automatically hidden if the module is not included in the tenant's current subscription plan.
*   **Suspension**: If a tenant is suspended, the API will only return "safe" menus (Billing, Support, Settings) to allow recovery.

---
**Status**: Ready for implementation.
**Endpoint available at**: `http://localhost:3000/api/v1/menus/me`
