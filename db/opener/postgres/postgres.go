package postgres

import (
	"cmp"
	"fmt"

	"go-slim.dev/infra/db/opener"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Opener struct{}

func init() {
	opener.Register(Opener{})
}

func (m Opener) Name() string {
	return "postgres"
}

func (m Opener) Open(c map[string]any) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cmp.Or(c["host"], "localhost"),
		cmp.Or(c["user"], "postgres"),
		cmp.Or(c["password"], "postgres"),
		cmp.Or(c["dbname"], "postgres"),
		cmp.Or(c["port"], 5432),
		cmp.Or(c["sslmode"], "disable"),
		cmp.Or(c["timezone"], "Asia/Shanghai"),
	)), &gorm.Config{})
}
