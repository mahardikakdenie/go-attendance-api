# Frontend Planning: System Menus Guardrails & Visibility

Dokumen ini berisi panduan dan *Action Plan* untuk Tim Frontend (FE) terkait manajemen visibilitas Menu, khususnya untuk menu dengan flag `is_system: true` (seperti **Platform Control**, **Menu Management**, dll).

## 🛑 Problem Statement
Saat ini di halaman Menu Management, FE mengizinkan assignment sembarang Role (contoh: `admin`, `hr`, `employee`) ke dalam menu yang memiliki `is_system: true`. Backend akan menerima dan menyimpannya (ke tabel `role_menu_visibility`), **TAPI** backend memiliki *guard/restriction* internal yang secara absolut menyembunyikan menu tersebut dari siapapun yang `base_role`-nya bukan `superadmin`.

Hal ini membingungkan *user* karena di halaman manajemen seolah-olah menu sudah di-assign, tapi menu tidak pernah muncul di *sidebar* (`/api/v1/menus/me`).

## 🧠 Backend Rules (Yang Harus FE Pahami)
Berikut adalah urutan *filtering* mutlak di backend (`filterMenus` di `menu_service.go`):
1. **Rule A (System Check)**: Jika menu memiliki `is_system: true` DAN `base_role` user yang login **bukan** `superadmin`, maka menu tersebut otomatis **dibuang (skipped)**, tidak peduli dia di-assign ke role tersebut atau tidak.
2. **Rule B (Empty Parent)**: Jika sebuah menu induk (contoh: *Platform Control*) tidak memiliki URL path (`path: ""`) dan semua *child*-nya tersembunyi karena kurang hak akses/role, maka menu induk tersebut juga **disembunyikan otomatis**.

---

## 🎯 Action Plan untuk Tim Frontend

Untuk menghindari kebingungan UI/UX dan agar FE sejalan dengan *guard* Backend, FE harus mengimplementasikan hal berikut:

### 1. Modifikasi Halaman "Edit Menu" (`/admin/menus`)
Ketika memanggil endpoint `GET /api/v1/superadmin/menus` untuk mendapatkan daftar menu, FE akan mendapatkan respons payload berikut:
```json
{
  "id": 1,
  "key": "platform-group",
  "label": "Platform Control",
  "is_system": true,     // <-- INI KUNCINYA
  "allowed_roles": [1]
}
```

**✅ FE Task:**
Pada form Edit Menu atau Assignment Role:
* Lakukan pengecekan: `if (menu.is_system === true)`
* **Disable/Lock** komponen *dropdown/checkbox* **"Allowed Roles"**.
* Jangan izinkan user memilih Role biasa. *Force-select* hanya role yang `base_role === "superadmin"`.
* **Tampilkan Alert/Pesan Informasi UI**: 
  > *"ℹ️ This is a System Menu. It is strictly reserved for Superadmin roles and cannot be assigned to regular tenant roles."*

### 2. Modifikasi Payload Update (`PUT /api/v1/superadmin/menus/:id`)
* **Endpoint**: `PUT /api/v1/superadmin/menus/:id`
* **Payload**: 
  ```json
  {
    "label": "Platform Control",
    "icon": "ShieldCheck",
    "allowed_roles": [1] // Pastikan FE tidak mengirim role ID selain superadmin jika is_system = true
  }
  ```
* **✅ FE Task**: Lakukan validasi sebelum submit form. Jika `is_system: true`, cegah payload `allowed_roles` berisi ID dari Role non-superadmin.

### 3. Modifikasi Halaman Role & Menu Overview
Daripada FE menebak-nebak menu mana yang akan muncul untuk Role tertentu, gunakan endpoint yang sudah disediakan backend khusus untuk melihat struktur akhir menu per-role (yang sudah dikalkulasi dengan segala *guard* dan *restriction*).

* **Endpoint**: `GET /api/v1/menus/overview`
* **Header/Auth**: Butuh akses base_role Superadmin atau Admin.
* **Response**:
  ```json
  [
    {
      "role_name": "Admin",
      "base_role": "admin",
      "menus": [ /* Struktur Tree Murni yang benar-benar akan muncul di role ini */ ]
    },
    {
      "role_name": "Super Admin",
      "base_role": "superadmin",
      "menus": [ /* ... */ ]
    }
  ]
  ```
* **✅ FE Task**: Gunakan API ini untuk halaman visualisasi "Siapa bisa akses apa" di Menu Settings, sehingga FE tidak membohongi user tentang hasil konfigurasi.

---

## 📝 Kesimpulan untuk Flow FE
1. **Fetch Menu Detail**: Cek `is_system`.
2. **If `is_system == true`**: Kunci dropdown Role, pasang alert peringatan. FE tidak boleh memberikan ilusi bahwa menu sistem bisa di-assign ke role HR/Employee.
3. **If `is_system == false`**: Bebas di-assign oleh user (seperti menu Attendance, Payroll, dll).
4. **Validation**: Jangan submit role non-superadmin di endpoint PUT jika menu adalah *system menu*.
