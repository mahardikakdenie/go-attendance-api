---
name: Backend Production Readiness Audit
about: Checklist tugas krusial Backend (Golang) sebelum rilis ke production SaaS
title: "[BE-AUDIT] Production Readiness & Architecture Gaps"
labels: backend, security, architecture, critical
assignees: ''
---

# 📋 ISSUE: Backend Production Readiness Audit & Feature Gaps
**Project:** Attendance API (Golang / SaaS Multi-Tenant)
**Status:** Pre-Alpha / MVP Development
**Objective:** Mengidentifikasi celah keamanan, skalabilitas, dan logika bisnis di sisi *Backend* (API) yang harus diselesaikan sebelum API ini dapat melayani ribuan *request* dari berbagai *Tenant* secara bersamaan.

---

## 🏢 1. Isolasi Multi-Tenant (Blocker Keamanan Data)
Sistem saat ini sudah memiliki `tenant_repo.go` dan entitas `tenant.go`. Namun, celah kebocoran data antar perusahaan sangat mungkin terjadi jika *query* tidak diproteksi secara absolut.

- [ ] **Global Tenant Scope (GORM/SQL):** Terapkan *Global Scope* pada ORM (jika menggunakan GORM) atau *middleware* di level *Repository* yang secara otomatis menyisipkan `WHERE tenant_id = ?` pada **setiap** *query* `SELECT`, `UPDATE`, dan `DELETE`. *Developer* tidak boleh mengandalkan pengecekan manual di setiap fungsi *Service*.
- [ ] **Validasi Relasi Data:** Pastikan API menolak *request* jika `Admin` dari Tenant A mencoba memodifikasi `user_id` yang terdaftar di Tenant B (ID Spoofing/Insecure Direct Object Reference - IDOR).

## 💰 2. Logika Bisnis Krusial yang Hilang (Payroll Engine)
Berdasarkan tinjauan struktur folder, saat ini hanya ada `overtime_service.go` dan `attendance_service.go`.
- [ ] **Pindahkan Kalkulasi Payroll ke Backend:** Kalkulasi gaji (Gaji Pokok, Tunjangan, PPh 21 TER, BPJS, Potongan Cuti) **TIDAK BOLEH** dilakukan di *Frontend* (Klien). Buat modul baru `internal/service/payroll_service.go`. Jika *Frontend* yang menghitung, *hacker* bisa memanipulasi *payload* API untuk mengganti total gaji mereka.
- [ ] **Presisi Tipe Data Keuangan:** Pastikan *database* dan *struct* Go menggunakan tipe data `DECIMAL` atau `Numeric` (misalnya *package* `shopspring/decimal`), BUKAN `float64`, untuk menghindari *floating-point error* pada perhitungan uang/pajak.

## 🔐 3. Autentikasi, Redis & Rate Limiting
Repositori sudah memiliki konfigurasi `internal/config/redis.go` dan `jwt_middleware.go`, namun penggunaannya harus dimaksimalkan untuk standar SaaS.

- [ ] **API Rate Limiting (DDoS Protection):** Gunakan Redis untuk membatasi jumlah *request* per IP atau per *User* (misal: maksimal 100 request/menit). Ini krusial untuk melindungi API dari serangan *Brute Force* atau *DDoS*.
- [ ] **Token Revocation (Blacklist):** Saat *User* melakukan *Logout*, JWT saat ini mungkin masih valid sampai masa kedaluwarsanya habis. Simpan token yang di-*logout* ke dalam Redis (*Blacklist*) agar tidak bisa digunakan lagi jika dicuri.
- [ ] **Refresh Token Mechanism:** Implementasikan *Refresh Token* agar *Access Token* JWT bisa dibuat berumur pendek (misal: 15 menit) demi keamanan, namun *User* tidak perlu *login* ulang setiap 15 menit.

## 🧪 4. Observabilitas & Pengujian (Quality Assurance)
Tidak terlihat adanya file pengujian (`_test.go`) pada struktur `internal/service/` atau `internal/handler/`.
- [ ] **Unit Testing (Wajib):** Tulis *Unit Test* di Golang menggunakan `testing` dan `testify/assert` atau `mockery`. Fokuskan pengujian pada modul krusial seperti `attendance_service.go` dan `payroll_service.go`. Targetkan *Code Coverage* minimal 70%.
- [ ] **Structured Logging:** Ganti fungsi `fmt.Println` atau `log.Println` bawaan dengan *Structured Logger* seperti `uber-go/zap` atau `sirupsen/logrus`. Log harus berformat JSON dan mencatat `tenant_id`, `user_id`, `endpoint`, dan `latency` untuk mempermudah *debugging* di *Production*.
- [ ] **Database Transaction Management:** Pada operasi yang memodifikasi lebih dari satu tabel (misal: *Clock Out* sekaligus meng-*update* total jam lembur), pastikan menggunakan `db.Begin()` dan `tx.Rollback()` jika terjadi *error*, agar data tidak korup (ACID Compliance).

## 🚀 5. CI/CD & Deployment
File `Dockerfile` dan `docker-compose.yaml` sudah ada, namun alur otomatisasi perlu disiapkan.
- [ ] **GitHub Actions / GitLab CI:** Buat *pipeline* yang otomatis menjalankan `go build`, `golangci-lint` (Linter), dan `go test` setiap kali ada *Pull Request* ke *branch* `main`.
- [ ] **Database Migration System:** Implementasikan sistem migrasi *database* terstruktur (seperti `golang-migrate/migrate` atau fitur bawaan ORM) yang otomatis berjalan saat proses *deployment*. Jangan melakukan alter tabel manual di server *Production*.
