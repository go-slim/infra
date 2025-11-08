package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var (
	_ driver.Valuer                  = (*NullInt)(nil)
	_ sql.Scanner                    = (*NullInt)(nil)
	_ json.Marshaler                 = (*NullInt)(nil)
	_ json.Unmarshaler               = (*NullInt)(nil)
	_ schema.GormDataTypeInterface   = (*NullInt)(nil)
	_ migrator.GormDataTypeInterface = (*NullInt)(nil)

	_ driver.Valuer                  = (*NullInt8)(nil)
	_ sql.Scanner                    = (*NullInt8)(nil)
	_ json.Marshaler                 = (*NullInt8)(nil)
	_ json.Unmarshaler               = (*NullInt8)(nil)
	_ schema.GormDataTypeInterface   = (*NullInt8)(nil)
	_ migrator.GormDataTypeInterface = (*NullInt8)(nil)

	_ driver.Valuer                  = (*NullInt16)(nil)
	_ sql.Scanner                    = (*NullInt16)(nil)
	_ json.Marshaler                 = (*NullInt16)(nil)
	_ json.Unmarshaler               = (*NullInt16)(nil)
	_ schema.GormDataTypeInterface   = (*NullInt16)(nil)
	_ migrator.GormDataTypeInterface = (*NullInt16)(nil)

	_ driver.Valuer                  = (*NullInt32)(nil)
	_ sql.Scanner                    = (*NullInt32)(nil)
	_ json.Marshaler                 = (*NullInt32)(nil)
	_ json.Unmarshaler               = (*NullInt32)(nil)
	_ schema.GormDataTypeInterface   = (*NullInt32)(nil)
	_ migrator.GormDataTypeInterface = (*NullInt32)(nil)

	_ driver.Valuer                  = (*NullInt64)(nil)
	_ sql.Scanner                    = (*NullInt64)(nil)
	_ json.Marshaler                 = (*NullInt64)(nil)
	_ json.Unmarshaler               = (*NullInt64)(nil)
	_ schema.GormDataTypeInterface   = (*NullInt64)(nil)
	_ migrator.GormDataTypeInterface = (*NullInt64)(nil)
)
