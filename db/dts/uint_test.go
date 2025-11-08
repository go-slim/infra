package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var (
	_ driver.Valuer                  = (*NullUint)(nil)
	_ sql.Scanner                    = (*NullUint)(nil)
	_ json.Marshaler                 = (*NullUint)(nil)
	_ json.Unmarshaler               = (*NullUint)(nil)
	_ schema.GormDataTypeInterface   = (*NullUint)(nil)
	_ migrator.GormDataTypeInterface = (*NullUint)(nil)

	_ driver.Valuer                  = (*NullUint8)(nil)
	_ sql.Scanner                    = (*NullUint8)(nil)
	_ json.Marshaler                 = (*NullUint8)(nil)
	_ json.Unmarshaler               = (*NullUint8)(nil)
	_ schema.GormDataTypeInterface   = (*NullUint8)(nil)
	_ migrator.GormDataTypeInterface = (*NullUint8)(nil)

	_ driver.Valuer                  = (*NullUint16)(nil)
	_ sql.Scanner                    = (*NullUint16)(nil)
	_ json.Marshaler                 = (*NullUint16)(nil)
	_ json.Unmarshaler               = (*NullUint16)(nil)
	_ schema.GormDataTypeInterface   = (*NullUint16)(nil)
	_ migrator.GormDataTypeInterface = (*NullUint16)(nil)

	_ driver.Valuer                  = (*NullUint32)(nil)
	_ sql.Scanner                    = (*NullUint32)(nil)
	_ json.Marshaler                 = (*NullUint32)(nil)
	_ json.Unmarshaler               = (*NullUint32)(nil)
	_ schema.GormDataTypeInterface   = (*NullUint32)(nil)
	_ migrator.GormDataTypeInterface = (*NullUint32)(nil)

	_ driver.Valuer                  = (*NullUint64)(nil)
	_ sql.Scanner                    = (*NullUint64)(nil)
	_ json.Marshaler                 = (*NullUint64)(nil)
	_ json.Unmarshaler               = (*NullUint64)(nil)
	_ schema.GormDataTypeInterface   = (*NullUint64)(nil)
	_ migrator.GormDataTypeInterface = (*NullUint64)(nil)
)
