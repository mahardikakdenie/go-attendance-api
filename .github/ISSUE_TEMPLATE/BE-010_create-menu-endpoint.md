# BE-010: Implementation of Create Menu Endpoint

## Prioritas: **HIGH**
## Label: `backend` `superadmin` `navigation`

---

## Konteks
Saat ini di UI Superadmin (`/admin/menus`), tombol **"Create Menu"** sudah tersedia namun fungsinya belum berjalan karena backend hanya menyediakan endpoint `PUT` (Update). Kita membutuhkan endpoint `POST` untuk memungkinkan Superadmin menambah elemen navigasi baru secara dinamis ke dalam sistem.

Endpoint ini harus selaras dengan arsitektur **BE-008**, di mana visibilitas menu ditentukan oleh kolom `required_permission`.

---

## Spesifikasi API

### 1) Endpoint
- **URL:** `/v1/superadmin/menus`
- **Method:** `POST`
- **Auth:** Required (Superadmin Only)

### 2) Request Body (JSON)
| Field | Type | Required | Keterangan |
|-------|------|----------|------------|
| `label` | `string` | **Yes** | Nama tampilan menu (e.g. "Payroll Dashboard") |
| `icon` | `string` | **Yes** | Nama icon Lucide (e.g. "DollarSign", "Users") |
| `path` | `string` | No | URL tujuan (null jika hanya sebagai parent group) |
| `sort_order` | `integer` | No | Urutan tampilan (default: 0) |
| `is_system` | `boolean` | No | Jika true, menu tidak bisa didelete via UI (default: false) |
| `required_permission` | `string` | No | Permission ID dari tabel `permissions` (Nullable) |
| `parent_id` | `integer` | No | ID menu parent jika ini adalah sub-menu (Nullable) |

**Contoh Payload:**
```json
{
  "label": "Performance Review",
  "icon": "TrendingUp",
  "path": "/performance",
  "sort_order": 10,
  "is_system": false,
  "required_permission": "performance.view",
  "parent_id": null
}
```

### 3) Business Rules & Validation
- Validasi bahwa `label` dan `icon` tidak kosong.
- Jika `required_permission` dikirim, pastikan permission tersebut eksis di database.
- Jika `parent_id` dikirim, pastikan parent ID tersebut valid dan bukan menu itu sendiri (prevent circular reference).
- Gunakan transaksi database untuk memastikan integritas data.

### 4) Expected Response
- **201 Created:** Mengembalikan object menu yang baru dibuat termasuk `id` yang di-generate.
- **400 Bad Request:** Jika validasi gagal.
- **403 Forbidden:** Jika user bukan superadmin.

---

## Acceptance Criteria
- [ ] Endpoint `POST /v1/superadmin/menus` dapat diakses oleh Superadmin.
- [ ] Data tersimpan dengan benar di tabel `menus`.
- [ ] Kolom `required_permission` terintegrasi dengan benar sesuai spek BE-008.
- [ ] Unit test untuk skenario success dan failure (invalid permission/parent).
