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
	_ driver.Valuer                  = (*Phone)(nil)
	_ sql.Scanner                    = (*Phone)(nil)
	_ fmt.Stringer                   = (*Phone)(nil)
	_ json.Marshaler                 = (*Phone)(nil)
	_ json.Unmarshaler               = (*Phone)(nil)
	_ schema.GormDataTypeInterface   = (*Phone)(nil)
	_ migrator.GormDataTypeInterface = (*Phone)(nil)

	_ driver.Valuer                  = (*NullPhone)(nil)
	_ sql.Scanner                    = (*NullPhone)(nil)
	_ json.Marshaler                 = (*NullPhone)(nil)
	_ json.Unmarshaler               = (*NullPhone)(nil)
	_ schema.GormDataTypeInterface   = (*NullPhone)(nil)
	_ migrator.GormDataTypeInterface = (*NullPhone)(nil)
)
