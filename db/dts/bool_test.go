package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var (
	_ driver.Valuer                  = (*NullBool)(nil)
	_ sql.Scanner                    = (*NullBool)(nil)
	_ json.Marshaler                 = (*NullBool)(nil)
	_ json.Unmarshaler               = (*NullBool)(nil)
	_ schema.GormDataTypeInterface   = (*NullBool)(nil)
	_ migrator.GormDataTypeInterface = (*NullBool)(nil)
)
