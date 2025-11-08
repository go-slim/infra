package jwt

import (
	"go-slim.dev/slim"
)

type AuthConfig struct {
	Skipper     func(c slim.Context) bool
	Anonymously bool // 是否允许匿名访问
	Finder      Finder
	Claims      func(c slim.Context, token string, claims *Claims) error
}

func (config AuthConfig) ToMiddleware() slim.MiddlewareFunc {
	finder := config.Finder
	if finder == nil {
		finder = DefaultFinder
	}
	return func(c slim.Context, next slim.HandlerFunc) error {
		if config.Skipper != nil && config.Skipper(c) {
			return next(c)
		}
		tokenString := finder(c)
		if tokenString == "" {
			if config.Anonymously {
				return next(c)
			}
			return ErrTokenNotFound
		}
		claims, err := Verify(tokenString)
		if err != nil {
			return err
		}
		if config.Claims != nil {
			err = config.Claims(c, tokenString, claims)
			if err != nil {
				return err
			}
		}
		if c.Written() {
			return nil
		}
		c.Set("jwt:token", tokenString)
		c.Set("jwt:claims", claims)
		return next(c)
	}
}

func Auth(c AuthConfig) slim.MiddlewareFunc {
	return c.ToMiddleware()
}
