# 数据库包

[![Go 参考文档](https://pkg.go.dev/badge/go-slim.dev/infra/db.svg)](https://pkg.go.dev/go-slim.dev/infra/db)
[![Go 代码质量报告](https://goreportcard.com/badge/go-slim.dev/infra/db)](https://goreportcard.com/report/go-slim.dev/infra/db)

基于 GORM 构建的 Go 应用程序数据库抽象层，提供简洁一致的数据库操作接口。

## 功能特性

- 🚀 **多数据库支持**：MySQL、PostgreSQL、SQLite、SQL Server
- 🔄 **连接管理**：基于环境变量的配置
- 🔍 **查询构建器**：流畅的 API 用于构建复杂查询
- 📊 **分页支持**：内置分页功能
- 🛡️ **事务支持**：简化事务管理
- 🏷️ **数据类型**：扩展的数据类型支持
- 🔒 **连接安全**：支持 SSL/TLS

## 安装

```bash
go get go-slim.dev/infra/db
```

## 快速开始

### 环境变量配置

使用环境变量配置数据库连接：

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

### 初始化

```go
import (
    "go-slim.dev/infra/db"
    "go-slim.dev/env"
)

// 使用默认环境变量初始化
conn, err := db.Open()
if err != nil {
    log.Fatalf("连接数据库失败: %v", err)
}

defer func() {
    if db, err := conn.DB(); err == nil {
        _ = db.Close()
    }
}()

// 测试连接
if err := conn.Exec("SELECT 1").Error; err != nil {
    log.Fatalf("数据库连接测试失败: %v", err)
}
```

### 初始化

```go
// 初始化数据库连接
db, err := db.New(config)
if err != nil {
    log.Fatalf("连接数据库失败: %v", err)
}
defer db.Close()

// 测试连接
if err := db.Ping(); err != nil {
    log.Fatalf("数据库连接测试失败: %v", err)
}
```

### 基本操作

```go
// 定义模型
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 创建表
err := conn.AutoMigrate(&User{})
if err != nil {
    log.Fatalf("数据库迁移失败: %v", err)
}

// 创建新用户
user := User{Name: "张三", Email: "zhangsan@example.com"}
result := conn.Create(&user)
if result.Error != nil {
    log.Printf("创建用户失败: %v", result.Error)
}

// 查询单个用户
var foundUser User
result = conn.First(&foundUser, "email = ?", "zhangsan@example.com")
if result.Error != nil {
    log.Printf("查询用户失败: %v", result.Error)
}

// 查询多个用户
var users []User
result = conn.Where("name LIKE ?", "张%").Find(&users)
if result.Error != nil {
    log.Printf("查询用户列表失败: %v", result.Error)
}

// 更新用户
result = conn.Model(&user).Update("Name", "张三丰")
if result.Error != nil {
    log.Printf("更新用户失败: %v", result.Error)
}

// 删除用户
result = conn.Delete(&user)
if result.Error != nil {
    log.Printf("删除用户失败: %v", result.Error)
}
```

### 使用查询构建器

```go
// 创建查询构建器
qb := db.NewQueryBuilder[User](conn)

// 分页查询用户
pager, err := qb.Where("name LIKE ?", "张%").
    OrderBy("created_at DESC").
    Paginate(1, 10)

if err != nil {
    log.Printf("查询用户失败: %v", err)
} else {
    log.Printf("找到 %d 个用户 (第 %d 页，共 %d 页)", 
        len(pager.Items), 
        pager.Page, 
        int(math.Ceil(float64(pager.Total)/float64(pager.Limit))))
}

// 更多查询示例
activeUsers, err := qb.Where("status = ?", "active").
    Where("last_login > ?", time.Now().Add(-24*time.Hour)).
    Find()

// 统计用户数
count, err := qb.Where("name LIKE ?", "张%").Count()

// 获取单个用户
user, err := qb.Where("email = ?", "zhangsan@example.com").First()
```

## 高级用法

### 事务处理

```go
tx := conn.Begin()
if tx.Error != nil {
    log.Fatalf("开始事务失败: %v", tx.Error)
}

defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        log.Printf("事务回滚(panic): %v", r)
    } else if tx.Error != nil {
        tx.Rollback()
        log.Printf("事务回滚(错误): %v", tx.Error)
    } else {
        if err := tx.Commit().Error; err != nil {
            log.Printf("提交事务失败: %v", err)
        }
    }
}()

// 在事务中执行操作
if err := tx.Create(&user1).Error; err != nil {
    return err
}

if err := tx.Model(&user2).Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
    return err
}

## 支持的数据库驱动

- MySQL
- PostgreSQL
- SQLite
- SQL Server

## 数据类型 (dts 包)

`dts` 包提供了用于数据库操作的扩展数据类型和工具：

### 可用类型

- **基本类型**
  - `Bool`: 可空的布尔值，支持 JSON
  - `Int`: 可空的整数，支持 JSON
  - `Uint`: 可空的无符号整数，支持 JSON
  - `Float`: 可空的浮点数，支持 JSON
  - `String`: 增强的字符串类型，支持验证
  - `Time`: 增强的时间类型，支持 JSON 和数据库

- **专用类型**
  - `Decimal`: 高精度十进制数
  - `Email`: 电子邮件地址，支持验证
  - `Phone`: 电话号码，支持验证和格式化
  - `IDCard`: 身份证号验证
  - `IP`: IP 地址处理
  - `URL`: URL 解析和验证
  - `Color`: 颜色代码验证和转换

- **集合类型**
  - `Slice`: 通用的切片类型，支持数据库/序列化
  - `Map`: 通用的映射类型，支持数据库/序列化

### 使用示例

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

// 创建带验证数据的新用户
user := UserProfile{
    Email:    dts.Email("user@example.com"),
    Phone:    dts.Phone("+8613812345678"),
    IsActive: dts.Bool(true),
    Score:    dts.Decimal("99.99"),
    Settings: dts.Map{"theme": "dark", "notifications": true},
    Tags:     dts.Slice{"vip", "early_adopter"},
}

// 验证字段
if err := user.Email.Validate(); err != nil {
    return fmt.Errorf("邮箱格式无效: %v", err)
}

// 在数据库操作中使用
if err := db.Create(&user).Error; err != nil {
    return fmt.Errorf("创建用户失败: %v", err)
}
```

### 特性

- **类型安全**：强类型字段，防止常见错误
- **内置验证**：每种类型都包含验证逻辑
- **数据库集成**：与 GORM 无缝集成
- **JSON 支持**：正确的 JSON 序列化/反序列化
- **空值安全**：优雅处理 NULL 值
- **格式化输出**：一致的字符串表示

### 验证

每种类型都包含验证方法：

```go
email := dts.Email("invalid-email")
if err := email.Validate(); err != nil {
    log.Printf("验证错误: %v", err)
}
```

### 数据库操作

所有类型都可以直接与 GORM 一起使用：

```go
// 使用 dts 类型查询
var user UserProfile
db.Where("email = ?", dts.Email("user@example.com")).First(&user)

// 使用 dts 类型更新
db.Model(&user).Update("score", dts.Decimal("100.00"))
```

## 最佳实践

1. **连接管理**：
   - 使用完毕后始终关闭数据库连接
   - 使用 `SetMaxOpenConns`、`SetMaxIdleConns` 和 `SetConnMaxLifetime` 有效管理连接池
   - 使用 `SetConnMaxIdleTime` 设置适当的超时时间

2. **错误处理**：
   - 始终检查并处理 GORM 操作的错误
   - 对多个相关操作使用事务
   - 为临时性故障实现适当的重试逻辑

3. **性能优化**：
   - 使用 `Select` 指定需要的列
   - 合理使用 `Preload` 和 `Joins` 避免 N+1 查询问题
   - 为频繁查询的列添加适当的索引

4. **安全性**：
   - 始终使用参数化查询（GORM 自动处理）
   - 不要记录敏感信息
   - 使用环境变量存储数据库凭据
   - 在生产环境中启用 SSL/TLS 加密数据库连接

## 许可证

MIT