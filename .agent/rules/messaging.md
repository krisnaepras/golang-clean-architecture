# Messaging Layer Rules (Kafka Producer & Consumer)

## Event Model Pattern

```go
package model

type {Name}Event struct {
    ID        string `json:"id,omitempty"`
    // ... field yang relevan
    CreatedAt int64  `json:"created_at,omitempty"`
    UpdatedAt int64  `json:"updated_at,omitempty"`
}

// WAJIB implement Event interface
func (e *{Name}Event) GetId() string {
    return e.ID
}
```

Event interface (sudah ada di `model/event.go`):

```go
type Event interface {
    GetId() string
}
```

## Producer Pattern

### Generic Base (sudah ada di `gateway/messaging/producer.go`)

```go
type Producer[T model.Event] struct {
    Producer sarama.SyncProducer
    Topic    string
    Log      *logrus.Logger
}

// Method bawaan: Send(event T) error, GetTopic() *string
```

### Specific Producer

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
            Topic:    "{topic_name_plural}",  // e.g. "users", "contacts"
            Log:      log,
        },
    }
}
```

## Consumer Pattern

### Specific Consumer

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
    return &{Name}Consumer{
        Log: log,
    }
}

func (c {Name}Consumer) Consume(message *sarama.ConsumerMessage) error {
    event := new(model.{Name}Event)
    if err := json.Unmarshal(message.Value, event); err != nil {
        c.Log.WithError(err).Error("error unmarshalling {name} event")
        return err
    }

    // TODO: process event
    c.Log.Infof("Received topic {topic} with event: %v from partition %d", event, message.Partition)
    return nil
}
```

### Consumer Registration (di `cmd/worker/main.go`)

```go
func Run{Name}Consumer(logger *logrus.Logger, viperConfig *viper.Viper, ctx context.Context) {
    logger.Info("setup {name} consumer")
    consumerGroup := config.NewKafkaConsumerGroup(viperConfig, logger)
    handler := messaging.New{Name}Consumer(logger)
    messaging.ConsumeTopic(ctx, consumerGroup, "{topic}", logger, handler.Consume)
}
```

Panggil consumer di main():

```go
go Run{Name}Consumer(logger, viperConfig, ctx)
```

## Converter Pattern (Entity → Event)

```go
// Di internal/model/converter/{name}_converter.go
func {Name}ToEvent(entity *entity.{Name}) *model.{Name}Event {
    return &model.{Name}Event{
        ID:        entity.ID,
        // ... map fields
        CreatedAt: entity.CreatedAt,
        UpdatedAt: entity.UpdatedAt,
    }
}
```

## Rules

1. **Event file**: `internal/model/{name}_event.go`
2. **Producer file**: `internal/gateway/messaging/{name}_producer.go`
3. **Consumer file**: `internal/delivery/messaging/{name}_consumer.go`
4. **Topic naming**: plural lowercase — `"users"`, `"contacts"`, `"addresses"`
5. **Event interface**: WAJIB implement `GetId() string` method
6. **Producer optional**: selalu cek `if producer != nil` sebelum Send (di usecase)
7. **Consumer goroutine**: consumer dijalankan sebagai goroutine terpisah di worker
8. **Message key**: gunakan `event.GetId()` sebagai message key (untuk partitioning)
9. **JSON encoding**: event di-serialize ke JSON untuk transport
