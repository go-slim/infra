package jwt

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/xid"
	"go-slim.dev/env"
)

var (
	ErrForbidden       = errors.New("jwt: forbidden")
	ErrInvalidToken    = errors.New("jwt: invalid token")
	ErrTokenExpired    = errors.New("jwt: token is expired")
	ErrTokenNotFound   = errors.New("jwt: token not found")
	ErrNotAllowRefresh = errors.New("jwt: not allow refresh")
	ErrRefreshDeadline = errors.New("jwt: refresh is deadline")
	ErrNotRefreshCount = errors.New("jwt: not refresh count")
)

// Generate 生成令牌
// 当客户端携带这个 token 访问接口提示 token 失效过期，
// 那么则需要携带这个过期的 token 去请求刷新 token 的接口，
// 刷新接口会判断是不是因为 token 过期而失效，如果是， 那么则会解析出这个 token 中的信息，
// 然后判断这个 token 的首次签名时间和当前时间对比，是不是小于刷新 token的时间，如果是，
// 那么就重新颁发 一个 token，但是需要注意的是，重新颁发的这个 token 中记录的首次签名时间还是之前失效的 token 的首次签名时间，
// 也就是首次签名时间不做变化，只更改了这个 token 的有效期。
// 这样就达到了通过一个 token 也可以做刷新的效果。
func Generate(claims Claims) (string, error) {
	if claims.AllowRefresh && claims.MaxRefreshCount == 0 {
		claims.MaxRefreshCount = uint(env.Int("JWT_MAX_REFRESH_COUNT", 10))
	}
	if claims.Issuer == "" {
		claims.Issuer = env.String("JWT_ISSUER")
	}
	if claims.Subject == "" {
		claims.Issuer = env.String("JWT_SUBJECT")
	}
	if claims.Audience == nil {
		claims.Audience = env.List("JWT_AUDIENCE")
		if len(claims.Audience) == 0 {
			claims.Audience = []string{"*"}
		}
	}
	if claims.IssuedAt == nil {
		claims.IssuedAt = jwt.NewNumericDate(time.Now())
	}
	if claims.ExpiresAt == nil {
		ttl := env.Duration("JWT_TTL", time.Hour)
		claims.ExpiresAt = jwt.NewNumericDate(claims.IssuedAt.Add(ttl))
	}
	if claims.NotBefore == nil {
		claims.NotBefore = jwt.NewNumericDate(claims.IssuedAt.Time)
	}
	if claims.ID == "" {
		claims.ID = xid.New().String()
	}
	var signingMethod *jwt.SigningMethodRSA
	switch algo := env.String("JWT_ALGO", "rs256"); strings.ToLower(algo) {
	case "rs256":
		signingMethod = jwt.SigningMethodRS256
	case "rs384":
		signingMethod = jwt.SigningMethodRS384
	case "rs512":
		signingMethod = jwt.SigningMethodRS512
	default:
		return "", fmt.Errorf("jwt: unsupported singing method %q", algo)
	}
	privateKey, err := PrivateKey.Load()
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(signingMethod, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString([]byte(signedToken)), nil
}

// Verify 验证令牌
func Verify(tokenString string) (*Claims, error) {
	bts, err := base64.RawURLEncoding.DecodeString(tokenString)
	if err != nil {
		return nil, err
	}
	var claims Claims
	token, err := jwt.ParseWithClaims(string(bts), &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		return PublicKey.Load()
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return &claims, ErrTokenExpired
		}
		return nil, err
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return &claims, nil
}

func Refresh(tokenString string) (string, *Claims, error) {
	bts, err := base64.RawURLEncoding.DecodeString(tokenString)
	if err != nil {
		return "", nil, err
	}
	var claims Claims
	_, err = jwt.ParseWithClaims(string(bts), &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		return PublicKey.Load()
	})
	// 如果没有错误，说明令牌尚未过期，就无须刷新
	if err == nil {
		return tokenString, &claims, nil
	}
	// 忽略非过期错误
	if !errors.Is(err, jwt.ErrTokenExpired) {
		return "", nil, err
	}
	if !claims.AllowRefresh {
		return "", nil, ErrNotAllowRefresh
	}
	if claims.RemainRefreshCount == 0 {
		return "", nil, ErrNotRefreshCount
	}
	// 计算当前令牌的有效时长
	lifetime := claims.ExpiresAt.Sub(claims.NotBefore.Time)
	// 计算令牌的刷新能力过期时间
	deadline := claims.IssuedAt.Add(lifetime * time.Duration(claims.MaxRefreshCount))
	if time.Now().After(deadline) {
		return "", nil, ErrRefreshDeadline
	}
	// 每次刷新，令牌的有效时长都是一致的
	token, err := Generate(claims.Refresh(lifetime))
	if err != nil {
		return "", nil, err
	}
	return token, &claims, nil
}
