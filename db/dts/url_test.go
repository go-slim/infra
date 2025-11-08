package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var (
	_ driver.Valuer                  = (*URL)(nil)
	_ sql.Scanner                    = (*URL)(nil)
	_ json.Marshaler                 = (*URL)(nil)
	_ json.Unmarshaler               = (*URL)(nil)
	_ schema.GormDataTypeInterface   = (*URL)(nil)
	_ migrator.GormDataTypeInterface = (*URL)(nil)

	_ driver.Valuer                  = (*NullURL)(nil)
	_ sql.Scanner                    = (*NullURL)(nil)
	_ json.Marshaler                 = (*NullURL)(nil)
	_ json.Unmarshaler               = (*NullURL)(nil)
	_ schema.GormDataTypeInterface   = (*NullURL)(nil)
	_ migrator.GormDataTypeInterface = (*NullURL)(nil)
)
