  ---

    1 ---
    2 name: "⚙️ [BE] API Integration: Individual Payroll Sync & Save (Stateful Calculator)"
    3 about: Kebutuhan endpoint untuk menghubungkan kalkulator payroll interaktif dengan data spesifik
      karyawan (user_id) agar dapat disimpan ke tabel Payroll utama.
    4 title: "[BE] Integrasi Kalkulator Payroll Individu & Penyimpanan Data (Stateful)"
    5 labels: backend, payroll, finance-ops
    6 assignees: ''
    7 ---
    8
    9 # ⚙️ Individual Payroll Sync & Integration Guide
   10
   11 **Context (From CFO / HR Director Perspective):**
   12 Saat ini, kalkulator payroll kita (`/payroll/calculator`) berfungsi sangat baik sebagai *engine*
      simulasi (stateless). Namun, untuk mempercepat operasional akhir bulan, tim Finance & HR
      membutuhkan fitur di mana mereka bisa **memilih karyawan spesifik (`user_id`)**, lalu sistem
      akan otomatis menarik data gaji pokok, status pajak, dan rekap absensi bulan berjalan.
   13 Setelah simulasi selesai dan angkanya valid, HR harus bisa **menyimpan hasil kalkulasi ini**
      agar langsung masuk ke tabel Dashboard Payroll Utama (`/payroll`).
   14
   15 Oleh karena itu, Frontend membutuhkan 3 endpoint krusial dari tim Backend.
   16
   17 ---
   18
   19 ## 📡 1. Endpoint: Fetch Employee Baseline Data
   20 Digunakan saat HR memilih nama karyawan dari dropdown di Kalkulator. Frontend butuh data
      komponen gaji tetap karyawan tersebut untuk melakukan *auto-fill* form.
   21
   22 - **Method**: `GET`
   23 - **URL**: `/api/v1/payroll/employee/{user_id}/baseline`
   24 - **Response**:
  {
    "success": true,
    "meta": { "message": "Employee baseline fetched successfully" },
    "data": {
      "user_id": 101,
      "employee_name": "Alex Johnson",
      "department": "Engineering",
      "ptkp_status": "K/1",
      "basic_salary": 15000000,
      "fixed_allowances": 2500000
    }
  }

    1
    2 ---
    3
    4 ## 📡 2. Endpoint: Fetch Employee Monthly Variables (Attendance Sync)
    5 Setelah karyawan dipilih, sistem harus menarik data variabel bulanannya (kehadiran, lembur,
      unpaid leave) berdasarkan periode yang sedang berjalan untuk *auto-fill* komponen
      pemotong/penambah gaji.
    6
    7 - **Method**: `GET`
    8 - **URL**: `/api/v1/payroll/employee/{user_id}/attendance-sync`
    9 - **Query Params**:
   10   - `period` (string): Format `YYYY-MM` (contoh: `2024-03`)
   11 - **Response**:
  {
    "success": true,
    "meta": { "message": "Attendance sync fetched" },
    "data": {
      "period": "2024-03",
      "working_days_in_month": 22,
      "attendance_days": 20,
      "unpaid_leave_days": 2,
      "overtime_hours": 5.5
    }
  }

    1 *Catatan BE*: Harap pastikan `unpaid_leave_days` dihitung secara akurat dari modul *Leaves*
      (Cuti) yang statusnya *approved* dan tipe cutinya *unpaid*, dikombinasikan dengan data Alpha
      dari modul *Attendance*.
    2
    3 ---
    4
    5 ## 📡 3. Endpoint: Save/Publish Individual Payroll
    6 Ini adalah endpoint utama. Setelah HR mereview hasil kalkulasi di layar (Digital Payslip
      Preview), mereka akan menekan tombol **"Save & Sync to Payroll Dashboard"**. Data ini akan
      divalidasi ulang dan masuk ke tabel `payrolls` sehingga muncul di layar
      `src/app/(admin)/payroll/page.tsx`.
    7
    8 - **Method**: `POST`
    9 - **URL**: `/api/v1/payroll/employee/{user_id}/save`
   10 - **Payload Request**:
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

    1 *(Catatan: `status` bisa dikirim sebagai "Draft" agar direview ulang oleh Manager, atau langsung
      "Published" jika sudah final).*
    2
    3 - **Behavior yang Diharapkan dari Backend**:
    4   1. Backend **WAJIB menghitung ulang (re-calculate)** payload tersebut di server menggunakan
      *Payroll Engine* (PPh 21 TER, BPJS, dll) yang sama dengan UI untuk mencegah manipulasi data dari
      sisi klien (Security First).
    5   2. Simpan hasil kalkulasi komplit (Gross, Net, Total Deductions, BPJS Breakdown) ke tabel
      `payrolls`.
    6   3. Jika untuk periode tersebut (`2024-03`) user sudah memiliki record gaji, lakukan **UPSERT**
      (Update jika ada, Insert jika belum ada).
    7
    8 ---
    9
   10 ## 🔐 4. Akses & Sekuritas
   11 1. Ketiga endpoint ini HANYA boleh diakses oleh user dengan Role `SUPERADMIN`, `ADMIN`, `HR`,
      atau `FINANCE`.
   12 2. Tolak request (return `403 Forbidden`) jika `EMPLOYEE` mencoba memukul endpoint ini.
   13 3. Berikan response validasi yang jelas (misal: `400 Bad Request`) jika periode penggajian
      (`period`) untuk bulan tersebut sudah di-"Lock" atau ditutup pembukuannya oleh Finance (Closed
      Book).
   14
   15 ---
   16 ## 🚀 5. Checklist Kesuksesan (DoD)
   17 - [ ] Endpoint `GET /baseline` mereturn data PTKP dan Gaji Pokok terkini dari tabel Users.
   18 - [ ] Endpoint `GET /attendance-sync` mengkalkulasi jam lembur dan unpaid leave secara akurat
      dari data mentah absensi.
   19 - [ ] Endpoint `POST /save` berhasil melakukan UPSERT ke tabel `payrolls` dan datanya langsung
      muncul di Dashboard Payroll Utama.
   20 - [ ] Perhitungan pajak (PPh 21 TER) di sisi Backend identik dengan hasil kalkulasi di Frontend.
