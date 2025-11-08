package dts

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Slice give a generic data type for json encoded slice data.
type Slice[T any] []T

func NewSlice[T any](s []T) Slice[T] {
	return s
}

// MarshalJSON implements the json.Marshaler interface.
func (j Slice[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal([]T(j))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (j *Slice[T]) UnmarshalJSON(data []byte) error {
	var slice []T
	err := json.Unmarshal(data, &slice)
	if err != nil {
		return err
	}
	*j = Slice[T](slice)
	return nil
}

// Value return json value, implement driver.Valuer interface
func (j Slice[T]) Value() (driver.Value, error) {
	return j.MarshalJSON()
}

// Scan scans a value into Slice[T], implements sql.Scanner interface
func (j *Slice[T]) Scan(value any) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	return j.UnmarshalJSON(bytes)
}

// GormDataType gorm common data type
func (Slice[T]) GormDataType() string {
	return "slices"
}

// GormDBDataType gorm db data type
func (Slice[T]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return ""
}

func (j Slice[T]) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	data, _ := json.Marshal(j)

	switch db.Dialector.Name() {
	case "mysql":
		if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
			return gorm.Expr("CAST(? AS JSON)", string(data))
		}
	}

	return gorm.Expr("?", string(data))
}

type NullSlice[T any] struct {
	Slice []T
	Valid bool
}

// Scan implements the sql.Scanner interface.
func (n *NullSlice[T]) Scan(value any) error {
	if value == nil {
		n.Slice, n.Valid = nil, false
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return n.UnmarshalJSON(v)
	case string:
		return n.UnmarshalJSON([]byte(v))
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
}

// Value implements the driver.Valuer interface.
func (n NullSlice[T]) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return json.Marshal(n.Slice)
}

// MarshalJSON implements the json.Marshaler interface.
func (n NullSlice[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Slice)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *NullSlice[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || len(data) == 0 {
		n.Valid = false
		return nil
	}

	var values []T
	err := json.Unmarshal(data, &values)
	if err != nil {
		return err
	}

	n.Valid = true
	n.Slice = values
	return nil
}

// GormDataType returns the GORM data type for NullSlice.
func (NullSlice[T]) GormDataType() string {
	return "json"
}

// GormDBDataType returns the database-specific data type for NullSlice.
func (NullSlice[T]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return ""
}
