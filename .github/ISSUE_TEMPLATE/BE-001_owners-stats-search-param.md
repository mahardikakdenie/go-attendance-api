# BE-001: Tambah Query Parameter `search` pada Endpoint Owners Stats

## Prioritas: **HIGH**
## Label: `backend` `enhancement` `superadmin`

---

## Konteks

Saat ini endpoint `GET /v1/superadmin/owners-stats` hanya mendukung `limit` dan `offset` untuk pagination. Frontend sudah mengimplementasi fitur **server-side search** yang membutuhkan parameter `search` agar pencarian berlaku di **seluruh data**, bukan hanya data yang ada di halaman saat ini (client-side filtering).

Tanpa perubahan ini, user yang mencari tenant "XYZ" hanya bisa menemukan jika data tersebut kebetulan ada di halaman yang sedang aktif.

---

## Endpoint yang Harus Diubah

```
GET /v1/superadmin/owners-stats
```

## Parameter Saat Ini (Existing)

| Param    | Type     | Required | Keterangan          |
|----------|----------|----------|---------------------|
| `limit`  | `number` | No       | Default: 10         |
| `offset` | `number` | No       | Default: 0          |

## Parameter yang Perlu Ditambahkan

| Param    | Type     | Required | Keterangan                              |
|----------|----------|----------|-----------------------------------------|
| `search` | `string` | No       | Keyword pencarian, case-insensitive     |

## Spesifikasi Pencarian

Parameter `search` harus melakukan pencarian **case-insensitive** dan **partial match** (LIKE/ILIKE) pada kolom-kolom berikut:

1. **`owner.name`** — Nama pemilik/owner
2. **`tenant.name`** — Nama perusahaan/organisasi
3. **`tenant.code`** — Kode unik tenant

Gunakan operator **OR** antar kolom, sehingga jika keyword cocok di salah satu kolom, row tersebut masuk hasil.

### Contoh SQL Logic (pseudocode)

```sql
WHERE (
  LOWER(owner.name) LIKE LOWER('%{search}%')
  OR LOWER(tenant.name) LIKE LOWER('%{search}%')
  OR LOWER(tenant.code) LIKE LOWER('%{search}%')
)
```

## Contoh Request

```
GET /v1/superadmin/owners-stats?limit=10&offset=0&search=acme
```

## Expected Response (Tidak Berubah Strukturnya)

```json
{
  "success": true,
  "meta": {
    "message": "Success",
    "code": 200,
    "status": "OK",
    "pagination": {
      "total": 3,
      "per_page": 10,
      "current_page": 1
    }
  },
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@acme.com",
      "tenant_id": 5,
      "tenant_name": "Acme Corp",
      "tenant_code": "ACME",
      "tenant_plan": "Pro",
      "tenant_status": "Active",
      "employee_count": 45,
      "attendance_count": 1200,
      "leave_count": 30,
      "overtime_count": 15,
      "payroll_count": 12,
      "expense_count": 8,
      "created_at": "2025-01-15T10:00:00Z"
    }
  ]
}
```

> **PENTING:** `pagination.total` harus merefleksikan jumlah total row yang match dengan filter `search`, BUKAN total keseluruhan tanpa filter.

## Edge Cases

- Jika `search` kosong atau tidak dikirim → abaikan, tampilkan semua data (behavior existing).
- Jika `search` tidak cocok dengan data apapun → kembalikan `data: []` dengan `pagination.total: 0`.
- Karakter spesial dalam search (contoh: `%`, `_`) harus di-escape agar tidak merusak query SQL.

## Frontend Reference

File: `src/service/support.ts` — line 14-22

```typescript
export const getOwnersStats = (limit = 10, offset = 0, search?: string, status?: string, plan?: string) => {
  return secureRequest<APIResponse<OwnerStats[]>>("get", "/v1/superadmin/owners-stats", {
    limit,
    offset,
    ...(search ? { search } : {}),
    ...(status && status !== "all" ? { status } : {}),
    ...(plan && plan !== "all" ? { plan } : {}),
  });
};
```

## Acceptance Criteria

- [ ] Endpoint menerima query parameter `search` (opsional)
- [ ] Pencarian case-insensitive + partial match pada `owner.name`, `tenant.name`, `tenant.code`
- [ ] `pagination.total` akurat sesuai hasil filter
- [ ] Tidak ada breaking change pada response structure
- [ ] SQL injection safe (parameterized query / escape special chars)
