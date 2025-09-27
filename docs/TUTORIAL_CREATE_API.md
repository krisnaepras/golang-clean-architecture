# Tutorial: Membuat REST API Baru - Step by Step

## Studi Kasus: Membuat API Product

Kita akan membuat REST API untuk mengelola **Product** dengan endpoints:
- `POST /api/products` - Create product
- `GET /api/products/{id}` - Get product by ID  
- `GET /api/products` - Search products

## Step 1: Buat Entity (Domain Model)

**File**: `internal/entity/product_entity.go`

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

**Penjelasan:**
- `gorm` tags menentukan mapping ke database columns
- `primaryKey` untuk primary key
- `autoCreateTime` dan `autoUpdateTime` untuk timestamp otomatis
- `foreignKey` untuk relasi ke tabel users

## Step 2: Buat Request/Response Models

**File**: `internal/model/product_model.go`

```go
package model

// Response model - data yang dikirim ke client
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

// Request model untuk create product
type CreateProductRequest struct {
    Name        string `json:"name" validate:"required,max=100"`
    Description string `json:"description" validate:"max=500"`
    Price       int64  `json:"price" validate:"required,min=0"`
    Stock       int    `json:"stock" validate:"required,min=0"`
}

// Request model untuk get product by ID
type GetProductRequest struct {
    ID string `validate:"required"`
}

// Request model untuk search products
type SearchProductRequest struct {
    Name        string `json:"name"`
    PriceMin    int64  `json:"price_min"`
    PriceMax    int64  `json:"price_max"`
    Page        int    `json:"page" validate:"min=1"`
    Size        int    `json:"size" validate:"min=1,max=100"`
}
```

**Penjelasan:**
- `json` tags untuk serialisasi JSON
- `validate` tags untuk validasi input
- `omitempty` membuat field tidak muncul jika kosong

## Step 3: Buat Model Converter

**File**: `internal/model/converter/product_converter.go`

```go
package converter

import (
    "golang-clean-architecture/internal/entity"
    "golang-clean-architecture/internal/model"
)

// Convert single entity to response
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

// Convert multiple entities to responses
func ProductsToResponses(products []entity.Product) []model.ProductResponse {
    var responses []model.ProductResponse
    for _, product := range products {
        responses = append(responses, *ProductToResponse(&product))
    }
    return responses
}
```

**Penjelasan:**
- Converter memisahkan Entity (database) dari Response (API)
- Memudahkan jika structure database berbeda dengan API response

## Step 4: Buat Repository (Data Access)

**File**: `internal/repository/product_repository.go`

```go
package repository

import (
    "golang-clean-architecture/internal/entity"
    "golang-clean-architecture/internal/model"
    "github.com/sirupsen/logrus"
    "gorm.io/gorm"
)

type ProductRepository struct {
    Repository[entity.Product]  // Inherit basic CRUD operations
    Log *logrus.Logger
}

func NewProductRepository(log *logrus.Logger) *ProductRepository {
    return &ProductRepository{
        Log: log,
    }
}

// Custom method untuk search products
func (r *ProductRepository) Search(db *gorm.DB, products *[]entity.Product, userID string, request *model.SearchProductRequest) (int64, error) {
    var total int64
    
    // Build query
    query := db.Model(products).Where("user_id = ?", userID)
    
    // Add filters
    if request.Name != "" {
        query = query.Where("name LIKE ?", "%"+request.Name+"%")
    }
    if request.PriceMin > 0 {
        query = query.Where("price >= ?", request.PriceMin)
    }
    if request.PriceMax > 0 {
        query = query.Where("price <= ?", request.PriceMax)
    }
    
    // Count total records
    if err := query.Count(&total).Error; err != nil {
        return 0, err
    }
    
    // Apply pagination
    offset := (request.Page - 1) * request.Size
    if err := query.Offset(offset).Limit(request.Size).Find(products).Error; err != nil {
        return 0, err
    }
    
    return total, nil
}
```

**Penjelasan:**
- `Repository[entity.Product]` memberikan method basic: Create, Update, Delete, FindById
- Method `Search` adalah custom method untuk query kompleks
- Return total count untuk pagination

## Step 5: Buat UseCase (Business Logic)

**File**: `internal/usecase/product_usecase.go`

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

// Create new product
func (c *ProductUseCase) Create(ctx context.Context, auth *model.Auth, request *model.CreateProductRequest) (*model.ProductResponse, error) {
    // Start transaction
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    // Validate request
    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request body: %+v", err)
        return nil, fiber.ErrBadRequest
    }

    // Create product entity
    product := &entity.Product{
        ID:          uuid.New().String(),
        Name:        request.Name,
        Description: request.Description,
        Price:       request.Price,
        Stock:       request.Stock,
        UserID:      auth.ID,
    }

    // Save to database
    if err := c.ProductRepository.Create(tx, product); err != nil {
        c.Log.Warnf("Failed to create product: %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    // Commit transaction
    if err := tx.Commit().Error; err != nil {
        c.Log.Warnf("Failed to commit transaction: %+v", err)
        return nil, fiber.ErrInternalServerError
    }

    return converter.ProductToResponse(product), nil
}

// Get product by ID
func (c *ProductUseCase) Get(ctx context.Context, auth *model.Auth, request *model.GetProductRequest) (*model.ProductResponse, error) {
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    // Validate request
    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request body: %+v", err)
        return nil, fiber.ErrBadRequest
    }

    // Find product
    product := new(entity.Product)
    if err := c.ProductRepository.FindById(tx, product, request.ID); err != nil {
        c.Log.Warnf("Failed to find product: %+v", err)
        return nil, fiber.ErrNotFound
    }

    // Check authorization
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

// Search products with pagination
func (c *ProductUseCase) Search(ctx context.Context, auth *model.Auth, request *model.SearchProductRequest) (*model.PageResponse[model.ProductResponse], error) {
    tx := c.DB.WithContext(ctx).Begin()
    defer tx.Rollback()

    // Validate request
    if err := c.Validate.Struct(request); err != nil {
        c.Log.Warnf("Invalid request body: %+v", err)
        return nil, fiber.ErrBadRequest
    }

    // Search products
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

    // Convert to response
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

**Penjelasan:**
- Setiap method menggunakan database transaction
- Validasi input menggunakan validator
- Authorization check untuk memastikan user hanya akses datanya sendiri
- Error handling dan logging yang proper

## Step 6: Buat Controller (HTTP Handler)

**File**: `internal/delivery/http/product_controller.go`

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

// POST /api/products
func (c *ProductController) Create(ctx *fiber.Ctx) error {
    // Get authenticated user
    auth := middleware.GetUser(ctx)
    
    // Parse request body
    request := new(model.CreateProductRequest)
    if err := ctx.BodyParser(request); err != nil {
        c.Log.Warnf("Failed to parse request body: %+v", err)
        return fiber.ErrBadRequest
    }

    // Call use case
    response, err := c.UseCase.Create(ctx.UserContext(), auth, request)
    if err != nil {
        c.Log.Warnf("Failed to create product: %+v", err)
        return err
    }

    // Return JSON response
    return ctx.JSON(model.WebResponse[*model.ProductResponse]{Data: response})
}

// GET /api/products/{productId}
func (c *ProductController) Get(ctx *fiber.Ctx) error {
    auth := middleware.GetUser(ctx)
    
    // Get path parameter
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

// GET /api/products?name=...&price_min=...&page=1&size=10
func (c *ProductController) Search(ctx *fiber.Ctx) error {
    auth := middleware.GetUser(ctx)
    
    // Parse query parameters
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

**Penjelasan:**
- Controller hanya handle HTTP request/response
- `ctx.BodyParser()` untuk parse JSON request body
- `ctx.Params()` untuk path parameters (/products/{id})
- `ctx.Query()` untuk query parameters (?page=1&size=10)
- `middleware.GetUser()` untuk mendapat user yang sedang login

## Step 7: Update Routes Configuration

**File**: `internal/delivery/http/route/route.go`

Tambahkan ProductController ke RouteConfig:

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
    
    // Existing routes...
    c.App.Delete("/api/users", c.UserController.Logout)
    c.App.Patch("/api/users/_current", c.UserController.Update)
    c.App.Get("/api/users/_current", c.UserController.Current)

    // Contact routes...
    c.App.Get("/api/contacts", c.ContactController.List)
    c.App.Post("/api/contacts", c.ContactController.Create)
    // ... other contact routes
    
    // Product routes (Tambahkan ini)
    c.App.Get("/api/products", c.ProductController.Search)      // Search with query params
    c.App.Post("/api/products", c.ProductController.Create)     // Create new product
    c.App.Get("/api/products/:productId", c.ProductController.Get)  // Get by ID
}
```

## Step 8: Update Dependency Injection

**File**: `internal/config/app.go`

Update function Bootstrap untuk include ProductController:

```go
func Bootstrap(config *BootstrapConfig) {
    // Setup repositories
    userRepository := repository.NewUserRepository(config.Log)
    contactRepository := repository.NewContactRepository(config.Log)
    addressRepository := repository.NewAddressRepository(config.Log)
    productRepository := repository.NewProductRepository(config.Log)  // Tambahkan ini

    // Setup producers (if needed)
    var userProducer *messaging.UserProducer
    var contactProducer *messaging.ContactProducer
    var addressProducer *messaging.AddressProducer

    if config.Producer != nil {
        userProducer = messaging.NewUserProducer(config.Producer, config.Log)
        contactProducer = messaging.NewContactProducer(config.Producer, config.Log)
        addressProducer = messaging.NewAddressProducer(config.Producer, config.Log)
    }

    // Setup use cases
    userUseCase := usecase.NewUserUseCase(config.DB, config.Log, config.Validate, userRepository, userProducer)
    contactUseCase := usecase.NewContactUseCase(config.DB, config.Log, config.Validate, contactRepository, contactProducer)
    addressUseCase := usecase.NewAddressUseCase(config.DB, config.Log, config.Validate, contactRepository, addressRepository, addressProducer)
    productUseCase := usecase.NewProductUseCase(config.DB, config.Log, config.Validate, productRepository)  // Tambahkan ini

    // Setup controllers
    userController := http.NewUserController(userUseCase, config.Log)
    contactController := http.NewContactController(contactUseCase, config.Log)
    addressController := http.NewAddressController(addressUseCase, config.Log)
    productController := http.NewProductController(productUseCase, config.Log)  // Tambahkan ini

    // Setup middleware
    authMiddleware := middleware.NewAuth(userUseCase)

    // Setup routes
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

## Step 9: Database Migration

**File**: `db/migrations/005_create_products_table.up.sql`

```sql
CREATE TABLE products (
    id VARCHAR(100) NOT NULL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price BIGINT NOT NULL,
    stock INT NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_products_user_id (user_id),
    INDEX idx_products_name (name),
    INDEX idx_products_price (price)
);
```

**File**: `db/migrations/005_create_products_table.down.sql`

```sql
DROP TABLE products;
```

## Step 10: Testing API

### 1. Register & Login User
```bash
# Register
POST /api/users
{
  "id": "john",
  "password": "secret",
  "name": "John Doe"
}

# Login
POST /api/users/_login  
{
  "id": "john", 
  "password": "secret"
}
# Response: {"data": {"token": "abc123..."}}
```

### 2. Create Product
```bash
POST /api/products
Authorization: abc123...
{
  "name": "Laptop Gaming",
  "description": "Gaming laptop with RTX 4080",
  "price": 25000000,
  "stock": 5
}
```

### 3. Get Product
```bash
GET /api/products/{productId}
Authorization: abc123...
```

### 4. Search Products  
```bash
GET /api/products?name=laptop&price_min=10000000&price_max=30000000&page=1&size=10
Authorization: abc123...
```

## Kesimpulan

Dengan mengikuti 10 langkah ini, kita telah berhasil membuat REST API lengkap dengan:

✅ **Clean Architecture** - kode terstruktur dan maintainable  
✅ **CRUD Operations** - Create, Read, Update, Delete  
✅ **Authentication** - user-specific data access  
✅ **Validation** - input validation dengan validator tags  
✅ **Pagination** - search dengan pagination support  
✅ **Error Handling** - proper error responses  
✅ **Transaction Management** - database consistency  
✅ **Logging** - structured logging untuk debugging

Pola ini bisa direplikasi untuk endpoint lainnya dengan mudah!