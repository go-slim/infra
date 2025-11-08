package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"go-slim.dev/misc"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Date represents a date without time components.
// It implements sql.Scanner, driver.Valuer, json.Marshaler/Unmarshaler, and GORM interfaces.
type Date time.Time

func (date *Date) Scan(value any) (err error) {
	nullTime := &sql.NullTime{}
	err = nullTime.Scan(value)
	*date = Date(nullTime.Time)
	return
}

func (date Date) Value() (driver.Value, error) {
	return time2date(time.Time(date)), nil
}

// GormDataType gorm common data type
func (date Date) GormDataType() string {
	return "date"
}

func (date Date) MarshalJSON() ([]byte, error) {
	str := time.Time(date).Format(time.DateOnly)
	return misc.StringToBytes(str), nil
}

func (date *Date) UnmarshalJSON(p []byte) error {
	var t time.Time
	err := t.UnmarshalJSON(p)
	if err != nil {
		return err
	}
	*date = Date(time2date(t))
	return nil
}

// String implements the fmt.Stringer interface.
func (date Date) String() string {
	return time.Time(date).Format(time.DateOnly)
}

// GormDBDataType returns the database-specific data type for Date.
func (Date) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "DATE"
}

func time2date(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

type NullDate struct {
	Date  time.Time
	Valid bool // Valid is true if Date is not NULL
}

func (nd NullDate) Any() any {
	if nd.Valid {
		return nd.Date
	}
	return nil
}

func (nd *NullDate) Scan(value any) error {
	if value == nil {
		nd.Date, nd.Valid = time.Time{}, false
		return nil
	}
	nt := new(sql.NullTime)
	err := nt.Scan(value)
	nd.Date = nt.Time
	nd.Valid = nt.Valid
	return err
}

func (nd NullDate) Value() (driver.Value, error) {
	if !nd.Valid {
		return nil, nil
	}
	y, m, d := nd.Date.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, nd.Date.Location()), nil
}

// GormDataType gorm common data type
func (nd NullDate) GormDataType() string {
	return "date"
}

func (nd NullDate) MarshalJSON() ([]byte, error) {
	if nd.Valid {
		return nd.Date.MarshalJSON()
	}
	return json.Marshal(nil)
}

func (nd *NullDate) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		nd.Valid = false
		return nil
	}
	err := nd.Date.UnmarshalJSON(b)
	if err == nil {
		nd.Valid = true
	}
	return err
}

// String implements the fmt.Stringer interface.
func (nd NullDate) String() string {
	if !nd.Valid {
		return ""
	}
	return nd.Date.Format(time.DateOnly)
}

// GormDBDataType returns the database-specific data type for NullDate.
func (NullDate) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "DATE"
}

// NullTime represents a [time.Time] that may be null.
// NullTime implements the [Scanner] interface so
// it can be used as a scan destination, similar to [NullString].
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

func NewNullTime(t time.Time) NullTime {
	return NullTime{
		Time:  t,
		Valid: !t.IsZero(),
	}
}

// Scan implements the sql.Scanner interface.
func (nt *NullTime) Scan(value any) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}
	st := new(sql.NullTime)
	err := st.Scan(value)
	nt.Time = st.Time
	nt.Valid = st.Valid
	return err
}

// Value implements the driver.Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return json.Marshal(nil)
	}
	return nt.Time.MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (nt *NullTime) UnmarshalJSON(bytes []byte) error {
	err := nt.Time.UnmarshalJSON(bytes)
	if err != nil {
		return err
	}
	nt.Valid = !nt.Time.IsZero()
	return nil
}

// String implements the fmt.Stringer interface.
func (nt NullTime) String() string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.String()
}

// GormDataType returns the GORM data type for NullTime.
func (NullTime) GormDataType() string {
	return "DATETIME"
}

// GormDBDataType returns the database-specific data type for NullTime.
func (NullTime) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "DATETIME"
}
