package dts

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"go-slim.dev/misc"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// NullUint represents a uint that may be null.
type NullUint struct {
	Uint  uint
	Valid bool
}

// NewNullUint creates a new NullUint with the given value.
func NewNullUint(u uint) NullUint {
	return NullUint{
		Uint:  u,
		Valid: true,
	}
}

// Scan implements the sql.Scanner interface.
func (n *NullUint) Scan(value any) (err error) {
	if value == nil {
		n.Uint, n.Valid = 0, false
		return nil
	}

	n.Uint, n.Valid, err = scan[uint](value)
	return
}

// Value implements the driver.Valuer interface.
func (n NullUint) Value() (driver.Value, error) {
	return toSqlValue(n.Uint, n.Valid)
}

func (NullUint) GormDataType() string {
	return "INTEGER UNSIGNED"
}

func (NullUint) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "INTEGER UNSIGNED"
}

func (n NullUint) String() string {
	if !n.Valid {
		return ""
	}
	return fmt.Sprintf("%d", n.Uint)
}

func (n NullUint) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Uint)
}

func (n *NullUint) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		n.Uint, n.Valid = 0, false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		n.Uint, n.Valid = 0, false
		return nil
	}

	var result uint
	err := json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("failed to parse NullUint: %w", err)
	}

	n.Uint, n.Valid = result, true
	return nil
}

// NullUint8 represents a uint8 that may be null.
type NullUint8 struct {
	Uint8 uint8
	Valid bool
}

// NewNullUint8 creates a new NullUint8 with the given value.
func NewNullUint8(u uint8) NullUint8 {
	return NullUint8{
		Uint8: u,
		Valid: true,
	}
}

func (n *NullUint8) Scan(value any) (err error) {
	if value == nil {
		n.Uint8, n.Valid = 0, false
		return nil
	}

	n.Uint8, n.Valid, err = scan[uint8](value)
	return
}

func (n NullUint8) Value() (driver.Value, error) {
	return toSqlValue(n.Uint8, n.Valid)
}

func (NullUint8) GormDataType() string {
	return "TINYINT UNSIGNED"
}

func (NullUint8) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "TINYINT UNSIGNED"
}

func (n NullUint8) String() string {
	if !n.Valid {
		return ""
	}
	return fmt.Sprintf("%d", n.Uint8)
}

func (n NullUint8) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Uint8)
}

func (n *NullUint8) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		n.Uint8, n.Valid = 0, false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		n.Uint8, n.Valid = 0, false
		return nil
	}

	var result uint8
	err := json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("failed to parse NullUint8: %w", err)
	}

	n.Uint8, n.Valid = result, true
	return nil
}

// NullUint16 represents a uint16 that may be null.
type NullUint16 struct {
	Uint16 uint16
	Valid  bool
}

// NewNullUint16 creates a new NullUint16 with the given value.
func NewNullUint16(u uint16) NullUint16 {
	return NullUint16{
		Uint16: u,
		Valid:  true,
	}
}

func (n *NullUint16) Scan(value any) (err error) {
	if value == nil {
		n.Uint16, n.Valid = 0, false
		return nil
	}

	n.Uint16, n.Valid, err = scan[uint16](value)
	return
}

func (n NullUint16) Value() (driver.Value, error) {
	return toSqlValue(n.Uint16, n.Valid)
}

func (NullUint16) GormDataType() string {
	return "SMALLINT UNSIGNED"
}

func (NullUint16) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "SMALLINT UNSIGNED"
}

func (n NullUint16) String() string {
	if !n.Valid {
		return ""
	}
	return fmt.Sprintf("%d", n.Uint16)
}

func (n NullUint16) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Uint16)
}

func (n *NullUint16) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		n.Uint16, n.Valid = 0, false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		n.Uint16, n.Valid = 0, false
		return nil
	}

	var result uint16
	err := json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("failed to parse NullUint16: %w", err)
	}

	n.Uint16, n.Valid = result, true
	return nil
}

// NullUint32 represents a uint32 that may be null.
type NullUint32 struct {
	Uint32 uint32
	Valid  bool
}

// NewNullUint32 creates a new NullUint32 with the given value.
func NewNullUint32(u uint32) NullUint32 {
	return NullUint32{
		Uint32: u,
		Valid:  true,
	}
}

func (n *NullUint32) Scan(value any) (err error) {
	if value == nil {
		n.Uint32, n.Valid = 0, false
		return nil
	}

	n.Uint32, n.Valid, err = scan[uint32](value)
	return
}

func (n NullUint32) Value() (driver.Value, error) {
	return toSqlValue(n.Uint32, n.Valid)
}

func (NullUint32) GormDataType() string {
	return "INT UNSIGNED"
}

func (NullUint32) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "INT UNSIGNED"
}

func (n NullUint32) String() string {
	if !n.Valid {
		return ""
	}
	return fmt.Sprintf("%d", n.Uint32)
}

func (n NullUint32) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Uint32)
}

func (n *NullUint32) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		n.Uint32, n.Valid = 0, false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		n.Uint32, n.Valid = 0, false
		return nil
	}

	var result uint32
	err := json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("failed to parse NullUint32: %w", err)
	}

	n.Uint32, n.Valid = result, true
	return nil
}

// NullUint64 represents a uint64 that may be null.
type NullUint64 struct {
	Uint64 uint64
	Valid  bool
}

// NewNullUint64 creates a new NullUint64 with the given value.
func NewNullUint64(u uint64) NullUint64 {
	return NullUint64{
		Uint64: u,
		Valid:  true,
	}
}

func (n *NullUint64) Scan(value any) (err error) {
	if value == nil {
		n.Uint64, n.Valid = 0, false
		return nil
	}

	n.Uint64, n.Valid, err = scan[uint64](value)
	return
}

func (n NullUint64) Value() (driver.Value, error) {
	return toSqlValue(n.Uint64, n.Valid)
}

func (NullUint64) GormDataType() string {
	return "BIGINT UNSIGNED"
}

func (NullUint64) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "BIGINT UNSIGNED"
}

func (n NullUint64) String() string {
	if !n.Valid {
		return ""
	}
	return fmt.Sprintf("%d", n.Uint64)
}

func (n NullUint64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Uint64)
}

func (n *NullUint64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		n.Uint64, n.Valid = 0, false
		return nil
	}

	str := misc.BytesToString(data)
	if str == "null" {
		n.Uint64, n.Valid = 0, false
		return nil
	}

	var result uint64
	err := json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("failed to parse NullUint64: %w", err)
	}

	n.Uint64, n.Valid = result, true
	return nil
}
