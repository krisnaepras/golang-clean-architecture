# Conventions — Go Clean Architecture

Read all files in `.agent/rules/` for detailed architecture patterns.
Read `.agent/skills/` for step-by-step guides on adding features.

## Quick Reference

- **Entity**: `internal/entity/{name}_entity.go` — GORM model with `TableName()` method
- **Model**: `internal/model/{name}_model.go` — Request/Response DTOs with validator tags
- **Repository**: `internal/repository/{name}_repository.go` — embeds generic `Repository[T]`
- **UseCase**: `internal/usecase/{name}_usecase.go` — tx per method, validate first, fiber errors
- **Controller**: `internal/delivery/http/{name}_controller.go` — Fiber handlers, wrap in `WebResponse[T]`
- **Producer**: `internal/gateway/messaging/{name}_producer.go` — embeds generic `Producer[T Event]`
- **Consumer**: `internal/delivery/messaging/{name}_consumer.go`
- **Converter**: `internal/model/converter/{name}_converter.go`
- **Migration**: `db/migrations/{timestamp}_{description}.up.sql`
- **Config wiring**: `internal/config/app.go`
- **Routes**: `internal/delivery/http/route/route.go`
- **Tests**: `test/{name}_test.go`
