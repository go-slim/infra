package mysql

import (
	"cmp"
	"fmt"

	"go-slim.dev/infra/db/opener"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Opener struct{}

func init() {
	opener.Register(Opener{})
}

func (m Opener) Name() string {
	return "mysql"
}

func (m Opener) Open(c map[string]any) (*gorm.DB, error) {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	return gorm.Open(mysql.Open(fmt.Sprintf(
		"%s:%s@%s(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
		c["user"],
		c["pass"],
		cmp.Or(c["network"], "tcp"),
		c["host"],
		c["port"],
		c["dbname"],
		cmp.Or(c["charset"], "utf8mb4"),
		cmp.Or(c["timezone"], "Local"),
	)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
}
