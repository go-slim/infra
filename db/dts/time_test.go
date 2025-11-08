package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var (
	_ driver.Valuer                  = (*Date)(nil)
	_ sql.Scanner                    = (*Date)(nil)
	_ fmt.Stringer                   = (*Date)(nil)
	_ json.Marshaler                 = (*Date)(nil)
	_ json.Unmarshaler               = (*Date)(nil)
	_ schema.GormDataTypeInterface   = (*Date)(nil)
	_ migrator.GormDataTypeInterface = (*Date)(nil)

	_ driver.Valuer                  = (*NullDate)(nil)
	_ sql.Scanner                    = (*NullDate)(nil)
	_ json.Marshaler                 = (*NullDate)(nil)
	_ json.Unmarshaler               = (*NullDate)(nil)
	_ schema.GormDataTypeInterface   = (*NullDate)(nil)
	_ migrator.GormDataTypeInterface = (*NullDate)(nil)

	_ driver.Valuer                  = (*NullTime)(nil)
	_ sql.Scanner                    = (*NullTime)(nil)
	_ json.Marshaler                 = (*NullTime)(nil)
	_ json.Unmarshaler               = (*NullTime)(nil)
	_ schema.GormDataTypeInterface   = (*NullTime)(nil)
	_ migrator.GormDataTypeInterface = (*NullTime)(nil)
)
