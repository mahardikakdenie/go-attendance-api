Jujur saya masih bingung dengan Validasi

{"message":"Feature not available in your current plan. Please upgrade."}

bagaimana cara menambahkan nya 

tolong Beritahu FE untuk membuatkan Halaman untuk Mengatur Ini dan Gunakan Endpoint Apa saja 

tolong update file di bawah ini sebagai balasan di tandai dengan 

<!-- Balasan -->

### 🛡️ Bagaimana Cara Kerja Validasi Fitur (Plan Enforcements)?

Validasi fitur ini terjadi di level **Middleware** (`internal/middleware/jwt_middleware.go`). Berikut penjelasannya:

1.  **Pengecekan Otomatis:** Setiap kali tim FE memanggil endpoint yang dijaga oleh middleware `HasPermission("nama_modul.action")`, sistem akan secara otomatis mengecek apakah Tenant tersebut memiliki akses ke modul tersebut di dalam paket (Plan) mereka.
2.  **Logic di Backend:**
    *   Sistem mengambil daftar `plan_features` dari field `Features` di tabel `subscription_plans`.
    *   Jika action adalah `attendance.view`, sistem mengecek apakah string `attendance` (modul utama) ada di dalam daftar fitur paket tersebut.
    *   Jika tidak ada, sistem langsung mengembalikan status **403 Forbidden** dengan pesan error yang Anda lihat.

### 🚀 Cara Menambahkan Fitur Baru ke Suatu Plan

Jika Anda ingin fitur baru (misal: "payroll") dibatasi hanya untuk paket tertentu:
1.  Buka file `internal/seeder/plan_seeder.go`.
2.  Tambahkan string nama modul (misal: `"payroll"`) ke dalam array `Features` pada paket yang diinginkan (contoh: Business atau Enterprise).
3.  Jalankan aplikasi dengan `RUN_SEEDER=true` (ini akan memperbarui data plan di database).

---

### 💻 Instruksi untuk Team Frontend (FE)

FE perlu membuat halaman **"Subscription & Plan Management"** agar Admin Tenant bisa melihat paket mereka saat ini dan melakukan upgrade.

#### 1. Endpoint yang Digunakan:
- **`GET /api/v1/subscriptions/me`**:
  - Digunakan untuk menampilkan paket saat ini, status (Active/Trial), tanggal tagihan berikutnya, serta daftar **Fitur** yang didapatkan.
- **`POST /api/v1/subscriptions/upgrade`**:
  - Digunakan untuk melakukan request upgrade.
  - Body: `{"plan_id": 3}` atau `{"plan": "Business"}`.

#### 2. Panduan UI/UX untuk Fitur Terkunci (Locked Features):
Agar pengalaman user (Admin) lebih baik, tim FE disarankan melakukan hal berikut:

1.  **Grey-out Sidebar/Menu:** 
    - Ambil data `plan_features` dari response login atau endpoint `/subscriptions/me`.
    - Jika modul (misal: Finance) tidak ada di daftar fitur, buat menu tersebut menjadi **Grey-out** (buram) dan tampilkan ikon **Gembok (Lock)**.
2.  **Upgrade Modal:**
    - Saat user mengklik menu yang terkunci, jangan biarkan mereka masuk. Alih-alih menampilkan error 500/403 yang mentah, tampilkan **Modal Popup** yang cantik berisi penawaran upgrade (misal: *"Fitur ini hanya tersedia di Paket Enterprise. Upgrade sekarang untuk mengaktifkan!"*).
3.  **Pricing Grid:**
    - Buat halaman yang menampilkan tabel perbandingan harga (Starter, Business, Enterprise) sehingga user tahu apa yang mereka lewatkan dan bisa langsung klik tombol "Upgrade".

#### 3. Error Handling Global:
Pastikan aplikasi FE memiliki interceptor global (misal di Axios) yang menangkap error 403 dengan pesan `"Feature not available..."`. Jika error ini muncul, arahkan user ke halaman Billing/Upgrade secara otomatis.

---

### 👑 Fitur untuk Superadmin (Platform Owner)

Superadmin memiliki akses penuh untuk mengatur master data paket (Plans) dan memanipulasi langganan tenant secara manual.

#### 1. Halaman "Master Plan Management":
Halaman ini digunakan untuk membuat atau mengubah paket (misal: merubah harga atau menambah fitur ke paket 'Business').
- **`GET /api/v1/superadmin/plans`**: List semua paket yang tersedia.
- **`POST /api/v1/superadmin/plans`**: Membuat paket baru.
- **`PUT /api/v1/superadmin/plans/:id`**: Update detail paket (nama, limit karyawan, daftar fitur).
- **`DELETE /api/v1/superadmin/plans/:id`**: Menghapus paket.

#### 2. Halaman "Tenant Subscription Management":
Halaman ini digunakan untuk melihat status langganan seluruh perusahaan (Tenant) dan melakukan intervensi manual.
- **`GET /api/v1/superadmin/subscriptions`**: List seluruh tenant dan paket yang mereka gunakan beserta statistik (MRR, dsb).
- **`PUT /api/v1/superadmin/subscriptions/:id`**: Mengubah paket tenant secara manual (misal: admin memberikan diskon atau perpanjangan masa trial).
- **`POST /api/v1/superadmin/subscriptions/:id/suspend`**: Menonaktifkan akses tenant jika melanggar ketentuan atau belum bayar.

<!-- end of Balasan -->
