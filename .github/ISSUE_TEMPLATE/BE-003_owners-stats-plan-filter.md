# BE-003: Tambah Query Parameter `plan` pada Endpoint Owners Stats

## Prioritas: **MEDIUM**
## Label: `backend` `enhancement` `superadmin`

---

## Konteks

Frontend sudah mengimplementasi dropdown filter berdasarkan **subscription plan** (Basic / Pro / Enterprise) pada halaman `/admin/tenants`. Backend perlu mendukung parameter `plan` agar filtering dilakukan di sisi server.

---

## Endpoint yang Harus Diubah

```
GET /v1/superadmin/owners-stats
```

## Parameter yang Perlu Ditambahkan

| Param  | Type     | Required | Nilai yang Diterima                  |
|--------|----------|----------|--------------------------------------|
| `plan` | `string` | No       | `Basic`, `Pro`, atau `Enterprise`    |

## Spesifikasi Filter

- Filter berdasarkan **nama plan** yang terasosiasi dengan subscription tenant.
- Matching harus **exact match** (bukan partial), case-insensitive.
- Jika parameter tidak dikirim atau kosong → tampilkan semua (behavior existing).

### Contoh SQL Logic (pseudocode)

```sql
-- Jika plan = "Pro"
WHERE LOWER(plan.name) = LOWER('Pro')

-- Atau jika pakai relasi:
JOIN subscriptions s ON s.tenant_id = t.id
JOIN plans p ON p.id = s.plan_id
WHERE LOWER(p.name) = LOWER('Pro')
```

## Contoh Request

```
GET /v1/superadmin/owners-stats?limit=10&offset=0&plan=Enterprise
```

## Kombinasi dengan Semua Parameter

Parameter `plan` harus bisa dikombinasikan dengan `search`, `status`, `limit`, `offset` menggunakan **AND**:

```
GET /v1/superadmin/owners-stats?limit=10&offset=0&search=corp&status=Active&plan=Pro
```

```sql
WHERE (
  LOWER(owner.name) LIKE '%corp%' 
  OR LOWER(tenant.name) LIKE '%corp%' 
  OR LOWER(tenant.code) LIKE '%corp%'
)
AND tenant.is_suspended = false
AND LOWER(plan.name) = LOWER('Pro')
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
      "total": 5,
      "per_page": 10,
      "current_page": 1
    }
  },
  "data": [
    {
      "tenant_plan": "Enterprise",
      "...": "..."
    }
  ]
}
```

## Edge Cases

- Jika value `plan` bukan salah satu dari plan yang ada di database → kembalikan `data: []` (bukan error, karena plan bisa bertambah di masa depan).
- Jika tenant belum punya subscription/plan → tenant tersebut **tidak masuk** hasil filter kecuali `plan` parameter tidak dikirim.

## Frontend Reference

File: `src/views/admin/OwnersStats.tsx`

```tsx
const [planFilter, setPlanFilter] = useState<string>("all");

// Dikirim ke API sebagai:
// plan=Basic | plan=Pro | plan=Enterprise | (tidak dikirim jika "all")
```

## Acceptance Criteria

- [ ] Endpoint menerima query parameter `plan` (opsional)
- [ ] Filter exact match berdasarkan nama plan
- [ ] Bisa dikombinasikan dengan `search`, `status`, `limit`, `offset`
- [ ] `pagination.total` akurat sesuai hasil filter
- [ ] Tidak ada breaking change pada response structure
