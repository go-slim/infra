package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var (
	_ driver.Valuer                  = (*Slice[any])(nil)
	_ sql.Scanner                    = (*Slice[any])(nil)
	_ json.Marshaler                 = (*Slice[any])(nil)
	_ json.Unmarshaler               = (*Slice[any])(nil)
	_ schema.GormDataTypeInterface   = (*Slice[any])(nil)
	_ migrator.GormDataTypeInterface = (*Slice[any])(nil)
)
