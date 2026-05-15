# Task Implementation: Halaman Tenant Info (Dashboard Owner)

**Deskripsi:**
Membuat halaman baru di `/tenant-settings/info` yang berfungsi sebagai pusat informasi bagi Owner (Admin) untuk melihat detail perusahaan, status langganan, dan transparansi menu yang terbuka untuk setiap role.

---

## 1. Spesifikasi Halaman
*   **Path:** `/tenant-settings/info`
*   **Akses:** Admin & Superadmin Only.
*   **Layout:** Dashboard style dengan beberapa Card/Section.

## 2. Daftar Pekerjaan (Checklist)

### A. Setup Page & Security
- [ ] Buat route baru `/tenant-settings/info`.
- [ ] Implementasikan *Page Guard* agar role lain (HR/Finance/Employee) yang mencoba mengakses langsung akan di-redirect ke Dashboard atau menampilkan *Access Denied*.

### B. Integrasi Data (API Calls)
Lakukan pemanggilan ke 3 endpoint berikut (bisa menggunakan `Promise.all` untuk efisiensi):
- [ ] **Data Langganan:** Ambil dari `GET /api/v1/users/me` (lihat properti `subscription` dan `plan_features`).
- [ ] **Data Profil Perusahaan:** Ambil dari `GET /api/v1/tenant-setting`.
- [ ] **Data Transparansi Menu:** Ambil dari `GET /api/v1/menus/overview` (Endpoint Baru).

### C. Pengembangan UI (Komponen)
- [ ] **Card: Status Langganan**
    - Tampilkan Nama Paket (misal: "Enterprise Plan").
    - Tampilkan Badge Status (Active/Trial/Past Due).
    - List fitur yang terbuka berdasarkan `plan_features` (misal: Attendance, Payroll, Performance).
- [ ] **Card: Profil Perusahaan**
    - Tampilkan Logo Perusahaan (Media URL).
    - Tampilkan Nama Perusahaan dan Detail Kontak.
- [ ] **Section: Matriks Akses Menu (Role Overview)**
    - Gunakan data dari `/menus/overview`.
    - Tampilkan daftar Role yang ada di tenant tersebut (HR, Finance, Employee, dll).
    - Untuk setiap Role, tampilkan pohon menu (tree view) atau list menu apa saja yang bisa mereka akses.
    - *Tujuan: Agar Owner tahu persis apa yang bisa dilihat oleh staffnya.*

## 3. Contoh Response Data `/menus/overview`
FE akan menerima array of objects:
```json
[
  {
    "role_name": "HR Manager",
    "base_role": "HR",
    "menus": [
      {
        "label": "Workforce Management",
        "children": [
          { "label": "Staff Directory", "path": "/employees" },
          { "label": "Attendance Logs", "path": "/attendances" }
        ]
      }
    ]
  }
]
```

## 4. Penanganan Error
- [ ] Jika API `/menus/overview` mengembalikan `403`, tampilkan pesan bahwa fitur ini memerlukan hak akses Owner.
- [ ] Tampilkan *Skeleton Loading* saat data sedang diambil.

---
**Catatan untuk Developer:**
Pastikan header keamanan (`X-Timestamp`, `X-Signature`, dll) sudah terkonfigurasi jika testing di environment staging/production.
