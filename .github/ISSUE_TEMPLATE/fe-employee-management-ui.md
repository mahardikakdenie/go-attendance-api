---
name: "👥 [FE] UI/UX: Employee Management & Role Assignment"
about: Implementasi antarmuka untuk pendaftaran karyawan baru, pengelolaan profil, dan penentuan jabatan.
title: "[FE] Pembuatan Dashboard Manajemen Karyawan"
labels: frontend, user-module, high-priority
assignees: ''
---

# 👥 Employee Management UI Guide

Tujuan task ini adalah membangun modul pengelolaan data karyawan (CRUD) yang terintegrasi dengan sistem Role & Position Backend.

---

## 🖼️ 1. Layout: Employee Directory
Gunakan layout **Data Table** yang modern dengan fitur pencarian dan filter.

- **Columns**: Nama, Email, Employee ID, Department, Role (Badge), Position, Status.
- **Filters**: Cari berdasarkan Nama/ID, Filter berdasarkan Department, dan Filter berdasarkan Role.
- **Empty State**: Tampilkan ilustrasi cantik jika belum ada data karyawan di tenant tersebut.

---

## 📝 2. Form: Create/Edit Employee
Modal atau halaman khusus untuk input data karyawan. Pastikan validasi berjalan di sisi client sebelum hit API.

### Field List (Backend Requirement):
| Field | Type | Required | Note |
| :--- | :--- | :--- | :--- |
| `name` | String | Yes | Min 3 chars |
| `email` | String | Yes | Valid email format |
| `password` | String | Yes | Min 6 chars |
| `role_id` | Number | Yes | Ambil dari endpoint `/api/v1/tenant-roles` |
| `position_id` | Number | No | Ambil dari endpoint `/api/v1/organization/positions` |
| `manager_id` | Number | No | List user lain dengan role Manager/Admin |
| `department` | String | No | Input teks (misal: IT, HR, Sales) |
| `employee_id` | String | No | ID unik internal (misal: EMP-001) |

---

## 📡 3. Integrasi API (The Payload)

### Create Employee:
- **Method**: `POST`
- **URL**: `/api/v1/users`
- **Body Example**:
```json
{
  "name": "Budi Santoso",
  "email": "budi@company.com",
  "password": "securepassword123",
  "role_id": 3,
  "position_id": 1,
  "manager_id": 2,
  "department": "Engineering",
  "address": "Jl. Sudirman No. 1",
  "phone_number": "08123456789"
}
```

---

## 🔐 4. Permission Logic pada UI
Pastikan tombol "Add Employee" atau menu "Edit/Delete" hanya muncul jika user yang login memiliki permission yang tepat (cek array `permissions` di profil `/me`).

- **Create**: `user.create`
- **Edit**: `user.edit`
- **Delete**: `user.delete`
- **View List**: `user.view`

---

## 🛠️ 5. Functional Requirements
1.  **Password Visibility**: Tambahkan icon mata untuk toggle password.
2.  **Role Selection**: Dropdown Role harus memuat data dinamis dari API, bukan hardcoded.
3.  **Avatar Upload**: Integrasikan dengan endpoint `/api/v1/media/upload` untuk pas foto karyawan.
4.  **Error Handling**: Tangani error `409 Conflict` jika email sudah terdaftar di database.

---
**Focus Scope**: ✅ Employee CRUD | ✅ Role Assignment | ✅ Position Mapping
**Target API**: `/api/v1/users`
