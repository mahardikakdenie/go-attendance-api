# Laporan Masalah: Ketidaksesuaian Data Timesheet

Kepada Tim Frontend,

Terima kasih atas laporannya. Kami telah melakukan pengecekan mendalam terhadap endpoint timesheet untuk memastikan field `task_name` dan `description` disimpan dan ditarik kembali dengan benar.

## 1. Analisis Teknis
Berdasarkan hasil investigasi kami:

1. **Penyimpanan Database**:
   - `task_name`: Sebenarnya field ini **tidak disimpan secara langsung** ke dalam tabel `timesheet_entries`. Backend menggunakan relasi ke tabel `tasks` melalui `task_id`.
   - `description`: Field ini disimpan di database dalam kolom `notes` pada tabel `timesheet_entries`.

2. **Endpoint GET `/v1/timesheet/me/report`**:
   - Saat ini endpoint tersebut mengembalikan object `TimesheetEntry` mentah dari database, yang memiliki struktur:
     ```json
     {
       "id": "...",
       "notes": "...",
       "task": { "name": "..." },
       ...
     }
     ```
   - Karena UI mengharapkan `task_name` dan `description` sebagai field tingkat atas (*top-level field*), data tersebut tidak muncul karena UI mencari field dengan nama yang berbeda.

## 2. Tindakan yang Akan Diambil
Untuk menyelesaikan masalah ini tanpa memecahkan kontrak API, kami menyarankan dua opsi:

**Opsi A (Rekomendasi Frontend)**:
Menyesuaikan cara pembacaan data di sisi frontend:
- Mengambil nama task dari `entry.task.name` (jika tersedia).
- Mengambil deskripsi dari `entry.notes`.

**Opsi B (Perubahan Backend)**:
Jika frontend tidak dapat menyesuaikan, kami akan membuat *Data Transfer Object* (DTO) baru khusus untuk respons laporan yang akan memetakan:
- `entry.notes` -> `description`
- `entry.task.name` -> `task_name`

Mohon informasinya apakah Opsi A dapat dilakukan, atau apakah tim frontend memerlukan perubahan pada struktur respons API (Opsi B). Kami siap membantu melakukan *mapping* data di sisi backend agar sinkron dengan kebutuhan UI.

Terima kasih.

---
Tim Backend
