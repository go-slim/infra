package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/golang-jwt/jwt/v5"
	"go-slim.dev/env"
	"go-slim.dev/misc"
)

var (
	PublicKey  *Key[rsa.PublicKey]
	PrivateKey *Key[rsa.PrivateKey]
)

func init() {
	PublicKey = NewKey(
		"JWT_PUBLIC_KEY",
		"file://public.key",
		jwt.ParseRSAPublicKeyFromPEM,
	)
	PrivateKey = NewKey(
		"JWT_PRIVATE_KEY",
		"file://private.key",
		jwt.ParseRSAPrivateKeyFromPEM,
	)
}

type Key[T any] struct {
	configKey          string
	defaultConfigValue string
	parseKey           func([]byte) (*T, error)
	once               sync.Once
	store              atomic.Pointer[T]
	err                atomic.Value
}

func NewKey[T any](configKey, defaultConfigValue string, parseKey func([]byte) (*T, error)) *Key[T] {
	return &Key[T]{
		configKey:          configKey,
		defaultConfigValue: defaultConfigValue,
		parseKey:           parseKey,
	}
}

func (k *Key[T]) Reset(configKey string, parseKey func([]byte) (*T, error)) {
	*k = Key[T]{
		configKey: configKey,
		parseKey:  parseKey,
	}
}

func (k *Key[T]) Load() (*T, error) {
	if v := k.store.Load(); v != nil {
		return v, nil
	}
	k.once.Do(func() {
		rawKey := env.String(k.configKey, k.defaultConfigValue)
		if rawKey == "" {
			k.err.Store(fmt.Errorf(`jwt: missing key "%s"`, k.configKey))
			return
		}
		var keyBuf []byte
		if strings.HasPrefix(rawKey, "file://") {
			file, err := os.Open(env.Path(rawKey))
			if err != nil {
				k.err.Store(err)
				return
			}
			defer file.Close()
			keyBuf, err = io.ReadAll(file)
			if err != nil {
				k.err.Store(err)
				return
			}
		} else {
			keyBuf = misc.StringToBytes(rawKey)
		}
		if v, e := k.parseKey(keyBuf); e != nil {
			k.err.Store(fmt.Errorf("jwt: failed to parse key, error: %w", e))
		} else {
			k.store.Store(v)
		}
	})
	if err := k.err.Load(); err != nil {
		return nil, err.(error)
	}
	return k.store.Load(), nil
}

func MakeKeys(dir string, bits int) error {
	// Generate the key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// Save private key
	privateKeyPath := dir + "/private.key"
	privateKeyFile, err := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return err
	}

	// Save public key
	publicKeyPath := dir + "/public.key"
	publicKeyFile, err := os.OpenFile(publicKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer publicKeyFile.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	if err := pem.Encode(publicKeyFile, publicKeyPEM); err != nil {
		return err
	}

	return nil
}
