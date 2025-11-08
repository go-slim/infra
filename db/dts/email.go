package dts

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go-slim.dev/is"
	"go-slim.dev/misc"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Email string

// Scan implements the sql.Scanner interface.
// Handles database NULL values by setting the email to empty string.
func (e *Email) Scan(value any) error {
	if value == nil {
		return errors.New("TODO")
	}

	var es string
	switch v := value.(type) {
	case []byte:
		es = string(v)
	case string:
		es = v
	default:
		return fmt.Errorf("failed to parse Email: unsupported type %T", v)
	}
	if es != "" && !is.Email(es) {
		return fmt.Errorf("failed to parse Email: invalid email format %v", es)
	}
	*e = Email(es)
	return nil
}

// Value implements the driver.Valuer interface.
func (e Email) Value() (driver.Value, error) {
	return string(e), nil
}

func (Email) GormDataType() string {
	return "VARCHAR(255)"
}

func (Email) GormDBDataType(*gorm.DB, *schema.Field) string {
	return "email"
}

func (e Email) String() string {
	return string(e)
}

func (e Email) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *Email) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*e = ""
		return nil
	}

	str := misc.BytesToString(data)
	// Handle JSON null value
	if str == "null" {
		*e = ""
		return nil
	}

	str = strings.Trim(str, `"'`)
	if str == "" {
		*e = ""
		return nil
	}

	if !is.Email(str) {
		return fmt.Errorf("failed to parse Email: invalid email format %v", str)
	}
	*e = Email(str)
	return nil
}

type NullEmail struct {
	Email Email
	Valid bool
}

func NewNullEmail(s string) NullEmail {
	if !is.Email(s) {
		return NullEmail{}
	}

	return NullEmail{
		Email: Email(s),
		Valid: true,
	}
}

// Scan implements the sql.Scanner interface.
func (e *NullEmail) Scan(value any) error {
	if value == nil {
		e.Email, e.Valid = "", false
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return fmt.Errorf("failed to parse NullEmail: %v", v)
	}

	// Handle empty string as null
	if str == "" {
		e.Email, e.Valid = "", false
		return nil
	}

	// Validate email format
	if !is.Email(str) {
		e.Email, e.Valid = "", false
		return fmt.Errorf("failed to parse NullEmail: invalid email format %v", str)
	}

	e.Email = Email(str)
	e.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (e NullEmail) Value() (driver.Value, error) {
	if !e.Valid {
		return nil, nil
	}
	return e.Email.String(), nil
}

func (NullEmail) GormDataType() string {
	return "VARCHAR(255)"
}

func (NullEmail) GormDBDataType(*gorm.DB, *schema.Field) string {
	return "email"
}

// MarshalJSON implements the json.Marshaler interface.
func (e NullEmail) MarshalJSON() ([]byte, error) {
	if !e.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(e.Email)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *NullEmail) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		e.Valid = false
		return nil
	}

	// Handle JSON null value
	str := misc.BytesToString(data)
	if str == "null" {
		e.Email, e.Valid = "", false
		return nil
	}

	err := e.Email.UnmarshalJSON(data)
	if err != nil {
		e.Email, e.Valid = "", false
		return err
	}
	e.Valid = string(e.Email) != ""

	return nil
}

func (e *NullEmail) String() string {
	if !e.Valid {
		return ""
	}

	return e.Email.String()
}
