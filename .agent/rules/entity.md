# Entity Layer Rules

Entity adalah representasi tabel database menggunakan GORM model.

## Pattern

```go
package entity

type {EntityName} struct {
    ID        string     `gorm:"column:id;primaryKey"`
    // ... field lain sesuai kolom tabel
    CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
    DeletedAt *time.Time `gorm:"column:deleted_at"`  // optional soft delete
    // relasi (optional)
    Parent    ParentEntity  `gorm:"foreignKey:ParentID;references:ID"`
    Children  []ChildEntity `gorm:"foreignKey:EntityID;references:ID"`
}

func (e *{EntityName}) TableName() string {
    return "{table_name_plural}"
}
```

## Rules

1. **Package**: `internal/entity`
2. **File naming**: `{name}_entity.go` (contoh: `user_entity.go`)
3. **Struct name**: PascalCase singular (contoh: `User`, `Contact`, `Address`)
4. **Primary key**: selalu `string` dengan tag `gorm:"column:id;primaryKey"`
5. **Timestamps**: selalu ada `CreatedAt` dan `UpdatedAt` sebagai `time.Time`
    - `CreatedAt`: tag `gorm:"column:created_at;autoCreateTime"`
    - `UpdatedAt`: tag `gorm:"column:updated_at;autoUpdateTime"`
    - `DeletedAt *time.Time` untuk soft delete (optional)
    - Nullable timestamp field: gunakan `*time.Time`
6. **Column mapping**: selalu explicit `gorm:"column:{snake_case}"` tag
7. **Foreign key**: gunakan `gorm:"foreignKey:{FieldName};references:ID"` tag (gunakan nama field Go, bukan nama kolom)
8. **TableName method**: WAJIB implement `TableName() string` yang return nama tabel (plural, snake_case)
9. **Tidak ada business logic** dalam entity — entity hanya data struct

## Contoh Relasi

### One-to-Many (Parent side)

```go
Contacts []Contact `gorm:"foreignKey:UserID;references:ID"`
```

### Many-to-One (Child side)

```go
UserID string `gorm:"column:user_id"`
User   User   `gorm:"foreignKey:UserID;references:ID"`
```

## Tipe Data Mapping

| Go Type      | SQL Type           | Keterangan                   |
| ------------ | ------------------ | ---------------------------- |
| `string`     | `VARCHAR(N)`       | Primary key, text            |
| `time.Time`  | `TIMESTAMPTZ`      | Timestamps (created/updated) |
| `*time.Time` | `TIMESTAMPTZ`      | Nullable timestamps          |
| `int`        | `INTEGER`          | Angka umum                   |
| `*int`       | `INTEGER`          | Nullable integer             |
| `float64`    | `DOUBLE PRECISION` | Angka desimal                |
| `bool`       | `BOOLEAN`          | Flag                         |
