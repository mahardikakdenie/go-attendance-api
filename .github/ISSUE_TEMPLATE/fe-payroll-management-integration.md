---
name: "📊 [FE] Integrasi: Payroll Management & Operations Dashboard"
about: Panduan integrasi UI/UX untuk pengelolaan siklus penggajian, summary statistik, dan penerbitan slip gaji.
title: "[FE] Integrasi Dashboard Payroll Operations"
labels: frontend, payroll, finance
assignees: ''
---

# 📊 Payroll Operations Integration Guide

Task ini bertujuan untuk membangun dashboard pengelolaan payroll bulanan yang digunakan oleh Admin/Finance untuk memantau pengeluaran gaji dan menerbitkan slip gaji karyawan.

---

## 🖼️ 1. Desain Dashboard (Modern Stats & Table)

Halaman ini terdiri dari dua bagian utama:

### A. Management Stats (Summary Cards)
Tampilkan 4 kartu di bagian atas dashboard untuk memberikan gambaran cepat:
- **Total Net Payout**: Total gaji bersih yang harus dibayarkan bulan ini.
- **Tax Liability**: Total hutang PPh 21 perusahaan.
- **BPJS Provision**: Total iuran BPJS yang harus disetor.
- **Sync Rate**: Persentase kehadiran yang sudah terekam.

### B. Payroll Table (Data List)
Tabel yang menampilkan rincian gaji per karyawan untuk periode yang dipilih.
- **Periode Filter**: DatePicker (Format Month-Year, e.g. March 2024).
- **Search Bar**: Cari berdasarkan nama atau NIK.
- **Status Badge**: `Draft` (Kuning), `Published` (Hijau).

---

## 📡 2. Integrasi API

### A. Get Summary Stats
Ambil data statistik untuk kartu di bagian atas.
- **Method**: `GET`
- **URL**: `/api/v1/payroll/summary?period=2024-03`

### B. Get Payroll List (Pagination)
Ambil daftar payroll karyawan.
- **Method**: `GET`
- **URL**: `/api/v1/payroll?period=2024-03&page=1&limit=10&search=Alex`
- **Response Structure**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "user": { "name": "Alex", "employee_id": "SA-001" },
      "period": "2024-03",
      "basic_salary": 12000000,
      "total_deductions": 772500,
      "net_salary": 11227500,
      "status": "Draft"
    }
  ]
}
```

### C. Generate New Payroll Cycle
Tombol "Start New Cycle" atau "Refresh Data" untuk memicu kalkulasi massal.
- **Method**: `POST`
- **URL**: `/api/v1/payroll/generate`
- **Body**: `{ "period": "2024-03" }`

### D. Publish Payroll (Review Selesai)
Mengubah status Draft menjadi Published agar karyawan bisa melihat slip gajinya.
- **Method**: `PATCH`
- **URL**: `/api/v1/payroll/{id}/publish`

---

## 🛠️ 3. Alur Kerja Finance (UX Workflow)
1. **Input User ID di Kalkulator**: Pada halaman Kalkulator Interaktif, kirimkan `user_id` ke endpoint `/calculate` agar BE otomatis menarik gaji pokok default.
2. **Hitung & Review**: Setelah Finance merasa perhitungan di kalkulator sudah oke, mereka pergi ke Dashboard Payroll.
3. **Generate**: Klik tombol "Generate" untuk menyimpan snapshot gaji seluruh karyawan ke database.
4. **Publish**: Klik tombol "Publish" pada baris karyawan yang sudah diverifikasi datanya.

---
**Focus Scope**: ✅ Payroll Table | ✅ Monthly Stats | ✅ State Sync
**Target API**: `/api/v1/payroll`
