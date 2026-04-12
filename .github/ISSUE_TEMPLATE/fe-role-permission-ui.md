---
name: "🎨 [FE] UI/UX: Role Management & Permission Matrix"
about: Desain dan implementasi antarmuka untuk manajemen jabatan (Role) dan pemetaan izin (Permission Mapping).
title: "[FE] Pembuatan UI Dashboard Role & Permission"
labels: frontend, ui/ux, security
assignees: ''
---

# 🎨 Role & Permission Management UI

Tujuan task ini adalah membangun dashboard manajemen akses yang intuitif, bersih, dan fungsional. Admin harus bisa membuat jabatan baru dan mencentang izin apa saja yang diberikan.

---

## 🖼️ 1. Layout Concept: Role Dashboard
Gunakan desain berbasis **Cards** atau **Clean Table** untuk daftar Role yang sudah ada.

- **Role Card**: Menampilkan Nama Role, Deskripsi, dan Badge `BaseRole` (misal: Badge Biru untuk HR, Hijau untuk Employee).
- **Actions**: Tombol Edit, Hapus, dan tombol khusus "Manage Hierarchy".

---

## 🔐 2. Permission Matrix (The Grid)
Saat membuat atau mengedit Role, gunakan **Categorized Checkbox Grid**. Jangan menumpuk semua permission dalam satu list panjang.

### Visual Grouping (Backend Expectation):
Backend mengirimkan list permission dengan format `module.action`. FE wajib mengelompokkannya berdasarkan **Module**:

| Category: Attendance | Category: User | Category: Support |
| :--- | :--- | :--- |
| [ ] View Attendance | [ ] View Users | [ ] Manage Support |
| [ ] Create Attendance | [ ] Create Users | |
| [ ] Export Reports | [ ] Delete Users | |

---

## 🌳 3. Hierarchy Builder (The Tree)
Antarmuka untuk mengatur hubungan atasan-bawahan (`internal/routes/api.go` endpoint: `/hierarchy`).

- **UI Suggestion**: Gunakan **Visual Tree View** atau **Nested Drag & Drop list**.
- **Logic**: Admin memilih satu Role Utama (Parent), lalu menarik Role lain ke bawahnya sebagai Bawahan (Child).

---

## 📡 4. API Data Mapping (The Payload)
Pastikan saat menekan tombol "Save", payload yang dikirim ke Backend sesuai dengan struktur berikut:

### Create/Update Role:
- **Endpoint**: `POST /api/v1/tenant-roles`
- **Body**:
```json
{
  "name": "Manager Operasional",
  "description": "Mengelola seluruh tim lapangan",
  "base_role": "HR",
  "permissions": [
    "attendance.view",
    "attendance.export",
    "user.view"
  ]
}
```

### Save Hierarchy:
- **Endpoint**: `POST /api/v1/tenant-roles/hierarchy`
- **Body**:
```json
{
  "parent_role_id": 2,
  "child_role_ids": [4, 5, 6]
}
```

---

## 🛠️ 5. Functional Requirements
1.  **Search & Filter**: Admin bisa mencari Role berdasarkan nama.
2.  **Form Validation**: Nama Role minimal 3 karakter.
3.  **Loading State**: Tampilkan skeleton atau spinner saat memproses "Save Hierarchy" karena transaksi di database cukup kompleks.
4.  **Feedback**: Notifikasi sukses/gagal menggunakan Toast yang cantik.

---
**Focus Scope**: ✅ Role Management | ✅ Permission Matrix | ✅ Hierarchy Tree
**Target API**: `/api/v1/tenant-roles`
