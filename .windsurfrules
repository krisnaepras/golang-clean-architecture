You are working on a Go Clean Architecture project. Before writing any code, read the architecture rules and patterns documented in the `.agent/rules/` directory.

## Architecture

This project follows Clean Architecture with these layers:

- **Entity** (`internal/entity/`) — GORM database models
- **Model** (`internal/model/`) — Request/Response DTOs, Events, Converters
- **Repository** (`internal/repository/`) — Data access with generic base Repository[T]
- **UseCase** (`internal/usecase/`) — Business logic with per-method transactions
- **Controller** (`internal/delivery/http/`) — Fiber HTTP handlers
- **Messaging** (`internal/gateway/messaging/` + `internal/delivery/messaging/`) — Kafka producers & consumers
- **Config** (`internal/config/`) — DI bootstrap & configuration

## Key Patterns

1. Every usecase method: `tx.Begin()` → validate → business logic → commit → publish event → return
2. Repository receives `*gorm.DB` as param (for transaction support from usecase)
3. Generic base: `Repository[T any]` and `Producer[T model.Event]`
4. All responses wrapped in `model.WebResponse[T]{Data: ...}`
5. Kafka is optional — always check `if producer != nil`
6. Auth via `middleware.GetUser(ctx)` returning `*model.Auth`
7. UUID for entity IDs: `uuid.New().String()`
8. Timestamps as epoch milliseconds (int64)

## Detailed Rules

Read `.agent/rules/` for complete patterns:

- architecture.md, entity.md, model.md, repository.md, usecase.md
- controller.md, messaging.md, config.md, migration.md, testing.md, converter.md

## Skills (Guides)

Read `.agent/skills/` for step-by-step guides:

- add-entity.md, add-crud-endpoint.md, add-kafka-event.md, add-migration.md
