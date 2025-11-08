package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestClaims_Refresh(t *testing.T) {
	tests := []struct {
		name           string
		originalClaims Claims
		lifetime       time.Duration
		expectedClaims Claims
	}{
		{
			name: "refresh basic claims",
			originalClaims: Claims{
				Extra:              map[string]any{"role": "admin"},
				AllowRefresh:       true,
				MaxRefreshCount:    5,
				RemainRefreshCount: 3,
				Issuer:             "test-issuer",
				Subject:            "test-subject",
				Audience:           jwt.ClaimStrings{"test-audience"},
				ExpiresAt:          jwt.NewNumericDate(time.Now().Add(time.Hour)),
				NotBefore:          jwt.NewNumericDate(time.Now().Add(-time.Minute)),
				IssuedAt:           jwt.NewNumericDate(time.Now().Add(-time.Minute)),
				ID:                 "test-id",
			},
			lifetime: time.Hour * 2,
			expectedClaims: Claims{
				Extra:              map[string]any{"role": "admin"},
				AllowRefresh:       true,
				MaxRefreshCount:    5,
				RemainRefreshCount: 2, // decremented by 1
				Issuer:             "test-issuer",
				Subject:            "test-subject",
				Audience:           jwt.ClaimStrings{"test-audience"},
				NotBefore:          nil, // will be set to now
				IssuedAt:           nil, // will be preserved from original
				ID:                 "",  // will be set to new ID
			},
		},
		{
			name: "refresh with zero remain count",
			originalClaims: Claims{
				Extra:              map[string]any{},
				AllowRefresh:       true,
				MaxRefreshCount:    5,
				RemainRefreshCount: 0,
				Issuer:             "test-issuer",
			},
			lifetime: time.Hour,
			expectedClaims: Claims{
				Extra:              map[string]any{},
				AllowRefresh:       true,
				MaxRefreshCount:    5,
				RemainRefreshCount: 0, // stays at 0 due to max()
				Issuer:             "test-issuer",
			},
		},
		{
			name: "refresh with nil extra",
			originalClaims: Claims{
				Extra:              nil,
				AllowRefresh:       false,
				MaxRefreshCount:    0,
				RemainRefreshCount: 0,
				Issuer:             "test-issuer",
			},
			lifetime: time.Hour,
			expectedClaims: Claims{
				Extra:              nil,
				AllowRefresh:       false,
				MaxRefreshCount:    0,
				RemainRefreshCount: 0,
				Issuer:             "test-issuer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalIssuedAt := tt.originalClaims.IssuedAt

			result := tt.originalClaims.Refresh(tt.lifetime)

			// Check that fields are properly updated
			assert.Equal(t, tt.expectedClaims.Extra, result.Extra)
			assert.Equal(t, tt.expectedClaims.AllowRefresh, result.AllowRefresh)
			assert.Equal(t, tt.expectedClaims.MaxRefreshCount, result.MaxRefreshCount)
			assert.Equal(t, tt.expectedClaims.RemainRefreshCount, result.RemainRefreshCount)
			assert.Equal(t, tt.expectedClaims.Issuer, result.Issuer)
			assert.Equal(t, tt.expectedClaims.Subject, result.Subject)
			assert.Equal(t, tt.expectedClaims.Audience, result.Audience)

			// Check that timestamps are properly set
			assert.True(t, result.ExpiresAt != nil, "ExpiresAt should be set")
			assert.True(t, result.NotBefore != nil, "NotBefore should be set")
			assert.True(t, result.IssuedAt != nil, "IssuedAt should be set")

			// NotBefore should be set to now (allow small tolerance)
			now := time.Now()
			assert.WithinDuration(t, now, result.NotBefore.Time, time.Second)

			// ExpiresAt should be NotBefore + lifetime
			expectedExpiresAt := result.NotBefore.Time.Add(tt.lifetime)
			assert.WithinDuration(t, expectedExpiresAt, result.ExpiresAt.Time, time.Second)

			// IssuedAt should be preserved from original
			if originalIssuedAt != nil {
				assert.Equal(t, originalIssuedAt.Time, result.IssuedAt.Time)
			}

			// ID should be new and not empty
			assert.NotEmpty(t, result.ID)
			assert.NotEqual(t, tt.originalClaims.ID, result.ID)

			// Extra map should be cloned, not the same reference
			if tt.originalClaims.Extra != nil && result.Extra != nil {
				// Modify result to ensure it's a different reference
				result.Extra["_test"] = "value"
				_, existsInOriginal := tt.originalClaims.Extra["_test"]
				assert.False(t, existsInOriginal, "Extra map should be cloned")
				delete(result.Extra, "_test")
			}

			// Audience slice should be cloned, not the same reference
			if len(tt.originalClaims.Audience) > 0 && len(result.Audience) > 0 {
				// Modify result to ensure it's a different reference
				originalFirstAud := result.Audience[0]
				result.Audience[0] = "modified"
				assert.NotEqual(t, "modified", tt.originalClaims.Audience[0], "Audience slice should be cloned")
				result.Audience[0] = originalFirstAud
			}
		})
	}
}

func TestClaims_InterfaceImplementation(t *testing.T) {
	claims := Claims{
		Issuer:    "test-issuer",
		Subject:   "test-subject",
		Audience:  jwt.ClaimStrings{"aud1", "aud2"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	// Test that Claims implements jwt.Claims interface
	var _ jwt.Claims = &claims

	// Test GetExpirationTime
	expTime, err := claims.GetExpirationTime()
	assert.NoError(t, err)
	assert.Equal(t, claims.ExpiresAt, expTime)

	// Test GetNotBefore
	notBefore, err := claims.GetNotBefore()
	assert.NoError(t, err)
	assert.Equal(t, claims.NotBefore, notBefore)

	// Test GetIssuedAt
	issuedAt, err := claims.GetIssuedAt()
	assert.NoError(t, err)
	assert.Equal(t, claims.IssuedAt, issuedAt)

	// Test GetAudience
	audience, err := claims.GetAudience()
	assert.NoError(t, err)
	assert.Equal(t, claims.Audience, audience)

	// Test GetIssuer
	issuer, err := claims.GetIssuer()
	assert.NoError(t, err)
	assert.Equal(t, claims.Issuer, issuer)

	// Test GetSubject
	subject, err := claims.GetSubject()
	assert.NoError(t, err)
	assert.Equal(t, claims.Subject, subject)
}

func TestClaims_NilFields(t *testing.T) {
	claims := Claims{}

	// Test all methods with nil fields
	expTime, err := claims.GetExpirationTime()
	assert.NoError(t, err)
	assert.Nil(t, expTime)

	notBefore, err := claims.GetNotBefore()
	assert.NoError(t, err)
	assert.Nil(t, notBefore)

	issuedAt, err := claims.GetIssuedAt()
	assert.NoError(t, err)
	assert.Nil(t, issuedAt)

	audience, err := claims.GetAudience()
	assert.NoError(t, err)
	assert.Empty(t, audience)

	issuer, err := claims.GetIssuer()
	assert.NoError(t, err)
	assert.Empty(t, issuer)

	subject, err := claims.GetSubject()
	assert.NoError(t, err)
	assert.Empty(t, subject)
}
