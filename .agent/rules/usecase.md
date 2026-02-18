# UseCase Layer Rules

UseCase berisi business logic utama aplikasi. Setiap usecase method mengelola transaction database sendiri.

## Pattern

```go
package usecase

import (
    "context"
    "golang-clean-architecture/internal/entity"
    "golang-clean-architecture/internal/gateway/messaging"
    "golang-clean-architecture/internal/model"
    "golang-clean-architecture/internal/model/converter"
    "golang-clean-architecture/internal/repository"

    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
    "gorm.io/gorm"
)

type {Name}UseCase struct {
    DB               *gorm.DB
    Log              *logrus.Logger
    Validate         *validator.Validate
    {Name}Repository *repository.{Name}Repository
    {Name}Producer   *messaging.{Name}Producer  // bisa nil jika Kafka disabled
}

func New{Name}UseCase(
    db *gorm.DB,
    logger *logrus.Logger,
    validate *validator.Validate,
    repo *repository.{Name}Repository,
    producer *messaging.{Name}Producer,
) *{Name}UseCase {
    return &{Name}UseCase{
        DB:               db,
        Log:              logger,
        Validate:         validate,
        {Name}Repository: repo,
        {Name}Producer:   producer,
    }
}
```

## Method Pattern — Create

```go
func (c *{Name}UseCase) Create(ctx context.Context, request *model.Create{Name}Request) (*model.{Name}Response, error) {
    // 1. Begin transaction
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    // 2. Validate request
    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request body : %+v", err)
        return nil, fiber.ErrBadRequest
    }

    // 3. Business logic (check duplicates, etc.)
    // ...

    // 4. Create entity
    entity := &entity.{Name}{
        ID: uuid.New().String(),
        // ... map from request
    }

    if err := c.{Name}Repository.Create(tx, entity); err != nil {
        c.Log.Warnf("Failed create {name} to database : %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    // 5. Commit
    if err := tx.Commit().Error; err != nil {
        c.Log.Warnf("Failed commit transaction : %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    // 6. Publish event (optional, after commit)
    if c.{Name}Producer != nil {
        event := converter.{Name}ToEvent(entity)
        if err := c.{Name}Producer.Send(event); err != nil {
            c.Log.Warnf("Failed publish {name} event : %+v", err)
            return nil, fiber.ErrInternalServerError
        }
    }

    // 7. Return response
    return converter.{Name}ToResponse(entity), nil
}
```

## Method Pattern — Get

```go
func (c *{Name}UseCase) Get(ctx context.Context, request *model.Get{Name}Request) (*model.{Name}Response, error) {
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request body : %+v", err)
        return nil, fiber.ErrBadRequest
    }

    entity := new(entity.{Name})
    if err := c.{Name}Repository.FindById(tx, entity, request.ID); err != nil {
        c.Log.Warnf("Failed find {name} by id : %+v", err)
        return nil, fiber.ErrNotFound
    }

    if err := tx.Commit().Error; err != nil {
        c.Log.Warnf("Failed commit transaction : %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    return converter.{Name}ToResponse(entity), nil
}
```

## Method Pattern — Update

```go
func (c *{Name}UseCase) Update(ctx context.Context, request *model.Update{Name}Request) (*model.{Name}Response, error) {
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request body : %+v", err)
        return nil, fiber.ErrBadRequest
    }

    entity := new(entity.{Name})
    if err := c.{Name}Repository.FindById(tx, entity, request.ID); err != nil {
        c.Log.Warnf("Failed find {name} by id : %+v", err)
        return nil, fiber.ErrNotFound
    }

    // Update fields
    entity.FieldName = request.FieldName

    if err := c.{Name}Repository.Update(tx, entity); err != nil {
        c.Log.Warnf("Failed save {name} : %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    if err := tx.Commit().Error; err != nil {
        c.Log.Warnf("Failed commit transaction : %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    // Publish event (optional)
    if c.{Name}Producer != nil {
        event := converter.{Name}ToEvent(entity)
        if err := c.{Name}Producer.Send(event); err != nil {
            c.Log.Warnf("Failed publish {name} event : %+v", err)
            return nil, fiber.ErrInternalServerError
        }
    }

    return converter.{Name}ToResponse(entity), nil
}
```

## Method Pattern — Delete

```go
func (c *{Name}UseCase) Delete(ctx context.Context, request *model.Delete{Name}Request) error {
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request body : %+v", err)
        return fiber.ErrBadRequest
    }

    entity := new(entity.{Name})
    if err := c.{Name}Repository.FindById(tx, entity, request.ID); err != nil {
        c.Log.Warnf("Failed find {name} by id : %+v", err)
        return fiber.ErrNotFound
    }

    if err := c.{Name}Repository.Delete(tx, entity); err != nil {
        c.Log.Warnf("Failed delete {name} : %+v", err)
        return fiber.ErrInternalServerError
    }

    if err := tx.Commit().Error; err != nil {
        c.Log.Warnf("Failed commit transaction : %+v", err)
        return fiber.ErrInternalServerError
    }

    return nil
}
```

## Rules

1. **Package**: `internal/usecase`
2. **File naming**: `{name}_usecase.go`
3. **Struct naming**: `{Name}UseCase`
4. **Dependencies**: inject via constructor (DB, Log, Validate, Repository, Producer)
5. **Transaction**: SETIAP method dimulai `tx := c.DB.WithContext(ctx).Begin()` + `defer tx.Rollback()`
6. **Validation**: SELALU validate request pertama kali dengan `c.Validate.Struct(request)`
7. **Error handling**: return `fiber.Err*` (ErrBadRequest, ErrNotFound, ErrInternalServerError, ErrConflict, ErrUnauthorized)
8. **Logging**: gunakan `c.Log.Warnf()` untuk warning, `c.Log.Info()` untuk info
9. **Kafka event**: publish SETELAH commit berhasil, SELALU cek `if c.Producer != nil`
10. **Converter**: gunakan fungsi dari `converter` package untuk entity ↔ model conversion
11. **UUID**: gunakan `uuid.New().String()` untuk generate ID baru
