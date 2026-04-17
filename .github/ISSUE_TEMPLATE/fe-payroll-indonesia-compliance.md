---
name: "💸 [FE] Integrasi: Payroll Compliance Update (Indonesian Regulation 2024)"
about: Update UI kalkulator dan dashboard untuk mendukung Tunjangan Tidak Tetap (Daily), BPJS Caps, dan PPh 21 TER 2024.
title: "[FE] Update Payroll UI: Indonesian Regulation Compliance"
labels: frontend, finance, payroll
assignees: ''
---

# 💸 Payroll Compliance Update (Indonesian Regulation 2024)

Kami telah memperbarui **Payroll Engine** di Backend agar sepenuhnya mematuhi regulasi terbaru di Indonesia (TER 2024 dan BPJS Caps). Terdapat perubahan pada struktur payload dan logic UI yang perlu disesuaikan oleh tim Frontend.

---

## 🏗️ 1. Perubahan Struktur Payload (Input Form)

Admin/Finance kini harus memisahkan antara **Tunjangan Tetap** (Fixed) dan **Tunjangan Tidak Tetap** (Daily/Variable).

### Payload Baru untuk `/calculate` dan `/save`:
| Field | Deskripsi | Jenis | Impact |
| :--- | :--- | :--- | :--- |
| `fixed_allowances` | Tunjangan tetap bulanan (Jabatan, dll). | Bulanan | **Masuk** Dasar BPJS & Lembur. |
| `daily_meal_allowance` | Uang makan harian. | Harian | **Tidak Masuk** Dasar BPJS & Lembur. |
| `daily_transport_allowance`| Uang transport harian. | Harian | **Tidak Masuk** Dasar BPJS & Lembur. |

**Logic FE**: `variable_allowances` di backend dihitung otomatis dari:
`(daily_meal + daily_transport) * attendance_days`.

---

## ⚙️ 2. Update Business Logic (Informasi untuk User)

### A. Perhitungan Lembur (Overtime)
Dasar pengali lembur sekarang lebih akurat:
`Basis = (Gaji Pokok + Tunjangan Tetap) / 173`.
*Sebelumnya hanya menggunakan Gaji Pokok.*

### B. Batas Maksimal BPJS (Caps)
Backend kini menerapkan batas atas sesuai regulasi:
- **BPJS Kesehatan**: Dasar pengali maksimal **Rp 12.000.000**. (Jika gaji 20jt, iuran tetap 1% dari 12jt).
- **BPJS JP (Pensiun)**: Dasar pengali maksimal **Rp 10.042.300** (Estimasi 2024).

### C. PPh 21 TER 2024
Backend telah mengimplementasikan kategori TER (A, B, C) berdasarkan `ptkp_status`:
- **Kategori A**: TK/0, TK/1, K/0
- **Kategori B**: TK/2, TK/3, K/1, K/2
- **Kategori C**: K/3
*Pajak dihitung langsung dari Bruto (Gross + BPJS Company Share).*

---

## 📡 3. Update Response Structure (JSON)

Response `/calculate` dan `/payroll` kini lebih detail untuk kebutuhan Slip Gaji:
- `earnings`: `fixed_allowances`, `variable_allowances`.
- `employer_contributions`: Rincian biaya yang ditanggung perusahaan (JKK, JKM, BPJS Company).
- `net_salary`: Take Home Pay (bersih).

---

## 🚀 Checklist Integrasi Frontend
- [ ] Update Form Kalkulator: Pisahkan input `Fixed Allowances`, `Daily Meal`, dan `Daily Transport`.
- [ ] Update Layout Preview Slip Gaji: Tampilkan rincian Tunjangan Tetap vs Tidak Tetap.
- [ ] Pastikan saat memanggil `/calculate`, parameter `fixed_allowances` diisi (bukan digabung ke basic salary).
- [ ] Tambahkan label informasi pada UI mengenai "BPJS Maximum Basis" agar HR tahu mengapa iuran tidak naik meskipun gaji sangat tinggi.

---
**Status Backend**: ✅ Compliant & Refactored
**Regulatory Reference**: PMK No. 168 Tahun 2023 (TER)
