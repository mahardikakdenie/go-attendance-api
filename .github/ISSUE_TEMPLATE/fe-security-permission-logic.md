---
name: "🔐 [FE] Integrasi V2: Security, Permissions & Hierarchical Scoping"
about: Panduan krusial implementasi Access Control (PBAC) dan pembatasan data berdasarkan Hierarki Organisasi.
title: "[FE] Implementasi Permission & Hierarki Logic"
labels: frontend, security, critical
assignees: ''
---

# 🛡️ Panduan Keamanan Frontend: Permission & Hierarki

Backend kini menggunakan sistem **Hybrid PBAC (Permission-Based Access Control)**. Tim Frontend **DILARANG** melakukan hardcode pengecekan akses berdasarkan Nama Role (misal: `user.role === 'admin'`).

---

## 🔑 1. Identity Data (The Profil Payload)
Gunakan data dari `GET /api/v1/users/me` sebagai sumber kebenaran tunggal.

**Struktur Data Penting**:
```json
{
  "data": {
    "id": 10,
    "name": "Budi HR",
    "base_role": "HR", 
    "permissions": [
      "attendance.view",
      "attendance.export",
      "user.view",
      "user.create",
      "support.manage"
    ],
    "tenant_id": 1,
    "tenant_setting": { ... }
  }
}
```

---

## 🛠️ 2. Implementasi Permission-Based UI (PBAC)
FE harus mengecek string di dalam array `permissions` untuk menampilkan/menyembunyikan elemen UI.

**Contoh Logic yang Benar**:
- **Tombol Export**: Muncul HANYA jika `permissions.includes('attendance.export')`.
- **Menu Support HQ**: Muncul HANYA jika `permissions.includes('support.manage')`.
- **Tambah Karyawan**: Muncul HANYA jika `permissions.includes('user.create')`.

> **Rekomendasi**: Buat Helper/Hook seperti `const { hasPermission } = useAuth()` untuk mempermudah pengecekan ini di seluruh komponen.

---

## 🌳 3. Hierarchical Scoping (Penting!)
Sistem kini memiliki **"Hierarki Dahan Pohon"**. Frontend harus siap dengan behavior list yang dinamis:

- **Kasus**: Akun dengan Role **HR** membuka menu `Daftar Karyawan`.
- **Behavior BE**: Backend secara otomatis hanya mengirimkan user dengan role **Employee** (Bawahan HR). User dengan role Admin/Owner tidak akan dikirim.
- **Tugas FE**: Jangan bingung jika daftar user berbeda-beda antar akun. Backend sudah melakukan filtering otomatis berdasarkan `RequesterID` dan hierarki yang diatur di database. FE cukup menampilkan apa yang dikirim oleh BE.

---

## 🏗️ 4. BaseRole vs Custom Role
Gunakan `base_role` hanya untuk logika fundamental sistem, bukan untuk tampilan menu spesifik.

| Base Role | Kegunaan di FE |
| :--- | :--- |
| `SUPERADMIN` | Akses ke global dashboard (Tenant 1) & Provisioning. |
| `ADMIN` | Akses penuh ke dashboard tenant sendiri. |
| `HR` | Fokus pada manajemen operasional & dashboard pulse. |
| `EMPLOYEE` | Tampilan dashboard terbatas (hanya data diri sendiri). |

---

## ⚙️ 5. Dynamic Configuration (Tenant Settings)
Gunakan objek `tenant_setting` yang ada di response `/me` untuk mengatur behavior aplikasi:

1.  **Geofencing**: Ambil `max_radius_meter` untuk validasi jarak absen di sisi klien (sebagai pre-check).
2.  **Branding**: Gunakan `tenant_logo` untuk Sidebar/Navbar.
3.  **Selfie Check**: Jika `require_selfie` true, aktifkan kamera saat proses absen.

---

## 🚀 Checklist Integrasi
- [ ] Buat Hook `usePermission('slug.action')` untuk proteksi komponen/tombol.
- [ ] Implementasi Interceptor/Guard pada Router FE berdasarkan array `permissions`.
- [ ] Pastikan tidak ada hardcode string seperti `"admin"` atau `"hr"` untuk proteksi menu.
- [ ] Sinkronisasi state global dengan objek `tenant_setting` terbaru.

---
**Status Backend**: ✅ Hardened & Data-Aware
**Security Standard**: Enterprise PBAC
