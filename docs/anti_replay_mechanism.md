# Anti-Replay Mechanism Documentation

## 1. Konsep Dasar Anti-Replay

**Anti-Replay Attack** adalah teknik di mana penyerang menangkap paket data yang valid (misalnya request absensi yang sudah ditandatangani) dan mengirimkannya kembali (replay) ke server untuk memanipulasi keadaan (misal: melakukan Clock-In berkali-kali dengan data yang sama).

Mekanisme Anti-Replay memastikan bahwa setiap request hanya dapat diproses **satu kali saja**.

### Komponen Utama:
1.  **Nonce (Number used once):** String acak unik untuk setiap request.
2.  **Timestamp:** Waktu saat request dibuat oleh client.
3.  **Signature/HMAC:** Hash dari (Payload + Nonce + Timestamp) menggunakan Secret Key untuk memastikan integritas data.

---

## 2. Cara Membuat & Mengimplementasikan

Untuk mengimplementasikan ini pada proyek `go-attendance-api`, kita perlu menambahkan layer keamanan pada Middleware.

### Langkah 1: Struktur Request di Client (FE)
Client harus mengirimkan header tambahan:
- `X-Nonce`: UUID unik.
- `X-Timestamp`: Unix timestamp saat ini.
- `X-Signature`: HMAC-SHA256 dari body request + nonce + timestamp.

### Langkah 2: Validasi di Server (Middleware)
Server akan melakukan pengecekan berikut:

1.  **Cek Window Waktu (Timeliness):**
    - Bandingkan `X-Timestamp` dengan waktu server. Jika selisihnya > 5 menit, tolak request. Ini membatasi masa berlaku paket data.
2.  **Cek Unik (Nonce Tracking):**
    - Simpan `X-Nonce` di Redis dengan TTL (misal 5 menit, sama dengan window waktu).
    - Jika `X-Nonce` sudah ada di Redis, berarti ini adalah request **Replay**. Tolak!
3.  **Verifikasi Signature:**
    - Hitung ulang HMAC menggunakan payload yang diterima dan Secret Key di server. Cocokkan dengan `X-Signature`.

### Contoh Implementasi Middleware (Go):
```go
func AntiReplayMiddleware(redis *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        nonce := c.GetHeader("X-Nonce")
        timestamp := c.GetHeader("X-Timestamp")
        signature := c.GetHeader("X-Signature")

        // 1. Validasi Window Waktu (misal 5 menit)
        ts, _ := strconv.ParseInt(timestamp, 10, 64)
        if time.Now().Unix() - ts > 300 {
            c.AbortWithStatusJSON(403, gin.H{"error": "Request expired"})
            return
        }

        // 2. Cek Nonce di Redis (Atomic check)
        nonceKey := "nonce:" + nonce
        isNew, _ := redis.SetNX(c, nonceKey, "1", 5*time.Minute).Result()
        if !isNew {
            c.AbortWithStatusJSON(403, gin.H{"error": "Duplicate request detected"})
            return
        }

        c.Next()
    }
}
```

---

## 3. Tantangan ke Depan

Implementasi Anti-Replay membawa beberapa tantangan teknis:

1.  **Sinkronisasi Waktu (Clock Skew):**
    - Jika waktu di HP karyawan tidak sinkron dengan server (misal beda beberapa menit), request akan ditolak terus menerus.
    - *Solusi:* Server harus memberikan toleransi window (misal ±5 menit) atau menyediakan endpoint `/ping` untuk sinkronisasi waktu client-server.

2.  **Overhead Performa Redis:**
    - Setiap request masuk harus melakukan operasi `SetNX` ke Redis. Pada jam puncak (peak hour) absensi, ini menambah beban I/O.
    - *Solusi:* Pastikan Redis menggunakan clustering atau optimalkan koneksi pool.

3.  **Manajemen Secret Key:**
    - Jika Secret Key bocor, penyerang bisa membuat signature baru.
    - *Solusi:* Gunakan rotasi kunci atau integrasikan dengan sistem autentikasi (misal: menggunakan hash dari JWT token sebagai key).

4.  **Skalabilitas Penyimpanan Nonce:**
    - Jika TTL nonce terlalu lama, memori Redis akan cepat penuh.
    - *Solusi:* Gunakan TTL yang sekecil mungkin (sama dengan window waktu validitas timestamp).

---

## 4. Kesimpulan
Anti-Replay sangat krusial untuk endpoint sensitif seperti **Attendance Record** dan **Payroll Approval**. Dengan kombinasi **Timestamp + Redis Nonce**, kita bisa menjamin integritas sistem dari serangan manipulasi request berulang.
