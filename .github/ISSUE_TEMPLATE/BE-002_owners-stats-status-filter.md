# BE-002: Tambah Query Parameter `status` pada Endpoint Owners Stats

## Prioritas: **MEDIUM**
## Label: `backend` `enhancement` `superadmin`

---

## Konteks

Frontend sudah mengimplementasi dropdown filter berdasarkan **status tenant** (Active / Suspended) pada halaman `/admin/tenants`. Backend perlu mendukung parameter `status` agar filtering dilakukan di sisi server, bukan client-side.

---

## Endpoint yang Harus Diubah

```
GET /v1/superadmin/owners-stats
```

## Parameter yang Perlu Ditambahkan

| Param    | Type     | Required | Nilai yang Diterima          |
|----------|----------|----------|------------------------------|
| `status` | `string` | No       | `Active` atau `Suspended`    |

## Spesifikasi Filter

- Jika `status=Active` → hanya tampilkan tenant dengan status aktif (`tenant.is_suspended = false` atau kolom status equivalent).
- Jika `status=Suspended` → hanya tampilkan tenant yang di-suspend (`tenant.is_suspended = true`).
- Jika parameter tidak dikirim atau kosong → tampilkan semua (behavior existing).

### Contoh SQL Logic (pseudocode)

```sql
-- Jika status = "Active"
WHERE tenant.is_suspended = false

-- Jika status = "Suspended"  
WHERE tenant.is_suspended = true
```

## Contoh Request

```
GET /v1/superadmin/owners-stats?limit=10&offset=0&status=Suspended
```

## Kombinasi dengan Parameter Lain

Parameter `status` harus bisa dikombinasikan dengan parameter lain menggunakan **AND**:

```
GET /v1/superadmin/owners-stats?limit=10&offset=0&search=acme&status=Active
```

```sql
WHERE (
  LOWER(owner.name) LIKE '%acme%' OR ...
)
AND tenant.is_suspended = false
```

## Expected Response

Struktur response **tidak berubah**. Hanya `data` yang terfilter dan `pagination.total` yang merefleksikan jumlah setelah filter.

```json
{
  "success": true,
  "meta": {
    "message": "Success",
    "code": 200,
    "status": "OK",
    "pagination": {
      "total": 2,
      "per_page": 10,
      "current_page": 1
    }
  },
  "data": [
    {
      "tenant_status": "Suspended",
      "...": "..."
    }
  ]
}
```

## Edge Cases

- Jika value `status` bukan `Active` atau `Suspended` → kembalikan error `400 Bad Request` dengan pesan yang jelas, atau abaikan filter (pilih salah satu, dokumentasikan).
- Jika tidak ada tenant dengan status tersebut → `data: []`, `pagination.total: 0`.

## Frontend Reference

File: `src/views/admin/OwnersStats.tsx`

```tsx
const [statusFilter, setStatusFilter] = useState<string>("all");

// Dikirim ke API sebagai:
// status=Active | status=Suspended | (tidak dikirim jika "all")
```

## Acceptance Criteria

- [ ] Endpoint menerima query parameter `status` (opsional)
- [ ] Filter berdasarkan `Active` / `Suspended`
- [ ] Bisa dikombinasikan dengan `search`, `plan`, `limit`, `offset`
- [ ] `pagination.total` akurat sesuai hasil filter
- [ ] Value invalid di-handle dengan baik (400 atau ignore)
- [ ] Tidak ada breaking change pada response structure
