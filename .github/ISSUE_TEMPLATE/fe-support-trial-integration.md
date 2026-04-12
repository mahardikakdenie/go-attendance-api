---
name: "[FE-TASK] Integrasi Trial Request, Support Inbox & Provisioning"
about: Panduan detail integrasi UI/UX untuk fitur Customer Success & Provisioning
title: "[FE] Integrasi Fitur CS & Automated Provisioning"
labels: frontend, integration, documentation
assignees: ''
---

# 🚀 Panduan Integrasi Frontend: CS & Provisioning System

Dokumen ini berisi spesifikasi teknis untuk tim Frontend dalam mengimplementasikan fitur pendaftaran Trial, Inbox bantuan, dan Manajemen Aktivasi Tenant.

---

## 📋 1. Fitur: Public Trial Request (Landing Page)
Endpoint ini **tidak memerlukan Token (No Auth)**.

- **Endpoint**: `POST /api/v1/public/trial-request`
- **Payload Request**:
```json
{
  "company_name": "Nama Perusahaan",
  "contact_name": "Nama Lengkap Penanggung Jawab",
  "email": "email@perusahaan.com",
  "phone_number": "08123456789",
  "employee_count_range": "11-50", 
  "industry": "Technology"
}
```
- **Constraint `employee_count_range`**: Wajib salah satu dari: `'1-10'`, `'11-50'`, `'51-200'`, `'201+'`.
- **Response Sukses**: `201 Created`

---

## 📥 2. Fitur: Support Management (Dashboard Admin Tenant 1)
Fitur ini khusus untuk admin di **SaaS HQ (Tenant ID 1)**.

### A. Trial Approval List
- **GET** `/api/v1/admin/support/trials` -> Mengambil semua list pendaftar.
- **PATCH** `/api/v1/admin/support/trials/{id}` -> Approve/Reject.
    - **Payload**: `{"status": "APPROVED"}` atau `{"status": "REJECTED"}`.
    - **Note**: Backend sudah support case-insensitive (bisa kirim "Approved" atau "APPROVED").

### B. Support Inbox (Messages)
- **GET** `/api/v1/admin/support/inbox` -> List pesan masuk dari tenant lain.
- **PATCH** `/api/v1/admin/support/inbox/{id}` -> Update status pesan (`PENDING`, `IN_PROGRESS`, `RESOLVED`).

---

## ⚡ 3. Fitur: Automated Provisioning (Superadmin Only)
Menu ini hanya muncul jika `base_role` user adalah `SUPERADMIN`.

- **GET** `/api/v1/admin/support/provisioning` -> Mengambil list antrean aktivasi (Tiket yang muncul otomatis saat trial di-approve).
- **POST** `/api/v1/admin/support/provisioning/{id}/execute` -> Tombol "Aktifkan Tenant".
    - **Behavior**: Backend akan membuat Tenant & Admin secara otomatis, lalu mengirim email welcome. FE disarankan menampilkan *loading spinner* atau *overlay* karena proses ini melibatkan transaksi DB yang berat.

---

## 🛠️ 4. Pengiriman Pesan Bantuan (Semua Tenant)
- **POST** `/api/v1/support/message`
- **Payload**:
```json
{
  "subject": "Judul Masalah",
  "message": "Detail pesan/kendala",
  "category": "TECHNICAL" 
}
```
- **Category Options**: `TECHNICAL`, `BILLING`, `FEATURE`, `OTHER`.

---

## ⚠️ Troubleshooting 405 Method Not Allowed
Jika Anda mencoba melakukan **PATCH** via proxy (misal: Next.js Port 3000) dan mendapatkan `405 Method Not Allowed`:
1. **Direct Call**: Pastikan URL diarahkan langsung ke port Backend (default `8080`) jika proxy belum dikonfigurasi untuk method PATCH.
2. **Trailing Slash**: Jangan tambahkan `/` di akhir URL (Gunakan `.../trials/{id}`, BUKAN `.../trials/{id}/`).
3. **CORS**: Periksa header `Origin` dan `Access-Control-Allow-Methods` pada server jika berjalan di domain berbeda.

---
**Status Backend**: ✅ Ready for Integration
**Base URL**: `http://localhost:8080/api/v1` (Default)
