# Converter Rules

Converter berisi fungsi untuk konversi antara Entity ↔ Response Model dan Entity → Event.

## File Location

`internal/model/converter/{name}_converter.go`

## Pattern

```go
package converter

import (
    "golang-clean-architecture/internal/entity"
    "golang-clean-architecture/internal/model"
)

// Entity → Response
func {Name}ToResponse(e *entity.{Name}) *model.{Name}Response {
    return &model.{Name}Response{
        ID:        e.ID,
        // ... map semua field yang diperlukan client
        CreatedAt: e.CreatedAt,
        UpdatedAt: e.UpdatedAt,
    }
}

// Entity → Event (untuk Kafka)
func {Name}ToEvent(e *entity.{Name}) *model.{Name}Event {
    return &model.{Name}Event{
        ID:        e.ID,
        // ... map field yang perlu di-publish
        CreatedAt: e.CreatedAt,
        UpdatedAt: e.UpdatedAt,
    }
}

// Optional: Entity → Response with special fields
func {Name}To{Variant}Response(e *entity.{Name}) *model.{Name}Response {
    return &model.{Name}Response{
        // ... only specific fields (e.g., token only)
    }
}
```

## Rules

1. **Package**: `converter` (di bawah `internal/model/converter/`)
2. **File naming**: `{name}_converter.go`
3. **Function naming**: `{EntityName}To{Target}` — contoh: `UserToResponse`, `ContactToEvent`
4. **Input**: selalu pointer to entity `*entity.{Name}`
5. **Output**: selalu pointer to model `*model.{Name}Response` atau `*model.{Name}Event`
6. **Tidak ada business logic** — hanya mapping field
7. **Field mapping**: explicit — tulis setiap field, jangan pakai reflection
8. **Password/sensitive**: JANGAN map field sensitif (password, token) ke response biasa
9. **Nested conversion**: untuk child entities, panggil converter child secara terpisah
