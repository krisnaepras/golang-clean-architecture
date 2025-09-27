# Analisis Architecture Project Golang Clean Architecture

## Overview Project

Project ini adalah implementasi **Clean Architecture** untuk REST API menggunakan Golang. Project sudah memiliki 3 domain utama:
1. **User Management** - registrasi, login, logout, update profile
2. **Contact Management** - CRUD contacts per user
3. **Address Management** - CRUD addresses per contact

## Struktur Architecture yang Digunakan

### 1. Domain Driven Design (DDD)
Project ini menerapkan DDD dengan memisahkan:
- **Domain Models** (`internal/entity/`) - User, Contact, Address
- **Business Rules** (`internal/usecase/`) - logic bisnis untuk setiap domain
- **Infrastructure** (`internal/repository/`, `internal/gateway/`) - technical implementation

### 2. Dependency Inversion Principle
- Layer tinggi tidak depend pada layer rendah
- Semua dependencies di-inject melalui constructor
- Use case tidak tahu detail database atau HTTP framework

### 3. Single Responsibility Principle  
- **Controller**: hanya handle HTTP request/response
- **UseCase**: hanya business logic
- **Repository**: hanya database operations
- **Entity**: hanya data structure

## Breakdown Per Layer

### Layer 1: Entity (Domain Models)

**Lokasi**: `internal/entity/`

**User Entity**:
```go
type User struct {
    ID        string    `gorm:"column:id;primaryKey"`
    Password  string    `gorm:"column:password"` 
    Name      string    `gorm:"column:name"`
    Token     string    `gorm:"column:token"`
    CreatedAt int64     `gorm:"column:created_at;autoCreateTime:milli"`
    UpdatedAt int64     `gorm:"column:updated_at;autoUpdateTime:milli"`
    Contacts  []Contact `gorm:"foreignKey:user_id;references:id"`
}
```

**Fungsi**:
- Merepresentasikan tabel database
- Mendefinisikan relasi antar tabel (User -> Contacts -> Addresses)
- Tidak memiliki business logic

### Layer 2: Model (Request/Response)

**Lokasi**: `internal/model/`

**Request Models**:
```go
type LoginUserRequest struct {
    ID       string `json:"id" validate:"required,max=100"`
    Password string `json:"password" validate:"required,max=100"`
}

type CreateContactRequest struct {
    FirstName string `json:"first_name" validate:"required,max=100"`
    LastName  string `json:"last_name" validate:"max=100"`
    Email     string `json:"email" validate:"max=100,email"`
    Phone     string `json:"phone" validate:"max=20"`
}
```

**Response Models**:
```go
type UserResponse struct {
    ID        string `json:"id,omitempty"`
    Name      string `json:"name,omitempty"`
    Token     string `json:"token,omitempty"`
    CreatedAt int64  `json:"created_at,omitempty"`
    UpdatedAt int64  `json:"updated_at,omitempty"`
}
```

**Fungsi**:
- Data Transfer Objects (DTO) untuk API
- Validasi input dengan struct tags
- Pemisahan antara database structure dan API structure

### Layer 3: Repository (Data Access)

**Lokasi**: `internal/repository/`

**Base Repository** (Generic):
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

**Specific Repository** (Custom):
```go
type UserRepository struct {
    Repository[entity.User]
    Log *logrus.Logger
}

func (r *UserRepository) FindByToken(db *gorm.DB, user *entity.User, token string) error {
    return db.Where("token = ?", token).First(user).Error
}
```

**Fungsi**:
- Abstraksi database operations
- Generic CRUD operations
- Custom queries untuk business needs specific

### Layer 4: UseCase (Business Logic)

**Lokasi**: `internal/usecase/`

**Structure**:
```go
type UserUseCase struct {
    DB             *gorm.DB
    Log            *logrus.Logger
    Validate       *validator.Validate
    UserRepository *repository.UserRepository
    UserProducer   *messaging.UserProducer  // Kafka producer
}
```

**Business Logic Example** (Login):
```go
func (c *UserUseCase) Login(ctx context.Context, request *model.LoginUserRequest) (*model.UserResponse, error) {
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    // 1. Validation
    if err := c.Validate.Struct(request); err != nil {
        return nil, fiber.ErrBadRequest
    }

    // 2. Find user
    user := new(entity.User)
    if err := c.UserRepository.FindById(tx, user, request.ID); err != nil {
        return nil, fiber.ErrUnauthorized
    }

    // 3. Check password  
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
        return nil, fiber.ErrUnauthorized
    }

    // 4. Generate new token
    user.Token = uuid.New().String()
    if err := c.UserRepository.Update(tx, user); err != nil {
        return nil, fiber.ErrInternalServerError
    }

    // 5. Publish event (optional)
    if c.UserProducer != nil {
        event := converter.UserToLoginEvent(user)
        c.UserProducer.Send(event)
    }

    // 6. Commit transaction
    if err := tx.Commit().Error; err != nil {
        return nil, fiber.ErrInternalServerError
    }

    return converter.UserToResponse(user), nil
}
```

**Fungsi**:
- Orchestrate business flow
- Transaction management
- Input validation
- Integration dengan external services (Kafka)
- Error handling

### Layer 5: Controller (HTTP Handlers)

**Lokasi**: `internal/delivery/http/`

**Structure**:
```go
type UserController struct {
    Log     *logrus.Logger
    UseCase *usecase.UserUseCase
}
```

**HTTP Handler Example**:
```go
func (c *UserController) Login(ctx *fiber.Ctx) error {
    // 1. Parse request
    request := new(model.LoginUserRequest)
    err := ctx.BodyParser(request)
    if err != nil {
        c.Log.Warnf("Failed to parse request body: %+v", err)
        return fiber.ErrBadRequest
    }

    // 2. Call use case
    response, err := c.UseCase.Login(ctx.UserContext(), request)
    if err != nil {
        c.Log.Warnf("Failed to login user: %+v", err)
        return err
    }

    // 3. Return JSON response
    return ctx.JSON(model.WebResponse[*model.UserResponse]{Data: response})
}
```

**Fungsi**:
- HTTP request/response handling
- Request parsing (JSON, query params, path params)
- Response formatting
- Minimal business logic

### Layer 6: Route Configuration

**Lokasi**: `internal/delivery/http/route/`

**Guest Routes** (No Authentication):
```go
func (c *RouteConfig) SetupGuestRoute() {
    c.App.Post("/api/users", c.UserController.Register)
    c.App.Post("/api/users/_login", c.UserController.Login)
}
```

**Authenticated Routes** (Require Token):
```go
func (c *RouteConfig) SetupAuthRoute() {
    c.App.Use(c.AuthMiddleware)  // Apply auth middleware
    
    // User routes
    c.App.Delete("/api/users", c.UserController.Logout)
    c.App.Patch("/api/users/_current", c.UserController.Update)
    c.App.Get("/api/users/_current", c.UserController.Current)

    // Contact routes  
    c.App.Get("/api/contacts", c.ContactController.List)
    c.App.Post("/api/contacts", c.ContactController.Create)
    c.App.Put("/api/contacts/:contactId", c.ContactController.Update)
    c.App.Get("/api/contacts/:contactId", c.ContactController.Get)
    c.App.Delete("/api/contacts/:contactId", c.ContactController.Delete)

    // Address routes (nested under contacts)
    c.App.Get("/api/contacts/:contactId/addresses", c.AddressController.List)
    c.App.Post("/api/contacts/:contactId/addresses", c.AddressController.Create)
    c.App.Put("/api/contacts/:contactId/addresses/:addressId", c.AddressController.Update)
    c.App.Get("/api/contacts/:contactId/addresses/:addressId", c.AddressController.Get)
    c.App.Delete("/api/contacts/:contactId/addresses/:addressId", c.AddressController.Delete)
}
```

## Authentication & Authorization

### Middleware Authentication

**Lokasi**: `internal/delivery/http/middleware/auth_middleware.go`

```go
func NewAuth(userUserCase *usecase.UserUseCase) fiber.Handler {
    return func(ctx *fiber.Ctx) error {
        // 1. Get token from Authorization header
        request := &model.VerifyUserRequest{Token: ctx.Get("Authorization", "NOT_FOUND")}
        
        // 2. Verify token with use case
        auth, err := userUserCase.Verify(ctx.UserContext(), request)
        if err != nil {
            return fiber.ErrUnauthorized
        }

        // 3. Store user info in context
        ctx.Locals("auth", auth)
        return ctx.Next()
    }
}

func GetUser(ctx *fiber.Ctx) *model.Auth {
    return ctx.Locals("auth").(*model.Auth)
}
```

**Flow Authentication**:
1. Client send request dengan header `Authorization: {token}`
2. Middleware extract token dan verify ke database
3. Jika valid, user info di-store di context
4. Controller bisa akses user info dengan `middleware.GetUser(ctx)`

## Configuration & Dependency Injection

**Lokasi**: `internal/config/app.go`

**Bootstrap Process**:
```go
func Bootstrap(config *BootstrapConfig) {
    // 1. Setup repositories (data layer)
    userRepository := repository.NewUserRepository(config.Log)
    contactRepository := repository.NewContactRepository(config.Log)
    addressRepository := repository.NewAddressRepository(config.Log)

    // 2. Setup producers (messaging layer)
    var userProducer *messaging.UserProducer
    if config.Producer != nil {
        userProducer = messaging.NewUserProducer(config.Producer, config.Log)
    }

    // 3. Setup use cases (business layer)
    userUseCase := usecase.NewUserUseCase(config.DB, config.Log, config.Validate, userRepository, userProducer)
    contactUseCase := usecase.NewContactUseCase(config.DB, config.Log, config.Validate, contactRepository, contactProducer)
    addressUseCase := usecase.NewAddressUseCase(config.DB, config.Log, config.Validate, contactRepository, addressRepository, addressProducer)

    // 4. Setup controllers (presentation layer)
    userController := http.NewUserController(userUseCase, config.Log)
    contactController := http.NewContactController(contactUseCase, config.Log)
    addressController := http.NewAddressController(addressUseCase, config.Log)

    // 5. Setup middleware
    authMiddleware := middleware.NewAuth(userUseCase)

    // 6. Setup routes
    routeConfig := route.RouteConfig{
        App:               config.App,
        UserController:    userController,
        ContactController: contactController,
        AddressController: addressController,
        AuthMiddleware:    authMiddleware,
    }
    routeConfig.Setup()
}
```

**Dependency Flow**: Repository → UseCase → Controller → Route

## External Integration

### Kafka Messaging

**Lokasi**: `internal/gateway/messaging/`

**Purpose**: Publish events ke Kafka untuk integration dengan microservices lain

**User Events**:
```go
type UserProducer struct {
    Producer sarama.SyncProducer
    Log      *logrus.Logger
}

func (p *UserProducer) Send(event *model.UserEvent) error {
    jsonBytes, _ := json.Marshal(event)
    message := &sarama.ProducerMessage{
        Topic: "user-events",
        Key:   sarama.StringEncoder(event.UserID),
        Value: sarama.ByteEncoder(jsonBytes),
    }
    
    _, _, err := p.Producer.SendMessage(message)
    return err
}
```

**Event Model**:
```go
type UserEvent struct {
    EventID   string `json:"event_id"`
    EventType string `json:"event_type"` // "USER_REGISTERED", "USER_LOGIN", etc
    UserID    string `json:"user_id"`
    Timestamp int64  `json:"timestamp"`
    Data      any    `json:"data"`
}
```

## API Response Format

### Standard Response

**Success Response**:
```json
{
  "data": {
    "id": "user123",
    "name": "John Doe",
    "token": "abc123...",
    "created_at": 1623456789000,
    "updated_at": 1623456789000
  }
}
```

**Error Response**:
```json
{
  "errors": "Invalid credentials"
}
```

### Paginated Response

**Search/List Response**:
```json
{
  "data": [
    {"id": "contact1", "name": "John"},
    {"id": "contact2", "name": "Jane"}
  ],
  "paging": {
    "page": 1,
    "size": 10,
    "total_item": 25,
    "total_page": 3
  }
}
```

## Database Design

### User Table
```sql
CREATE TABLE users (
    id VARCHAR(100) PRIMARY KEY,
    password VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL, 
    token VARCHAR(100),
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);
```

### Contact Table  
```sql
CREATE TABLE contacts (
    id VARCHAR(100) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    email VARCHAR(100),
    phone VARCHAR(20),
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### Address Table
```sql  
CREATE TABLE addresses (
    id VARCHAR(100) PRIMARY KEY,
    contact_id VARCHAR(100) NOT NULL,
    street VARCHAR(200),
    city VARCHAR(100),
    province VARCHAR(100), 
    country VARCHAR(100),
    postal_code VARCHAR(10),
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    FOREIGN KEY (contact_id) REFERENCES contacts(id)
);
```

## Testing Strategy

**Lokasi**: `test/`

### Integration Tests
- Test full HTTP flow dari request sampai database
- Menggunakan real database (test database)
- Menggunakan httptest untuk simulate HTTP requests

**Example Test**:
```go
func TestCreateContact(t *testing.T) {
    TestLogin(t)  // Setup authenticated user
    
    user := new(entity.User)
    err := db.Where("id = ?", "khannedy").First(user).Error
    assert.Nil(t, err)

    requestBody := model.CreateContactRequest{
        FirstName: "Eko Kurniawan",
        LastName:  "Khannedy", 
        Email:     "eko@example.com",
        Phone:     "088888888888",
    }
    bodyJson, _ := json.Marshal(requestBody)

    request := httptest.NewRequest(http.MethodPost, "/api/contacts", strings.NewReader(string(bodyJson)))
    request.Header.Set("Content-Type", "application/json")
    request.Header.Set("Authorization", user.Token)

    response, err := app.Test(request)
    assert.Nil(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
}
```

## Keunggulan Architecture Ini

### 1. **Maintainability**
- Kode terorganisir dengan baik per layer
- Mudah locate bugs dan modify functionality
- Clear separation of concerns

### 2. **Testability** 
- Easy to mock dependencies
- Unit test per layer
- Integration test end-to-end

### 3. **Scalability**
- Horizontal scaling dengan multiple instances
- Event-driven architecture dengan Kafka
- Database sharding friendly

### 4. **Flexibility**
- Bisa ganti database (MySQL -> PostgreSQL) tanpa ubah business logic
- Bisa ganti HTTP framework (Fiber -> Gin) tanpa ubah use cases
- Bisa tambah messaging system lain

### 5. **Team Development**
- Different teams bisa work di different layers
- Contract-based development dengan interfaces
- Independent deployment per service

## Best Practices yang Diterapkan

1. **Database Transaction** - Setiap use case method menggunakan transaction
2. **Input Validation** - Menggunakan validator tags dan struct validation  
3. **Error Handling** - Consistent error responses dan logging
4. **Security** - Password hashing dengan bcrypt, token-based authentication
5. **Logging** - Structured logging dengan logrus untuk debugging
6. **Configuration** - External configuration dengan Viper (JSON config)
7. **Generic Repository** - Reusable CRUD operations dengan Go generics

Project ini adalah excellent example dari Clean Architecture implementation di Golang untuk production-ready REST API!