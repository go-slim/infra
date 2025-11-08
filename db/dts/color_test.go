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
	_ driver.Valuer                  = (*Color)(nil)
	_ sql.Scanner                    = (*Color)(nil)
	_ fmt.Stringer                   = (*Color)(nil)
	_ json.Marshaler                 = (*Color)(nil)
	_ json.Unmarshaler               = (*Color)(nil)
	_ schema.GormDataTypeInterface   = (*Color)(nil)
	_ migrator.GormDataTypeInterface = (*Color)(nil)

	_ driver.Valuer                  = (*NullColor)(nil)
	_ sql.Scanner                    = (*NullColor)(nil)
	_ json.Marshaler                 = (*NullColor)(nil)
	_ json.Unmarshaler               = (*NullColor)(nil)
	_ schema.GormDataTypeInterface   = (*NullColor)(nil)
	_ migrator.GormDataTypeInterface = (*NullColor)(nil)
)
