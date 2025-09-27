# Ringkasan Clean Architecture - Golang REST API

## Apa itu Clean Architecture?

Clean Architecture adalah pola desain yang memisahkan kode menjadi beberapa layer dengan tanggung jawab yang jelas. Tujuannya adalah membuat kode yang:
- **Mudah ditest** 
- **Mudah dimaintain**
- **Independen dari framework**
- **Independen dari database**

## Struktur Folder Project

```
internal/
├── entity/          # Domain models (struct untuk database)
├── model/           # Request/Response models untuk API
├── repository/      # Data access layer (database operations)
├── usecase/         # Business logic layer
├── delivery/        # HTTP controllers dan routes
├── config/          # Configuration dan dependency injection
└── gateway/         # External services (Kafka, etc.)
```

## Flow Eksekusi REST API

```
HTTP Request → Controller → UseCase → Repository → Database
                    ↓
HTTP Response ← Controller ← UseCase ← Repository ← Database
```

## Penjelasan Setiap Layer

### 1. Entity Layer (`internal/entity/`)

**Fungsi**: Model data yang merepresentasikan tabel database

**Contoh**: User Entity
```go
type User struct {
    ID        string `gorm:"column:id;primaryKey"`
    Password  string `gorm:"column:password"`
    Name      string `gorm:"column:name"`
    Token     string `gorm:"column:token"`
    CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
    UpdatedAt int64  `gorm:"column:updated_at;autoCreateTime:milli"`
}

func (u *User) TableName() string {
    return "users"
}
```

### 2. Model Layer (`internal/model/`)

**Fungsi**: Struktur data untuk komunikasi API (Request & Response)

**Request Model** - Data yang diterima dari client:
```go
type LoginUserRequest struct {
    ID       string `json:"id" validate:"required,max=100"`
    Password string `json:"password" validate:"required,max=100"`
}
```

**Response Model** - Data yang dikirim ke client:
```go
type UserResponse struct {
    ID        string `json:"id,omitempty"`
    Name      string `json:"name,omitempty"`
    Token     string `json:"token,omitempty"`
    CreatedAt int64  `json:"created_at,omitempty"`
    UpdatedAt int64  `json:"updated_at,omitempty"`
}
```

**Generic Response Wrapper**:
```go
type WebResponse[T any] struct {
    Data   T      `json:"data"`
    Errors string `json:"errors,omitempty"`
}
```

### 3. Repository Layer (`internal/repository/`)

**Fungsi**: Mengakses database - CRUD operations

**Base Repository** (sudah ada):
```go
type Repository[T any] struct {
    DB *gorm.DB
}

func (r *Repository[T]) Create(db *gorm.DB, entity *T) error {
    return db.Create(entity).Error
}

func (r *Repository[T]) FindById(db *gorm.DB, entity *T, id any) error {
    return db.Where("id = ?", id).Take(entity).Error
}
```

**Custom Repository** untuk operasi khusus:
```go
type UserRepository struct {
    Repository[entity.User]
    Log *logrus.Logger
}

func (r *UserRepository) FindByToken(db *gorm.DB, user *entity.User, token string) error {
    return db.Where("token = ?", token).First(user).Error
}
```

### 4. UseCase Layer (`internal/usecase/`)

**Fungsi**: Business logic - validasi, transaction, orchestration

```go
type UserUseCase struct {
    DB             *gorm.DB
    Log            *logrus.Logger
    Validate       *validator.Validate
    UserRepository *repository.UserRepository
}

func (c *UserUseCase) Login(ctx context.Context, request *model.LoginUserRequest) (*model.UserResponse, error) {
    // 1. Start transaction
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    // 2. Validate request
    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request: %+v", err)
        return nil, fiber.ErrBadRequest
    }

    // 3. Business logic
    user := new(entity.User)
    if err := c.UserRepository.FindById(tx, user, request.ID); err != nil {
        return nil, fiber.ErrUnauthorized
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
        return nil, fiber.ErrUnauthorized
    }

    // 4. Update token
    user.Token = uuid.New().String()
    if err := c.UserRepository.Update(tx, user); err != nil {
        return nil, fiber.ErrInternalServerError
    }

    // 5. Commit transaction
    if err := tx.Commit().Error; err != nil {
        return nil, fiber.ErrInternalServerError
    }

    // 6. Convert to response
    return converter.UserToResponse(user), nil
}
```

### 5. Controller Layer (`internal/delivery/http/`)

**Fungsi**: Handle HTTP request/response - parsing, calling usecase

```go
type UserController struct {
    Log     *logrus.Logger
    UseCase *usecase.UserUseCase
}

func (c *UserController) Login(ctx *fiber.Ctx) error {
    // 1. Parse request
    request := new(model.LoginUserRequest)
    if err := ctx.BodyParser(request); err != nil {
        return fiber.ErrBadRequest
    }

    // 2. Call usecase
    response, err := c.UseCase.Login(ctx.UserContext(), request)
    if err != nil {
        return err
    }

    // 3. Return JSON response
    return ctx.JSON(model.WebResponse[*model.UserResponse]{Data: response})
}
```

### 6. Route Configuration (`internal/delivery/http/route/`)

**Fungsi**: Mapping URL endpoints ke controller methods

```go
type RouteConfig struct {
    App            *fiber.App
    UserController *http.UserController
    AuthMiddleware fiber.Handler
}

func (c *RouteConfig) SetupGuestRoute() {
    c.App.Post("/api/users", c.UserController.Register)
    c.App.Post("/api/users/_login", c.UserController.Login)
}

func (c *RouteConfig) SetupAuthRoute() {
    c.App.Use(c.AuthMiddleware)  // Require authentication
    c.App.Get("/api/users/_current", c.UserController.Current)
    c.App.Delete("/api/users", c.UserController.Logout)
}
```

## Dependency Injection (`internal/config/app.go`)

**Fungsi**: Menghubungkan semua layer dengan dependency injection

```go
func Bootstrap(config *BootstrapConfig) {
    // 1. Setup repositories
    userRepository := repository.NewUserRepository(config.Log)
    
    // 2. Setup use cases  
    userUseCase := usecase.NewUserUseCase(config.DB, config.Log, config.Validate, userRepository, nil)
    
    // 3. Setup controllers
    userController := http.NewUserController(userUseCase, config.Log)
    
    // 4. Setup middleware
    authMiddleware := middleware.NewAuth(userUseCase)
    
    // 5. Setup routes
    routeConfig := route.RouteConfig{
        App:            config.App,
        UserController: userController,
        AuthMiddleware: authMiddleware,
    }
    routeConfig.Setup()
}
```

## Keuntungan Clean Architecture

### 1. **Separation of Concerns**
- Setiap layer punya tanggung jawab yang jelas
- Controller hanya handle HTTP
- UseCase hanya business logic  
- Repository hanya database access

### 2. **Testable**
```go
// Easy to mock dependencies
func TestUserLogin(t *testing.T) {
    mockRepo := &MockUserRepository{}
    useCase := NewUserUseCase(db, log, validator, mockRepo, nil)
    
    // Test business logic tanpa database
    result, err := useCase.Login(ctx, request)
    assert.NoError(t, err)
}
```

### 3. **Independent dari Framework**
- Bisa ganti dari Fiber ke Gin tanpa ubah business logic
- Bisa ganti dari MySQL ke PostgreSQL tanpa ubah UseCase

### 4. **Maintainable**
- Code terorganisir dengan baik
- Mudah cari bugs
- Mudah tambah fitur baru

## Tips Implementasi

1. **Mulai dari Entity** - definisikan domain model dulu
2. **Buat Repository** - implementasikan database operations
3. **Implement UseCase** - tulis business logic
4. **Buat Controller** - handle HTTP requests
5. **Setup Routes** - mapping endpoints
6. **Wire Dependencies** - connect semua di config

## Struktur Response API

**Success Response:**
```json
{
  "data": {
    "id": "user123",
    "name": "John Doe",
    "token": "abc123"
  }
}
```

**Error Response:**
```json
{
  "errors": "Invalid credentials"
}
```

**Paginated Response:**
```json
{
  "data": [...],
  "paging": {
    "page": 1,
    "size": 10,
    "total_item": 100,
    "total_page": 10
  }
}
```

Clean Architecture memberikan struktur yang solid untuk membangun REST API yang scalable dan maintainable!