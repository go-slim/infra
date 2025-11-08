package dts

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"go-slim.dev/is"
	"go-slim.dev/misc"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Phone 手机号码（仅支持中国大陆手机号）
type Phone string

// Scan implements the sql.Scanner interface.
func (e *Phone) Scan(value any) error {
	if value == nil {
		*e = ""
		return nil
	}
	var ps string
	switch v := value.(type) {
	case []byte:
		ps = string(v)
	case string:
		ps = v
	default:
		return fmt.Errorf("failed to parse phone number: unsupported type %T", v)
	}
	if ps != "" && !is.PhoneNumber(ps) {
		return fmt.Errorf("failed to parse phone number: invalid phone format %v", ps)
	}
	*e = Phone(ps)
	return nil
}

// Value implements the driver.Valuer interface.
func (e Phone) Value() (driver.Value, error) {
	return string(e), nil
}

func (Phone) GormDataType() string {
	return "phone"
}

func (Phone) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "VARCHAR(14)"
}

func (e Phone) String() string {
	return string(e)
}

func (e Phone) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

func (e *Phone) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*e = ""
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		*e = ""
		return nil
	}

	str = strings.Trim(str, `"'`)
	if str == "" {
		*e = ""
		return nil
	}

	if !is.PhoneNumber(str) {
		return fmt.Errorf("failed to parse phone number: invalid phone format %v", str)
	}
	*e = Phone(str)
	return nil
}

func (e Phone) MarshalBinary() ([]byte, error) {
	return []byte(e.String()), nil
}

func (e *Phone) UnmarshalBinary(data []byte) error {
	return e.UnmarshalText(data)
}

func (e Phone) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *Phone) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*e = ""
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		*e = ""
		return nil
	}
	str = strings.Trim(str, `"'`)
	if str == "" {
		*e = ""
		return nil
	}
	if !is.PhoneNumber(str) {
		return fmt.Errorf("failed to parse phone number: invalid phone format %v", str)
	}
	*e = Phone(str)
	return nil
}

type NullPhone struct {
	Phone Phone
	Valid bool
}

// Scan implements the sql.Scanner interface.
func (e *NullPhone) Scan(value any) error {
	if value == nil {
		e.Phone, e.Valid = "", false
		return nil
	}

	var ps string
	switch v := value.(type) {
	case []byte:
		ps = string(v)
	case string:
		ps = v
	default:
		return fmt.Errorf("failed to parse NullPhone: unsupported type %T", v)
	}

	// Handle empty string as null
	if ps == "" {
		e.Phone, e.Valid = "", false
		return nil
	}

	// Validate phone number format
	if !is.PhoneNumber(ps) {
		e.Phone, e.Valid = "", false
		return fmt.Errorf("failed to parse NullPhone: invalid phone format %v", ps)
	}

	e.Phone = Phone(ps)
	e.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (e NullPhone) Value() (driver.Value, error) {
	if !e.Valid {
		return nil, nil
	}
	return e.Phone.String(), nil
}

func (e NullPhone) GormDataType() string {
	return "phone"
}

func (e NullPhone) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "VARCHAR(14)"
}

// MarshalJSON implements the json.Marshaler interface.
func (e NullPhone) MarshalJSON() ([]byte, error) {
	if !e.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(e.Phone)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *NullPhone) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		e.Phone, e.Valid = "", false
		return nil
	}

	// Handle JSON null value
	str := misc.BytesToString(data)
	if str == "null" {
		e.Phone, e.Valid = "", false
		return nil
	}

	err := e.Phone.UnmarshalJSON(data)
	if err != nil {
		e.Phone, e.Valid = "", false
		return err
	}
	e.Valid = string(e.Phone) != ""
	return nil
}
