# JWT 认证包

[![Go 参考文档](https://pkg.go.dev/badge/go-slim.dev/infra/jwt.svg)](https://pkg.go.dev/go-slim.dev/infra/jwt)
[![Go 代码质量](https://goreportcard.com/badge/go-slim.dev/infra/jwt)](https://goreportcard.com/report/go-slim.dev/infra/jwt)
[![测试状态](https://github.com/go-slim/jwt/workflows/Test/badge.svg)](https://github.com/go-slim/jwt/actions?query=workflow%3ATest)

一个健壮的 JWT (JSON Web Token) 实现，提供安全的令牌生成、解析和验证功能，支持多种签名方法。

## 功能特性

- 🔐 支持多种签名方法 (HMAC, RSA, ECDSA, EdDSA)
- ⏱️ 令牌过期和验证
- 🔄 令牌刷新机制
- 🛡️ 安全默认值和最佳实践
- 🧪 全面的测试覆盖
- 🚀 高性能

## 安装

```bash
go get go-slim.dev/infra/jwt
```

## 快速开始

### 生成令牌

```go
package main

import (
	"fmt"
	"time"

	"go-slim.dev/infra/jwt"
)

func main() {
	// 使用 HMAC 签名方法创建新令牌
	token := jwt.New(jwt.SigningMethodHS256)

	// 设置声明(claims)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = "1234567890"
	claims["name"] = "张三"
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// 生成签名后的令牌字符串
	tokenString, err := token.SignedString([]byte("你的密钥"))
	if err != nil {
		panic(err)
	}

	fmt.Println("生成的令牌:", tokenString)
}
```

### 验证令牌

```go
// 解析并验证令牌
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    // 验证签名方法
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("不支持的签名方法: %v", token.Header["alg"])
    }
    return []byte("你的密钥"), nil
})

if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
    fmt.Println("用户ID:", claims["sub"])
    fmt.Println("过期时间:", time.Unix(int64(claims["exp"].(float64)), 0))
} else {
    fmt.Println("无效的令牌:", err)
}
```

## 高级用法

### 使用 RSA 签名

```go
// 生成 RSA 密钥对
privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
if err != nil {
    panic(err)
}

// 使用 RSA 签名创建令牌
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = "user123"
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

// 签名并获取完整的编码后令牌字符串
tokenString, err := token.SignedString(privateKey)
```

### 使用自定义声明验证令牌

```go
type CustomClaims struct {
    UserID string `json:"user_id"`
    jwt.StandardClaims
}

// 使用自定义声明解析令牌
token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
    return []byte("你的密钥"), nil
})

if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
    fmt.Printf("用户ID: %v\n", claims.UserID)
    fmt.Printf("过期时间: %v\n", time.Unix(claims.ExpiresAt, 0))
} else {
    fmt.Println("无效的令牌:", err)
}
```

## 安全最佳实践

1. 始终使用强大且唯一的密钥
2. 设置适当的令牌过期时间
3. 所有令牌传输都使用 HTTPS
4. 安全地存储令牌（web 应用使用 httpOnly cookies）
5. 实现令牌刷新机制
6. 定期轮换签名密钥
7. 在服务端验证所有令牌声明

## 许可证

MIT

## 贡献

欢迎贡献代码！请随时提交 Pull Request。