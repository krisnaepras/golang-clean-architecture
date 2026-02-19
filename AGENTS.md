# AGENTS.md — Agent Instructions

> File ini dibaca oleh AI coding agents (GitHub Copilot, Claude Code, Antigravity, dll.)
> untuk memahami arsitektur dan pattern project ini.

## Project Overview

Go Clean Architecture project menggunakan Fiber (HTTP), GORM (ORM), Sarama (Kafka), Viper (Config), dan Logrus (Logging).

## Architecture Rules

Baca semua file di `.agent/rules/` untuk memahami pattern setiap layer:

- [.agent/rules/architecture.md](.agent/rules/architecture.md) — Arsitektur keseluruhan, folder structure, tech stack
- [.agent/rules/entity.md](.agent/rules/entity.md) — GORM entity/domain model pattern
- [.agent/rules/model.md](.agent/rules/model.md) — Request/Response DTO pattern
- [.agent/rules/repository.md](.agent/rules/repository.md) — Generic repository pattern
- [.agent/rules/usecase.md](.agent/rules/usecase.md) — Business logic & transaction pattern
- [.agent/rules/controller.md](.agent/rules/controller.md) — Fiber HTTP controller pattern
- [.agent/rules/messaging.md](.agent/rules/messaging.md) — Kafka producer/consumer pattern
- [.agent/rules/config.md](.agent/rules/config.md) — Bootstrap & config pattern
- [.agent/rules/migration.md](.agent/rules/migration.md) — Database migration pattern
- [.agent/rules/testing.md](.agent/rules/testing.md) — Integration test pattern
- [.agent/rules/converter.md](.agent/rules/converter.md) — Entity ↔ Model converter pattern

## Skills (Step-by-Step Guides)

- [.agent/skills/add-entity.md](.agent/skills/add-entity.md) — Menambah entity baru
- [.agent/skills/add-crud-endpoint.md](.agent/skills/add-crud-endpoint.md) — Menambah CRUD endpoint lengkap
- [.agent/skills/add-kafka-event.md](.agent/skills/add-kafka-event.md) — Menambah Kafka event
- [.agent/skills/add-migration.md](.agent/skills/add-migration.md) — Menambah database migration
- [.agent/skills/add-api-testing.md](.agent/skills/add-api-testing.md) — Menulis integration test + manual.http setelah membuat endpoint

## Critical Rules Summary

1. **Dependency flow**: Controller → UseCase → Repository → DB (TIDAK boleh terbalik)
2. **Transaction**: Setiap usecase method buat transaction sendiri (`tx.Begin()` + `defer tx.Rollback()`)
3. **Validation**: Selalu validate request pertama di usecase dengan `c.Validate.Struct(request)`
4. **Error handling**: Return `fiber.Err*` types (ErrBadRequest, ErrNotFound, ErrInternalServerError)
5. **Kafka optional**: Selalu cek `if producer != nil` sebelum publish event
6. **Generics**: Repository dan Producer menggunakan Go generics sebagai base
7. **No interfaces**: Dependency injection menggunakan concrete structs
8. **Naming**: file `snake_case`, struct `PascalCase`, constructor `New{Name}`
9. **Response**: Selalu wrap dalam `model.WebResponse[T]{Data: ...}`
10. **UUID**: Gunakan `uuid.New().String()` untuk generate entity ID
