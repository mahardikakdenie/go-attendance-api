---
name: "[FE-TASK] Integrasi Advanced HR Ops, Dynamic Attendance & Tenant Branding"
about: Panduan detail integrasi untuk modul Shift, Roster, Kalender, dan Branding Kustom
title: "[FE] Integrasi Advanced HR & Dynamic Configuration"
labels: frontend, integration, high-priority
assignees: ''
---

# 🚀 Panduan Integrasi: Advanced HR Ops & Dynamic Configuration

Dokumen ini mencakup 4 poin besar pembaruan Backend yang harus diintegrasikan ke sisi Frontend.

---

## 🏗️ 1. Modul Advanced HR (Shift, Roster & Calendar)
Kami telah menyediakan endpoint master data untuk operasional HR.

### A. Shift Management
- **GET** `/api/v1/hr/shifts` -> List semua shift (Pagi, Siang, Malam).
- **POST** `/api/v1/hr/shifts` -> Tambah shift baru.
    - **Payload**: `{"name": "...", "startTime": "HH:mm", "endTime": "HH:mm", "type": "Morning|Afternoon|Night", "color": "bg-blue-500"}`

### B. Weekly Rostering (Penjadwalan)
- **GET** `/api/v1/hr/roster?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD`
    - Mengambil jadwal mingguan karyawan. Jika hari tersebut `OFF`, value-nya adalah `"off"`.
- **POST** `/api/v1/hr/roster/save`
    - Menyimpan plot jadwal.
    - **Payload**: Lihat `SaveRosterRequest` di Swagger.

### C. Holiday Calendar
- **GET** `/api/v1/hr/calendar?year=2026` -> List hari libur nasional/kantor.
- **POST** `/api/v1/hr/calendar` -> Daftarkan hari libur baru.

---

## ⚙️ 2. Pembaruan Logic Absensi (Shift-Aware)
Endpoint `POST /api/v1/attendance` kini jauh lebih pintar. FE harus siap menangani pesan error baru:

1.  **Blocker Hari Libur**: Jika user clock-in di tanggal merah, API return `400` dengan pesan `"Hari ini adalah hari libur: [Nama Libur]"`.
2.  **Blocker Roster OFF**: Jika jadwal user hari ini adalah `OFF`, API return `400` dengan pesan `"Hari ini adalah jadwal libur (OFF) anda"`.
3.  **Dynamic Lateness**: Status `LATE` (Terlambat) kini tidak lagi mengikuti setting global, melainkan mengikuti **StartTime pada Shift** yang sedang dijalankan user hari ini.

---

## 🎨 3. Tenant Branding (Logo Dinamis)
Admin kini bisa mengunggah/mengatur logo perusahaan sendiri.

- **Field Baru**: `tenant_logo` (string/URL) telah ditambahkan ke tabel `TenantSetting`.
- **Endpoint Update**: Gunakan `PUT /api/v1/tenant-setting` untuk memperbarui logo perusahaan.
- **Implementasi FE**: Gunakan logo ini di Navbar atau Sidebar sebagai branding dinamis per perusahaan.

---

## 👤 4. Update Response `GetMe` (Optimasi Load)
Kami telah memperbarui endpoint profil untuk mengurangi jumlah hit API saat pertama kali login.

- **Endpoint**: `GET /api/v1/users/me`
- **Response Baru**: Sekarang menyertakan objek `tenant_setting` secara langsung.
- **Data yang tersedia**:
    - `user.tenant_setting.tenant_logo` -> Branding.
    - `user.tenant_setting.max_radius_meter` -> Radius Geofencing.
    - `user.tenant_setting.allow_remote` -> Izin kerja remote.
    - `user.tenant_setting.require_selfie` -> Wajib foto.

---

## 🚀 Langkah Integrasi FE
1.  **Update Global State**: Pastikan state `user` di FE menyimpan data `tenant_setting` dari response `/me`.
2.  **Branding**: Ganti logo static di header dengan `user.tenant_setting.tenant_logo`.
3.  **Validation**: Saat halaman absen dimuat, cek apakah hari ini user punya jadwal (Roster). Jika `"off"`, tampilkan banner "Hari ini Anda Libur".
4.  **Error Handling**: Tambahkan interceptor untuk menangani pesan error 400 khusus hari libur/roster agar pesan yang tampil user-friendly.

---
**Status Backend**: ✅ Completed & Tested
**Swagger**: Updated at `/swagger/index.html`
