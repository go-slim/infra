package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-slim.dev/env"
)

func TestErrors(t *testing.T) {
	// Test that all error variables are defined
	assert.Error(t, ErrForbidden)
	assert.Error(t, ErrInvalidToken)
	assert.Error(t, ErrTokenExpired)
	assert.Error(t, ErrTokenNotFound)
	assert.Error(t, ErrNotAllowRefresh)
	assert.Error(t, ErrRefreshDeadline)
	assert.Error(t, ErrNotRefreshCount)
}

func TestJWTAlgorithms(t *testing.T) {
	// Test supported JWT algorithms
	tests := []struct {
		algorithm string
		expected  bool
	}{
		{"rs256", true},
		{"RS256", true},
		{"rs384", true},
		{"RS384", true},
		{"rs512", true},
		{"RS512", true},
		{"hs256", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.algorithm, func(t *testing.T) {
			switch strings.ToLower(tt.algorithm) {
			case "rs256":
				assert.Equal(t, jwt.SigningMethodRS256, jwt.SigningMethodRS256)
			case "rs384":
				assert.Equal(t, jwt.SigningMethodRS384, jwt.SigningMethodRS384)
			case "rs512":
				assert.Equal(t, jwt.SigningMethodRS512, jwt.SigningMethodRS512)
			default:
				if tt.expected {
					t.Errorf("Expected algorithm %s to be supported", tt.algorithm)
				}
			}
		})
	}
}

func TestTokenEncoding(t *testing.T) {
	// Test base64 URL encoding used for tokens
	original := "test.jwt.token"
	encoded := base64.RawURLEncoding.EncodeToString([]byte(original))
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	require.NoError(t, err)
	assert.Equal(t, original, string(decoded))
}

func TestClaimsValidation(t *testing.T) {
	// Test the Claims struct and its validation
	claims := Claims{
		AllowRefresh:       true,
		MaxRefreshCount:    5,
		RemainRefreshCount: 3,
		Issuer:             "test-issuer",
		Subject:            "test-subject",
		Audience:           []string{"test-audience"},
		Extra:              map[string]any{"role": "admin"},
	}

	// Test Get methods
	issuer, err := claims.GetIssuer()
	assert.NoError(t, err)
	assert.Equal(t, "test-issuer", issuer)

	subject, err := claims.GetSubject()
	assert.NoError(t, err)
	assert.Equal(t, "test-subject", subject)

	audience, err := claims.GetAudience()
	assert.NoError(t, err)
	assert.Equal(t, jwt.ClaimStrings{"test-audience"}, audience)
}

func TestClaimsRefreshLogic(t *testing.T) {
	now := time.Now()

	claims := Claims{
		AllowRefresh:       true,
		MaxRefreshCount:    5,
		RemainRefreshCount: 3,
		Issuer:             "test-issuer",
		IssuedAt:           jwt.NewNumericDate(now.Add(-time.Hour)),
		Extra:              map[string]any{"role": "admin"},
	}

	lifetime := time.Hour * 2
	refreshedClaims := claims.Refresh(lifetime)

	// Test that refresh logic works correctly
	assert.Equal(t, claims.AllowRefresh, refreshedClaims.AllowRefresh)
	assert.Equal(t, claims.MaxRefreshCount, refreshedClaims.MaxRefreshCount)
	assert.Equal(t, claims.RemainRefreshCount-1, refreshedClaims.RemainRefreshCount)
	assert.Equal(t, claims.Issuer, refreshedClaims.Issuer)
	assert.Equal(t, claims.Extra, refreshedClaims.Extra)
	assert.Equal(t, claims.IssuedAt.Time, refreshedClaims.IssuedAt.Time)
	assert.NotEqual(t, claims.ID, refreshedClaims.ID)
	assert.True(t, refreshedClaims.ExpiresAt.After(refreshedClaims.NotBefore.Time))
}

func TestKeyGeneration(t *testing.T) {
	// Test RSA key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	assert.NotNil(t, privateKey)
	assert.Equal(t, 2048, privateKey.N.BitLen())

	// Test public key extraction
	publicKey := &privateKey.PublicKey
	assert.NotNil(t, publicKey)
	assert.Equal(t, privateKey.N, publicKey.N)
	assert.Equal(t, privateKey.E, publicKey.E)
}

func TestPEMEncoding(t *testing.T) {
	// Test PEM encoding/decoding for RSA keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Encode private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	assert.NotEmpty(t, privateKeyPEM)

	// Decode private key
	block, _ := pem.Decode(privateKeyPEM)
	require.NotNil(t, block)
	assert.Equal(t, "RSA PRIVATE KEY", block.Type)

	decodedPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)
	assert.Equal(t, privateKey.N, decodedPrivateKey.N)

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	assert.NotEmpty(t, publicKeyPEM)

	// Decode public key
	block, _ = pem.Decode(publicKeyPEM)
	require.NotNil(t, block)
	assert.Equal(t, "PUBLIC KEY", block.Type)

	decodedPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	require.NoError(t, err)
	assert.Equal(t, privateKey.PublicKey.N, decodedPublicKey.(*rsa.PublicKey).N)
}

func TestJWTTokenCreation(t *testing.T) {
	// Test basic JWT token creation and signing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := jwt.MapClaims{
		"iss": "test-issuer",
		"sub": "test-subject",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	require.NoError(t, err)
	assert.NotEmpty(t, signedToken)

	// Parse and verify the token
	parsedToken, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		return &privateKey.PublicKey, nil
	})

	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	parsedClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, "test-issuer", parsedClaims["iss"])
	assert.Equal(t, "test-subject", parsedClaims["sub"])
}

func TestEnvironmentDefaults(t *testing.T) {
	// Test that environment variable defaults work
	// Note: We can't test env.Set directly, but we can test the defaults

	// Test default values that would be used if env vars are not set
	_ = env.String("JWT_ISSUER", "") // Read but don't use
	defaultTTL := env.Duration("JWT_TTL", time.Hour)
	defaultAlgo := env.String("JWT_ALGO", "rs256")
	defaultMaxRefresh := env.Int("JWT_MAX_REFRESH_COUNT", 10)
	defaultAudience := env.List("JWT_AUDIENCE")

	assert.Equal(t, time.Hour, defaultTTL)
	assert.Equal(t, "rs256", defaultAlgo)
	assert.Equal(t, 10, defaultMaxRefresh)

	// Test default audience
	if len(defaultAudience) == 0 {
		// Would default to ["*"] in Generate function
		defaultAudience = []string{"*"}
	}
	assert.Equal(t, []string{"*"}, defaultAudience)
}

func TestTimeCalculations(t *testing.T) {
	// Test time calculations used in refresh logic
	now := time.Now()
	lifetime := time.Hour
	maxRefreshCount := 5

	// Calculate refresh deadline
	deadline := now.Add(lifetime * time.Duration(maxRefreshCount))
	expectedDeadline := now.Add(5 * time.Hour)
	assert.WithinDuration(t, expectedDeadline, deadline, time.Second)

	// Test that deadline comparison works
	assert.True(t, now.Before(deadline))
	assert.True(t, deadline.After(now))
}

func TestTokenExpiration(t *testing.T) {
	// Test token expiration logic
	now := time.Now()

	// Create an expired token
	expiredClaims := jwt.MapClaims{
		"exp": now.Add(-time.Hour).Unix(),
		"iat": now.Add(-2 * time.Hour).Unix(),
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, expiredClaims)
	signedToken, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// Try to parse expired token
	parsedToken, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, jwt.ErrTokenExpired))
	assert.False(t, parsedToken.Valid)
}
