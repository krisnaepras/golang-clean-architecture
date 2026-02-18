# Repository Layer Rules

Repository bertanggung jawab untuk akses data ke database via GORM.

## Generic Base Repository

```go
// Sudah ada di repository.go — JANGAN buat ulang
package repository

type Repository[T any] struct {
    DB *gorm.DB
}

// Method bawaan: Create, Update, Delete, CountById, FindById
```

## Specific Repository Pattern

```go
package repository

import (
    "golang-clean-architecture/internal/entity"
    "github.com/sirupsen/logrus"
    "gorm.io/gorm"
)

type {Name}Repository struct {
    Repository[entity.{Name}]
    Log *logrus.Logger
}

func New{Name}Repository(log *logrus.Logger) *{Name}Repository {
    return &{Name}Repository{
        Log: log,
    }
}

// Custom query methods
func (r *{Name}Repository) FindBy{Field}(db *gorm.DB, entity *entity.{Name}, value string) error {
    return db.Where("{column} = ?", value).First(entity).Error
}

// Search with pagination
func (r *{Name}Repository) Search(db *gorm.DB, request *model.Search{Name}Request) ([]entity.{Name}, int64, error) {
    var entities []entity.{Name}
    // Build query...
    if err := db.Scopes(/* ... */).Offset(offset).Limit(request.Size).Find(&entities).Error; err != nil {
        return nil, 0, err
    }
    // Count total...
    return entities, total, nil
}
```

## Rules

1. **Package**: `internal/repository`
2. **File naming**: `{name}_repository.go`
3. **Embed generic**: setiap repository WAJIB embed `Repository[entity.{Name}]`
4. **Constructor**: `New{Name}Repository(log *logrus.Logger) *{Name}Repository`
5. **DB parameter**: method menerima `*gorm.DB` sebagai parameter pertama (bukan field struct) — ini memungkinkan transaction dari usecase
6. **Entity pointer**: output entity dikirim sebagai pointer parameter, bukan return value
7. **Error handling**: return `error` langsung, biarkan usecase yang handle
8. **Tidak ada transaction** di repository — transaction di-manage usecase

## Built-in Methods (dari Repository[T])

| Method                             | Fungsi               |
| ---------------------------------- | -------------------- |
| `Create(db, entity) error`         | Insert record        |
| `Update(db, entity) error`         | Update record (Save) |
| `Delete(db, entity) error`         | Delete record        |
| `CountById(db, id) (int64, error)` | Count by ID          |
| `FindById(db, entity, id) error`   | Find by ID           |

## Pattern untuk Custom Queries

```go
// Find with preload (relasi)
func (r *{Name}Repository) FindByIdWithRelation(db *gorm.DB, entity *entity.{Name}, id string) error {
    return db.Preload("Relation").Where("id = ?", id).First(entity).Error
}

// Find by foreign key
func (r *{Name}Repository) FindByParentId(db *gorm.DB, parentId string) ([]entity.{Name}, error) {
    var entities []entity.{Name}
    err := db.Where("parent_id = ?", parentId).Find(&entities).Error
    return entities, err
}
```
