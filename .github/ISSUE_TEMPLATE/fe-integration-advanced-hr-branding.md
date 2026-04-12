---
name: "📦 [FE] Integrasi V2: Advanced HR, Dynamic Logo & Profile Context"
about: Dokumentasi lengkap integrasi UI/UX untuk modul Shift, Roster, Branding, dan Profil.
title: "[FE] Integrasi Fitur Advanced HR & Dynamic Branding"
labels: frontend, enhancement, documentation
assignees: ''
---

# 🚀 Frontend Integration Guide: Advanced HR & Branding

Dokumen ini memuat spesifikasi teknis untuk mengintegrasikan 4 pilar perubahan utama di sisi Backend.

---

## 🏗️ 1. Modul Operasional HR (Shift & Calendar)
Digunakan oleh Admin/HR untuk mengatur ritme kerja perusahaan.

### A. Endpoint Master Data
| Method | Endpoint | Fungsi |
| :--- | :--- | :--- |
| `GET` | `/api/v1/hr/shifts` | List semua template shift kerja. |
| `POST` | `/api/v1/hr/shifts` | Membuat shift baru (Pagi/Siang/Malam). |
| `GET` | `/api/v1/hr/calendar?year=2026` | List hari libur nasional/perusahaan. |

### B. Weekly Rostering (Penjadwalan)
Gunakan endpoint ini untuk mengisi tabel jadwal mingguan.
- **GET** `/api/v1/hr/roster?start_date=2026-04-12&end_date=2026-04-18`
- **POST** `/api/v1/hr/roster/save`
- **Response Format**:
```json
{
  "status": "success",
  "data": [
    {
      "id": 1,
      "name": "Bagus Fikri",
      "department": "Engineering",
      "weeklyRoster": {
        "monday": "uuid-shift-1",
        "tuesday": "off",
        "wednesday": "uuid-shift-1"
      }
    }
  ]
}
```

---

## ⚙️ 2. Dynamic Attendance Logic (Smart Blocker)
Logic absensi kini bersifat **Shift-Aware**. FE harus menangani state "Block" berdasarkan jadwal user.

### Alur Kerja FE:
1. User membuka halaman Absen.
2. Jika hari ini di Roster adalah `"off"`, tampilkan pesan: **"Hari ini Anda Libur (OFF)"**.
3. Jika user mencoba Clock-in di Hari Libur Kalender, Backend akan me-return error:
```json
{
  "status": "error",
  "code": 400,
  "message": "Hari ini adalah hari libur: Idul Fitri 1447H"
}
```
4. **Late Calculation**: Status terlambat (`LATE`) sekarang otomatis dihitung berdasarkan `startTime` pada Shift yang sedang berjalan, bukan lagi jam 08:00 default.

---

## 🎨 3. Tenant Customization (Branding)
Setiap perusahaan (Tenant) kini memiliki identitas visual sendiri.

- **Field Baru**: `tenant_logo` (URL String).
- **Cara Update**: `PUT /api/v1/tenant-setting`.
- **Implementasi**: Gunakan URL logo ini di Sidebar atau Navbar untuk menggantikan logo default aplikasi.

---

## 👤 4. Profil User & Konfigurasi (`/me`)
Kami telah melakukan **Response Flattening** pada profil user untuk mempercepat load aplikasi.

- **Endpoint**: `GET /api/v1/users/me`
- **Struktur Response Terbaru**:
```json
{
  "data": {
    "id": 1,
    "name": "Super Admin",
    "email": "admin@hq.com",
    "tenant_id": 1,
    "tenant_setting": {
      "tenant_logo": "https://cdn.example.com/logos/my-company.png",
      "max_radius_meter": 100,
      "allow_remote": false,
      "require_selfie": true,
      "clock_in_start_time": "07:00"
    }
  }
}
```
> **FE Note**: Simpan objek `tenant_setting` ke dalam global state (Redux/Zustand/Context). Gunakan data ini untuk mengatur validasi Geofencing dan tampilan UI secara dinamis tanpa perlu hit API tambahan.

---

## 🚀 Checklist Integrasi
- [ ] Implementasi Sidebar Logo menggunakan `user.tenant_setting.tenant_logo`.
- [ ] Update validasi Geofencing FE menggunakan `user.tenant_setting.max_radius_meter`.
- [ ] Menambahkan banner "OFF" pada dashboard jika jadwal hari ini adalah libur.
- [ ] Sinkronisasi warna shift (Emerald untuk Pagi, Amber untuk Sore, Slate untuk Malam).

---
**Dokumentasi Teknis**: ✅ Selesai
**Status Backend**: ✅ Production Ready
