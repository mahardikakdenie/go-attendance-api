# Frontend Task: Dynamic Sidebar Navigation (API Driven)

## 📌 Context
We are migrating the sidebar menu from a static, hardcoded configuration to a dynamic, database-driven system. The backend now calculates the exact menu tree a user is authorized to see based on:
1. Their **Base Role** (Admin, HR, User, etc.).
2. Their explicit **Permissions**.
3. Their tenant's **Subscription Plan** (Feature enforcement).

This ensures that "locked" features (not in the current plan) are automatically filtered out from the sidebar.

---

## 🎯 API Integration

### `GET /v1/menus/me` (NEW)
**Description:** Fetches the hierarchical menu tree for the current logged-in user.
**Authentication:** Required (Bearer Token).

**Response Type:**
```typescript
interface MenuResponse {
  id: number;
  key: string;       // Unique identifier (e.g., "manage-tenants")
  label: string;     // Display text
  icon: string;      // Name of the icon (e.g., "ShieldCheck")
  path?: string;     // URL path (optional for group items)
  children?: MenuResponse[]; // Nested sub-menus
}

interface ApiResponse {
  success: boolean;
  data: MenuResponse[];
}
```

---

## 🎨 UI/UX Requirements

### 1. Dynamic Rendering
- Replace the static `const MENUS` with a state variable that fetches from `/v1/menus/me` upon login or application load.
- Ensure the rendering logic is recursive to support nested `children`.

### 2. Icon Mapping
The API returns the icon name as a string (e.g., `"Building2"`). You need to map these strings to your icon library components (e.g., Lucide React).

**Example Mapping Helper:**
```typescript
import { Building2, CreditCard, LayoutDashboard, ... } from 'lucide-react';

const ICON_MAP: Record<string, any> = {
  "Building2": Building2,
  "CreditCard": CreditCard,
  "LayoutDashboard": LayoutDashboard,
  // ... add all icons used in the seeder
};

// Inside your SidebarItem component:
const IconComponent = ICON_MAP[item.icon] || DefaultIcon;
return <IconComponent />;
```

### 3. Loading State
Show a skeleton sidebar or a loading spinner while the menus are being fetched to prevent a "jumpy" UI experience.

---

## ✅ Success Criteria
- [ ] Sidebar menus are purely driven by the API response.
- [ ] Restricted features (e.g., "Finance" on a Basic plan) do not appear in the sidebar.
- [ ] Nested menu groups (children) expand and collapse correctly.
- [ ] Icons are rendered correctly using the mapping helper.
- [ ] No `any` types used in the menu data interfaces.

---

## 🛠️ References
- Backend PR: Implemented Hierarchical Menu Engine.
- Database Table: `menus`.
- Seeder Source: `internal/seeder/menu_seeder.go`.
