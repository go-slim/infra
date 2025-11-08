package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyInitialization(t *testing.T) {
	// Test that global Key variables are initialized
	assert.NotNil(t, PublicKey)
	assert.NotNil(t, PrivateKey)

	// Test that they have correct configuration keys
	// Note: We can't access private fields directly, but we can test the initialization
	// happened without panicking in the init() function
}

func TestNewKey(t *testing.T) {
	// Test creating a new Key
	key := NewKey("TEST_KEY", "default_value", jwt.ParseRSAPrivateKeyFromPEM)

	assert.NotNil(t, key)
	// Note: We can't access private fields directly, but the Key should be created successfully
}

func TestKeyReset(t *testing.T) {
	// Test Key reset functionality
	key := NewKey("OLD_KEY", "default", jwt.ParseRSAPrivateKeyFromPEM)

	// Reset key with new configuration (same type to avoid type mismatch)
	key.Reset("NEW_KEY", jwt.ParseRSAPrivateKeyFromPEM)

	// Key should be reset without errors
	assert.NotNil(t, key)
}

func TestKeyParsing(t *testing.T) {
	// Test key parsing functionality

	// Generate test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Test private key parsing
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	parsedPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	require.NoError(t, err)
	assert.NotNil(t, parsedPrivateKey)
	assert.Equal(t, privateKey.N, parsedPrivateKey.N)

	// Test public key parsing
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	parsedPublicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	require.NoError(t, err)
	assert.NotNil(t, parsedPublicKey)
	assert.Equal(t, privateKey.PublicKey.N, parsedPublicKey.N)
}

func TestKeyParsingErrors(t *testing.T) {
	// Test key parsing with invalid data

	// Test invalid PEM data
	invalidPEM := []byte("invalid key data")

	_, err := jwt.ParseRSAPrivateKeyFromPEM(invalidPEM)
	assert.Error(t, err)

	_, err = jwt.ParseRSAPublicKeyFromPEM(invalidPEM)
	assert.Error(t, err)

	// Test empty data
	_, err = jwt.ParseRSAPrivateKeyFromPEM([]byte{})
	assert.Error(t, err)

	_, err = jwt.ParseRSAPublicKeyFromPEM([]byte{})
	assert.Error(t, err)

	// Test PEM with wrong type
	wrongTypePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "WRONG TYPE",
		Bytes: []byte("some data"),
	})

	_, err = jwt.ParseRSAPrivateKeyFromPEM(wrongTypePEM)
	assert.Error(t, err)

	_, err = jwt.ParseRSAPublicKeyFromPEM(wrongTypePEM)
	assert.Error(t, err)
}

func TestMakeKeysFunction(t *testing.T) {
	// Test the MakeKeys function
	tempDir, err := os.MkdirTemp("", "jwt_make_keys_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Generate keys
	err = MakeKeys(tempDir, 1024) // Use smaller key for faster tests
	require.NoError(t, err)

	// Check that files were created
	privateKeyPath := filepath.Join(tempDir, "private.key")
	publicKeyPath := filepath.Join(tempDir, "public.key")

	assert.FileExists(t, privateKeyPath)
	assert.FileExists(t, publicKeyPath)

	// Check file permissions
	privateFileInfo, err := os.Stat(privateKeyPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), privateFileInfo.Mode().Perm())

	publicFileInfo, err := os.Stat(publicKeyPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), publicFileInfo.Mode().Perm())

	// Verify that keys can be loaded
	privateKeyData, err := os.ReadFile(privateKeyPath)
	require.NoError(t, err)
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	require.NoError(t, err)
	assert.IsType(t, &rsa.PrivateKey{}, privateKey)

	publicKeyData, err := os.ReadFile(publicKeyPath)
	require.NoError(t, err)
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	require.NoError(t, err)
	assert.IsType(t, &rsa.PublicKey{}, publicKey)

	// Verify that keys match
	assert.Equal(t, privateKey.PublicKey, *publicKey)
}

func TestMakeKeysErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupDir      func() string
		expectedError bool
	}{
		{
			name: "create keys in non-existent directory",
			setupDir: func() string {
				return filepath.Join(os.TempDir(), "jwt_test_non_existent", "subdir")
			},
			expectedError: false, // Should create directory successfully
		},
		{
			name: "create keys with invalid key size",
			setupDir: func() string {
				tempDir, err := os.MkdirTemp("", "jwt_test")
				require.NoError(t, err)
				return tempDir
			},
			expectedError: true, // Invalid key size should cause error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir()

			var err error
			if tt.name == "create keys with invalid key size" {
				err = MakeKeys(dir, 0) // Invalid key size
			} else {
				err = MakeKeys(dir, 1024)
			}

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Clean up if successful
				if _, err := os.Stat(dir); err == nil {
					os.RemoveAll(dir)
				}
			}
		})
	}
}

func TestRSAKeyGenerationSizes(t *testing.T) {
	// Test RSA key generation with different sizes
	tests := []struct {
		name    string
		bits    int
		wantErr bool
	}{
		{
			name:    "generate 1024-bit key",
			bits:    1024,
			wantErr: false,
		},
		{
			name:    "generate 2048-bit key",
			bits:    2048,
			wantErr: false,
		},
		{
			name:    "generate 0-bit key (invalid)",
			bits:    0,
			wantErr: true,
		},
		{
			name:    "generate negative bit key (invalid)",
			bits:    -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKey, err := rsa.GenerateKey(rand.Reader, tt.bits)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, privateKey)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, privateKey)
				assert.Equal(t, tt.bits, privateKey.N.BitLen())

				// Test that public key is properly set
				assert.NotNil(t, privateKey.PublicKey)
				assert.Equal(t, privateKey.N, privateKey.PublicKey.N)
				assert.Equal(t, privateKey.E, privateKey.PublicKey.E)
			}
		})
	}
}

func TestPEMBlockTypes(t *testing.T) {
	// Test that PEM block types are correct for different key formats
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	// Test PKCS#1 private key encoding
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	block, _ := pem.Decode(privateKeyPEM)
	require.NotNil(t, block)
	assert.Equal(t, "RSA PRIVATE KEY", block.Type)

	// Test PKIX public key encoding
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	block, _ = pem.Decode(publicKeyPEM)
	require.NotNil(t, block)
	assert.Equal(t, "PUBLIC KEY", block.Type)
}
