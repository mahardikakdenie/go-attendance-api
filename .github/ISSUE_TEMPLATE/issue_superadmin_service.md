# Issues: `superadmin_service.go` — Performance & Safety Fixes

> File: `internal/service/superadmin_service.go`
> Reviewed: 2026-04-20
> Total issues: 7 (2 kritikal, 4 penting, 1 saran)

---

## 🔴 Kritikal

---

### [ISSUE-001] Error `bcrypt` diabaikan di `CreatePlatformAccount`

**Severity:** Critical
**Function:** `CreatePlatformAccount`
**Label:** `bug`, `security`

#### Deskripsi

`bcrypt.GenerateFromPassword` mengembalikan error yang diabaikan dengan blank identifier `_`. Jika hashing gagal (misalnya password > 72 bytes atau resource constraint), user tetap dibuat ke DB dengan password kosong — celah keamanan serius.

#### Kode saat ini

```go
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

#### Perbaikan

```go
// Validasi panjang sebelum hashing
if len([]byte(password)) > 72 {
    return model.UserResponse{}, errors.New("password exceeds maximum length of 72 bytes")
}

// Handle error
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
    return model.UserResponse{}, fmt.Errorf("failed to hash password: %w", err)
}
```

#### Acceptance Criteria

- [ ] Error dari `bcrypt.GenerateFromPassword` tidak diabaikan
- [ ] Validasi panjang password (max 72 bytes) ditambahkan sebelum hashing
- [ ] Jika hashing gagal, function return error dan user **tidak** dibuat ke DB

---

### [ISSUE-002] Nil slice response di `GetPlatformAccounts`

**Severity:** Critical
**Function:** `GetPlatformAccounts`
**Label:** `bug`, `api`

#### Deskripsi

Saat `users` kosong, `var responses []model.UserResponse` tetap `nil`. JSON encode nil slice menghasilkan `null`, bukan `[]`. Client API yang mengharapkan array bisa crash atau perlu handle dua kondisi sekaligus.

#### Kode saat ini

```go
var responses []model.UserResponse
for _, user := range users {
    responses = append(responses, ...)
}
```

#### Perbaikan

```go
// Pre-alokasi dengan kapasitas diketahui
responses := make([]model.UserResponse, 0, len(users))
for _, user := range users {
    responses = append(responses, toUserResponse(user))
}
```

#### Acceptance Criteria

- [ ] Response tidak pernah `null` pada success path — selalu `[]` saat data kosong
- [ ] Slice di-pre-alokasi dengan `make([]T, 0, len(users))` untuk efisiensi memori
- [ ] Helper `toUserResponse` dibuat terpusat (lihat ISSUE-004)

---

## 🟡 Penting

---

### [ISSUE-003] Silent-fail saat role invalid di `UpdatePlatformAccount`

**Severity:** Warning
**Function:** `UpdatePlatformAccount`
**Label:** `bug`, `ux`

#### Deskripsi

Jika `RoleID` pada request invalid atau bukan system role, update role di-skip tanpa error. User mengira update berhasil, padahal role tidak berubah.

#### Kode saat ini

```go
if req.RoleID != 0 {
    role, err := s.roleRepo.FindByID(ctx, req.RoleID)
    if err == nil && role != nil {
        if role.BaseRole == ... {
            user.RoleID = req.RoleID
        }
        // tidak ada error jika kondisi tidak terpenuhi
    }
}
```

#### Perbaikan

```go
if req.RoleID != 0 {
    role, err := s.roleRepo.FindByID(ctx, req.RoleID)
    if err != nil || role == nil {
        return model.UserResponse{}, errors.New("invalid role")
    }
    if !isSystemRole(role.BaseRole) {
        return model.UserResponse{}, errors.New("role must be a system role")
    }
    user.RoleID = req.RoleID
}
```

#### Acceptance Criteria

- [ ] Role invalid atau non-system role mengembalikan error eksplisit
- [ ] Response 4xx dikirim ke client saat validasi role gagal

---

### [ISSUE-004] Mapping `User → UserResponse` duplikat, tidak terpusat

**Severity:** Warning
**Function:** `GetPlatformAccounts`, `CreatePlatformAccount`, `UpdatePlatformAccount`
**Label:** `maintainability`, `refactor`

#### Deskripsi

Logika mapping field dari `model.User` ke `model.UserResponse` diulang inline di beberapa fungsi. Perubahan di struct `UserResponse` harus diupdate di banyak tempat.

#### Perbaikan

```go
func toUserResponse(user model.User) model.UserResponse {
    resp := model.UserResponse{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        TenantID:  user.TenantID,
        IsActive:  user.IsActive,
        CreatedAt: user.CreatedAt,
        BaseRole:  user.Role.BaseRole,
    }
    if user.Role.ID != 0 {
        resp.Role = &model.RoleResponse{
            ID:       user.Role.ID,
            Name:     user.Role.Name,
            BaseRole: user.Role.BaseRole,
        }
    }
    return resp
}
```

#### Acceptance Criteria

- [ ] Helper `toUserResponse` dibuat dan dipakai di semua tempat yang relevan
- [ ] Tidak ada duplikasi mapping inline

---

### [ISSUE-005] Urutan response `ListAllPermissions` tidak deterministik

**Severity:** Warning
**Function:** `ListAllPermissions`
**Label:** `bug`, `api`

#### Deskripsi

Iterasi `for moduleKey := range modulesMap` menghasilkan urutan acak tiap request karena Go map tidak menjamin urutan. Response API menjadi tidak konsisten dan menyulitkan client.

#### Perbaikan

```go
import "sort"

sort.Slice(result, func(i, j int) bool {
    return result[i].Key < result[j].Key
})
```

#### Acceptance Criteria

- [ ] Response `ListAllPermissions` selalu dalam urutan yang sama (sort by key)
- [ ] Unit test memverifikasi konsistensi urutan

---

### [ISSUE-006] Race condition antara create role dan assign permission di `CreateSystemRole`

**Severity:** Warning
**Function:** `CreateSystemRole`
**Label:** `bug`, `data-integrity`

#### Deskripsi

Jika `roleRepo.Create` berhasil tapi `roleRepo.UpdatePermissions` gagal, role terbuat di DB tanpa permission — state tidak konsisten. Perlu database transaction.

#### Kode saat ini

```go
if err := s.roleRepo.Create(ctx, role); err != nil {
    return model.Role{}, err
}
if len(req.PermissionIDs) > 0 {
    if err := s.roleRepo.UpdatePermissions(ctx, role.ID, req.PermissionIDs); err != nil {
        return model.Role{}, err  // role sudah terbuat, tidak di-rollback
    }
}
```

#### Perbaikan (pendekatan transaction)

```go
func (r *roleRepository) CreateWithPermissions(ctx context.Context, role *model.Role, permissionIDs []uint) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(role).Error; err != nil {
            return err
        }
        if len(permissionIDs) > 0 {
            // assign permissions dalam transaksi yang sama
        }
        return nil
    })
}
```

#### Acceptance Criteria

- [ ] Create role dan assign permission berada dalam satu DB transaction
- [ ] Jika salah satu gagal, keduanya di-rollback
- [ ] Perilaku sama diterapkan di `UpdateSystemRole`

---

## 🟢 Saran

---

### [ISSUE-007] Goroutine email tanpa timeout dan error logging

**Severity:** Suggestion
**Function:** `CreatePlatformAccount`
**Label:** `reliability`, `observability`

#### Deskripsi

Email dikirim via goroutine tanpa context timeout atau error logging. Jika SMTP lambat atau gagal, tidak ada jejak di log dan goroutine bisa leak.

#### Perbaikan

```go
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := utils.SendEmail(ctx, []string{user.Email}, subject, emailHtml); err != nil {
        log.Printf("warn: failed to send welcome email to %s: %v", user.Email, err)
    }
}()
```

#### Acceptance Criteria

- [ ] Goroutine email menggunakan context dengan timeout (minimal 30 detik)
- [ ] Error pengiriman email di-log (warn level)
```

Atau ambil dari context auth jika sudah tersedia:

```go
performerID := auth.UserIDFromContext(ctx)
```

#### Acceptance Criteria

- [ ] `CreatePlatformAccount` menerima `performerID` atau mengambilnya dari context
- [ ] Audit log tidak lagi menggunakan hardcoded `UserID: 1`
- [ ] Interface `SuperadminService` diupdate sesuai perubahan signature

---

## Ringkasan

| ID | Fungsi | Severity | Label |
|----|--------|----------|-------|
| ISSUE-001 | `CreatePlatformAccount` | 🔴 Critical | `bug`, `security` |
| ISSUE-002 | `GetPlatformAccounts` | 🔴 Critical | `bug`, `api` |
| ISSUE-003 | `UpdatePlatformAccount` | 🟡 Warning | `bug`, `ux` |
| ISSUE-004 | Multiple | 🟡 Warning | `maintainability`, `refactor` |
| ISSUE-005 | `ListAllPermissions` | 🟡 Warning | `bug`, `api` |
| ISSUE-006 | `CreateSystemRole` | 🟡 Warning | `bug`, `data-integrity` |
| ISSUE-007 | `CreatePlatformAccount` | 🟢 Suggestion | `reliability` |
| ISSUE-008 | `CreatePlatformAccount` | 🟢 Suggestion | `audit` |
