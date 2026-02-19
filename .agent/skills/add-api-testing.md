# Skill: Add API Testing & Manual HTTP

## Trigger

Gunakan skill ini setelah menambah endpoint baru (controller + route), untuk:

1. Menulis integration test di `test/`
2. Menambah endpoint ke `test/manual.http`

## Prerequisites

- Endpoint sudah terdaftar di `route.go`
- Controller sudah selesai diimplementasikan
- Skema response (`model.WebResponse[T]`) sudah fix

---

## Part 1 — Integration Test

### Step 1: Pilih file test yang tepat

- Gunakan file yang sudah ada jika domain sama — `test/auth_test.go`, dll.
- Buat file baru `test/{domain}_test.go` jika domain berbeda.

### Step 2: Pastikan helpers tersedia

Di `test/auth_test.go` atau file test utama, helpers berikut **harus ada**:

```go
// POST tanpa auth
func doPost(t *testing.T, path string, body interface{}) *http.Response

// POST dengan Bearer token
func doPostAuth(t *testing.T, path, token string, body interface{}) *http.Response

// GET dengan optional Bearer token
func doGet(t *testing.T, path, token string) *http.Response

// Baca response body sebagai string
func readBody(resp *http.Response) string

// Extract token pair dari body
func parseTokens(t *testing.T, body string) model.TokenResponse
```

Jika belum ada, tambahkan ke file test yang bersangkutan.

### Step 3: Struktur test function

Gunakan satu `Test{Domain}` parent dengan banyak `t.Run` subtest. State dibagikan antar subtest via **closure variable**:

```go
func Test{Domain}(t *testing.T) {
    // 1. Setup: cleanup di awal + Cleanup hook di akhir
    cleanupUser(testEmail)
    t.Cleanup(func() { cleanupUser(testEmail) })

    // 2. Shared state via closure
    var (
        accessToken  string
        refreshToken string
    )

    // ── Subtest: happy path ──────────────────────────────────────────────────

    t.Run("Create_Success", func(t *testing.T) {
        resp := doPost(t, "/api/{resource}", map[string]string{
            "field": "value",
        })
        body := readBody(resp)
        assert.Equal(t, http.StatusOK, resp.StatusCode, "body: %s", body)

        var wrapper struct {
            Data model.{Name}Response `json:"data"`
        }
        require.NoError(t, json.Unmarshal([]byte(body), &wrapper))
        assert.NotEmpty(t, wrapper.Data.ID)
    })

    // ── Subtest: error cases ─────────────────────────────────────────────────

    t.Run("Create_InvalidRequest", func(t *testing.T) {
        resp := doPost(t, "/api/{resource}", map[string]string{})
        assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
    })

    // ── Subtest: auth protected ──────────────────────────────────────────────

    t.Run("Update_Unauthenticated", func(t *testing.T) {
        resp := doPostAuth(t, "/api/{resource}/update", "", map[string]string{})
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
    })

    t.Run("Update_Success", func(t *testing.T) {
        resp := doPostAuth(t, "/api/{resource}/update", accessToken, map[string]string{
            "field": "new_value",
        })
        body := readBody(resp)
        assert.Equal(t, http.StatusOK, resp.StatusCode, "body: %s", body)
    })
}
```

### Step 4: Cleanup function pattern

Hapus **child → parent** dan gunakan `db.Exec` dengan `?` placeholder:

```go
func cleanupUser(email string) {
    // Hapus dependents terlebih dahulu (safeguard jika cascade tidak berjalan)
    db.Exec("DELETE FROM user_auth_providers WHERE email = ?", email)
    db.Exec("DELETE FROM users WHERE email = ?", email)
}
```

> **Jangan** gunakan `db.Where("id is not null").Delete(...)` — ini membutuhkan primary key dan bisa gagal.

### Step 5: Tips menulis subtest

| Skenario           | HTTP Method | Expected Status  |
| ------------------ | ----------- | ---------------- |
| Sukses             | POST/GET    | 200 OK           |
| Field wajib kosong | POST        | 400 BadRequest   |
| Email duplikat     | POST        | 409 Conflict     |
| Belum login        | GET/POST    | 401 Unauthorized |
| Forbidden          | GET/POST    | 403 Forbidden    |
| Data tidak ada     | GET         | 404 NotFound     |

### Step 6: Jalankan test

```bash
go test ./test/... -v -run Test{Domain} -count=1
```

Jalankan minimal **dua kali berturut-turut** untuk memastikan cleanup idempoten.

### Bugs Umum yang Perlu Dicek

- **Unique constraint gagal**: pastikan field yang dipakai sebagai `provider_user_id` atau unique key di DB adalah nilai yang benar-benar unik per user (misalnya email), bukan string kosong.
- **State antar subtest**: closure variable (`accessToken`, dll.) harus di-set oleh subtest sebelumnya; jika subtest gagal, downstream akan ikut gagal.
- **Soft delete**: jika entity punya `deleted_at`, cleanup dengan `db.Exec("DELETE FROM ...")` perlu menghapus hard (tanpa GORM soft delete).

---

## Part 2 — Manual HTTP

File: `test/manual.http`

### Struktur file

```http
@baseUrl = http://localhost:3000

# ─────────────────────────────────────────────────────────────────────────────
# {DOMAIN} — Guest endpoints
# ─────────────────────────────────────────────────────────────────────────────

### {Deskripsi singkat endpoint}
# @name {requestName}   ← tambahkan hanya pada request yang response-nya dipakai request lain
POST {{baseUrl}}/api/{resource}
Content-Type: application/json

{
  "field": "value"
}

# ─────────────────────────────────────────────────────────────────────────────
# {DOMAIN} — Protected endpoints
# ─────────────────────────────────────────────────────────────────────────────

### {Deskripsi singkat endpoint}
GET {{baseUrl}}/api/{resource}
Authorization: Bearer {{login.response.body.$.data.access_token}}
```

### Cara referensi response dari request sebelumnya

1. Beri nama request dengan `# @name login` (di atas baris method)
2. Pakai `{{login.response.body.$.data.{field}}}` di request lain

Contoh lengkap:

```http
### Login
# @name login
POST {{baseUrl}}/api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password"
}

### Endpoint yang butuh token
GET {{baseUrl}}/api/auth/me
Authorization: Bearer {{login.response.body.$.data.access_token}}

### Endpoint yang butuh refresh token
POST {{baseUrl}}/api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "{{login.response.body.$.data.refresh_token}}"
}
```

### Tips manual.http

- Urutan request dari atas ke bawah mengikuti alur kerja user
- Guest endpoints (tanpa auth) duluan, protected endpoints kemudian
- `# @name` hanya pada request yang hasilnya akan dipakai request lain
- Tulis komentar di atas tiap request: `### {deskripsi}`
- Pisahkan grup dengan separator: `# ─────────────────────`

---

## Checklist

- [ ] Test file dibuat/diupdate di `test/`
- [ ] Cleanup function menghapus semua data test (child → parent)
- [ ] Happy path subtest: 200 OK + assert field penting
- [ ] Error case subtests: 400/401/403/404/409
- [ ] Auth-protected endpoint ditest dengan dan tanpa token
- [ ] Test dijalankan 2x berturut-turut: hasil sama (idempoten)
- [ ] `test/manual.http` ditambah request untuk semua endpoint baru
- [ ] Request yang response-nya dipakai request lain diberi `# @name`

## Referensi Rules

- [testing.md](../rules/testing.md)
- [controller.md](../rules/controller.md)
