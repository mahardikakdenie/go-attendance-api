# BE-013: Real-time Menu Cache Invalidation (Redis & SSE)

## Prioritas: **HIGH**
## Label: `backend` `cache` `redis` `real-time`

---

## Deskripsi Masalah
Saat Superadmin melakukan update pada **System Role Permissions** atau **Menu Architecture**, perubahan tersebut tidak langsung tercermin di sidebar user (endpoint `/v1/menus/me`). User harus melakukan logout-login ulang atau menunggu waktu yang lama agar menu berubah.

Hal ini terindikasi karena:
1. Cache Redis pada endpoint `/v1/menus/me` tidak di-invalidate saat data master role berubah.
2. Tidak ada mekanisme pemberitahuan ke Client (Frontend) bahwa state menu telah berubah.

---

## Perubahan yang Dibutuhkan

### 1) Redis Cache Invalidation (The "Ripple" Effect)
Setiap kali terjadi perubahan pada:
- Tabel `role_permissions`
- Tabel `menus`
- Endpoint `PUT /v1/superadmin/system-roles/:id`

Backend **WAJIB** menghapus (flush/delete) key cache Redis yang berkaitan dengan navigasi. Karena navigasi bersifat user-specific, pastikan pola key-nya dapat diidentifikasi (misal: `user_nav:*`).

### 2) Real-time Sync via SSE (Server-Sent Events)
Untuk memberikan pengalaman "Live Update" tanpa refresh halaman:
- Setelah cache di-invalidate, Backend harus mengirimkan event melalui SSE (Server-Sent Events) ke client yang sedang aktif.
- **Event Name:** `RELOAD_NAV` atau `SYNC_PERMISSIONS`.
- **Payload:** `{ "action": "refresh_sidebar" }`.

### 3) Optimized Query
Pastikan endpoint `/v1/menus/me` selalu melakukan pengecekan permission terbaru terhadap session yang aktif sebelum menyajikan data dari cache (atau gunakan mekanisme *cache-tagging*).

---

## Acceptance Criteria
- [ ] Update pada System Role Permissions langsung menghapus cache navigasi di Redis.
- [ ] Backend mengirimkan signal SSE saat terjadi perubahan arsitektur menu/role.
- [ ] User mendapatkan update menu di sidebar secara instan (atau via trigger event di FE) tanpa perlu re-login.
- [ ] Perubahan pada Master Blueprint (BE-012) ter-propagate dengan benar ke seluruh sesi aktif.

---

## Catatan untuk Developer
Gunakan mekanisme `Observer` atau `Event Dispatcher` di level aplikasi agar logic invalidasi cache tidak mengotori Controller utama.
