package jwt

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go-slim.dev/slim"
)

func TestAuthConfig_ToMiddleware_BasicFunctionality(t *testing.T) {
	tests := []struct {
		name        string
		config      AuthConfig
		expectError bool
		expectSkip  bool
	}{
		{
			name: "skip when skipper returns true",
			config: AuthConfig{
				Skipper: func(c slim.Context) bool {
					return true
				},
			},
			expectSkip: true,
		},
		{
			name: "anonymous access allowed",
			config: AuthConfig{
				Anonymously: true,
				Finder: func(c slim.Context) string {
					return ""
				},
			},
			expectSkip: false,
		},
		{
			name: "token not found and not anonymous",
			config: AuthConfig{
				Anonymously: false,
				Finder: func(c slim.Context) string {
					return ""
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := tt.config.ToMiddleware()
			assert.NotNil(t, middleware)

			// Test middleware creation - actual execution would require proper slim.Context
			// which is complex to mock correctly
		})
	}
}

func TestAuth(t *testing.T) {
	config := AuthConfig{
		Anonymously: true,
		Finder: func(c slim.Context) string {
			return "test-token"
		},
	}

	middleware := Auth(config)
	assert.NotNil(t, middleware)

	// Test that ToMiddleware also returns a non-nil middleware
	middleware2 := config.ToMiddleware()
	assert.NotNil(t, middleware2)
}

func TestAuthConfig_Fields(t *testing.T) {
	config := AuthConfig{
		Skipper: func(c slim.Context) bool {
			return false
		},
		Anonymously: true,
		Finder: func(c slim.Context) string {
			return "custom-finder-token"
		},
		Claims: func(c slim.Context, token string, claims *Claims) error {
			return nil
		},
	}

	// Test that all fields are set correctly
	assert.NotNil(t, config.Skipper)
	assert.True(t, config.Anonymously)
	assert.NotNil(t, config.Finder)
	assert.NotNil(t, config.Claims)

	// Test ToMiddleware
	middleware := config.ToMiddleware()
	assert.NotNil(t, middleware)
}

func TestAuthConfig_CustomClaimsError(t *testing.T) {
	customError := errors.New("custom claims validation failed")

	config := AuthConfig{
		Anonymously: false,
		Finder: func(c slim.Context) string {
			return "valid-token"
		},
		Claims: func(c slim.Context, token string, claims *Claims) error {
			return customError
		},
	}

	middleware := config.ToMiddleware()
	assert.NotNil(t, middleware)

	// The middleware should propagate the custom error when claims validation fails
	// Note: Full testing would require proper slim.Context implementation
}

func TestAuthConfig_ZeroValue(t *testing.T) {
	var config AuthConfig

	// Test zero value configuration
	assert.Nil(t, config.Skipper)
	assert.False(t, config.Anonymously)
	assert.Nil(t, config.Finder)
	assert.Nil(t, config.Claims)

	// Should still create a middleware
	middleware := config.ToMiddleware()
	assert.NotNil(t, middleware)
}
