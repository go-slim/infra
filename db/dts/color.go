package dts

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"go-slim.dev/is"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Color 颜色，仅支持16进制
type Color string

// Set 设置颜色值
func (c *Color) Set(s string) bool {
	if s == "" || is.HEXColor(s) {
		*c = Color(s)
		return true
	}
	return false
}

func (c Color) Value() (driver.Value, error) {
	if len(c) == 0 {
		return nil, errors.New("TODO")
	}
	return string(c), nil
}

func (c *Color) Scan(value any) error {
	if value == nil {
		return errors.New("TODO")
	}

	var ok bool
	switch v := value.(type) {
	case []byte:
		ok = c.Set(string(v))
	case string:
		ok = c.Set(v)
	}
	if !ok {
		return fmt.Errorf("failed to unmarshal value: %v", value)
	}
	return nil
}

func (c Color) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(c))
}

func (c *Color) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	if !c.Set(str) {
		return errors.New("failed to unmarshal value")
	}
	return nil
}

func (c Color) String() string {
	return string(c)
}

// GormDBDataType gorm db data type
func (Color) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "varchar(9)"
}

// GormDataType returns the GORM data type for Color
func (Color) GormDataType() string {
	return "varchar(9)"
}

func (c Color) GormValue(_ context.Context, _ *gorm.DB) clause.Expr {
	if len(c) == 0 {
		return gorm.Expr("NULL")
	}
	data, _ := c.MarshalJSON()
	return gorm.Expr("?", string(data))
}

// NullColor represents a Color that may be null.
// It implements sql.Scanner, driver.Valuer, json.Marshaler/Unmarshaler, and GORM interfaces.
type NullColor struct {
	Color Color
	Valid bool
}

// NewNullColor creates a new NullColor with the given color value.
func NewNullColor(c Color) NullColor {
	return NullColor{
		Color: c,
		Valid: true,
	}
}

// Scan implements the sql.Scanner interface.
func (n *NullColor) Scan(value any) error {
	if value == nil {
		n.Color, n.Valid = "", false
		return nil
	}

	var ok bool
	switch v := value.(type) {
	case []byte:
		ok = n.Color.Set(string(v))
	case string:
		ok = n.Color.Set(v)
	default:
		return fmt.Errorf("failed to unmarshal value: %v", value)
	}

	n.Valid = ok
	if !ok {
		return fmt.Errorf("failed to unmarshal value: %v", value)
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullColor) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Color.Value()
}

// MarshalJSON implements the json.Marshaler interface.
func (n NullColor) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.Color.MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *NullColor) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		n.Valid = false
		return nil
	}
	err := n.Color.UnmarshalJSON(b)
	if err == nil {
		n.Valid = true
	}
	return err
}

// String implements the fmt.Stringer interface.
func (n NullColor) String() string {
	if !n.Valid {
		return ""
	}
	return n.Color.String()
}

// GormDataType returns the GORM data type for NullColor.
func (NullColor) GormDataType() string {
	return "varchar(9)"
}

// GormDBDataType returns the database-specific data type for NullColor.
func (NullColor) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "varchar(9)"
}
