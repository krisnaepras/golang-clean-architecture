# Architecture Rules — Go Clean Architecture

## Overview

Project ini menggunakan **Clean Architecture** pattern dengan Go (Golang). Arsitektur dipisahkan ke dalam layer yang jelas dengan dependency rule: layer dalam TIDAK BOLEH tahu tentang layer luar.

## Dependency Flow

```
HTTP Request → Controller → UseCase → Repository → Database
                  ↓
              Converter
                  ↓
          Producer (Kafka)
```

## Folder Structure

```
cmd/
├── web/main.go          # Entry point HTTP server (Fiber)
├── worker/main.go       # Entry point Kafka consumer worker

internal/
├── config/              # Dependency injection & configuration
│   ├── app.go           # Bootstrap — wiring semua dependency
│   ├── fiber.go         # Fiber HTTP framework setup
│   ├── gorm.go          # GORM database connection
│   ├── kafka.go         # Kafka producer & consumer group setup
│   ├── logrus.go        # Logger setup
│   ├── validator.go     # Validator setup
│   └── viper.go         # Config file & env loader
│
├── entity/              # Domain entities (GORM models)
│   └── {name}_entity.go
│
├── model/               # Request/Response DTOs & Events
│   ├── {name}_model.go  # Request & Response structs
│   ├── {name}_event.go  # Kafka event structs
│   ├── model.go         # WebResponse, PageResponse generics
│   ├── event.go         # Event interface
│   ├── auth.go          # Auth model
│   └── converter/       # Entity ↔ Model converters
│       └── {name}_converter.go
│
├── repository/          # Data access layer
│   ├── repository.go    # Generic Repository[T any] base struct
│   └── {name}_repository.go
│
├── usecase/             # Business logic layer
│   └── {name}_usecase.go
│
├── delivery/            # Presentation/transport layer
│   ├── http/
│   │   ├── {name}_controller.go
│   │   ├── middleware/
│   │   │   └── auth_middleware.go
│   │   └── route/
│   │       └── route.go
│   └── messaging/
│       ├── consumer.go           # Generic consumer handler
│       └── {name}_consumer.go
│
└── gateway/             # External service gateway
    └── messaging/
        ├── producer.go           # Generic Producer[T Event] base struct
        └── {name}_producer.go

db/
└── migrations/          # SQL migration files
    ├── {timestamp}_{description}.up.sql
    └── {timestamp}_{description}.down.sql

test/                    # Integration tests
    ├── init.go          # Test bootstrap (shared across all test files)
    ├── helper_test.go   # Test helper functions
    └── {name}_test.go
```

## Key Principles

1. **Tidak ada interface** — project ini menggunakan concrete struct dependency injection, bukan interface
2. **Generic base struct** — Repository dan Producer menggunakan Go generics sebagai base
3. **Transaction per usecase method** — setiap method di usecase dimulai dengan `tx.Begin()` dan di-defer `tx.Rollback()`
4. **Fiber error types** — gunakan `fiber.ErrBadRequest`, `fiber.ErrNotFound`, dll. untuk error HTTP
5. **Validator struct tags** — validasi menggunakan `go-playground/validator/v10` via struct tags
6. **Viper + .env config** — config.json sebagai base, .env sebagai override (env var takes priority)
7. **Kafka optional** — Kafka producer bisa nil, selalu cek `if producer != nil` sebelum Send
8. **GORM timestamps** — gunakan `time.Time` dengan tag `autoCreateTime` dan `autoUpdateTime` (bukan epoch milli)
    - DSN PostgreSQL: jika password kosong, **jangan sertakan** `password=` di DSN (pgx misparse `dbname` sebagai bagian password)
9. **UUID** — gunakan `github.com/google/uuid` untuk generate ID
10. **Logrus** — semua logging menggunakan `*logrus.Logger`

## Tech Stack

| Component     | Library                       |
| ------------- | ----------------------------- |
| HTTP          | `github.com/gofiber/fiber/v2` |
| ORM           | `gorm.io/gorm`                |
| DB (Postgres) | `gorm.io/driver/postgres`     |
| Validation    | `go-playground/validator/v10` |
| Config        | `github.com/spf13/viper`      |
| Env loader    | `github.com/joho/godotenv`    |
| Logging       | `github.com/sirupsen/logrus`  |
| Kafka         | `github.com/IBM/sarama`       |
| UUID          | `github.com/google/uuid`      |
| Password hash | `golang.org/x/crypto/bcrypt`  |
| Testing       | `github.com/stretchr/testify` |

## Naming Convention

- File: `snake_case` — `user_usecase.go`, `contact_entity.go`
- Package: singular lowercase — `entity`, `model`, `usecase`, `repository`
- Struct: PascalCase — `UserUseCase`, `ContactRepository`
- Constructor: `New{StructName}` — `NewUserUseCase()`, `NewContactRepository()`
- Table method: `TableName() string` pada setiap entity

## Agent Behavior Rules

- **TIDAK BOLEH** melakukan `git commit` atau `git push` secara otomatis
- Commit hanya boleh dilakukan jika user **secara eksplisit** mengizinkan di chat (contoh: "boleh langsung commit", "commit sekarang", "silakan commit")
- Setelah implementasi selesai, cukup informasikan bahwa kode sudah siap dan minta konfirmasi sebelum commit