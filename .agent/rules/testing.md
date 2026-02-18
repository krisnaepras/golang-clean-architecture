# Testing Rules

Project ini menggunakan integration test dengan Fiber HTTP test dan testify assertions.

## Test Setup

Test bootstrap ada di `test/init.go` — shared across all test files:

```go
package test

var app *fiber.App
var db *gorm.DB
var viperConfig *viper.Viper
var log *logrus.Logger
var validate *validator.Validate

func init() {
    viperConfig = config.NewViper()
    log = config.NewLogger(viperConfig)
    validate = config.NewValidator(viperConfig)
    app = config.NewFiber(viperConfig)
    db = config.NewDatabase(viperConfig, log)
    producer := config.NewKafkaProducer(viperConfig, log)

    config.Bootstrap(&config.BootstrapConfig{...})
}
```

## Helper Functions Pattern

```go
// Clear data — urut dari child ke parent
func ClearAll() {
    Clear{Children}()
    Clear{Parents}()
}

func Clear{Name}() {
    err := db.Where("id is not null").Delete(&entity.{Name}{}).Error
    if err != nil {
        log.Fatalf("Failed clear {name} data : %+v", err)
    }
}

// Create test data
func Create{Name}(parent *entity.Parent, total int) {
    for i := 0; i < total; i++ {
        entity := &entity.{Name}{
            ID: uuid.NewString(),
            // ... field
        }
        err := db.Create(entity).Error
        if err != nil {
            log.Fatalf("Failed create {name} data : %+v", err)
        }
    }
}

// Get first record for assertions
func GetFirst{Name}(t *testing.T) *entity.{Name} {
    entity := new(entity.{Name})
    err := db.First(entity).Error
    assert.Nil(t, err)
    return entity
}
```

## Test Function Pattern

```go
func Test{Action}{Name}(t *testing.T) {
    // Setup
    ClearAll()
    // Create prerequisite data...

    // Build request
    requestBody := model.Create{Name}Request{
        // ... fields
    }
    bodyJson, _ := json.Marshal(requestBody)

    // Create HTTP request
    request := httptest.NewRequest(http.MethodPost, "/api/{resources}", bytes.NewBuffer(bodyJson))
    request.Header.Set("Content-Type", "application/json")
    request.Header.Set("Authorization", user.Token)

    // Execute
    response, err := app.Test(request)
    assert.Nil(t, err)

    // Read response body
    bytes, err := io.ReadAll(response.Body)
    assert.Nil(t, err)

    // Parse response
    responseBody := new(model.WebResponse[model.{Name}Response])
    err = json.Unmarshal(bytes, responseBody)
    assert.Nil(t, err)

    // Assert
    assert.Equal(t, http.StatusOK, response.StatusCode)
    assert.NotEmpty(t, responseBody.Data.ID)
}
```

## Rules

1. **Package**: `test` (folder `test/`)
2. **File naming**: `{name}_test.go`
3. **Test function**: `Test{Action}{Name}` — contoh: `TestCreateContact`, `TestLoginUser`
4. **Setup**: SELALU mulai dengan `ClearAll()` untuk clean state
5. **HTTP testing**: gunakan `app.Test(request)` dari Fiber
6. **Assertions**: gunakan `testify/assert`
7. **Auth**: set `Authorization` header dengan token user
8. **Content-Type**: selalu `application/json`
9. **Helper functions**: buat di `helper_test.go`
10. **Init**: JANGAN ubah `init.go` kecuali menambah dependency baru
11. **Clear order**: child dulu, baru parent (addresses → contacts → users)
