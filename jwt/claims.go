package jwt

import (
	"maps"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/xid"
)

var _ jwt.Claims = (*Claims)(nil)

type Claims struct {
	Extra              map[string]any   `json:"eta,omitempty"` // 扩展信息
	AllowRefresh       bool             `json:"arf,omitempty"` // 是否允许刷新
	MaxRefreshCount    uint             `json:"mrc,omitempty"` // 允许的最大刷新次数
	RemainRefreshCount uint             `json:"rrc,omitempty"` // 剩余的刷新次数
	Issuer             string           `json:"iss,omitempty"` // 签发人
	Subject            string           `json:"sub,omitempty"` // 主题
	Audience           jwt.ClaimStrings `json:"aud,omitempty"` // 受众
	ExpiresAt          *jwt.NumericDate `json:"exp,omitempty"` // 过期时间
	NotBefore          *jwt.NumericDate `json:"nbf,omitempty"` // 生效时间
	IssuedAt           *jwt.NumericDate `json:"iat,omitempty"` // 签发时间
	ID                 string           `json:"jti,omitempty"` // 令牌编号
}

func (c Claims) Refresh(lifetime time.Duration) Claims {
	now := time.Now()
	c.Extra = maps.Clone(c.Extra)
	c.Audience = slices.Clone(c.Audience)
	c.ExpiresAt = jwt.NewNumericDate(now.Add(lifetime)) // 新的过期时间
	c.NotBefore = jwt.NewNumericDate(now)               // 立即生效
	// 保持原 IssuedAt 不变，如果为 nil 则设置为当前时间
	if c.IssuedAt != nil {
		c.IssuedAt = jwt.NewNumericDate(c.IssuedAt.Time)
	} else {
		c.IssuedAt = jwt.NewNumericDate(now)
	}
	c.ID = xid.New().String()
	// 减少剩余刷新次数，但不能低于0
	if c.RemainRefreshCount > 0 {
		c.RemainRefreshCount--
	}
	return c
}

// GetExpirationTime implements the Claims interface.
func (c Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.ExpiresAt, nil
}

// GetNotBefore implements the Claims interface.
func (c Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.NotBefore, nil
}

// GetIssuedAt implements the Claims interface.
func (c Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.IssuedAt, nil
}

// GetAudience implements the Claims interface.
func (c Claims) GetAudience() (jwt.ClaimStrings, error) {
	return c.Audience, nil
}

// GetIssuer implements the Claims interface.
func (c Claims) GetIssuer() (string, error) {
	return c.Issuer, nil
}

// GetSubject implements the Claims interface.
func (c Claims) GetSubject() (string, error) {
	return c.Subject, nil
}
