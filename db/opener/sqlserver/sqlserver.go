package sqlserver

import (
	"cmp"
	"fmt"

	"go-slim.dev/infra/db/opener"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type Opener struct{}

func init() {
	opener.Register(Opener{})
}

func (Opener) Name() string {
	return "mysql"
}

func (Opener) Open(c map[string]any) (*gorm.DB, error) {
	return gorm.Open(sqlserver.Open(fmt.Sprintf(
		"sqlserver://%s:%s@%s:%d?database=%s",
		c["user"],
		c["password"],
		cmp.Or(c["host"], "localhost"),
		cmp.Or(c["port"], 9930),
		c["dbname"],
	)), &gorm.Config{})
}
