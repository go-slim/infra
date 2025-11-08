package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var (
	_ driver.Valuer                  = (*Map)(nil)
	_ sql.Scanner                    = (*Map)(nil)
	_ json.Marshaler                 = (*Map)(nil)
	_ json.Unmarshaler               = (*Map)(nil)
	_ schema.GormDataTypeInterface   = (*Map)(nil)
	_ migrator.GormDataTypeInterface = (*Map)(nil)
)
