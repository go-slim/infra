# Database Package

[![Go Reference](https://pkg.go.dev/badge/go-slim.dev/infra/db.svg)](https://pkg.go.dev/go-slim.dev/infra/db)
[![Go Report Card](https://goreportcard.com/badge/go-slim.dev/infra/db)](https://goreportcard.com/report/go-slim.dev/infra/db)

A robust database abstraction layer for Go applications built on top of GORM, providing a clean and consistent interface for database operations.

## Features

- 🚀 **Multiple Database Support**: MySQL, PostgreSQL, SQLite, SQL Server
- 🔄 **Connection Management**: Environment-based configuration
- 🔍 **Query Builder**: Fluent API for building complex queries
- 📊 **Pagination**: Built-in pagination support
- 🔄 **Transaction Support**: Easy transaction management
- 🏷️ **Data Types**: Extended data types support
- 🔒 **Connection Security**: SSL/TLS support

## Installation

```bash
go get go-slim.dev/infra/db
```

## Quick Start

### Environment Configuration

Configure your database connection using environment variables:

```env
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=user
DB_PASSWORD=password
DB_DATABASE=mydb
DB_CHARSET=utf8mb4
DB_TIMEZONE=Local
DB_SSLMODE=disable
```

### Initialization

```go
import (
    "go-slim.dev/infra/db"
    "go-slim.dev/env"
)

// Initialize with default environment
conn, err := db.Open()
if err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}

defer func() {
    if db, err := conn.DB(); err == nil {
        _ = db.Close()
    }
}()

// Ping to verify connection
if err := conn.Exec("SELECT 1").Error; err != nil {
    log.Fatalf("Failed to ping database: %v", err)
}
```

### Basic Operations

```go
// Define your model
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Create table
err := conn.AutoMigrate(&User{})
if err != nil {
    log.Fatalf("Failed to migrate database: %v", err)
}

// Create a new user
user := User{Name: "John Doe", Email: "john@example.com"}
result := conn.Create(&user)
if result.Error != nil {
    log.Printf("Error creating user: %v", result.Error)
}

// Query a single user
var foundUser User
result = conn.First(&foundUser, "email = ?", "john@example.com")
if result.Error != nil {
    log.Printf("Error finding user: %v", result.Error)
}

// Query multiple users
var users []User
result = conn.Where("name LIKE ?", "John%").Find(&users)
if result.Error != nil {
    log.Printf("Error finding users: %v", result.Error)
}

// Update a user
result = conn.Model(&user).Update("Name", "John Updated")
if result.Error != nil {
    log.Printf("Error updating user: %v", result.Error)
}

// Delete a user
result = conn.Delete(&user)
if result.Error != nil {
    log.Printf("Error deleting user: %v", result.Error)
}
```

### Using Query Builder

```go
// Create a new query builder
qb := db.NewQueryBuilder[User](conn)

// Find users with pagination
pager, err := qb.Where("name LIKE ?", "John%").
    OrderBy("created_at DESC").
    Paginate(1, 10)

if err != nil {
    log.Printf("Error querying users: %v", err)
} else {
    log.Printf("Found %d users (page %d of %d)", 
        len(pager.Items), 
        pager.Page, 
        int(math.Ceil(float64(pager.Total)/float64(pager.Limit))))
}

// More query examples
activeUsers, err := qb.Where("status = ?", "active").
    Where("last_login > ?", time.Now().Add(-24*time.Hour)).
    Find()

// Count users
count, err := qb.Where("name LIKE ?", "John%").Count()

// Get single user
user, err := qb.Where("email = ?", "john@example.com").First()
```

## Advanced Usage

### Transactions

```go
tx := conn.Begin()
if tx.Error != nil {
    log.Fatalf("Error starting transaction: %v", tx.Error)
}

defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        log.Printf("Transaction rolled back due to panic: %v", r)
    } else if tx.Error != nil {
        tx.Rollback()
        log.Printf("Transaction rolled back due to error: %v", tx.Error)
    } else {
        if err := tx.Commit().Error; err != nil {
            log.Printf("Error committing transaction: %v", err)
        }
    }
}()

// Perform operations within the transaction
if err := tx.Create(&user1).Error; err != nil {
    return err
}

if err := tx.Model(&user2).Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
    return err
}

## Supported Database Drivers

- MySQL
- PostgreSQL
- SQLite
- SQL Server

## Data Types (dts package)

The `dts` package provides extended data types and utilities for database operations:

### Available Types

- **Basic Types**
  - `Bool`: Nullable boolean with JSON support
  - `Int`: Nullable integer with JSON support
  - `Uint`: Nullable unsigned integer with JSON support
  - `Float`: Nullable float with JSON support
  - `String`: Enhanced string type with validation
  - `Time`: Enhanced time type with JSON and database support

- **Specialized Types**
  - `Decimal`: High-precision decimal number
  - `Email`: Email address with validation
  - `Phone`: Phone number with validation and formatting
  - `IDCard`: ID card number validation
  - `IP`: IP address handling
  - `URL`: URL parsing and validation
  - `Color`: Color code validation and conversion

- **Collection Types**
  - `Slice`: Generic slice type with database/serialization support
  - `Map`: Generic map type with database/serialization support

### Usage Example

```go
import "go-slim.dev/infra/db/dts"

type UserProfile struct {
    ID        dts.Uint    `gorm:"primaryKey"`
    IsActive  dts.Bool    `gorm:"default:true"`
    Email     dts.Email   `gorm:"size:100"`
    Phone     dts.Phone   
    Score     dts.Decimal `gorm:"type:decimal(10,2)"`
    Settings  dts.Map     `gorm:"type:json"`
    Tags      dts.Slice   `gorm:"type:json"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Create a new user with validated data
user := UserProfile{
    Email:    dts.Email("user@example.com"),
    Phone:    dts.Phone("+8613812345678"),
    IsActive: dts.Bool(true),
    Score:    dts.Decimal("99.99"),
    Settings: dts.Map{"theme": "dark", "notifications": true},
    Tags:     dts.Slice{"vip", "early_adopter"},
}

// Validate fields
if err := user.Email.Validate(); err != nil {
    return fmt.Errorf("invalid email: %v", err)
}

// Use in database operations
if err := db.Create(&user).Error; err != nil {
    return fmt.Errorf("failed to create user: %v", err)
}
```

### Features

- **Type Safety**: Strongly typed fields prevent common errors
- **Built-in Validation**: Each type includes validation logic
- **Database Integration**: Seamless integration with GORM
- **JSON Support**: Proper JSON serialization/deserialization
- **Null Safety**: Handle NULL values gracefully
- **Formatted Output**: Consistent string representation

### Validation

Each type includes validation methods:

```go
email := dts.Email("invalid-email")
if err := email.Validate(); err != nil {
    log.Printf("Validation error: %v", err)
}
```

### Database Operations

All types work with GORM out of the box:

```go
// Query with dts types
var user UserProfile
db.Where("email = ?", dts.Email("user@example.com")).First(&user)

// Update with dts types
db.Model(&user).Update("score", dts.Decimal("100.00"))
```

## Best Practices

1. **Connection Management**:
   - Always close the database connection when done
   - Use connection pooling effectively with `SetMaxOpenConns`, `SetMaxIdleConns`, and `SetConnMaxLifetime`
   - Set appropriate timeouts using `SetConnMaxIdleTime`

2. **Error Handling**:
   - Always check and handle errors from GORM operations
   - Use transactions for multiple related operations
   - Implement proper retry logic for transient failures

3. **Performance**:
   - Use `Select` to specify only needed columns
   - Use `Preload` and `Joins` wisely to avoid N+1 queries
   - Add appropriate indexes for frequently queried columns

4. **Security**:
   - Always use parameterized queries (handled automatically by GORM)
   - Never log sensitive information
   - Use environment variables for database credentials
   - Enable SSL/TLS for production database connections

## License

MIT