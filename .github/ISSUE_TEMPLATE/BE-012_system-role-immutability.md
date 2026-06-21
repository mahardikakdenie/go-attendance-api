# BE-012: System Role Immutability Issue on Update

## Prioritas: **HIGH**
## Label: `backend` `bug` `rbac`

---

## Deskripsi Masalah
Saat mencoba melakukan update pada System Role melalui endpoint `PUT/PATCH /v1/superadmin/system-roles/:id`, API mengembalikan error **500 Internal Server Error** dengan pesan bahwa role tersebut bersifat *immutable*.

**Endpoint:** `PUT /v1/superadmin/system-roles/4` (atau ID lainnya)

**Response Error:**
```json
{
    "success": false,
    "meta": {
        "message": "Failed to patch system role",
        "code": 500,
        "status": "error"
    },
    "data": "this system role is immutable and cannot be modified"
}
```

---

## Analisa & Ekspektasi
Sesuai dengan alur **Master Blueprint**, Superadmin harus dapat memodifikasi *Capabilities Matrix* (Permissions) pada System Roles agar perubahan tersebut dapat merembes (ripple) ke seluruh tenant. 

Jika semua System Role di-set sebagai `immutable`, maka fitur dinamisasi menu dan permission di level platform tidak dapat berjalan.

### Rekomendasi Solusi:
1. **Relax Immutability:** Izinkan Superadmin untuk tetap mengupdate permission pada system roles, meskipun field fundamental (seperti `slug` atau `base_role`) mungkin tetap terkunci.
2. **Selective Lock:** Jika ada role tertentu yang memang HARUS dikunci total (misal: ID 1 untuk Superadmin inti), mohon dokumentasikan ID mana saja yang tidak boleh disentuh. Role blueprint lainnya (Admin, HR, Employee) seharusnya tetap bisa di-update permission-nya.
3. **Correct Status Code:** Jika aksi ditolak karena aturan bisnis, gunakan **403 Forbidden** atau **422 Unprocessable Entity**, bukan **500**.

---

## Acceptance Criteria
- [ ] Superadmin dapat mengupdate permission pada System Role (Blueprint) tanpa terhalang error "immutable".
- [ ] Perubahan permission pada System Role berhasil tersimpan di database.
- [ ] Error message lebih deskriptif jika ada field tertentu yang memang dilarang diubah.
