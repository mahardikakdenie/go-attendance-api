# Delivery: Finance Expenses & Quota Management API

## 1. Overview
Backend untuk modul Finance Expenses (Reimbursement) dan Management Quota telah selesai diimplementasikan. Fitur ini mencakup pengajuan klaim, pengecekan kuota bulanan, summary statistik, serta sistem approval/rejection.

## 2. API Endpoints (Prefix: `/api/v1/finance`)

### A. Daftar Pengajuan Klaim
`GET /expenses`
- **Query Params:** 
  - `status`: `Pending`, `Approved`, atau `Rejected`.
  - `search`: Filter berdasarkan Nama Karyawan atau ID.
  - `page` & `limit`: Untuk pagination.

### B. Summary Statistik
`GET /expenses/summary`
- **Output:** `pendingAmount`, `approvedThisMonthAmount`, `topCategory` (name & percentage).

### C. Submit Klaim Baru
`POST /expenses`
- **Payload:**
```json
{
  "category": "Travel",
  "amount": 500000,
  "date": "2024-03-20",
  "description": "Bensin luar kota",
  "receipt": "https://url-ke-file.com"
}
```
- **Quota Logic:** API akan mengembalikan error `400` jika (Total Approved + Pending + New Amount) melebihi kuota bulanan user.

### D. Approval & Rejection (Admin/Finance Only)
- **Approve:** `PATCH /expenses/:id/approve`
- **Reject:** `PATCH /expenses/:id/reject` -> Payload: `{"reason": "string"}`

### E. Quota Management (Admin/Finance Only)
`PATCH /quotas/:user_id`
- **Payload:** `{"quota": 5000000}`
- **Fungsi:** Mengatur limit reimbursement maksimal per bulan untuk user tersebut.

## 3. Data Constants

### Expense Categories
`Travel`, `Medical`, `Supplies`, `Equipment`, `Other`.

### Expense Status
`Pending`, `Approved`, `Rejected`.

## 4. Integration Notes
1. **User Quota:** Data kuota user tersedia di field `expense_quota` pada objek User (Endpoint `/users/me` atau `/users`).
2. **Error Handling:** Jika kuota habis, API mengirimkan pesan detail: *"quota tidak cukup. Sisa kuota bulan ini: X"*.
3. **Role Check:** Pastikan UI untuk Approval dan Set Quota hanya muncul untuk role `admin`, `superadmin`, atau `finance`.
