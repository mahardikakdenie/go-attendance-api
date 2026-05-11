# FE Task: Implementasi Endpoint Paginasi Timesheet & Fix Task Association

## 1. Deskripsi
Kami telah melakukan optimasi pada endpoint Timesheet untuk mendukung data dalam jumlah besar melalui **Paginasi** dan memperbaiki masalah sinkronisasi data *Task* yang sebelumnya kosong di laporan. Tim Frontend diharapkan untuk menyesuaikan *fetch data* dari endpoint lama ke endpoint baru ini.

## 2. Perubahan Endpoint
Ganti pemanggilan endpoint `GET /v1/timesheet/me/report` dengan endpoint baru berikut:

- **Endpoint**: `GET /v1/timesheet/me/entries`
- **Query Parameters**:
  - `page`: (int) Nomor halaman (default: 1)
  - `limit`: (int) Jumlah data per halaman (default: 10)
  - `start_date`: (string, YYYY-MM-DD) Wajib diisi
  - `end_date`: (string, YYYY-MM-DD) Wajib diisi

## 3. Struktur Data Baru (Response)
Respons sekarang sudah *paginated* dan memiliki field *flat* yang memudahkan *mapping* ke komponen Tabel:

```json
{
  "success": true,
  "data": {
    "entries": [
      {
        "id": "uuid-string",
        "project_name": "WEBSITE",
        "task_name": "Review Document",
        "description": "Menyelesaikan peninjauan dokumen.",
        "duration_hours": 8.0,
        "date": "2026-05-12T00:00:00Z",
        "created_at": "..."
      }
    ],
    "total": 105,
    "page": 1,
    "limit": 10,
    "total_hours": 8.0
  }
}
```

## 4. Checklist Penyesuaian FE
- [ ] **Migrasi Endpoint**: Ubah pemanggilan `GET` ke `/v1/timesheet/me/entries`.
- [ ] **Implementasi Pagination**: Tambahkan kontrol *pagination* (Previous/Next page) di UI menggunakan field `total`, `page`, dan `limit` dari respons API.
- [ ] **Mapping Data**: Update fungsi *rendering* tabel untuk menggunakan `task_name` dan `description` langsung dari object entry (tidak perlu mengakses relasi nested).
- [ ] **Validasi Input**: Pastikan input `duration_hours` pada form `POST` dibatasi maksimal **999.99** di sisi Frontend untuk menghindari error *numeric overflow* dari database.

---
*Catatan: Pastikan saat melakukan `POST /v1/timesheet/entries`, Anda mengirimkan `task_id` (atau `task_name` untuk task baru) agar data terekam dengan benar dan muncul di laporan.*
