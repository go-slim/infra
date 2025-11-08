package dts

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// NullFloat32 represents a float32 that may be null.
// It implements sql.Scanner, driver.Valuer, json.Marshaler/Unmarshaler, and GORM interfaces.
type NullFloat32 struct {
	Float32 float32
	Valid   bool
}

// NewNullFloat32 creates a new NullFloat32 with the given float32 value.
func NewNullFloat32(f float32) NullFloat32 {
	return NullFloat32{
		Float32: f,
		Valid:   true,
	}
}

// Scan implements the sql.Scanner interface.
func (n *NullFloat32) Scan(value any) error {
	if value == nil {
		n.Float32, n.Valid = 0, false
		return nil
	}

	switch v := value.(type) {
	case float64:
		n.Float32, n.Valid = float32(v), true
	case float32:
		n.Float32, n.Valid = v, true
	case int64:
		n.Float32, n.Valid = float32(v), true
	case int:
		n.Float32, n.Valid = float32(v), true
	case []byte:
		var f float32
		_, err := fmt.Sscanf(string(v), "%f", &f)
		if err != nil {
			return fmt.Errorf("failed to scan NullFloat32: %w", err)
		}
		n.Float32, n.Valid = f, true
	case string:
		var f float32
		_, err := fmt.Sscanf(v, "%f", &f)
		if err != nil {
			return fmt.Errorf("failed to scan NullFloat32: %w", err)
		}
		n.Float32, n.Valid = f, true
	default:
		return fmt.Errorf("failed to scan NullFloat32: unsupported type %T", v)
	}

	return nil
}

// Value implements the driver.Valuer interface.
func (n NullFloat32) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Float32, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (n NullFloat32) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Float32)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *NullFloat32) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		n.Valid = false
		return nil
	}
	var f float32
	err := json.Unmarshal(b, &f)
	if err != nil {
		return fmt.Errorf("failed to unmarshal NullFloat32: %w", err)
	}
	n.Float32, n.Valid = f, true
	return nil
}

// String implements the fmt.Stringer interface.
func (n NullFloat32) String() string {
	if !n.Valid {
		return ""
	}
	return fmt.Sprintf("%f", n.Float32)
}

// GormDataType returns the GORM data type for NullFloat32.
func (NullFloat32) GormDataType() string {
	return "FLOAT"
}

// GormDBDataType returns the database-specific data type for NullFloat32.
func (NullFloat32) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "FLOAT"
}

// NullFloat64 represents a float64 that may be null.
// It implements sql.Scanner, driver.Valuer, json.Marshaler/Unmarshaler, and GORM interfaces.
type NullFloat64 struct {
	Float64 float64
	Valid   bool
}

// NewNullFloat64 creates a new NullFloat64 with the given float64 value.
func NewNullFloat64(f float64) NullFloat64 {
	return NullFloat64{
		Float64: f,
		Valid:   true,
	}
}

// Scan implements the sql.Scanner interface.
func (n *NullFloat64) Scan(value any) error {
	if value == nil {
		n.Float64, n.Valid = 0, false
		return nil
	}

	switch v := value.(type) {
	case float64:
		n.Float64, n.Valid = v, true
	case float32:
		n.Float64, n.Valid = float64(v), true
	case int64:
		n.Float64, n.Valid = float64(v), true
	case int:
		n.Float64, n.Valid = float64(v), true
	case []byte:
		var f float64
		_, err := fmt.Sscanf(string(v), "%f", &f)
		if err != nil {
			return fmt.Errorf("failed to scan NullFloat64: %w", err)
		}
		n.Float64, n.Valid = f, true
	case string:
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		if err != nil {
			return fmt.Errorf("failed to scan NullFloat64: %w", err)
		}
		n.Float64, n.Valid = f, true
	default:
		return fmt.Errorf("failed to scan NullFloat64: unsupported type %T", v)
	}

	return nil
}

// Value implements the driver.Valuer interface.
func (n NullFloat64) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Float64, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (n NullFloat64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Float64)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *NullFloat64) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		n.Valid = false
		return nil
	}
	var f float64
	err := json.Unmarshal(b, &f)
	if err != nil {
		return fmt.Errorf("failed to unmarshal NullFloat64: %w", err)
	}
	n.Float64, n.Valid = f, true
	return nil
}

// String implements the fmt.Stringer interface.
func (n NullFloat64) String() string {
	if !n.Valid {
		return ""
	}
	return fmt.Sprintf("%f", n.Float64)
}

// GormDataType returns the GORM data type for NullFloat64.
func (NullFloat64) GormDataType() string {
	return "DOUBLE"
}

// GormDBDataType returns the database-specific data type for NullFloat64.
func (NullFloat64) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "DOUBLE"
}
