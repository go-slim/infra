package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var (
	_ driver.Valuer                  = (*NullFloat32)(nil)
	_ sql.Scanner                    = (*NullFloat32)(nil)
	_ json.Marshaler                 = (*NullFloat32)(nil)
	_ json.Unmarshaler               = (*NullFloat32)(nil)
	_ schema.GormDataTypeInterface   = (*NullFloat32)(nil)
	_ migrator.GormDataTypeInterface = (*NullFloat32)(nil)

	_ driver.Valuer                  = (*NullFloat64)(nil)
	_ sql.Scanner                    = (*NullFloat64)(nil)
	_ json.Marshaler                 = (*NullFloat64)(nil)
	_ json.Unmarshaler               = (*NullFloat64)(nil)
	_ schema.GormDataTypeInterface   = (*NullFloat64)(nil)
	_ migrator.GormDataTypeInterface = (*NullFloat64)(nil)
)
