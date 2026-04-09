---
name: "[BE-TASK] API & Logic Core untuk User Dashboard"
about: Kebutuhan endpoint, tabel, dan validasi logika untuk melayani FE User Dashboard
title: "[BE-TASK] Pembuatan API & Integrasi Logika Dashboard Karyawan"
labels: backend, api, database, enhancement
assignees: ''
---

# 📋 ISSUE: Pembuatan API & Logika Dashboard Karyawan (User Dashboard)
**Module:** Backend API (Golang)
**Target FE:** `src/views/dashboard/UserDashboard.tsx`
**Objective:** Melengkapi tabel, *Service*, *Handler*, dan validasi logika *Tenant Settings* agar komponen *User Dashboard* di Frontend bisa terintegrasi secara dinamis dan aman.

---

## 🏗️ 1. Missing Database Tables (Kekurangan Tabel)
Berdasarkan kebutuhan UI FE (Cuti, Lembur, Riwayat), Backend masih kekurangan beberapa tabel krusial untuk melayani data secara persisten:

- [ ] **Tabel `leave_types`:** Menyimpan jenis cuti per *tenant* (Tahunan, Sakit, Melahirkan, Unpaid).
  - Kolom: `id`, `tenant_id`, `name`, `default_quota`, `is_paid`.
- [ ] **Tabel `leave_balances`:** Menyimpan sisa saldo cuti masing-masing *user*.
  - Kolom: `id`, `user_id`, `tenant_id`, `leave_type_id`, `allocated`, `used`, `year`.
- [ ] **Tabel `leaves` (Pengajuan Cuti):** Menyimpan riwayat pengajuan cuti.
  - Kolom: `id`, `user_id`, `tenant_id`, `leave_type_id`, `start_date`, `end_date`, `reason`, `proof_url`, `status` (Pending, Approved, Rejected).

---

## ⚙️ 2. Validasi Kritis: Tenant Settings pada Service Absensi
Fungsi POST *Clock In* dan *Clock Out* tidak boleh hanya menerima *request* buta dari Klien. Harus ada pengecekan silang dengan tabel `tenant_settings`.

- [ ] **Validasi `is_multi_attendance` (Sangat Penting):**
  - **Jika `false` (Standard):** Sistem HANYA menerima 1x *Clock In* dan 1x *Clock Out* dalam satu hari. Jika Karyawan A mencoba *Clock In* lagi di hari yang sama, API harus me-return HTTP 400 *"Anda sudah absen masuk hari ini"*.
  - **Jika `true` (Multi/Shift/Sales):** Sistem mengizinkan beberapa *Clock In/Out* dalam sehari. Logika penyimpanan harus diubah (jangan me-replace data `clock_out` yang sudah ada, tapi buat *row* baru atau simpan dalam format JSON/array di kolom `logs`).
- [ ] **Validasi Radius Lokasi (Geofencing):**
  - API Absensi wajib menerima *payload* `latitude` dan `longitude` dari FE.
  - *Service* harus menghitung jarak (menggunakan rumus Haversine) antara kordinat Karyawan dengan `office_latitude` dan `office_longitude` dari tabel `tenant_settings`.
  - Jika jarak > `attendance_radius` (misal 100 meter), API wajib me-return HTTP 400 *"Anda berada di luar jangkauan area kantor"*.
- [ ] **Validasi IP Address / Jaringan (Opsional/Configurable):**
  - Cek apakah `tenant_settings.require_office_wifi` aktif. Jika ya, validasi *IP Address* *request* Klien dengan daftar *Whitelist IP* di *Tenant*.

---

## 🚀 3. Endpoint Handlers yang Wajib Dibuat / Diperbarui
Untuk menyuplai data ke UI Bento Grid di FE, buatkan *endpoints* RESTful berikut:

### A. Modul Kehadiran (Attendance)
- [ ] **`GET /api/v1/attendances/today`** -> Untuk komponen `TodayStatusCard`
  - **Logic:** Cari data absen `WHERE user_id = ? AND DATE(created_at) = CURDATE()`.
  - **Response:** Jam *Clock In*, Jam *Clock Out*, Total Jam Kerja, Status (Terlambat/Tepat Waktu).
- [ ] **`POST /api/v1/attendances/clock`** -> Untuk komponen `ClockCard`
  - **Payload:** `{ type: "in"|"out", latitude: float, longitude: float, photo_proof: file/base64 }`.
  - **Logic:** Terapkan validasi Poin 2 di atas. Jika FE menangani Face Recognition, BE tetap harus menyimpan `photo_proof` ke AWS S3 / *Local Storage* sebagai bukti audit (jika `tenant_settings.require_photo` aktif).

### B. Modul Cuti & Lembur (Leaves & Overtime)
- [ ] **`GET /api/v1/leaves/balances`** -> Untuk komponen `LeaveBalanceCard`
  - **Response:** Array sisa cuti (contoh: `{ annual: 10, sick: 5 }`).
- [ ] **`POST /api/v1/leaves/request`** -> Untuk komponen `LeaveRequestCard`
  - **Logic:** Cek apakah saldo cuti Karyawan mencukupi. Jika kurang, tolak dengan HTTP 400. Jika cukup, *insert* status `Pending` dan otomatis kurangi `allocated` di tabel sementara.
- [ ] **`POST /api/v1/overtimes/request`** -> Untuk komponen `OvertimeRequestCard`
  - **Logic:** Validasi apakah jam lembur bertabrakan dengan jadwal *shift* Karyawan (jika ada fitur Shift).

### C. Modul Ringkasan & Aktivitas (Summary & Activity)
- [ ] **`GET /api/v1/activities/recent`** -> Untuk komponen `RecentActivityCard` & `QuickInfoCard`
  - **Logic:** Lakukan *query UNION* atau ambil data dari tabel pengajuan Cuti, Lembur, dan *Profile Update* milik Karyawan yang berstatus `Pending`, `Approved`, atau `Rejected` yang diurutkan berdasarkan `created_at DESC` (limit 5).

---

## 🔐 4. Strict Security & Data Binding
- [ ] Semua *endpoint* di atas **WAJIB** mengekstrak `user_id` dan `tenant_id` dari token JWT Karyawan di sisi *Backend*, BUKAN mengambil dari *Payload Body* JSON. Ini mencegah *User A* mengubah data cuti *User B* dengan mengganti ID di Postman.
