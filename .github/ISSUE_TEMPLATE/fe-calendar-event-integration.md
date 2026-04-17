# Task: Integration of Refactored Calendar Events & Scheduling System

## 1. Overview
Kita telah melakukan refactor pada modul **Holiday** menjadi **Calendar Events**. Perubahan ini memungkinkan HR tidak hanya mencatat hari libur nasional, tetapi juga agenda internal seperti Meeting, Company Event, atau Informasi penting lainnya tanpa memblokir absensi karyawan (kecuali jika dikategorikan sebagai hari libur). 

Selain itu, sistem sekarang mendukung **Targeted Notifications**, di mana HR bisa memilih apakah event ditujukan untuk **Seluruh Karyawan** atau **Karyawan Tertentu Saja**. Sistem akan secara otomatis mengirimkan email pengingat pada H-1 jam 08:00 WIB kepada user yang relevan.

## 2. Backend Changes (Context)
- **Model Name:** `Holiday` -> `CalendarEvent`
- **New Fields:**
  - `category`: `OFFICE_CLOSED` (Absensi dikunci) | `INFORMATION` (Hanya agenda/info, absensi tetap buka).
  - `description`: Text area untuk detail agenda.
  - `type`: Penambahan type `Meeting`.
  - `is_all_users`: Boolean (Default: `true`). Jika `false`, sistem hanya akan mengirim notifikasi ke user yang dipilih.
  - `user_ids`: Array of integers. ID user yang dipilih jika `is_all_users` bernilai `false`.
- **Logic:** 
  - Endpoint `/api/v1/hr/calendar` mengembalikan data event lengkap dengan daftar user yang di-assign.
  - Cron Job akan mengirim email otomatis H-1 event.

## 3. UI/UX Requirements

### A. Calendar View Enhancement
- **Visual Distinction:** Bedakan warna badge/event di kalender berdasarkan `category`.
  - `OFFICE_CLOSED`: Gunakan warna **Red/Rose** (Indikasi kantor tutup).
  - `INFORMATION` (Meeting/Event): Gunakan warna **Blue/Indigo** atau **Amber**.
- **Target Indicator:** Tambahkan label "All Users" atau "Targeted" pada detail event.

### B. Event Creation & Edit Modal
Update form pada Modal Kalender untuk mencakup field baru:
1. **Event Type (Dropdown):** Tambahkan opsi `Meeting`.
2. **Category (Radio/Toggle):**
   - **Office Closed:** Jika dipilih, sistem akan menampilkan warning: *"Karyawan tidak dapat melakukan absensi pada hari ini."*
   - **Information Only:** Agenda biasa, absensi tetap diizinkan.
3. **Audience Selection:**
   - Tambahkan Checkbox: **"Send to all employees?"** (Default: Checked).
   - Jika Checkbox **Unchecked**, tampilkan **Multi-select User Search** (Bisa mengambil data dari `/api/v1/users`).
4. **Description (Textarea):** Tambahkan input untuk detail agenda/catatan meeting.

### C. Attendance Screen (Mobile/User Side)
- Jika hari ini ada event dengan category `INFORMATION` (misal: "HR Monthly Meeting") dan user tersebut masuk dalam daftar audience, tampilkan banner informasi kecil di Dashboard.

## 4. API Integration Notes
- **Endpoint:** `GET /api/v1/hr/calendar?year=2026`
- **Payload Post/Put (Targeted Event):**
```json
{
  "date": "2026-04-20",
  "name": "Strategy Meeting Q2",
  "type": "Meeting",
  "category": "INFORMATION",
  "is_paid": true,
  "description": "Meeting koordinasi tahunan di ruang aula utama.",
  "is_all_users": false,
  "user_ids": [2, 4, 8]
}
```

## 5. Acceptance Criteria
- [ ] User dapat membuat event dengan type 'Meeting' dan category 'Information'.
- [ ] HR dapat memilih user tertentu untuk sebuah event.
- [ ] Data `is_all_users` dan `user_ids` terkirim dengan benar ke API.
- [ ] Kalender menampilkan warna yang berbeda antara Libur Nasional dan Meeting Internal.
- [ ] Tooltip atau Modal Detail Kalender menampilkan isi dari field `description`.

## 6. Resources
- File Model: `internal/model/hr_ops.go`
- File Service: `internal/service/calendar_cron_service.go`
