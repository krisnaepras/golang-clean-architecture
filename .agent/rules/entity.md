# Entity Layer Rules

Entity adalah representasi tabel database menggunakan GORM model.

## Pattern

```go
package entity

type {EntityName} struct {
    ID        string    `gorm:"column:id;primaryKey"`
    // ... field lain sesuai kolom tabel
    CreatedAt int64     `gorm:"column:created_at;autoCreateTime:milli"`
    UpdatedAt int64     `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
    // relasi (optional)
    Parent    ParentEntity  `gorm:"foreignKey:parent_id;references:id"`
    Children  []ChildEntity `gorm:"foreignKey:entity_id;references:id"`
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
5. **Timestamps**: selalu ada `CreatedAt` dan `UpdatedAt` sebagai `int64` (epoch milli)
    - `CreatedAt`: tag `gorm:"column:created_at;autoCreateTime:milli"`
    - `UpdatedAt`: tag `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
6. **Column mapping**: selalu explicit `gorm:"column:{snake_case}"` tag
7. **Foreign key**: gunakan `gorm:"foreignKey:{fk_column};references:id"` tag
8. **TableName method**: WAJIB implement `TableName() string` yang return nama tabel (plural, snake_case)
9. **Tidak ada business logic** dalam entity — entity hanya data struct

## Contoh Relasi

### One-to-Many (Parent side)

```go
Contacts []Contact `gorm:"foreignKey:user_id;references:id"`
```

### Many-to-One (Child side)

```go
UserId string `gorm:"column:user_id"`
User   User   `gorm:"foreignKey:user_id;references:id"`
```

## Tipe Data Mapping

| Go Type   | SQL Type          | Keterangan        |
| --------- | ----------------- | ----------------- |
| `string`  | `varchar(N)`      | Primary key, text |
| `int64`   | `bigint`          | Timestamps        |
| `int`     | `int`             | Angka umum        |
| `float64` | `decimal/double`  | Angka desimal     |
| `bool`    | `boolean/tinyint` | Flag              |
