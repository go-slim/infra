package db

import (
	"go-slim.dev/env"
	"go-slim.dev/infra/db/opener"
	"gorm.io/gorm"
)

type Opener interface {
	Name() string
	Open(map[string]any) (*gorm.DB, error)
}

type Config struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	Schema   string
	Network  string
	Charset  string
	TimeZone string
	SSLMode  string
}

func Open() (*gorm.DB, error) {
	return OpenWithEnvironment(env.Default())
}

func OpenWithEnvironment(e env.Environ) (*gorm.DB, error) {
	return OpenWithConfig(&Config{
		Driver:   e.String("DB_DRIVER"),
		Host:     e.String("DB_HOST"),
		Port:     e.Int("DB_PORT"),
		User:     e.String("DB_USER"),
		Password: e.String("DB_PASSWORD"),
		DBName:   e.String("DB_DATABASE"),
		Schema:   e.String("DB_SCHEMA"),
		Network:  e.String("DB_NETWORK"),
		Charset:  e.String("DB_CHARSET"),
		TimeZone: e.String("DB_TIMEZONE"),
		SSLMode:  e.String("DB_SSLMODE"),
	})
}

func OpenWithConfig(c *Config) (*gorm.DB, error) {
	if c == nil {
		return nil, ErrConfigRequired
	}

	if c.Driver == "" {
		return nil, ErrDriverRequired
	}

	return opener.Open(c.Driver, map[string]any{
		"user":     c.User,
		"pass":     c.Password,
		"network":  c.Network,
		"host":     c.Host,
		"port":     c.Port,
		"dbname":   c.DBName,
		"charset":  c.Charset,
		"timezone": c.TimeZone,
		"sslmode":  c.SSLMode,
	})
}
