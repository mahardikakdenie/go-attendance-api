# Backend Task: Employee Behavioral DNA & Analytics API

## 1. Latar Belakang
Halaman **HR Dashboard** saat ini memiliki fitur "Deep Dive" atau **Employee DNA Profile** yang muncul saat sebuah baris karyawan di klik. Data yang ditampilkan saat ini sebagian besar masih bersifat *mock* atau kalkulasi sederhana di sisi Frontend (FE). Diperlukan sebuah endpoint khusus yang menyediakan data analitik mendalam terkait perilaku dan performa individu karyawan untuk mendukung pengambilan keputusan HR yang berbasis data.

## 2. Deskripsi Tugas
Mengembangkan API endpoint baru untuk mengambil data statistik perilaku karyawan ("Behavioral DNA") yang mencakup metrik ketepatan waktu, efisiensi lembur, pola cuti, dan skor kepatuhan (*compliance*).

## 3. API Contract

### Endpoint
`GET /v1/dashboards/hr/employee-dna/{user_id}`

### Header
- `Authorization: Bearer <token>`
- `X-Tenant-ID: <tenant_id>`

### Request Parameters (Path)
- `user_id` (integer, required): ID Karyawan yang ingin dianalisis.

### Success Response (200 OK)
```json
{
  "meta": {
    "status": "success",
    "message": "Employee DNA retrieved successfully"
  },
  "data": {
    "user": {
      "id": 101,
      "name": "Budi Santoso",
      "avatar": "https://api.dicebear.com/7.x/avataaars/svg?seed=Budi",
      "department": "Engineering",
      "position": "Senior Backend Developer",
      "joined_at": "2023-01-15T00:00:00Z"
    },
    "performance_score": 92.5,
    "radar_metrics": {
      "punctuality": 98,
      "overtime_efficiency": 75,
      "leave_regularity": 60,
      "productivity_index": 85,
      "compliance_rate": 100
    },
    "punctuality_dna": {
      "arrival_consistency": 98.2,
      "late_incident_rate": 2,
      "avg_clock_in": "08:05",
      "avg_clock_out": "17:15"
    },
    "workspace_balance": {
      "remaining_leave": 12,
      "total_leave_taken": 3,
      "overtime_hours_30d": 15.5
    },
    "insights": [
      "Karyawan sangat konsisten dalam waktu kedatangan.",
      "Terdapat kecenderungan lembur di hari Jumat.",
      "Kepatuhan pengisian timesheet mencapai 100%."
    ]
  }
}
```

## 4. Logika Bisnis & Kalkulasi (Metrik)

### A. Radar Metrics (Skala 0-100)
1. **Punctuality**: Persentase kehadiran tepat waktu dalam 30 hari terakhir.
2. **Overtime Efficiency**: Rasio antara jam lembur yang disetujui vs total jam kerja (targetkan agar tidak terjadi *burnout*).
3. **Leave Regularity**: Frekuensi pengambilan cuti yang terencana (bukan mendadak).
4. **Productivity Index**: Diambil dari skor penyelesaian tugas/proyek (jika terintegrasi dengan modul Project) atau rasio jam kerja efektif.
5. **Compliance Rate**: Persentase kepatuhan terhadap aturan sistem (misal: selalu selfie saat absen, tidak pernah lupa clock-out).

### B. Punctuality DNA
1. **Arrival Consistency**: Deviasi standar dari waktu kedatangan. Semakin kecil deviasi, semakin tinggi persentase konsistensi.
2. **Late Incident Rate**: Total frekuensi keterlambatan dalam periode aktif (bulan berjalan).
3. **Avg Clock-In**: Rata-rata waktu `clock_in` yang dilakukan karyawan.

### C. Workspace Balance
1. **Remaining Leave**: Sisa kuota cuti tahunan karyawan.
2. **Overtime Hours 30d**: Akumulasi jam lembur dalam 30 hari terakhir.

## 5. Persyaratan Keamanan
- Pastikan user yang memanggil adalah Admin atau memiliki permission `analytics.view` atau `user.view.detail`.
- User hanya bisa melihat data karyawan yang berada dalam **Tenant** yang sama.

## 6. Definisi Selesai (Definition of Done)
- [ ] Endpoint sukses dibuat dan bisa diakses via Postman.
- [ ] Dokumentasi Swagger/OpenAPI diperbarui.
- [ ] Unit Test untuk kalkulasi metrik (Radar & Punctuality) telah dibuat.
- [ ] Response time API di bawah 500ms untuk kalkulasi agregat.

---
**Status FE:** *Waiting for BE Implementation*
FE akan mengintegrasikan endpoint ini ke dalam modal `setSelectedEmployee` di `HrDashboard.tsx` segera setelah API siap.
