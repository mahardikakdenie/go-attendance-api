# Panduan Integrasi Frontend: Modul Subscription (Langganan)

Dokumen ini berisi panduan *step-by-step* untuk tim Frontend (FE) dalam mengimplementasikan integrasi API Subscription, khususnya untuk endpoint `/api/v1/subscriptions/me`. 

Sistem backend menggunakan struktur Role-Based Access Control (RBAC) yang ketat berbasis `BaseRole`.

---

## 1. Menyembunyikan Menu Berdasarkan Role (Role-Based UI)

Halaman dan menu langganan (Subscription/Billing) **tidak boleh** dimunculkan untuk role HR atau Employee biasa. Hanya pemilik bisnis (Admin) atau departemen keuangan (Finance) yang memiliki wewenang.

**Langkah Implementasi:**
1. Tangkap response dari API Login (`/api/v1/auth/login`).
2. Simpan nilai `base_role` dari user ke dalam *Global State* (Zustand/Redux/Context).
3. Buat kondisi di komponen Sidebar/Navbar.

```javascript
// Contoh Logika Sidebar
const user = useGlobalStore((state) => state.user);

// Daftar base_role yang diizinkan
const allowedRoles = ['SUPERADMIN', 'ADMIN', 'FINANCE'];
const canAccessSubscription = allowedRoles.includes(user?.base_role);

return (
  <nav>
    <MenuItem to="/dashboard">Dashboard</MenuItem>
    <MenuItem to="/attendance">Absensi</MenuItem>
    
    {/* Render bersyarat */}
    {canAccessSubscription && (
      <MenuItem to="/subscription">Paket & Langganan</MenuItem>
    )}
  </nav>
);
```

---

## 2. Page-Level Guard & Pemanggilan API

Jika user "memaksa" masuk dengan mengetik URL secara manual (misal: `/subscription`), halaman harus menangkap error `403 Forbidden` dari API dan menampilkan komponen *Access Denied*.

**Endpoint:** `GET /api/v1/subscriptions/me`

**Langkah Implementasi:**
1. Lakukan *fetch* ke API saat komponen di-mount (`useEffect`).
2. Tangkap error response. Jika statusnya `403`, tampilkan layar penolakan akses.

```javascript
import { useEffect, useState } from 'react';

export default function SubscriptionPage() {
  const [subscription, setSubscription] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchSubscription = async () => {
      try {
        const response = await api.get('/api/v1/subscriptions/me');
        // Backend membungkus data di dalam properti `data`
        setSubscription(response.data.data);
      } catch (err) {
        if (err.response?.status === 403) {
          setError("Akses Ditolak: Halaman ini hanya untuk Admin dan Finance.");
        } else {
          setError("Terjadi kesalahan saat memuat data langganan.");
        }
      } finally {
        setLoading(false);
      }
    };

    fetchSubscription();
  }, []);

  if (loading) return <LoadingSpinner />;
  if (error) return <AccessDeniedCard message={error} />;

  // Render komponen detail langganan jika berhasil
  return <SubscriptionDetail data={subscription} />;
}
```

---

## 3. Penanganan State Kosong (Empty State)

Backend dirancang sedemikian rupa sehingga jika tenant (perusahaan) **belum memiliki data langganan aktif**, API *tidak* akan mengembalikan error 500, melainkan mengembalikan *success (200)* dengan data berupa *empty object* `{}`.

**Langkah Implementasi:**
Lakukan pengecekan validitas objek sebelum melakukan render UI.

```javascript
// Di dalam komponen render
{subscription && Object.keys(subscription).length > 0 ? (
  <SubscriptionDetail data={subscription} />
) : (
  <EmptySubscriptionPrompt 
    title="Belum Ada Paket Aktif" 
    message="Silakan pilih paket berlangganan untuk membuka semua fitur."
    actionButton={<button>Pilih Paket</button>}
  />
)}
```

---

## 4. Keamanan & Headers (Penting untuk Production)

Backend API memiliki fitur *Anti-Replay Mechanism* dan *HMAC Signature*. Jika aplikasi dijalankan di *production* (yaitu ketika backend `APP_ENV != development`), FE **wajib** mengirimkan header khusus pada setiap *protected request*.

Jika header ini tidak ada, request akan ditolak dengan status `403`.

**Header yang dibutuhkan:**
*   `X-Timestamp`: Waktu request saat ini dalam milidetik (Unix Milliseconds). *Request yang usianya lebih dari 10 detik akan ditolak.*
*   `X-Request-ID`: UUID unik untuk mencegah *replay attack* (request dikirim ulang).
*   `X-Signature`: Hash HMAC-SHA256 dari kombinasi `Request Body` + `Timestamp` + `Request ID`.

*Catatan: Pastikan `axios` interceptor atau fetch wrapper Anda sudah dikonfigurasi untuk menyisipkan header-header ini secara otomatis jika berjalan di environment production.*
