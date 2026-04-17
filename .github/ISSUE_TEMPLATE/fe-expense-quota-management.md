# Task: Expense Quota Management UI & Integration

## 1. Overview
Karyawan sekarang memiliki **Kuota Reimbursement (Expense Quota)** bulanan. Kita perlu menyediakan UI bagi Admin/HR/Finance untuk mengatur nominal kuota ini per karyawan, serta menampilkan sisa kuota bagi karyawan saat mereka ingin mengajukan klaim (Expenses).

## 2. User Interface Requirements

### A. Employee Management (Admin Side)
- **Employee List Table:** Tambahkan kolom baru **"Monthly Quota"** yang menampilkan nilai `expense_quota`.
- **Edit Modal / Detail Page:** 
  - Tambahkan field input `Expense Quota (IDR)`.
  - Gunakan numeric input (Currency Formatter).
  - Tambahkan tombol **"Update Quota"** yang memicu API Patch.

### B. Expense Submission Form (Employee Side)
- **Current Balance Indicator:** Saat karyawan membuka form "Submit Expense", tampilkan informasi kecil: 
  > *"Sisa kuota bulan ini: Rp 2.500.000"*
- **Validation:** 
  - Jika `amount` yang diinput melebihi sisa kuota, tampilkan warning secara realtime (Client-side validation).
  - Tampilkan pesan error dari backend jika request ditolak karena kuota habis.

## 3. API Integration

### A. Update Quota (Admin/Finance Only)
- **Endpoint:** `PATCH /api/v1/finance/quotas/:user_id`
- **Method:** `PATCH`
- **Payload:**
```jsonz
{
  "quota": 5000000
}
```

### B. Fetching Quota Data
- Data `expense_quota` sekarang otomatis ter-include dalam objek User:
  - `GET /api/v1/users/me` (Untuk sisa kuota sendiri).
  - `GET /api/v1/users` (Untuk daftar kuota semua karyawan).

## 4. Implementation Logic
1. **Security:** Pastikan input kuota hanya bisa diedit oleh user dengan role `admin`, `superadmin`, atau `finance`.
2. **Persistence:** Setelah sukses update kuota, refresh state user atau panggil kembali API list agar data di tabel sinkron.
3. **Empty Quota:** Jika data `expense_quota` bernilai `0`, artinya karyawan tersebut belum diberikan hak reimbursement.

## 5. Acceptance Criteria
- [ ] Admin dapat mengubah nominal kuota karyawan di halaman Employee Management.
- [ ] Perubahan kuota tersimpan di database dan muncul di log aktivitas.
- [ ] Karyawan dapat melihat sisa kuota mereka sebelum mengajukan klaim.
- [ ] Sistem menolak (error 400) jika pengajuan klaim melebihi kuota.
