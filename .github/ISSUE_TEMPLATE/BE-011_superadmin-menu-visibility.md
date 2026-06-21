# BE-011: Superadmin Menu Visibility & Seeding

## Prioritas: **CRITICAL**
## Label: `backend` `superadmin` `rbac`

---

## Konteks
Halaman **Menu Management** (`/admin/menus`) sudah diimplementasikan di Frontend, namun menu navigasi tersebut belum muncul di Sidebar. Hal ini dikarenakan endpoint `/v1/menus/me` belum mengembalikan item menu tersebut untuk user dengan role Superadmin.

Sesuai dengan arsitektur **BE-008**, visibilitas menu harus berbasis `required_permission`. Kita perlu memastikan data seeder dan logic filter di backend sudah mencakup menu infrastruktur ini.

---

## Perubahan yang Dibutuhkan

### 1) Seeding Navigation Data
Pastikan menu berikut ada di database (tabel `menus`):

| Label | Icon | Path | Required Permission | Is System |
|-------|------|------|---------------------|-----------|
| Menu Management | LayoutGrid | `/admin/menus` | `superadmin.access` | `true` |

### 2) Permission Assignment
Pastikan role `superadmin` memiliki permission `superadmin.access` di tabel `role_permissions` (atau seeder terkait).

### 3) Verifikasi Endpoint `/v1/menus/me`
Pastikan logic pada `/v1/menus/me` mengembalikan menu tersebut jika user memiliki permission `superadmin.access`. 

**Pseudo-logic Check:**
```sql
SELECT * FROM menus 
WHERE required_permission IS NULL 
   OR required_permission IN (list_of_user_permissions);
```

### 4) Grouping (Opsional - Jika diperlukan)
Jika menu ini ingin dikelompokkan dalam Group tertentu (misal: "Governance"), pastikan `parent_id` diatur ke ID group yang sesuai di seeder.

---

## Acceptance Criteria
- [ ] Menu "Menu Management" muncul di response `/v1/menus/me` untuk user Superadmin.
- [ ] Menu "Menu Management" **TIDAK** muncul untuk user non-Superadmin (e.g. Employee/HR).
- [ ] Link `/admin/menus` dapat diakses dan muncul di sidebar setelah login sebagai Superadmin.
- [ ] Script migrasi/seeder diperbarui untuk menyertakan entry ini.
