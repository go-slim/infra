package sqlite

import (
	"go-slim.dev/infra/db/opener"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Opener struct{}

func init() {
	opener.Register(Opener{})
}

func (Opener) Name() string {
	return "sqlite"
}

func (Opener) Open(c map[string]any) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(c["dbname"].(string)), &gorm.Config{})
}
