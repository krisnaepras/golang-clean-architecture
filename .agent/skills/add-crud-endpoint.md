# Skill: Add Full CRUD Endpoint

## Trigger

Gunakan skill ini ketika diminta menambah CRUD endpoint lengkap untuk sebuah entity. Skill ini mencakup SEMUA layer dari entity sampai route.

## Prerequisites

- Entity sudah dibuat (atau gunakan skill `add-entity` terlebih dahulu)
- Nama entity (singular, PascalCase)
- URL path yang diinginkan
- Apakah perlu auth middleware

## Steps

### Step 1: Pastikan Entity & Model Sudah Ada

Jika belum, jalankan skill [add-entity.md](add-entity.md) dulu.

### Step 2: Create Repository

**File**: `internal/repository/{name}_repository.go`

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

// Tambahkan custom query methods sesuai kebutuhan
// Contoh: FindByField, Search, dll.
```

### Step 3: Create Producer (Optional — jika pakai Kafka)

**File**: `internal/gateway/messaging/{name}_producer.go`

```go
package messaging

import (
    "golang-clean-architecture/internal/model"
    "github.com/IBM/sarama"
    "github.com/sirupsen/logrus"
)

type {Name}Producer struct {
    Producer[*model.{Name}Event]
}

func New{Name}Producer(producer sarama.SyncProducer, log *logrus.Logger) *{Name}Producer {
    return &{Name}Producer{
        Producer: Producer[*model.{Name}Event]{
            Producer: producer,
            Topic:    "{table_name}",
            Log:      log,
        },
    }
}
```

### Step 4: Create UseCase

**File**: `internal/usecase/{name}_usecase.go`

Implement methods:

- `Create(ctx, request) (*Response, error)` — validate → create entity → commit → publish event → return response
- `Get(ctx, request) (*Response, error)` — validate → find by id → commit → return response
- `Update(ctx, request) (*Response, error)` — validate → find → update fields → save → commit → publish event → return response
- `Delete(ctx, request) (bool, error)` — validate → find → delete → commit → return true
- `List/Search(ctx, request) ([]Response, int64, error)` — validate → search → commit → return responses + total

Lihat [usecase.md](../rules/usecase.md) untuk pattern detail setiap method.

### Step 5: Create Controller

**File**: `internal/delivery/http/{name}_controller.go`

Implement handlers:

- `Create(ctx *fiber.Ctx) error`
- `Get(ctx *fiber.Ctx) error`
- `Update(ctx *fiber.Ctx) error`
- `Delete(ctx *fiber.Ctx) error`
- `List(ctx *fiber.Ctx) error`

Lihat [controller.md](../rules/controller.md) untuk pattern detail setiap handler.

### Step 6: Create Consumer (Optional — jika pakai Kafka)

**File**: `internal/delivery/messaging/{name}_consumer.go`

```go
package messaging

type {Name}Consumer struct {
    Log *logrus.Logger
}

func New{Name}Consumer(log *logrus.Logger) *{Name}Consumer {
    return &{Name}Consumer{Log: log}
}

func (c {Name}Consumer) Consume(message *sarama.ConsumerMessage) error {
    event := new(model.{Name}Event)
    if err := json.Unmarshal(message.Value, event); err != nil {
        c.Log.WithError(err).Error("error unmarshalling {name} event")
        return err
    }
    c.Log.Infof("Received topic {topic} with event: %v", event)
    return nil
}
```

### Step 7: Register di Bootstrap (app.go)

Update `internal/config/app.go`:

```go
// Di dalam func Bootstrap():

// 1. Tambah repository
{name}Repository := repository.New{Name}Repository(config.Log)

// 2. Tambah producer (jika Kafka)
var {name}Producer *messaging.{Name}Producer
if config.Producer != nil {
    {name}Producer = messaging.New{Name}Producer(config.Producer, config.Log)
}

// 3. Tambah usecase
{name}UseCase := usecase.New{Name}UseCase(config.DB, config.Log, config.Validate, {name}Repository, {name}Producer)

// 4. Tambah controller
{name}Controller := http.New{Name}Controller({name}UseCase, config.Log)

// 5. Tambah ke RouteConfig
routeConfig := route.RouteConfig{
    // ... existing
    {Name}Controller: {name}Controller,
}
```

### Step 8: Register Routes

Update `internal/delivery/http/route/route.go`:

1. Tambah field di `RouteConfig`:

```go
{Name}Controller *http.{Name}Controller
```

2. Tambah route di `SetupAuthRoute()` atau `SetupGuestRoute()`:

```go
// CRUD routes
c.App.Get("/api/{resources}", c.{Name}Controller.List)
c.App.Post("/api/{resources}", c.{Name}Controller.Create)
c.App.Get("/api/{resources}/:{name}Id", c.{Name}Controller.Get)
c.App.Put("/api/{resources}/:{name}Id", c.{Name}Controller.Update)
c.App.Delete("/api/{resources}/:{name}Id", c.{Name}Controller.Delete)
```

### Step 9: Register Consumer di Worker (Optional)

Update `cmd/worker/main.go`:

```go
go Run{Name}Consumer(logger, viperConfig, ctx)
```

Dan tambah function:

```go
func Run{Name}Consumer(logger *logrus.Logger, viperConfig *viper.Viper, ctx context.Context) {
    logger.Info("setup {name} consumer")
    consumerGroup := config.NewKafkaConsumerGroup(viperConfig, logger)
    handler := messaging.New{Name}Consumer(logger)
    messaging.ConsumeTopic(ctx, consumerGroup, "{topic}", logger, handler.Consume)
}
```

### Step 10: Create Tests

**File**: `test/{name}_test.go`

Buat test untuk setiap endpoint. Lihat [testing.md](../rules/testing.md) untuk pattern.

Update `test/helper_test.go` dengan:

- `Clear{Name}()` function
- `Create{Name}s()` function
- `GetFirst{Name}()` function
- Update `ClearAll()` untuk include `Clear{Name}()`

## Checklist

- [ ] Entity + Migration (skill: add-entity)
- [ ] Model (Request/Response/Event)
- [ ] Converter
- [ ] Repository
- [ ] Producer (optional)
- [ ] UseCase
- [ ] Controller
- [ ] Consumer (optional)
- [ ] Bootstrap wiring (app.go)
- [ ] Route registration (route.go)
- [ ] Worker registration (main.go) — optional
- [ ] Tests + helpers

## Referensi Rules

- [architecture.md](../rules/architecture.md)
- [repository.md](../rules/repository.md)
- [usecase.md](../rules/usecase.md)
- [controller.md](../rules/controller.md)
- [config.md](../rules/config.md)
- [testing.md](../rules/testing.md)
