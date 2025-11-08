package dts

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"go-slim.dev/misc"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// NullString represents a string that may be null.
// NullString implements the sql.Scanner interface so
// it can be used as a scan destination, similar to sql.NullString.
type NullString struct {
	String string
	Valid  bool // Valid is true if String is not NULL
}

// NewNullString creates a new NullString with the given value.
func NewNullString(s string) NullString {
	return NullString{
		String: s,
		Valid:  true,
	}
}

// Scan implements the sql.Scanner interface.
func (ns *NullString) Scan(value any) error {
	if value == nil {
		ns.String, ns.Valid = "", false
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		str = fmt.Sprintf("%v", v)
	default:
		return fmt.Errorf("failed to parse NullString: unsupported type %T", v)
	}

	ns.String = str
	ns.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.Value, nil
}

func (NullString) GormDataType() string {
	return "VARCHAR(255)"
}

func (NullString) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "VARCHAR(255)"
}

func (ns NullString) MarshalText() ([]byte, error) {
	if !ns.Valid {
		return nil, nil
	}
	return []byte(ns.String), nil
}

func (ns *NullString) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		ns.String, ns.Valid = "", false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		ns.String, ns.Valid = "", false
		return nil
	}

	ns.String, ns.Valid = str, true
	return nil
}

func (ns NullString) MarshalBinary() ([]byte, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.MarshalText()
}

func (ns *NullString) UnmarshalBinary(data []byte) error {
	return ns.UnmarshalText(data)
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(ns.Value)
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		ns.String, ns.Valid = "", false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		ns.String, ns.Valid = "", false
		return nil
	}

	// Handle quoted JSON strings
	if len(str) >= 2 && (str[0] == '"' && str[len(str)-1] == '"' || str[0] == '\'' && str[len(str)-1] == '\'') {
		str = str[1 : len(str)-1]
	}

	ns.String, ns.Valid = str, true
	return nil
}
