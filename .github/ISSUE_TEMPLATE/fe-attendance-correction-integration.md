---
name: "✍️ [FE] Integrasi: Request Attendance (Koreksi Absensi)"
about: Panduan integrasi UI/UX untuk fitur pengajuan koreksi absensi bagi karyawan dan persetujuan bagi HR/Admin.
title: "[FE] Integrasi Fitur Request Attendance Correction"
labels: frontend, user-module, attendance
assignees: ''
---

# ✍️ Request Attendance Correction UI Guide

Task ini bertujuan untuk mengimplementasikan fitur pengajuan koreksi absensi (misal: lupa absen masuk/pulang) dan dashboard persetujuan untuk level HR/Admin.

---

## 🖼️ 1. Layout Concept

### A. Form Pengajuan (Karyawan)
Berada di dashboard absen atau halaman riwayat absen.
- **Fields**:
    - `attendance_id`: (Hidden/Optional) ID absen yang ingin dikoreksi (jika ada).
    - `date`: DatePicker (Format: YYYY-MM-DD).
    - `clock_in_time`: TimePicker (Format: HH:mm:ss).
    - `clock_out_time`: TimePicker (Format: HH:mm:ss).
    - `reason`: Textarea (Alasan koreksi).

### B. Dashboard Approval (HR/Admin)
Tabel khusus untuk melihat daftar pengajuan koreksi yang masuk.
- **Columns**: Nama Karyawan, Tanggal, Jam Masuk Baru, Jam Pulang Baru, Alasan, Status.
- **Actions**: Tombol **Approve** (Warna Hijau) & **Reject** (Warna Merah). Saat klik aksi, muncul modal untuk mengisi `admin_notes`.

---

## 📡 2. Integrasi API

### A. Submit Request Correction (Employee)
- **Method**: `POST`
- **URL**: `/api/v1/attendance/corrections`
- **Payload**:
```json
{
  "attendance_id": "uuid-optional",
  "date": "2026-04-12",
  "clock_in_time": "08:00:00",
  "clock_out_time": "17:00:00",
  "reason": "Lupa clock out karena buru-buru"
}
```

### B. Fetch Corrections List
- **Method**: `GET`
- **URL**: `/api/v1/attendance/corrections?status=PENDING&page=1&limit=10`
- **Behavior**: Jika login sebagai karyawan biasa, hanya akan muncul data miliknya sendiri. Jika Admin/HR, akan muncul semua data satu tenant.

### C. Approve/Reject (HR/Admin)
- **Method**: `POST`
- **URL**: `/api/v1/attendance/corrections/{id}/approve` (atau `/reject`)
- **Payload**:
```json
{
  "admin_notes": "Disetujui setelah konfirmasi dengan atasan langsung."
}
```

---

## 🔐 3. Aturan Bisnis & Logika
1. **Auto-Update**: Saat HR menekan **Approve**, Backend secara otomatis akan mengupdate tabel `attendances` yang bersangkutan atau membuat record absen baru jika sebelumnya tidak ada.
2. **Future Date Blocker**: User dilarang mengajukan koreksi untuk tanggal di masa depan.
3. **Status Lock**: Pengajuan yang sudah `APPROVED` atau `REJECTED` tidak dapat diubah lagi statusnya.

---

## 🚀 Checklist Integrasi
- [ ] Implementasi Form Pengajuan Koreksi.
- [ ] Implementasi Dashboard Review untuk Admin/HR.
- [ ] Implementasi parameter `page` dan `limit` untuk pagination.
- [ ] Tampilkan status pengajuan (Badge: Orange untuk Pending, Green untuk Approved, Red untuk Rejected).

- [ ] Handle error response dari BE (misal: "request is already processed").

---
**Focus Scope**: ✅ Correction Form | ✅ Admin Approval Dashboard | ✅ Attendance Table Sync
**Target API**: `/api/v1/attendance/corrections`
