# Model Layer Rules

Model berisi Request DTO, Response DTO, Event DTO, dan converter functions.

## Response Model Pattern

```go
package model

type {Name}Response struct {
    ID        string `json:"id"`
    // ... field lain
    CreatedAt int64  `json:"created_at"`
    UpdatedAt int64  `json:"updated_at"`
}
```

## Request Model Pattern

```go
// Create request
type Create{Name}Request struct {
    // Field dari auth context (tidak dari body JSON)
    UserId string `json:"-" validate:"required"`

    // Field dari request body
    FieldName string `json:"field_name" validate:"required,max=100"`
}

// Update request — ID dari URL param, bukan body
type Update{Name}Request struct {
    UserId string `json:"-" validate:"required"`
    ID     string `json:"-" validate:"required,max=100,uuid"`

    FieldName string `json:"field_name" validate:"required,max=100"`
}

// Get/Delete request — semua dari param/context
type Get{Name}Request struct {
    UserId string `json:"-" validate:"required"`
    ID     string `json:"-" validate:"required,max=100,uuid"`
}

type Delete{Name}Request struct {
    UserId string `json:"-" validate:"required"`
    ID     string `json:"-" validate:"required,max=100,uuid"`
}

// Search/List request
type Search{Name}Request struct {
    UserId string `json:"-" validate:"required"`
    Name   string `json:"name" validate:"max=100"`
    Page   int    `json:"page" validate:"min=1"`
    Size   int    `json:"size" validate:"min=1,max=100"`
}
```

## Rules

1. **Package**: `internal/model`
2. **File naming**: `{name}_model.go` (contoh: `contact_model.go`)
3. **Response struct**: `{Name}Response` — field yang dikirim ke client
4. **Request struct**: `{Action}{Name}Request` — field yang diterima dari client
5. **JSON tag**: selalu `snake_case` — `json:"field_name"`
6. **Validate tag**: gunakan `go-playground/validator` tags
7. **Field dari auth/param**: gunakan `json:"-"` (tidak di-parse dari body)
8. **omitempty**: gunakan pada response field yang opsional

## Generic Response Wrapper

Semua response di-wrap dalam `WebResponse[T]`:

```go
// Sudah ada di model.go — JANGAN buat ulang
type WebResponse[T any] struct {
    Data   T             `json:"data"`
    Paging *PageMetadata `json:"paging,omitempty"`
    Errors string        `json:"errors,omitempty"`
}

type PageResponse[T any] struct {
    Data         []T          `json:"data,omitempty"`
    PageMetadata PageMetadata `json:"paging,omitempty"`
}

type PageMetadata struct {
    Page      int   `json:"page"`
    Size      int   `json:"size"`
    TotalItem int64 `json:"total_item"`
    TotalPage int64 `json:"total_page"`
}
```

## Auth Model

```go
// Sudah ada di auth.go — JANGAN buat ulang
type Auth struct {
    ID string
}
```

## Validation Tag Cheat Sheet

| Tag         | Keterangan                  |
| ----------- | --------------------------- |
| `required`  | Wajib diisi                 |
| `max=N`     | Maksimum N karakter         |
| `min=N`     | Minimum N karakter/angka    |
| `email`     | Format email valid          |
| `uuid`      | Format UUID valid           |
| `omitempty` | Skip validation jika kosong |
