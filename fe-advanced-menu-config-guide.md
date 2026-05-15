# Panduan Implementasi FE: Konfigurasi Menu & Matriks Akses

Dokumen ini menjelaskan integrasi untuk dua fitur baru: **Halaman Tenant Info** (untuk Owner) dan **Manajemen Menu** (untuk Superadmin).

---

## 1. Halaman Tenant Info (Dashboard Owner)

**Route FE:** `/tenant-settings/info`
**Tujuan:** Memberikan transparansi kepada Owner mengenai siapa saja (role apa saja) yang bisa melihat menu tertentu di perusahaan mereka.

### API Utama:
*   **Endpoint:** `GET /api/v1/menus/overview`
*   **Akses:** Admin & Superadmin.
*   **Logic:** API ini secara otomatis menyaring role sistem (Superadmin, Support, Engineer). Hanya role yang relevan bagi tenant tersebut yang akan muncul.

### Struktur Data (JSON):
```json
[
  {
    "role_name": "HR Manager",
    "base_role": "HR",
    "menus": [
      {
        "label": "Workforce Management",
        "children": [
          { "label": "Staff Directory" },
          { "label": "Attendance Logs" }
        ]
      }
    ]
  }
]
```

### Tugas FE:
- Buat tampilan matriks atau tab. Setiap tab mewakili satu Role.
- Di dalam tab tersebut, tampilkan *tree view* atau list menu yang bisa dilihat oleh role tersebut sesuai response API.

---

## 2. Manajemen Menu (Dashboard Superadmin)

**Route FE:** `/admin/menus`
**Tujuan:** Memungkinkan Superadmin untuk mengatur secara dinamis role mana saja yang boleh melihat menu tertentu.

### API List Menu:
*   **Endpoint:** `GET /api/v1/superadmin/menus`
*   **Akses:** Superadmin Only.
*   **Tampilan:** Tampilkan semua menu dalam bentuk tabel atau list.

### API Update Konfigurasi:
*   **Endpoint:** `PUT /api/v1/superadmin/menus/:id`
*   **Method:** `PUT`
*   **Payload:**
```json
{
  "label": "Staff Directory",
  "icon": "Users",
  "allowed_roles": ["ADMIN", "HR", "FINANCE"], // Masukkan array role di sini
  "sort_order": 1,
  "is_system": false
}
```

### Tugas FE:
- Buat tombol "Edit" pada setiap baris menu.
- Gunakan **Multi-Select Checkbox** untuk field `allowed_roles`.
- Opsi role yang tersedia untuk dipilih: `SUPERADMIN`, `ADMIN`, `HR`, `FINANCE`, `EMPLOYEE`.
- Setelah sukses Update, backend akan otomatis menghapus cache, sehingga perubahan akan langsung terlihat di sidebar user setelah mereka refresh halaman atau pindah route.

---

## Ringkasan Endpoint Backend

| Fitur | Method | URL | Deskripsi |
| :--- | :--- | :--- | :--- |
| **Menu Overview** | `GET` | `/api/v1/menus/overview` | Data matriks role vs menu (untuk Owner). |
| **List All Menus** | `GET` | `/api/v1/superadmin/menus` | List semua menu untuk konfigurasi. |
| **Update Menu** | `PUT` | `/api/v1/superadmin/menus/:id` | Mengubah role yang diizinkan untuk suatu menu. |

**Catatan Keamanan:**
Jangan lupa menyertakan header standar (`Authorization` cookie) dan security headers (`X-Timestamp`, `X-Signature`) jika Anda mengetes di server staging/production.
