# Controller Layer Rules

Controller adalah HTTP handler yang menerima request dari client dan mengembalikan response. Menggunakan Fiber framework.

## Pattern

```go
package http

import (
    "golang-clean-architecture/internal/delivery/http/middleware"
    "golang-clean-architecture/internal/model"
    "golang-clean-architecture/internal/usecase"

    "github.com/gofiber/fiber/v2"
    "github.com/sirupsen/logrus"
)

type {Name}Controller struct {
    Log     *logrus.Logger
    UseCase *usecase.{Name}UseCase
}

func New{Name}Controller(useCase *usecase.{Name}UseCase, logger *logrus.Logger) *{Name}Controller {
    return &{Name}Controller{
        Log:     logger,
        UseCase: useCase,
    }
}
```

## Handler Pattern — Create (dari request body)

```go
func (c *{Name}Controller) Create(ctx *fiber.Ctx) error {
    auth := middleware.GetUser(ctx)

    request := new(model.Create{Name}Request)
    if err := ctx.BodyParser(request); err != nil {
        c.Log.Warnf("Failed to parse request body : %+v", err)
        return fiber.ErrBadRequest
    }

    request.UserId = auth.ID
    response, err := c.UseCase.Create(ctx.UserContext(), request)
    if err != nil {
        c.Log.WithError(err).Warnf("Failed to create {name}")
        return err
    }

    return ctx.JSON(model.WebResponse[*model.{Name}Response]{Data: response})
}
```

## Handler Pattern — Get (dari URL param)

```go
func (c *{Name}Controller) Get(ctx *fiber.Ctx) error {
    auth := middleware.GetUser(ctx)

    request := &model.Get{Name}Request{
        UserId: auth.ID,
        ID:     ctx.Params("{name}Id"),
    }

    response, err := c.UseCase.Get(ctx.UserContext(), request)
    if err != nil {
        c.Log.WithError(err).Warnf("Failed to get {name}")
        return err
    }

    return ctx.JSON(model.WebResponse[*model.{Name}Response]{Data: response})
}
```

## Handler Pattern — Update (body + param)

```go
func (c *{Name}Controller) Update(ctx *fiber.Ctx) error {
    auth := middleware.GetUser(ctx)

    request := new(model.Update{Name}Request)
    if err := ctx.BodyParser(request); err != nil {
        c.Log.Warnf("Failed to parse request body : %+v", err)
        return fiber.ErrBadRequest
    }

    request.UserId = auth.ID
    request.ID = ctx.Params("{name}Id")

    response, err := c.UseCase.Update(ctx.UserContext(), request)
    if err != nil {
        c.Log.WithError(err).Warnf("Failed to update {name}")
        return err
    }

    return ctx.JSON(model.WebResponse[*model.{Name}Response]{Data: response})
}
```

## Handler Pattern — Delete

```go
func (c *{Name}Controller) Delete(ctx *fiber.Ctx) error {
    auth := middleware.GetUser(ctx)

    request := &model.Delete{Name}Request{
        UserId: auth.ID,
        ID:     ctx.Params("{name}Id"),
    }

    err := c.UseCase.Delete(ctx.UserContext(), request)
    if err != nil {
        c.Log.WithError(err).Warnf("Failed to delete {name}")
        return err
    }

    return ctx.JSON(model.WebResponse[bool]{Data: true})
}
```

## Handler Pattern — List/Search (dari query params)

```go
func (c *{Name}Controller) List(ctx *fiber.Ctx) error {
    auth := middleware.GetUser(ctx)

    request := &model.Search{Name}Request{
        UserId: auth.ID,
        Name:   ctx.Query("name", ""),
        Page:   ctx.QueryInt("page", 1),
        Size:   ctx.QueryInt("size", 10),
    }

    responses, total, err := c.UseCase.Search(ctx.UserContext(), request)
    if err != nil {
        c.Log.WithError(err).Warnf("Failed to search {name}")
        return err
    }

    paging := &model.PageMetadata{
        Page:      request.Page,
        Size:      request.Size,
        TotalItem: total,
        TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
    }

    return ctx.JSON(model.WebResponse[[]model.{Name}Response]{
        Data:   responses,
        Paging: paging,
    })
}
```

## Rules

1. **Package**: `internal/delivery/http`
2. **File naming**: `{name}_controller.go`
3. **Struct**: `{Name}Controller` dengan field `Log` dan `UseCase`
4. **Constructor**: `New{Name}Controller(useCase, logger)`
5. **Auth**: gunakan `middleware.GetUser(ctx)` untuk dapatkan user dari context
6. **Body parsing**: `ctx.BodyParser(request)` untuk parse JSON body
7. **URL params**: `ctx.Params("paramName")` untuk path parameter
8. **Query params**: `ctx.Query("name", "")` atau `ctx.QueryInt("page", 1)`
9. **Response**: selalu wrap dalam `model.WebResponse[T]{Data: response}`
10. **Error**: return error langsung — Fiber error handler yang akan format response
11. **Context**: gunakan `ctx.UserContext()` untuk pass ke usecase

## Route Registration

Routes didaftarkan di `internal/delivery/http/route/route.go`:

```go
// Guest routes (tanpa auth)
c.App.Post("/api/{resources}", c.{Name}Controller.Create)

// Auth routes (dengan middleware)
c.App.Use(c.AuthMiddleware)
c.App.Get("/api/{resources}", c.{Name}Controller.List)
c.App.Post("/api/{resources}", c.{Name}Controller.Create)
c.App.Get("/api/{resources}/:{name}Id", c.{Name}Controller.Get)
c.App.Put("/api/{resources}/:{name}Id", c.{Name}Controller.Update)
c.App.Delete("/api/{resources}/:{name}Id", c.{Name}Controller.Delete)
```

## URL Pattern Convention

- Resource plural: `/api/contacts`, `/api/addresses`
- Param camelCase: `:contactId`, `:addressId`
- Nested: `/api/contacts/:contactId/addresses/:addressId`
- Special actions: `/api/users/_login`, `/api/users/_current`
