# 📘 Developer Instructions: Go Attendance API

Selamat datang di project **Go Attendance API**. Dokumentasi ini bertujuan untuk memberikan panduan standar pengerjaan bagi developer baru agar menjaga konsistensi codebase dan keamanan data.

---

## 🏗️ Project Architecture (Clean Layout)

Project ini mengikuti struktur Go yang standar namun dengan beberapa penyesuaian untuk multi-tenancy:

- `cmd/api`: Entry point aplikasi.
- `internal/model`: Definisi struct database (GORM).
- `internal/dto`: Data Transfer Object untuk request dan response API.
- `internal/repository`: Layer akses database (Query GORM).
- `internal/service`: Layer logika bisnis utama.
- `internal/handler`: Layer HTTP (Gin Framework) untuk handling request/response.
- `internal/middleware`: Middleware Gin (JWT, Rate Limit, dll).
- `internal/config`: Konfigurasi database, redis, dan plugin.

---

## 👥 Roles & Permission System

Sistem role di project ini menggunakan kombinasi **Custom Role** (per Tenant) dan **BaseRole** (Global Logic).

### 🔑 Apa itu BaseRole?

`BaseRole` adalah "identity core" yang menentukan logika fundamental di dalam kode. Jika Custom Role menentukan *nama* jabatan (misal: "Manajer IT"), maka `BaseRole` menentukan *hak akses dasar* di level sistem.

Definisi BaseRole (`internal/model/role.go`):
1. **`SUPERADMIN`**: Akses total ke seluruh sistem, melewati filter tenant. Biasanya hanya untuk tim internal pengembang.
2. **`ADMIN`**: Pemilik/Owner dari sebuah Tenant. Memiliki akses penuh terhadap data di dalam tenant-nya sendiri.
3. **`HR`**: Mengelola absensi, cuti, dan data karyawan di level operasional.
4. **`FINANCE`**: Mengelola payroll dan biaya terkait lembur.
5. **`EMPLOYEE`**: Role dasar dengan akses terbatas hanya untuk data diri sendiri (absensi mandiri, pengajuan cuti).

**Penting:** Selalu gunakan `BaseRole` untuk pengecekan logika di `Service` layer (misal: `if user.Role.BaseRole == model.BaseRoleAdmin`).

---

## 🛡️ Multi-Tenancy Rules

Project ini menggunakan **Data Isolation** menggunakan `TenantPlugin` (`internal/config/tenant_plugin.go`).

1. **Model Requirement**: Setiap tabel yang bersifat data tenant (User, Attendance, Leave, dll) **WAJIB** memiliki field `TenantID uint`.
2. **Automatic Filtering**: GORM secara otomatis akan menambahkan `WHERE tenant_id = ?` pada setiap query jika model memiliki field tersebut.
3. **Context Mandatory**: Pastikan context yang dikirim ke repository membawa nilai `tenant_id`. Ini biasanya ditangani otomatis oleh `JWT Middleware`.

---

## 📝 Coding Standards & Conventions

### 1. Repository Pattern
Gunakan `ApplyPreloads` utility untuk menangani relasi yang dinamis:
```go
// Contoh di repository
query = utils.ApplyPreloads(query, includes, preloadMap)
```

### 2. Service Layer & WIB Timezone
Aplikasi ini di-lock menggunakan timezone **Asia/Jakarta (WIB)**.
- Gunakan variabel `WIB` yang sudah di-define di `internal/service/attendance_service.go`.
- Selalu gunakan `.In(WIB)` saat mengambil `time.Now()`.

### 3. Response Formatting
Gunakan `internal/utils/response.go` untuk standarisasi JSON response:
```go
// Success
c.JSON(http.StatusOK, utils.BuildResponse("Message", code, "success", data))

// Error
c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Error Message", code, "error", detail))
```

### 4. Search Implementation
Untuk fitur pencarian di list (Attendance/User), gunakan join ke tabel User dan gunakan `ILIKE` untuk pencarian case-insensitive pada nama atau ID karyawan.

---

## 🚀 Workflow Pengerjaan
1. **Mulai dari Model**: Definisikan struct di `internal/model`.
2. **Migration**: Tambahkan di `internal/config/database.go` pada Stage yang sesuai.
3. **Seeder**: Buat data dummy di `internal/seeder` untuk testing.
4. **Logic**: Implementasikan di Repository -> Service -> Handler.
5. **Documentation**: Selalu update komentar Swagger (`@Summary`, `@Tags`, dll) di Handler sebelum running `swag init`.

---
Jika ada pertanyaan mengenai arsitektur, silakan hubungi lead developer. Selamat ngoding! 🚀
