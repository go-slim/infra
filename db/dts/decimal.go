package dts

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// NullDecimal represents a decimal.Decimal that may be null.
// It implements sql.Scanner, driver.Valuer, json.Marshaler/Unmarshaler, and GORM interfaces.
type NullDecimal struct {
	Decimal decimal.Decimal
	Valid   bool
}

// NewNullDecimal creates a new NullDecimal with the given decimal value.
func NewNullDecimal(d decimal.Decimal) NullDecimal {
	return NullDecimal{
		Decimal: d,
		Valid:   true,
	}
}

// Scan implements the sql.Scanner interface.
func (n *NullDecimal) Scan(value any) error {
	if value == nil {
		n.Decimal, n.Valid = decimal.Zero, false
		return nil
	}
	var d decimal.Decimal
	err := d.Scan(value)
	if err != nil {
		return err
	}
	n.Decimal = d
	n.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullDecimal) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Decimal.Value()
}

// MarshalJSON implements the json.Marshaler interface.
func (n NullDecimal) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Decimal)
	}
	return []byte("null"), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *NullDecimal) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		n.Valid = false
		return nil
	}
	err := n.Decimal.UnmarshalJSON(b)
	if err == nil {
		n.Valid = true
	}
	return err
}

// String implements the fmt.Stringer interface.
func (n NullDecimal) String() string {
	if !n.Valid {
		return ""
	}
	return n.Decimal.String()
}

// GormDataType returns the GORM data type for NullDecimal.
func (NullDecimal) GormDataType() string {
	return "DECIMAL(65,30)"
}

// GormDBDataType returns the database-specific data type for NullDecimal.
func (NullDecimal) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "DECIMAL(65,30)"
}
