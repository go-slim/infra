package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var (
	_ driver.Valuer                  = (*NullString)(nil)
	_ sql.Scanner                    = (*NullString)(nil)
	_ json.Marshaler                 = (*NullString)(nil)
	_ json.Unmarshaler               = (*NullString)(nil)
	_ schema.GormDataTypeInterface   = (*NullString)(nil)
	_ migrator.GormDataTypeInterface = (*NullString)(nil)
)
