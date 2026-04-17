---
name: "⚙️ [FE] Integrasi: Individual Payroll Sync & Save (Stateful)"
about: Panduan integrasi kalkulator payroll individu yang terhubung dengan database (Save & Sync).
title: "[FE] Integrasi Kalkulator Payroll Stateful & Sync Data"
labels: frontend, payroll, finance-ops
assignees: ''
---

# ⚙️ Individual Payroll Sync & Integration Guide

Task ini bertujuan untuk menghubungkan UI Kalkulator dengan data riil karyawan dan menyimpan hasilnya ke database.

---

## 🏗️ 1. Workflow di Frontend (Step-by-Step)

1.  **Pilih Karyawan**: User memilih karyawan dari dropdown.
2.  **Auto-fill Baseline**: FE memanggil `GET /baseline` untuk mengisi form `Basic Salary` dan `PTKP Status`.
3.  **Auto-fill Variables**: FE memanggil `GET /attendance-sync?period=YYYY-MM` untuk mengisi `Attendance Days`, `Overtime Hours`, dan `Unpaid Leave`.
4.  **Live Calculation**: FE tetap melakukan live calculation (stateless) menggunakan endpoint `/calculate` saat user mengubah angka.
5.  **Save to DB**: Saat user menekan tombol **"Save to Dashboard"**, FE mengirim payload lengkap ke endpoint `/save`.

---

## 📡 2. Endpoint Spesifikasi

### A. Fetch Baseline Data
- **URL**: `/api/v1/payroll/employee/{user_id}/baseline`
- **Tujuan**: Auto-fill komponen gaji tetap.

### B. Sync Attendance & Variables
- **URL**: `/api/v1/payroll/employee/{user_id}/attendance-sync?period=2024-03`
- **Tujuan**: Mengambil data kehadiran dan lembur bulan berjalan.

### C. Save/Update Payroll
- **Method**: `POST`
- **URL**: `/api/v1/payroll/employee/{user_id}/save`
- **Payload**:
```json
{
  "period": "2024-03",
  "basic_salary": 15000000,
  "allowances": 2500000,
  "attendance_days": 20,
  "working_days_in_month": 22,
  "overtime_hours": 5.5,
  "unpaid_leave_days": 2,
  "ptkp_status": "K/1",
  "status": "Draft"
}
```

---

## 🔐 3. Aturan Keamanan & UI
1.  **RBAC**: Menu ini wajib disembunyikan untuk role `EMPLOYEE`.
2.  **Validation**: Jika `period` yang dipilih sudah memiliki status `Published` di Dashboard, disarankan memberi warning "Data periode ini sudah dipublish dan akan diupdate".
3.  **Visual Sync**: Setelah berhasil "Save", arahkan user kembali ke Dashboard Payroll Utama atau tampilkan toast sukses.

---
**Status Backend**: ✅ Stateful Engine Ready
**Target API**: `/api/v1/payroll/employee/*`
