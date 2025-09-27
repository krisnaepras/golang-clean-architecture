# Panduan Clean Architecture dan Cara Membuat REST API

## Pengantar

Proyek ini mengimplementasikan **Clean Architecture** (Arsitektur Bersih) yang dikembangkan oleh Robert C. Martin (Uncle Bob). Clean Architecture adalah pola arsitektur perangkat lunak yang menekankan pada pemisahan tanggung jawab dan ketergantungan yang tepat antar layer.

## Struktur Arsitektur

```
┌─────────────────────────────────────────────────────────────┐
│                    External Systems                         │
│                  (HTTP, gRPC, Database)                     │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                    Delivery Layer                           │
│              (HTTP Controllers, Routes)                     │
│               internal/delivery/http/                       │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                    Use Case Layer                           │
│                (Business Logic)                             │
│                internal/usecase/                            │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                  Repository Layer                           │
│              (Data Access Layer)                            │
│               internal/repository/                          │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                    Entity Layer                             │
│                 (Domain Models)                             │
│                internal/entity/                             │
└─────────────────────────────────────────────────────────────┘
```

## Layer-Layer dalam Clean Architecture

### 1. Entity Layer (`internal/entity/`)
- **Tujuan**: Berisi definisi struktur data domain (model data)
- **Karakteristik**: 
  - Tidak bergantung pada layer lain
  - Berisi business rules tingkat tinggi
  - Representasi database table dalam bentuk struct

**Contoh**: `internal/entity/user_entity.go`
```go
type User struct {
    ID        string    `gorm:"column:id;primaryKey"`
    Password  string    `gorm:"column:password"`
    Name      string    `gorm:"column:name"`
    Token     string    `gorm:"column:token"`
    CreatedAt int64     `gorm:"column:created_at;autoCreateTime:milli"`
    UpdatedAt int64     `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
    Contacts  []Contact `gorm:"foreignKey:user_id;references:id"`
}
```

### 2. Repository Layer (`internal/repository/`)
- **Tujuan**: Mengatur akses data ke database atau sumber data lain
- **Karakteristik**:
  - Implementasi pola Repository Pattern
  - Abstraksi untuk database operations
  - Bekerja dengan Entity objects

**Contoh**: `internal/repository/user_repository.go`
```go
type UserRepository struct {
    Repository[entity.User]
    Log *logrus.Logger
}

func (r *UserRepository) FindByToken(db *gorm.DB, user *entity.User, token string) error {
    return db.Where("token = ?", token).First(user).Error
}
```

### 3. Use Case Layer (`internal/usecase/`)
- **Tujuan**: Berisi business logic aplikasi
- **Karakteristik**:
  - Orchestrates flow data antara entities dan repositories
  - Validasi business rules
  - Transaction management
  - Integration dengan external services

**Contoh**: `internal/usecase/user_usecase.go`
```go
type UserUseCase struct {
    DB             *gorm.DB
    Log            *logrus.Logger
    Validate       *validator.Validate
    UserRepository *repository.UserRepository
    UserProducer   *messaging.UserProducer
}

func (c *UserUseCase) Login(ctx context.Context, request *model.LoginUserRequest) (*model.UserResponse, error) {
    // Validasi request
    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request body: %+v", err)
        return nil, fiber.ErrBadRequest
    }

    // Business logic untuk login
    // ...
}
```

### 4. Delivery Layer (`internal/delivery/http/`)
- **Tujuan**: Menangani input dari external systems (HTTP requests)
- **Karakteristik**:
  - HTTP Controllers
  - Request/Response handling
  - Routing configuration
  - Middleware implementation

**Contoh**: `internal/delivery/http/user_controller.go`
```go
type UserController struct {
    Log     *logrus.Logger
    UseCase *usecase.UserUseCase
}

func (c *UserController) Login(ctx *fiber.Ctx) error {
    request := new(model.LoginUserRequest)
    err := ctx.BodyParser(request)
    if err != nil {
        c.Log.Warnf("Failed to parse request body: %+v", err)
        return fiber.ErrBadRequest
    }

    response, err := c.UseCase.Login(ctx.UserContext(), request)
    if err != nil {
        c.Log.Warnf("Failed to login user: %+v", err)
        return err
    }

    return ctx.JSON(model.WebResponse[*model.UserResponse]{Data: response})
}
```

## Model Layer (`internal/model/`)

Model layer berisi struktur data yang digunakan untuk komunikasi antar layer:

### Request/Response Models
```go
// Request model untuk API
type LoginUserRequest struct {
    ID       string `json:"id" validate:"required,max=100"`
    Password string `json:"password" validate:"required,max=100"`
}

// Response model untuk API
type UserResponse struct {
    ID        string `json:"id,omitempty"`
    Name      string `json:"name,omitempty"`
    Token     string `json:"token,omitempty"`
    CreatedAt int64  `json:"created_at,omitempty"`
    UpdatedAt int64  `json:"updated_at,omitempty"`
}

// Generic web response wrapper
type WebResponse[T any] struct {
    Data   T             `json:"data"`
    Paging *PageMetadata `json:"paging,omitempty"`
    Errors string        `json:"errors,omitempty"`
}
```

## Cara Membuat REST API Baru

Mari kita buat contoh REST API untuk entity `Product`. Ikuti langkah-langkah berikut:

### Langkah 1: Buat Entity

Buat file `internal/entity/product_entity.go`:

```go
package entity

type Product struct {
    ID          string `gorm:"column:id;primaryKey"`
    Name        string `gorm:"column:name"`
    Description string `gorm:"column:description"`
    Price       int64  `gorm:"column:price"`
    Stock       int    `gorm:"column:stock"`
    UserID      string `gorm:"column:user_id"`
    CreatedAt   int64  `gorm:"column:created_at;autoCreateTime:milli"`
    UpdatedAt   int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
    User        User   `gorm:"foreignKey:user_id;references:id"`
}

func (p *Product) TableName() string {
    return "products"
}
```

### Langkah 2: Buat Model Request/Response

Buat file `internal/model/product_model.go`:

```go
package model

type ProductResponse struct {
    ID          string `json:"id,omitempty"`
    Name        string `json:"name,omitempty"`
    Description string `json:"description,omitempty"`
    Price       int64  `json:"price,omitempty"`
    Stock       int    `json:"stock,omitempty"`
    UserID      string `json:"user_id,omitempty"`
    CreatedAt   int64  `json:"created_at,omitempty"`
    UpdatedAt   int64  `json:"updated_at,omitempty"`
}

type CreateProductRequest struct {
    Name        string `json:"name" validate:"required,max=100"`
    Description string `json:"description" validate:"max=500"`
    Price       int64  `json:"price" validate:"required,min=0"`
    Stock       int    `json:"stock" validate:"required,min=0"`
}

type UpdateProductRequest struct {
    ID          string `json:"-" validate:"required"`
    Name        string `json:"name,omitempty" validate:"max=100"`
    Description string `json:"description,omitempty" validate:"max=500"`
    Price       int64  `json:"price,omitempty" validate:"min=0"`
    Stock       int    `json:"stock,omitempty" validate:"min=0"`
}

type GetProductRequest struct {
    ID string `validate:"required"`
}

type SearchProductRequest struct {
    Name        string `json:"name"`
    PriceMin    int64  `json:"price_min"`
    PriceMax    int64  `json:"price_max"`
    Page        int    `json:"page" validate:"min=1"`
    Size        int    `json:"size" validate:"min=1,max=100"`
}
```

### Langkah 3: Buat Model Converter

Buat file `internal/model/converter/product_converter.go`:

```go
package converter

import (
    "golang-clean-architecture/internal/entity"
    "golang-clean-architecture/internal/model"
)

func ProductToResponse(product *entity.Product) *model.ProductResponse {
    return &model.ProductResponse{
        ID:          product.ID,
        Name:        product.Name,
        Description: product.Description,
        Price:       product.Price,
        Stock:       product.Stock,
        UserID:      product.UserID,
        CreatedAt:   product.CreatedAt,
        UpdatedAt:   product.UpdatedAt,
    }
}

func ProductsToResponses(products []entity.Product) []model.ProductResponse {
    var responses []model.ProductResponse
    for _, product := range products {
        responses = append(responses, *ProductToResponse(&product))
    }
    return responses
}
```

### Langkah 4: Buat Repository

Buat file `internal/repository/product_repository.go`:

```go
package repository

import (
    "golang-clean-architecture/internal/entity"
    "golang-clean-architecture/internal/model"
    "github.com/sirupsen/logrus"
    "gorm.io/gorm"
)

type ProductRepository struct {
    Repository[entity.Product]
    Log *logrus.Logger
}

func NewProductRepository(log *logrus.Logger) *ProductRepository {
    return &ProductRepository{
        Log: log,
    }
}

func (r *ProductRepository) Search(db *gorm.DB, products *[]entity.Product, userID string, request *model.SearchProductRequest) (int64, error) {
    var total int64
    query := db.Model(products).Where("user_id = ?", userID)
    
    if request.Name != "" {
        query = query.Where("name LIKE ?", "%"+request.Name+"%")
    }
    if request.PriceMin > 0 {
        query = query.Where("price >= ?", request.PriceMin)
    }
    if request.PriceMax > 0 {
        query = query.Where("price <= ?", request.PriceMax)
    }
    
    if err := query.Count(&total).Error; err != nil {
        return 0, err
    }
    
    offset := (request.Page - 1) * request.Size
    if err := query.Offset(offset).Limit(request.Size).Find(products).Error; err != nil {
        return 0, err
    }
    
    return total, nil
}
```

### Langkah 5: Buat Use Case

Buat file `internal/usecase/product_usecase.go`:

```go
package usecase

import (
    "context"
    "golang-clean-architecture/internal/entity"
    "golang-clean-architecture/internal/model"
    "golang-clean-architecture/internal/model/converter"
    "golang-clean-architecture/internal/repository"
    
    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
    "gorm.io/gorm"
)

type ProductUseCase struct {
    DB                *gorm.DB
    Log               *logrus.Logger
    Validate          *validator.Validate
    ProductRepository *repository.ProductRepository
}

func NewProductUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate, 
    productRepository *repository.ProductRepository) *ProductUseCase {
    return &ProductUseCase{
        DB:                db,
        Log:               logger,
        Validate:          validate,
        ProductRepository: productRepository,
    }
}

func (c *ProductUseCase) Create(ctx context.Context, auth *model.Auth, request *model.CreateProductRequest) (*model.ProductResponse, error) {
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request body: %+v", err)
        return nil, fiber.ErrBadRequest
    }

    product := &entity.Product{
        ID:          uuid.New().String(),
        Name:        request.Name,
        Description: request.Description,
        Price:       request.Price,
        Stock:       request.Stock,
        UserID:      auth.ID,
    }

    if err := c.ProductRepository.Create(tx, product); err != nil {
        c.Log.Warnf("Failed to create product: %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    if err := tx.Commit().Error; err != nil {
        c.Log.Warnf("Failed to commit transaction: %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    return converter.ProductToResponse(product), nil
}

func (c *ProductUseCase) Get(ctx context.Context, auth *model.Auth, request *model.GetProductRequest) (*model.ProductResponse, error) {
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request body: %+v", err)
        return nil, fiber.ErrBadRequest
    }

    product := new(entity.Product)
    if err := c.ProductRepository.FindById(tx, product, request.ID); err != nil {
        c.Log.Warnf("Failed to find product: %+v", err)
        return nil, fiber.ErrNotFound
    }

    if product.UserID != auth.ID {
        c.Log.Warnf("Unauthorized access to product: %s", request.ID)
        return nil, fiber.ErrForbidden
    }

    if err := tx.Commit().Error; err != nil {
        c.Log.Warnf("Failed to commit transaction: %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    return converter.ProductToResponse(product), nil
}

func (c *ProductUseCase) Search(ctx context.Context, auth *model.Auth, request *model.SearchProductRequest) (*model.PageResponse[model.ProductResponse], error) {
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request body: %+v", err)
        return nil, fiber.ErrBadRequest
    }

    products := make([]entity.Product, 0)
    total, err := c.ProductRepository.Search(tx, &products, auth.ID, request)
    if err != nil {
        c.Log.Warnf("Failed to search products: %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    if err := tx.Commit().Error; err != nil {
        c.Log.Warnf("Failed to commit transaction: %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    responses := converter.ProductsToResponses(products)
    
    return &model.PageResponse[model.ProductResponse]{
        Data: responses,
        PageMetadata: model.PageMetadata{
            Page:      request.Page,
            Size:      request.Size,
            TotalItem: total,
            TotalPage: (total + int64(request.Size) - 1) / int64(request.Size),
        },
    }, nil
}
```

### Langkah 6: Buat Controller

Buat file `internal/delivery/http/product_controller.go`:

```go
package http

import (
    "golang-clean-architecture/internal/delivery/http/middleware"
    "golang-clean-architecture/internal/model"
    "golang-clean-architecture/internal/usecase"
    "strconv"
    
    "github.com/gofiber/fiber/v2"
    "github.com/sirupsen/logrus"
)

type ProductController struct {
    Log     *logrus.Logger
    UseCase *usecase.ProductUseCase
}

func NewProductController(useCase *usecase.ProductUseCase, logger *logrus.Logger) *ProductController {
    return &ProductController{
        Log:     logger,
        UseCase: useCase,
    }
}

func (c *ProductController) Create(ctx *fiber.Ctx) error {
    auth := middleware.GetUser(ctx)
    request := new(model.CreateProductRequest)
    
    if err := ctx.BodyParser(request); err != nil {
        c.Log.Warnf("Failed to parse request body: %+v", err)
        return fiber.ErrBadRequest
    }

    response, err := c.UseCase.Create(ctx.UserContext(), auth, request)
    if err != nil {
        c.Log.Warnf("Failed to create product: %+v", err)
        return err
    }

    return ctx.JSON(model.WebResponse[*model.ProductResponse]{Data: response})
}

func (c *ProductController) Get(ctx *fiber.Ctx) error {
    auth := middleware.GetUser(ctx)
    request := &model.GetProductRequest{
        ID: ctx.Params("productId"),
    }

    response, err := c.UseCase.Get(ctx.UserContext(), auth, request)
    if err != nil {
        c.Log.Warnf("Failed to get product: %+v", err)
        return err
    }

    return ctx.JSON(model.WebResponse[*model.ProductResponse]{Data: response})
}

func (c *ProductController) Search(ctx *fiber.Ctx) error {
    auth := middleware.GetUser(ctx)
    
    page, _ := strconv.Atoi(ctx.Query("page", "1"))
    size, _ := strconv.Atoi(ctx.Query("size", "10"))
    priceMin, _ := strconv.ParseInt(ctx.Query("price_min", "0"), 10, 64)
    priceMax, _ := strconv.ParseInt(ctx.Query("price_max", "0"), 10, 64)
    
    request := &model.SearchProductRequest{
        Name:     ctx.Query("name"),
        PriceMin: priceMin,
        PriceMax: priceMax,
        Page:     page,
        Size:     size,
    }

    response, err := c.UseCase.Search(ctx.UserContext(), auth, request)
    if err != nil {
        c.Log.Warnf("Failed to search products: %+v", err)
        return err
    }

    return ctx.JSON(response)
}
```

### Langkah 7: Update Route Configuration

Update file `internal/delivery/http/route/route.go`:

```go
type RouteConfig struct {
    App               *fiber.App
    UserController    *http.UserController
    ContactController *http.ContactController
    AddressController *http.AddressController
    ProductController *http.ProductController  // Tambahkan ini
    AuthMiddleware    fiber.Handler
}

func (c *RouteConfig) SetupAuthRoute() {
    c.App.Use(c.AuthMiddleware)
    
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

    // Address routes
    c.App.Get("/api/contacts/:contactId/addresses", c.AddressController.List)
    c.App.Post("/api/contacts/:contactId/addresses", c.AddressController.Create)
    c.App.Put("/api/contacts/:contactId/addresses/:addressId", c.AddressController.Update)
    c.App.Get("/api/contacts/:contactId/addresses/:addressId", c.AddressController.Get)
    c.App.Delete("/api/contacts/:contactId/addresses/:addressId", c.AddressController.Delete)
    
    // Product routes (Tambahkan ini)
    c.App.Get("/api/products", c.ProductController.Search)
    c.App.Post("/api/products", c.ProductController.Create)
    c.App.Get("/api/products/:productId", c.ProductController.Get)
}
```

### Langkah 8: Update Bootstrap Configuration

Update file `internal/config/app.go`:

```go
func Bootstrap(config *BootstrapConfig) {
    // setup repositories
    userRepository := repository.NewUserRepository(config.Log)
    contactRepository := repository.NewContactRepository(config.Log)
    addressRepository := repository.NewAddressRepository(config.Log)
    productRepository := repository.NewProductRepository(config.Log)  // Tambahkan ini

    // setup producer
    var userProducer *messaging.UserProducer
    var contactProducer *messaging.ContactProducer
    var addressProducer *messaging.AddressProducer

    if config.Producer != nil {
        userProducer = messaging.NewUserProducer(config.Producer, config.Log)
        contactProducer = messaging.NewContactProducer(config.Producer, config.Log)
        addressProducer = messaging.NewAddressProducer(config.Producer, config.Log)
    }

    // setup use cases
    userUseCase := usecase.NewUserUseCase(config.DB, config.Log, config.Validate, userRepository, userProducer)
    contactUseCase := usecase.NewContactUseCase(config.DB, config.Log, config.Validate, contactRepository, contactProducer)
    addressUseCase := usecase.NewAddressUseCase(config.DB, config.Log, config.Validate, contactRepository, addressRepository, addressProducer)
    productUseCase := usecase.NewProductUseCase(config.DB, config.Log, config.Validate, productRepository)  // Tambahkan ini

    // setup controller
    userController := http.NewUserController(userUseCase, config.Log)
    contactController := http.NewContactController(contactUseCase, config.Log)
    addressController := http.NewAddressController(addressUseCase, config.Log)
    productController := http.NewProductController(productUseCase, config.Log)  // Tambahkan ini

    // setup middleware
    authMiddleware := middleware.NewAuth(userUseCase)

    routeConfig := route.RouteConfig{
        App:               config.App,
        UserController:    userController,
        ContactController: contactController,
        AddressController: addressController,
        ProductController: productController,  // Tambahkan ini
        AuthMiddleware:    authMiddleware,
    }
    routeConfig.Setup()
}
```

## Best Practices

### 1. Dependency Injection
- Gunakan constructor pattern untuk inject dependencies
- Avoid global variables
- Use interfaces untuk loose coupling

### 2. Error Handling
- Gunakan structured logging
- Return appropriate HTTP status codes
- Handle database errors gracefully

### 3. Validation
- Validate request di use case layer
- Gunakan struct tags untuk validation rules
- Return clear error messages

### 4. Transaction Management
- Gunakan database transactions untuk consistency
- Always defer rollback
- Commit hanya setelah semua operations berhasil

### 5. Testing
- Write unit tests untuk setiap layer
- Use test doubles/mocks untuk dependencies
- Test happy path dan error scenarios

## Struktur Database Migration

Untuk membuat tabel product, buat migration file:

```sql
-- db/migrations/xxxxx_create_products_table.up.sql
CREATE TABLE products (
    id VARCHAR(100) NOT NULL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price BIGINT NOT NULL,
    stock INT NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- db/migrations/xxxxx_create_products_table.down.sql
DROP TABLE products;
```

## Kesimpulan

Clean Architecture menyediakan struktur yang jelas dan maintainable untuk aplikasi Go. Dengan mengikuti pola ini, Anda dapat:

1. **Separation of Concerns**: Setiap layer memiliki tanggung jawab yang jelas
2. **Testability**: Easy to test karena dependencies yang loose coupling
3. **Maintainability**: Code yang mudah diubah dan dipelihara
4. **Scalability**: Architecture yang dapat berkembang seiring kebutuhan

Pola ini sangat cocok untuk aplikasi enterprise yang membutuhkan maintainability tinggi dan tim development yang besar.