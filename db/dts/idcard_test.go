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
	_ driver.Valuer                  = (*Idcard)(nil)
	_ sql.Scanner                    = (*Idcard)(nil)
	_ fmt.Stringer                   = (*Idcard)(nil)
	_ json.Marshaler                 = (*Idcard)(nil)
	_ json.Unmarshaler               = (*Idcard)(nil)
	_ schema.GormDataTypeInterface   = (*Idcard)(nil)
	_ migrator.GormDataTypeInterface = (*Idcard)(nil)

	_ driver.Valuer                  = (*NullIdcard)(nil)
	_ sql.Scanner                    = (*NullIdcard)(nil)
	_ json.Marshaler                 = (*NullIdcard)(nil)
	_ json.Unmarshaler               = (*NullIdcard)(nil)
	_ schema.GormDataTypeInterface   = (*NullIdcard)(nil)
	_ migrator.GormDataTypeInterface = (*NullIdcard)(nil)
)
