# Skill: Add Kafka Event (Producer & Consumer)

## Trigger

Gunakan skill ini ketika diminta menambah event publishing (Kafka) ke entity yang sudah ada atau entity baru.

## Prerequisites

- Entity sudah ada di `internal/entity/`
- Model Response sudah ada
- Nama topic Kafka (biasanya sama dengan nama tabel, plural)

## Steps

### Step 1: Create Event Model

**File**: `internal/model/{name}_event.go`

```go
package model

type {Name}Event struct {
    ID        string `json:"id,omitempty"`
    // field yang perlu di-publish ke Kafka
    CreatedAt int64  `json:"created_at,omitempty"`
    UpdatedAt int64  `json:"updated_at,omitempty"`
}

func (e *{Name}Event) GetId() string {
    return e.ID
}
```

> **PENTING**: WAJIB implement `GetId() string` agar comply dengan `Event` interface

### Step 2: Add Event Converter

**File**: `internal/model/converter/{name}_converter.go`

Tambahkan function:

```go
func {Name}ToEvent(e *entity.{Name}) *model.{Name}Event {
    return &model.{Name}Event{
        ID:        e.ID,
        // ... map relevant fields
        CreatedAt: e.CreatedAt,
        UpdatedAt: e.UpdatedAt,
    }
}
```

### Step 3: Create Producer

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
            Topic:    "{topic_plural}",
            Log:      log,
        },
    }
}
```

### Step 4: Create Consumer

**File**: `internal/delivery/messaging/{name}_consumer.go`

```go
package messaging

import (
    "encoding/json"
    "golang-clean-architecture/internal/model"

    "github.com/IBM/sarama"
    "github.com/sirupsen/logrus"
)

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

    // TODO: process event (e.g., update cache, sync to another service)
    c.Log.Infof("Received topic {topic} with event: %v from partition %d", event, message.Partition)
    return nil
}
```

### Step 5: Wire Producer di Bootstrap

Update `internal/config/app.go`:

```go
// Di dalam Bootstrap(), setelah config.Producer != nil check:
var {name}Producer *messaging.{Name}Producer
if config.Producer != nil {
    {name}Producer = messaging.New{Name}Producer(config.Producer, config.Log)
}
```

Pass producer ke usecase constructor yang membutuhkan.

### Step 6: Add Event Publishing di UseCase

Di method Create/Update/Delete di usecase, tambahkan SETELAH commit:

```go
// Publish event (setelah tx.Commit)
if c.{Name}Producer != nil {
    event := converter.{Name}ToEvent(entity)
    c.Log.Info("Publishing {name} created event")
    if err := c.{Name}Producer.Send(event); err != nil {
        c.Log.Warnf("Failed publish {name} event : %+v", err)
        return nil, fiber.ErrInternalServerError
    }
} else {
    c.Log.Info("Kafka producer is disabled, skipping {name} event")
}
```

### Step 7: Register Consumer di Worker

Update `cmd/worker/main.go`:

```go
// Di main():
go Run{Name}Consumer(logger, viperConfig, ctx)

// Function baru:
func Run{Name}Consumer(logger *logrus.Logger, viperConfig *viper.Viper, ctx context.Context) {
    logger.Info("setup {name} consumer")
    consumerGroup := config.NewKafkaConsumerGroup(viperConfig, logger)
    handler := messaging.New{Name}Consumer(logger)
    messaging.ConsumeTopic(ctx, consumerGroup, "{topic}", logger, handler.Consume)
}
```

## Checklist

- [ ] Event model (`{name}_event.go`) dengan `GetId()`
- [ ] Event converter (`{Name}ToEvent`)
- [ ] Producer (`{name}_producer.go`)
- [ ] Consumer (`{name}_consumer.go`)
- [ ] Bootstrap wiring (producer di `app.go`)
- [ ] UseCase event publishing (setelah commit)
- [ ] Worker registration (`cmd/worker/main.go`)

## Referensi Rules

- [messaging.md](../rules/messaging.md)
- [converter.md](../rules/converter.md)
- [config.md](../rules/config.md)
