package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"go-slim.dev/misc"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// NullBool represents a bool that may be null.
// NullBool implements the sql.Scanner interface so
// it can be used as a scan destination, similar to sql.NullString.
type NullBool struct {
	Bool  bool
	Valid bool // Valid is true if Bool is not NULL
}

// NewNullBool creates a new NullBool with the given value.
func NewNullBool(b bool) NullBool {
	return NullBool{
		Bool:  b,
		Valid: true,
	}
}

// Scan implements the sql.Scanner interface.
func (n *NullBool) Scan(value any) error {
	var nb sql.NullBool
	err := nb.Scan(value)
	n.Bool = nb.Bool
	n.Valid = nb.Valid
	return err
}

// Value implements the driver.Valuer interface.
func (n NullBool) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Bool, nil
}

func (NullBool) GormDataType() string {
	return "BOOLEAN"
}

func (NullBool) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
	}
	return "BOOLEAN"
}

func (n NullBool) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Bool)
}

func (n *NullBool) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		n.Bool, n.Valid = false, false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		n.Bool, n.Valid = false, false
		return nil
	}

	var result bool
	err := json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("failed to parse NullBool: %w", err)
	}

	n.Bool, n.Valid = result, true
	return nil
}
