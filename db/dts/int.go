package dts

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"go-slim.dev/misc"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// NullInt represents an int that may be null.
type NullInt struct {
	Int   int
	Valid bool
}

// NewNullInt creates a new NullInt with the given value.
func NewNullInt(i int) NullInt {
	return NullInt{
		Int:   i,
		Valid: true,
	}
}

func (n *NullInt) Scan(value any) (err error) {
	if value == nil {
		n.Int, n.Valid = 0, false
		return nil
	}

	n.Int, n.Valid, err = scan[int](value)
	return
}

func (n NullInt) Value() (driver.Value, error) {
	return toSqlValue(n.Int, n.Valid)
}

func (NullInt) GormDataType() string {
	return "INTEGER"
}

func (NullInt) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "INTEGER"
}

func (n NullInt) String() string {
	if !n.Valid {
		return ""
	}
	return fmt.Sprintf("%d", n.Int)
}

func (n NullInt) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Int)
}

func (n *NullInt) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		n.Int, n.Valid = 0, false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		n.Int, n.Valid = 0, false
		return nil
	}

	var result int
	err := json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("failed to parse NullInt: %w", err)
	}

	n.Int, n.Valid = result, true
	return nil
}

// NullInt8 represents an int8 that may be null.
type NullInt8 struct {
	Int8  int8
	Valid bool
}

// NewNullInt8 creates a new NullInt8 with the given value.
func NewNullInt8(i int8) NullInt8 {
	return NullInt8{
		Int8:  i,
		Valid: true,
	}
}

func (n *NullInt8) Scan(value any) (err error) {
	if value == nil {
		n.Int8, n.Valid = 0, false
		return nil
	}

	n.Int8, n.Valid, err = scan[int8](value)
	return
}

func (n NullInt8) Value() (driver.Value, error) {
	return toSqlValue(n.Int8, n.Valid)
}

func (NullInt8) GormDataType() string {
	return "TINYINT"
}

func (NullInt8) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "TINYINT"
}

func (n NullInt8) String() string {
	if !n.Valid {
		return ""
	}
	return fmt.Sprintf("%d", n.Int8)
}

func (n NullInt8) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Int8)
}

func (n *NullInt8) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		n.Int8, n.Valid = 0, false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		n.Int8, n.Valid = 0, false
		return nil
	}

	var result int8
	err := json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("failed to parse NullInt8: %w", err)
	}

	n.Int8, n.Valid = result, true
	return nil
}

// NullInt16 represents an int16 that may be null.
type NullInt16 struct {
	Int16 int16
	Valid bool
}

// NewNullInt16 creates a new NullInt16 with the given value.
func NewNullInt16(i int16) NullInt16 {
	return NullInt16{
		Int16: i,
		Valid: true,
	}
}

func (n *NullInt16) Scan(value any) (err error) {
	if value == nil {
		n.Int16, n.Valid = 0, false
		return nil
	}

	n.Int16, n.Valid, err = scan[int16](value)
	return
}

func (n NullInt16) Value() (driver.Value, error) {
	return toSqlValue(n.Int16, n.Valid)
}

func (NullInt16) GormDataType() string {
	return "SMALLINT"
}

func (NullInt16) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "SMALLINT"
}

func (n NullInt16) String() string {
	if !n.Valid {
		return ""
	}
	return fmt.Sprintf("%d", n.Int16)
}

func (n NullInt16) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Int16)
}

func (n *NullInt16) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		n.Int16, n.Valid = 0, false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		n.Int16, n.Valid = 0, false
		return nil
	}

	var result int16
	err := json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("failed to parse NullInt16: %w", err)
	}

	n.Int16, n.Valid = result, true
	return nil
}

// NullInt32 represents an int32 that may be null.
type NullInt32 struct {
	Int32 int32
	Valid bool
}

// NewNullInt32 creates a new NullInt32 with the given value.
func NewNullInt32(i int32) NullInt32 {
	return NullInt32{
		Int32: i,
		Valid: true,
	}
}

func (n *NullInt32) Scan(value any) (err error) {
	if value == nil {
		n.Int32, n.Valid = 0, false
		return nil
	}

	n.Int32, n.Valid, err = scan[int32](value)
	return
}

func (n NullInt32) Value() (driver.Value, error) {
	return toSqlValue(n.Int32, n.Valid)
}

func (NullInt32) GormDataType() string {
	return "INT"
}

func (NullInt32) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "INT"
}

func (n NullInt32) String() string {
	if !n.Valid {
		return ""
	}
	return fmt.Sprintf("%d", n.Int32)
}

func (n NullInt32) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Int32)
}

func (n *NullInt32) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		n.Int32, n.Valid = 0, false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		n.Int32, n.Valid = 0, false
		return nil
	}

	var result int32
	err := json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("failed to parse NullInt32: %w", err)
	}

	n.Int32, n.Valid = result, true
	return nil
}

// NullInt64 represents an int64 that may be null.
type NullInt64 struct {
	Int64 int64
	Valid bool
}

// NewNullInt64 creates a new NullInt64 with the given value.
func NewNullInt64(i int64) NullInt64 {
	return NullInt64{
		Int64: i,
		Valid: true,
	}
}

func (n *NullInt64) Scan(value any) (err error) {
	if value == nil {
		n.Int64, n.Valid = 0, false
		return nil
	}

	n.Int64, n.Valid, err = scan[int64](value)
	return
}

func (n NullInt64) Value() (driver.Value, error) {
	return toSqlValue(n.Int64, n.Valid)
}

func (NullInt64) GormDataType() string {
	return "BIGINT"
}

func (NullInt64) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "BIGINT"
}

func (n NullInt64) String() string {
	if !n.Valid {
		return ""
	}
	return fmt.Sprintf("%d", n.Int64)
}

func (n NullInt64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Int64)
}

func (n *NullInt64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		n.Int64, n.Valid = 0, false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		n.Int64, n.Valid = 0, false
		return nil
	}

	var result int64
	err := json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("failed to parse NullInt64: %w", err)
	}

	n.Int64, n.Valid = result, true
	return nil
}
