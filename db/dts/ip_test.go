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
	_ driver.Valuer                  = (*IP)(nil)
	_ sql.Scanner                    = (*IP)(nil)
	_ fmt.Stringer                   = (*IP)(nil)
	_ json.Marshaler                 = (*IP)(nil)
	_ json.Unmarshaler               = (*IP)(nil)
	_ schema.GormDataTypeInterface   = (*IP)(nil)
	_ migrator.GormDataTypeInterface = (*IP)(nil)

	_ driver.Valuer                  = (*NullIP)(nil)
	_ sql.Scanner                    = (*NullIP)(nil)
	_ json.Marshaler                 = (*NullIP)(nil)
	_ json.Unmarshaler               = (*NullIP)(nil)
	_ schema.GormDataTypeInterface   = (*NullIP)(nil)
	_ migrator.GormDataTypeInterface = (*NullIP)(nil)
)
