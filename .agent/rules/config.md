# Config & Bootstrap Rules

## Bootstrap Pattern (Dependency Injection)

File `internal/config/app.go` adalah tempat wiring semua dependency:

```go
func Bootstrap(config *BootstrapConfig) {
    // 1. Setup repositories
    {name}Repository := repository.New{Name}Repository(config.Log)

    // 2. Setup producers (Kafka optional)
    var {name}Producer *messaging.{Name}Producer
    if config.Producer != nil {
        {name}Producer = messaging.New{Name}Producer(config.Producer, config.Log)
    }

    // 3. Setup use cases
    {name}UseCase := usecase.New{Name}UseCase(
        config.DB, config.Log, config.Validate,
        {name}Repository, {name}Producer,
    )

    // 4. Setup controllers
    {name}Controller := http.New{Name}Controller({name}UseCase, config.Log)

    // 5. Setup middleware
    authMiddleware := middleware.NewAuth(userUseCase)

    // 6. Setup routes
    routeConfig := route.RouteConfig{
        App:               config.App,
        {Name}Controller:  {name}Controller,
        AuthMiddleware:    authMiddleware,
    }
    routeConfig.Setup()
}
```

## Rules

1. **Wiring order**: Repository → Producer → UseCase → Controller → Middleware → Route
2. **BootstrapConfig**: tambahkan field baru jika perlu dependency dari luar
3. **Kafka check**: selalu wrap producer creation dengan `if config.Producer != nil`
4. **Tidak ada singleton** — setiap Bootstrap call membuat instance baru

## Config Loading

Konfigurasi menggunakan **Viper** dengan prioritas:

1. Environment variables (.env / system) — **TERTINGGI**
2. config.json — **BASE**

### config.json Structure

```json
{
    "app": { "name": "..." },
    "web": { "prefork": false, "port": 3000 },
    "log": { "level": 6 },
    "database": {
        "driver": "postgres",
        "username": "...",
        "password": "...",
        "host": "localhost",
        "port": 5432,
        "name": "...",
        "pool": { "idle": 10, "max": 100, "lifetime": 300 }
    },
    "kafka": {
        "bootstrap": { "servers": "localhost:9092" },
        "group": { "id": "..." },
        "auto": { "offset": { "reset": "earliest" } },
        "producer": { "enabled": false }
    }
}
```

### .env Variables

```
APP_NAME=...
WEB_PORT=3000
DATABASE_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=...
DB_NAME=...
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
KAFKA_GROUP_ID=...
KAFKA_PRODUCER_ENABLED=false
```

## Menambah Config Baru

1. Tambah key di `config.json`
2. Tambah env binding di `internal/config/viper.go`:
    ```go
    _ = config.BindEnv("new.key", "NEW_KEY")
    ```
3. Dokumentasikan env var di `.env.example`

## Entry Points

### Web Server (`cmd/web/main.go`)

```go
func main() {
    viperConfig := config.NewViper()
    log := config.NewLogger(viperConfig)
    db := config.NewDatabase(viperConfig, log)
    validate := config.NewValidator(viperConfig)
    app := config.NewFiber(viperConfig)
    producer := config.NewKafkaProducer(viperConfig, log)

    config.Bootstrap(&config.BootstrapConfig{...})

    app.Listen(fmt.Sprintf(":%d", webPort))
}
```

### Worker (`cmd/worker/main.go`)

```go
func main() {
    viperConfig := config.NewViper()
    logger := config.NewLogger(viperConfig)

    ctx, cancel := context.WithCancel(context.Background())

    go Run{Name}Consumer(logger, viperConfig, ctx)
    // ... more consumers

    // Graceful shutdown
    signal.Notify(terminateSignals, syscall.SIGINT, syscall.SIGTERM)
    // ...
}
```
