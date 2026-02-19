# Skill: Add New Entity

## Trigger

Gunakan skill ini ketika diminta menambah entity/tabel/domain baru ke project.

## Prerequisites

- Nama entity (singular, PascalCase) — contoh: `Product`
- Fields yang dibutuhkan (nama, tipe, constraint)
- Relasi ke entity lain (jika ada)

## Steps

### Step 1: Create Migration Files

Buat 2 file migration (up + down):

**File**: `db/migrations/{timestamp}_create_table_{table_name}.up.sql`

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE {table_name}
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- kolom sesuai kebutuhan
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

**File**: `db/migrations/{timestamp}_create_table_{table_name}.down.sql`

```sql
DROP TABLE IF EXISTS {table_name};
DROP EXTENSION IF EXISTS pgcrypto;
```

### Step 2: Create Entity

**File**: `internal/entity/{name}_entity.go`

```go
package entity

import "time"

type {Name} struct {
    ID        string     `gorm:"column:id;primaryKey"`
    // ... fields sesuai migration
    CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (e *{Name}) TableName() string {
    return "{table_name}"
}
```

### Step 3: Create Models

**File**: `internal/model/{name}_model.go`

Buat struct untuk:

- `{Name}Response` — field untuk response ke client
- `Create{Name}Request` — field dari request body + auth context
- `Update{Name}Request` — field dari body + URL param + auth context
- `Get{Name}Request` — ID dari param + auth context
- `Delete{Name}Request` — ID dari param + auth context
- `Search{Name}Request` (optional) — filter + pagination

### Step 4: Create Event Model

**File**: `internal/model/{name}_event.go`

```go
package model

type {Name}Event struct {
    ID        string `json:"id,omitempty"`
    // ... relevant fields
    CreatedAt int64  `json:"created_at,omitempty"`
    UpdatedAt int64  `json:"updated_at,omitempty"`
}

func (e *{Name}Event) GetId() string {
    return e.ID
}
```

### Step 5: Create Converter

**File**: `internal/model/converter/{name}_converter.go`

Buat fungsi:

- `{Name}ToResponse(entity) *model.{Name}Response`
- `{Name}ToEvent(entity) *model.{Name}Event`

## Checklist

- [ ] Migration files (up + down, PostgreSQL)
- [ ] Entity struct dengan TableName()
- [ ] Response model
- [ ] Request models (Create, Update, Get, Delete)
- [ ] Event model dengan GetId()
- [ ] Converter functions

## Referensi Rules

- [entity.md](../rules/entity.md)
- [model.md](../rules/model.md)
- [migration.md](../rules/migration.md)
- [converter.md](../rules/converter.md)
