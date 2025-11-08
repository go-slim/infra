# JWT Authentication Package

[![Go Reference](https://pkg.go.dev/badge/go-slim.dev/infra/jwt.svg)](https://pkg.go.dev/go-slim.dev/infra/jwt)
[![Go Report Card](https://goreportcard.com/badge/go-slim.dev/infra/jwt)](https://goreportcard.com/report/go-slim.dev/infra/jwt)
[![Test Status](https://github.com/go-slim/jwt/workflows/Test/badge.svg)](https://github.com/go-slim/jwt/actions?query=workflow%3ATest)

A robust JWT (JSON Web Token) implementation for Go, providing secure token generation, parsing, and validation with support for multiple signing methods.

## Features

- 🔐 Support for multiple signing methods (HMAC, RSA, ECDSA, EdDSA)
- ⏱️ Token expiration and validation
- 🔄 Token refresh mechanism
- 🛡️ Secure defaults and best practices
- 🧪 Comprehensive test coverage
- 🚀 High performance

## Installation

```bash
go get go-slim.dev/infra/jwt
```

## Quick Start

### Generating Tokens

```go
package main

import (
	"fmt"
	"time"

	"go-slim.dev/infra/jwt"
)

func main() {
	// Create a new token with HMAC signing method
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = "1234567890"
	claims["name"] = "John Doe"
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// Generate encoded token
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated token:", tokenString)
}
```

### Validating Tokens

```go
// Parse and validate token
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    // Validate the signing method
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
    }
    return []byte("your-secret-key"), nil
})

if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
    fmt.Println("User ID:", claims["sub"])
    fmt.Println("Expires at:", time.Unix(int64(claims["exp"].(float64)), 0))
} else {
    fmt.Println("Invalid token:", err)
}
```

## Advanced Usage

### Using RSA for Signing

```go
// Generate RSA key pair
privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
if err != nil {
    panic(err)
}

// Create token with RSA signing
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = "user123"
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

// Sign and get the complete encoded token as a string
tokenString, err := token.SignedString(privateKey)
```

### Token Validation with Custom Claims

```go
type CustomClaims struct {
    UserID string `json:"user_id"`
    jwt.StandardClaims
}

// Parse token with custom claims
token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
    return []byte("your-secret-key"), nil
})

if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
    fmt.Printf("User ID: %v\n", claims.UserID)
    fmt.Printf("Expires at: %v\n", time.Unix(claims.ExpiresAt, 0))
} else {
    fmt.Println("Invalid token:", err)
}
```

## Security Best Practices

1. Always use strong, unique secret keys
2. Set appropriate token expiration times
3. Use HTTPS for all token transmissions
4. Store tokens securely (httpOnly cookies for web)
5. Implement token refresh mechanism
6. Rotate signing keys periodically
7. Validate all token claims on the server

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.