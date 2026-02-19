# Testing Rules

Project ini menggunakan integration test dengan Fiber HTTP test dan testify assertions.

## Test Setup

Test bootstrap ada di `test/init.go` — shared across all test files:

```go
package test

import (
    "os"

    "golang-clean-architecture/internal/config"

    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v2"
    "github.com/sirupsen/logrus"
    "github.com/spf13/viper"
    "gorm.io/gorm"
)

var app *fiber.App
var db *gorm.DB
var viperConfig *viper.Viper
var log *logrus.Logger
var validate *validator.Validate

func init() {
    // Kurangi log noise saat test
    _ = os.Setenv("LOG_LEVEL", "3")

    viperConfig = config.NewViper()
    log = config.NewLogger(viperConfig)
    validate = config.NewValidator(viperConfig)
    app = config.NewFiber(viperConfig)
    db = config.NewDatabase(viperConfig, log)

    config.Bootstrap(&config.BootstrapConfig{
        DB:       db,
        App:      app,
        Log:      log,
        Validate: validate,
        Config:   viperConfig,
    })
}
```

## Helper Functions Pattern

```go
// Clear data — urut dari child ke parent
func cleanupUser(t *testing.T, email string) {
    err := db.Exec("DELETE FROM users WHERE email = ?", email).Error
    if err != nil {
        t.Fatalf("Failed cleanup user: %v", err)
    }
}

// HTTP helpers
func doPost(t *testing.T, path string, body interface{}) *http.Response {
    bodyJson, _ := json.Marshal(body)
    req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(bodyJson))
    req.Header.Set("Content-Type", "application/json")
    resp, err := app.Test(req)
    assert.Nil(t, err)
    return resp
}

func doPostAuth(t *testing.T, path string, token string, body interface{}) *http.Response {
    bodyJson, _ := json.Marshal(body)
    req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(bodyJson))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)
    resp, err := app.Test(req)
    assert.Nil(t, err)
    return resp
}

func readBody(t *testing.T, resp *http.Response) []byte {
    b, err := io.ReadAll(resp.Body)
    assert.Nil(t, err)
    return b
}

func parseTokens(t *testing.T, b []byte) (accessToken, refreshToken string) {
    var responseBody model.WebResponse[model.TokenResponse]
    err := json.Unmarshal(b, &responseBody)
    assert.Nil(t, err)
    return responseBody.Data.AccessToken, responseBody.Data.RefreshToken
}
```

## Test Function Pattern

```go
func TestAuth(t *testing.T) {
    // Capture shared state via closure
    var accessToken, refreshToken string

    t.Run("Register_Success", func(t *testing.T) {
        defer cleanupUser(t, "test@example.com")

        resp := doPost(t, "/api/auth/register", model.RegisterUserRequest{
            Email:    "test@example.com",
            Password: "password123",
        })

        b := readBody(t, resp)
        assert.Equal(t, http.StatusOK, resp.StatusCode)

        var responseBody model.WebResponse[model.UserResponse]
        err := json.Unmarshal(b, &responseBody)
        assert.Nil(t, err)
        assert.NotEmpty(t, responseBody.Data.ID)
    })

    t.Run("Login_Success", func(t *testing.T) {
        resp := doPost(t, "/api/auth/login", model.LoginUserRequest{
            Email:    "test@example.com",
            Password: "password123",
        })

        b := readBody(t, resp)
        assert.Equal(t, http.StatusOK, resp.StatusCode)

        accessToken, refreshToken = parseTokens(t, b)
        assert.NotEmpty(t, accessToken)
        assert.NotEmpty(t, refreshToken)
    })

    t.Run("Logout_Success", func(t *testing.T) {
        // Auth-protected routes use doPostAuth
        resp := doPostAuth(t, "/api/auth/logout", accessToken, nil)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
    })
}
```

## Rules

1. **Package**: `test` (folder `test/`)
2. **File naming**: `{name}_test.go`
3. **Test function**: `Test{Domain}` dengan subtests via `t.Run` — contoh: `TestAuth` berisi `Register_Success`, `Login_Success`
4. **Shared state**: gunakan closure variable (`var accessToken string`) antar subtest dalam satu parent test
5. **Cleanup**: `defer cleanupUser(t, email)` di awal subtest, gunakan `db.Exec("DELETE FROM ... WHERE ... = ?", val)`
6. **HTTP testing**: gunakan `app.Test(request)` dari Fiber
7. **Assertions**: gunakan `testify/assert`
8. **Auth protected routes**: gunakan `doPostAuth(t, path, token, body)` yang set `Authorization: Bearer {token}`
9. **Content-Type**: selalu `application/json`
10. **Helper functions**: `doPost`, `doPostAuth`, `doGet`, `readBody`, `parseTokens` di `helper_test.go` atau inline
11. **Init**: JANGAN ubah `init.go` kecuali menambah dependency baru
