# 📚 Dokumentasi Lengkap: Clean Architecture & REST API

Selamat datang! Repository ini berisi implementasi **Clean Architecture** untuk REST API menggunakan Golang, dilengkapi dengan dokumentasi lengkap dalam Bahasa Indonesia.

## 📋 Daftar Dokumentasi

### 1. 🏗️ [Analisis Architecture Project](ARCHITECTURE_ANALYSIS.md)
**Untuk memahami struktur project yang sudah ada**
- Breakdown detail setiap layer (Entity, Repository, UseCase, Controller)  
- Analisis flow eksekusi REST API
- Penjelasan authentication & authorization
- Database design dan integration testing
- Best practices yang sudah diterapkan

### 2. 📖 [Ringkasan Clean Architecture](CLEAN_ARCHITECTURE_SUMMARY.md)  
**Untuk pemahaman cepat konsep Clean Architecture**
- Penjelasan singkat setiap layer dengan contoh kode
- Flow eksekusi HTTP Request → Controller → UseCase → Repository
- Structure response API dan error handling
- Keuntungan menggunakan Clean Architecture

### 3. 🛠️ [Tutorial Step-by-Step: Membuat REST API Baru](TUTORIAL_CREATE_API.md)
**Panduan praktis membuat endpoint REST API dari nol**
- 10 langkah lengkap membuat API Product (Create, Read, Search)
- Contoh kode lengkap untuk setiap layer
- Database migration dan testing API
- Best practices implementation

### 4. 📚 [Panduan Lengkap Clean Architecture](CLEAN_ARCHITECTURE_GUIDE.md)
**Deep dive into Clean Architecture concepts**
- Penjelasan mendalam setiap layer dengan filosofi di baliknya
- Contoh implementasi lengkap Product API
- Advanced patterns dan best practices
- Tips untuk team development

## 🚀 Quick Start

### 1. Clone Repository
```bash
git clone https://github.com/krisnaepras/golang-clean-architecture.git
cd golang-clean-architecture
```

### 2. Setup Dependencies
```bash
go mod tidy
```

### 3. Run Application
```bash
go run cmd/web/main.go
```

### 4. Test API Endpoints
```bash
# Register user
POST http://localhost:3000/api/users
{
  "id": "john",
  "password": "secret", 
  "name": "John Doe"
}

# Login
POST http://localhost:3000/api/users/_login
{
  "id": "john",
  "password": "secret"
}
```

## 🏛️ Struktur Clean Architecture

```
┌─────────────────────┐
│   HTTP Request      │
└─────────────────────┘
           │
┌─────────────────────┐
│   Controller        │  ← HTTP handlers (Fiber)
│  (Delivery Layer)   │
└─────────────────────┘
           │
┌─────────────────────┐  
│   UseCase          │  ← Business Logic
│ (Business Layer)    │
└─────────────────────┘
           │
┌─────────────────────┐
│   Repository        │  ← Data Access (GORM)
│ (Data Layer)        │
└─────────────────────┘
           │
┌─────────────────────┐
│   Database          │  ← MySQL
└─────────────────────┘
```

## 🎯 Fitur yang Sudah Ada

### User Management
- ✅ Register user baru
- ✅ Login dengan password hashing (bcrypt)
- ✅ Logout dengan token invalidation  
- ✅ Get current user profile
- ✅ Update user profile

### Contact Management  
- ✅ Create contact (per user)
- ✅ Get contact by ID
- ✅ Update contact
- ✅ Delete contact  
- ✅ Search contacts dengan pagination

### Address Management
- ✅ Create address (per contact)
- ✅ Get address by ID
- ✅ Update address
- ✅ Delete address
- ✅ List addresses per contact

### Infrastructure
- ✅ Authentication middleware dengan token
- ✅ Input validation dengan struct tags
- ✅ Database transaction management
- ✅ Error handling dan logging
- ✅ Kafka integration untuk event publishing
- ✅ Generic repository pattern dengan Go generics
- ✅ Comprehensive integration testing

## 📁 Struktur Project

```
.
├── cmd/
│   ├── web/main.go          # Web server entry point
│   └── worker/main.go       # Background worker
├── internal/
│   ├── entity/              # Domain models (database)
│   ├── model/               # Request/Response models
│   ├── repository/          # Data access layer
│   ├── usecase/             # Business logic layer  
│   ├── delivery/http/       # HTTP controllers & routes
│   ├── config/              # Configuration & DI
│   └── gateway/             # External services (Kafka)
├── test/                    # Integration tests
├── db/migrations/           # Database migrations
├── api/                     # OpenAPI specification
└── docs/                    # Documentation (ini)
```

## 🛠️ Tech Stack

- **Language**: Go 1.25
- **HTTP Framework**: [Fiber](https://github.com/gofiber/fiber) v2
- **Database**: MySQL dengan [GORM](https://gorm.io/) ORM
- **Validation**: [Go Playground Validator](https://github.com/go-playground/validator) v10
- **Logging**: [Logrus](https://github.com/sirupsen/logrus)
- **Configuration**: [Viper](https://github.com/spf13/viper)
- **Messaging**: [Apache Kafka](https://kafka.apache.org/) dengan [Sarama](https://github.com/IBM/sarama)
- **Testing**: [Testify](https://github.com/stretchr/testify)

## 📝 API Endpoints

### Guest Endpoints (No Auth)
```
POST   /api/users            # Register
POST   /api/users/_login     # Login
```

### Authenticated Endpoints (Require Token)
```
# User
GET    /api/users/_current   # Get profile
PATCH  /api/users/_current   # Update profile  
DELETE /api/users            # Logout

# Contact
GET    /api/contacts          # Search contacts
POST   /api/contacts          # Create contact
GET    /api/contacts/{id}     # Get contact
PUT    /api/contacts/{id}     # Update contact
DELETE /api/contacts/{id}     # Delete contact

# Address  
GET    /api/contacts/{id}/addresses           # List addresses
POST   /api/contacts/{id}/addresses           # Create address
GET    /api/contacts/{id}/addresses/{id}      # Get address
PUT    /api/contacts/{id}/addresses/{id}      # Update address
DELETE /api/contacts/{id}/addresses/{id}      # Delete address
```

## 🧪 Testing

```bash
# Run all tests
go test -v ./test/

# Run specific test
go test -v ./test/ -run TestCreateContact

# Run with coverage
go test -v ./test/ -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 📊 Database Schema

**Users** → **Contacts** → **Addresses** (One-to-Many relationships)

- Users can have multiple Contacts
- Contacts can have multiple Addresses  
- All operations are user-scoped (multi-tenant)

## 🔐 Authentication

- **Token-based authentication** menggunakan UUID
- **Password hashing** dengan bcrypt  
- **Authorization middleware** untuk protected endpoints
- **User isolation** - setiap user hanya bisa akses data miliknya

## 📈 Keuntungan Clean Architecture

### ✅ **Maintainable**
- Kode terorganisir per layer dengan tanggung jawab jelas
- Mudah locate bugs dan add new features
- Separation of concerns yang strict

### ✅ **Testable**  
- Easy mocking dengan dependency injection
- Unit testing per layer  
- Comprehensive integration testing

### ✅ **Scalable**
- Horizontal scaling dengan stateless design
- Event-driven architecture dengan Kafka
- Database sharding ready

### ✅ **Flexible**
- Framework agnostic business logic
- Database agnostic dengan repository pattern
- Easy integration dengan external services

## 💡 Mulai dari Mana?

### Untuk Pemula
1. 📖 Baca [Ringkasan Clean Architecture](CLEAN_ARCHITECTURE_SUMMARY.md) untuk overview
2. 🏗️ Baca [Analisis Architecture Project](ARCHITECTURE_ANALYSIS.md) untuk memahami code yang ada
3. 🛠️ Ikuti [Tutorial Create API](TUTORIAL_CREATE_API.md) untuk hands-on practice

### Untuk Advanced
1. 📚 Baca [Panduan Lengkap](CLEAN_ARCHITECTURE_GUIDE.md) untuk deep understanding
2. Implement fitur baru mengikuti pattern yang ada
3. Explore integration dengan external services

## 🤝 Contributing

Feel free untuk:
- Add documentation improvements
- Fix bugs atau add features  
- Share best practices
- Create examples untuk use cases lain

---

**Happy coding!** 🚀 

Semoga dokumentasi ini membantu Anda memahami dan mengimplementasikan Clean Architecture dengan Golang!