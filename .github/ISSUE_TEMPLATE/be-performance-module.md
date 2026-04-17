# Backend Task: Implementation of Performance Management API

## 1. Overview
Kita memerlukan implementasi API untuk modul **Performance Management**. Modul ini mencakup dua fitur utama: **Goal Setting (KPI/OKR)** dan **Performance Appraisals**. Frontend sudah diimplementasikan (Phase 1) dan saat ini melakukan *mocking* atau memanggil endpoint berikut yang perlu disediakan oleh backend.

## 2. Model Requirements

### A. PerformanceGoal
- `id`: Integer (PK)
- `user_id`: Integer (FK to Users)
- `title`: String
- `description`: Text
- `type`: Enum (`KPI`, `OKR`)
- `target_value`: Float
- `current_progress`: Float (Default: 0)
- `unit`: String (e.g., "%", "IDR", "Tasks")
- `start_date`: Date
- `end_date`: Date
- `status`: Enum (`IN_PROGRESS`, `COMPLETED`, `CANCELLED`)

### B. PerformanceCycle
- `id`: Integer (PK)
- `name`: String (e.g., "Annual Review 2026")
- `start_date`: Date
- `end_date`: Date
- `status`: Enum (`DRAFT`, `ACTIVE`, `CLOSED`)

### C. Appraisal
- `id`: Integer (PK)
- `cycle_id`: Integer (FK to PerformanceCycle)
- `user_id`: Integer (FK to Users)
- `self_score`: Float
- `manager_score`: Float
- `final_score`: Float
- `final_rating`: String
- `status`: Enum (`PENDING`, `SELF_REVIEW`, `MANAGER_REVIEW`, `COMPLETED`)
- `comments`: Text

## 3. API Endpoints to Implement

### 3.1. Goals (KPI/OKR)
- **`GET /api/v1/performance/goals/me`**
  - Mengambil daftar goal untuk user yang sedang login.
- **`GET /api/v1/performance/goals/user/{userId}`**
  - Mengambil daftar goal user tertentu (Hanya untuk Manajer atau Admin/HR).
- **`POST /api/v1/performance/goals`**
  - Membuat goal baru untuk user (Hanya untuk Manajer atau Admin/HR).
- **`PUT /api/v1/performance/goals/{id}/progress`**
  - Payload: `{ "current_progress": float }`
  - Update progres sebuah goal.

### 3.2. Cycles & Appraisals
- **`GET /api/v1/performance/cycles`**
  - Mengambil semua siklus evaluasi.
- **`GET /api/v1/performance/appraisals/cycle/{cycleId}`**
  - Mengambil daftar appraisal dalam satu siklus (Filter by tenant).
- **`PUT /api/v1/performance/appraisals/{id}/self-review`**
  - Payload: `{ "self_score": float, "comments": string }`
  - Submit evaluasi diri oleh karyawan.

## 4. Business Logic Notes
1. **Manager Access:** User dengan `manager_id` yang cocok dengan ID anggota tim bisa melihat dan membuat goal untuk timnya.
2. **Tenant Isolation:** Pastikan data `PerformanceGoal` dan `PerformanceCycle` terisolasi berdasarkan `tenant_id`.
3. **Email Notification (Optional):** Kirim email notifikasi ketika goal baru ditugaskan.

## 5. Integration with Existing Systems
- Data **Attendance** (Late count) di masa depan akan digunakan untuk mengisi salah satu metrik KPI secara otomatis. Mohon sediakan hook atau helper function untuk sinkronisasi data ini.
