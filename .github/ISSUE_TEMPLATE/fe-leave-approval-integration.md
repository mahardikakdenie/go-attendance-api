---
name: "✅ [FE] Integrasi: Leave Approval & Rejection System"
about: Panduan detail integrasi UI/UX untuk modul persetujuan cuti karyawan bagi HR/Admin.
title: "[FE] Integrasi Fitur Approval Cuti"
labels: frontend, hr-ops, leaves
assignees: ''
---

# ✅ Leave Approval System UI Guide

Task ini bertujuan untuk mengimplementasikan fitur persetujuan (`Approve`) dan penolakan (`Reject`) pengajuan cuti karyawan. Fitur ini ditujukan bagi user dengan akses Manager/HR/Admin.

---

## 🖼️ 1. Layout Concept: Review Modal/Page
Saat Admin memilih pengajuan cuti dari daftar "Pending", tampilkan modal review yang berisi detail pengajuan dan form input untuk catatan (Notes).

- **Actions**: 
    - Tombol **Approve** (Warna Hijau).
    - Tombol **Reject** (Warna Merah).
- **Input Field**: `Notes` (Textarea) - Opsional, digunakan untuk memberikan alasan persetujuan atau penolakan.

---

## 📡 2. Integrasi API

### A. Approve Leave
- **Method**: `POST`
- **URL**: `/api/v1/leaves/approve/{id}`
- **Payload**:
```json
{
  "notes": "Dokumen pendukung lengkap, silakan ambil cuti."
}
```

### B. Reject Leave
- **Method**: `POST`
- **URL**: `/api/v1/leaves/reject/{id}`
- **Payload**:
```json
{
  "notes": "Mohon maaf, kuota personil di departemen Anda sudah penuh pada tanggal tersebut."
}
```

---

## 🔐 3. Aturan Keamanan & Akses
1.  **Role Requirement**: Endpoint ini hanya dapat dipanggil oleh user dengan Base Role `SUPERADMIN`, `ADMIN`, atau `HR`.
2.  **UI State**: Sembunyikan tombol Approve/Reject jika status pengajuan sudah bukan `PENDING` (misal: sudah `APPROVED` atau `REJECTED`).
3.  **Automatic Balance Check**: Backend secara otomatis akan mengembalikan saldo cuti karyawan jika pengajuan ditolak, atau memotong saldo secara permanen jika disetujui.

---

## 🚀 Checklist Integrasi
- [ ] Implementasi Modal Review dengan field input `Notes`.
- [ ] Integrasi tombol Approve ke endpoint `/leaves/approve/{id}`.
- [ ] Integrasi tombol Reject ke endpoint `/leaves/reject/{id}`.
- [ ] Refresh data table secara otomatis setelah status berubah.
- [ ] Tampilkan Toast Notification untuk feedback sukses/gagal.

---
**Focus Scope**: ✅ Review Modal | ✅ Approval Logic | ✅ Rejection Logic
**Target API**: `/api/v1/leaves/approve` & `/api/v1/leaves/reject`
