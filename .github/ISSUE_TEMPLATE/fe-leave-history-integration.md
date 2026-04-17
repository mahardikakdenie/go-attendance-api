---
name: "📋 [FE] Integrasi: Leave History & Approval Dashboard"
about: Panduan detail integrasi UI/UX untuk menampilkan riwayat cuti dan daftar persetujuan cuti beserta filternya.
title: "[FE] Integrasi Daftar Cuti & Filter Status"
labels: frontend, hr-ops, leaves
assignees: ''
---

# 📋 Leave History & Approval Dashboard UI Guide

Task ini bertujuan untuk mengintegrasikan daftar pengajuan cuti. Endpoint ini digunakan oleh dua sisi:
1. **Karyawan (Employee)**: Untuk melihat riwayat pengajuan cuti mereka sendiri.
2. **HR / Admin**: Untuk melihat daftar pengajuan cuti dari seluruh karyawan yang membutuhkan persetujuan (Approval Dashboard).

---

## 📡 1. Endpoint Utama
Endpoint ini menangani list cuti dengan fitur **Hierarchical Scoping** otomatis (Karyawan hanya melihat miliknya, HR melihat semua).

- **Method**: `GET`
- **URL**: `/api/v1/leaves`

---

## 🎛️ 2. Query Parameters (Filter & Pagination)
Gunakan query parameter berikut untuk memanipulasi data yang ditampilkan:

| Parameter | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `status` | String | (empty) | Filter berdasarkan status. Pilihan: `pending`, `approved`, `rejected`. |
| `page` | Number | `1` | Nomor halaman data yang ingin diambil. |
| `limit` | Number | `10` | Jumlah data per halaman. |

### Contoh Penggunaan:
- **Tab "Menunggu Persetujuan" (HR)**: `/api/v1/leaves?status=pending&page=1&limit=10`
- **Tab "Riwayat Disetujui"**: `/api/v1/leaves?status=approved&page=1&limit=10`
- **Halaman 2 (Semua Status)**: `/api/v1/leaves?page=2&limit=10`

---

## 📦 3. Struktur Response
Response kini sudah dilengkapi dengan data **User (Nama & Avatar)** dan **Pagination Meta**.

```json
{
  "status": "success",
  "message": "History fetched successfully",
  "data": [
    {
      "id": 12,
      "user_id": 5,
      "user_name": "Bagus Fikri",
      "user_avatar": "https://ui-avatars.com/api/?name=Bagus+Fikri",
      "leave_type_id": 1,
      "leave_type_name": "Annual Leave",
      "start_date": "2026-04-20T00:00:00Z",
      "end_date": "2026-04-22T00:00:00Z",
      "total_days": 3,
      "reason": "Acara keluarga",
      "status": "pending",
      "created_at": "2026-04-14T10:00:00Z"
    }
  ],
  "meta": {
    "total": 45,
    "limit": 10,
    "offset": 0
  },
  "pagination": {
    "total": 45,
    "per_page": 10,
    "current_page": 1,
    "last_page": 5
  }
}
```

---

## 🖼️ 4. Panduan Implementasi UI (Frontend)

### A. Dashboard HR / Admin (Approval)
1. **Tabs System**: Buat 3 tab utama: `Pending`, `Approved`, dan `Rejected`.
2. **Tab Pending (Aktif Default)**: Saat tab ini aktif, tembak API dengan `?status=pending`. Di tab ini, tampilkan tombol **Approve** dan **Reject** (Lihat integrasi Leave Approval sebelumnya).
3. **Data Grid/Table**: Tampilkan avatar dan nama karyawan (`user_name`, `user_avatar`) agar HR mudah mengidentifikasi siapa yang mengajukan cuti tanpa harus membuka halaman profil.
4. **Pagination Controls**: Gunakan objek `pagination` dari response untuk merender tombol *Next*, *Previous*, atau angka halaman.

### B. Halaman Karyawan (My Leaves)
1. Karyawan tidak membutuhkan Tab "Pending" yang terpisah seperti HR. Cukup tampilkan satu tabel riwayat cuti (semua status).
2. Berikan label/badge warna pada kolom `status` (Misal: 🟠 Pending, 🟢 Approved, 🔴 Rejected).
3. FE tidak perlu mengirimkan `user_id`, backend otomatis mengenali identitas dari Token JWT.

---

## 🚀 Checklist Integrasi
- [ ] Implementasi parameter `status` untuk filter tab pengajuan.
- [ ] Implementasi parameter `page` dan `limit` untuk komponen Pagination.
- [ ] Tampilkan `user_name` dan `user_avatar` di tabel Approval HR.
- [ ] Tampilkan Badge Status dengan warna yang sesuai.

---
**Focus Scope**: ✅ Leave History Table | ✅ Filter Status | ✅ Pagination Data
**Target API**: `/api/v1/leaves`
