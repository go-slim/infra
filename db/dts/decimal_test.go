package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var (
	_ driver.Valuer                  = (*NullDecimal)(nil)
	_ sql.Scanner                    = (*NullDecimal)(nil)
	_ json.Marshaler                 = (*NullDecimal)(nil)
	_ json.Unmarshaler               = (*NullDecimal)(nil)
	_ schema.GormDataTypeInterface   = (*NullDecimal)(nil)
	_ migrator.GormDataTypeInterface = (*NullDecimal)(nil)
)
