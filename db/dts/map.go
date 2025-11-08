package dts

import (
	"bytes"
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

// Map represents a JSON map data type for database storage.
// It implements driver.Valuer, sql.Scanner, json.Marshaler/Unmarshaler, and GORM interfaces.
// This type is useful for storing structured JSON data in databases with JSON support.
type Map map[string]any

// NewMap creates a new Map from a map[string]any.
func NewMap(m map[string]any) Map {
	if m == nil {
		return make(Map)
	}
	return Map(m)
}

// Value implements the driver.Valuer interface.
// It returns the JSON string representation of the map for database storage.
func (m Map) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	ba, err := m.MarshalJSON()
	return string(ba), err
}

// Scan implements the sql.Scanner interface.
// It scans database values into Map, handling JSON strings and byte arrays.
func (m *Map) Scan(val any) error {
	if val == nil {
		*m = make(Map)
		return nil
	}
	var ba []byte
	switch v := val.(type) {
	case []byte:
		ba = v
	case string:
		ba = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", val))
	}
	t := map[string]any{}
	rd := bytes.NewReader(ba)
	decoder := json.NewDecoder(rd)
	decoder.UseNumber()
	err := decoder.Decode(&t)
	*m = Map(t)
	return err
}

// MarshalJSON implements the json.Marshaler interface.
// It returns JSON representation of the map, using "null" for nil maps.
func (m Map) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	t := (map[string]any)(m)
	return json.Marshal(t)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It deserializes JSON data into the Map type.
func (m *Map) UnmarshalJSON(b []byte) error {
	t := map[string]any{}
	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}
	*m = Map(t)
	return nil
}

// GormDataType gorm common data type
func (m Map) GormDataType() string {
	return "map"
}

// GormDBDataType gorm db data type
func (Map) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	}
	return ""
}

// GormValue returns the GORM expression for database operations.
// It handles MySQL-specific casting for JSON types when appropriate.
func (m Map) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	data, _ := m.MarshalJSON()
	switch db.Dialector.Name() {
	case "mysql":
		if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
			return gorm.Expr("CAST(? AS JSON)", string(data))
		}
	}
	return gorm.Expr("?", string(data))
}

// String implements the fmt.Stringer interface.
func (m Map) String() string {
	data, err := m.MarshalJSON()
	if err != nil {
		return "{}"
	}
	return string(data)
}
