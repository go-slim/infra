package dts

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"go-slim.dev/misc"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type URL url.URL

// Scan implements the sql.Scanner interface.
func (u *URL) Scan(value any) error {
	var us string
	switch v := value.(type) {
	case []byte:
		us = misc.BytesToString(v)
	case string:
		us = v
	default:
		return fmt.Errorf("failed to parse URL: %T", value)
	}
	uu, err := url.Parse(us)
	if err != nil {
		return err
	}
	*u = URL(*uu)
	return nil
}

// Value implements the driver.Valuer interface.
func (u URL) Value() (driver.Value, error) {
	return u.String(), nil
}

func (URL) GormDataType() string {
	return "TEXT"
}

func (URL) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "url"
}

func (u *URL) String() string {
	return (*url.URL)(u).String()
}

func (u URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

func (u *URL) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	uu, err := url.Parse(strings.Trim(string(data), `"'`))
	if err != nil {
		return err
	}
	*u = URL(*uu)
	return nil
}

// NullURL represents a URL that may be null.
// It implements sql.Scanner, driver.Valuer, json.Marshaler/Unmarshaler, and GORM interfaces.
type NullURL struct {
	URL   URL
	Valid bool
}

// NewNullURL creates a new NullURL with the given URL value.
func NewNullURL(u URL) NullURL {
	return NullURL{
		URL:   u,
		Valid: true,
	}
}

// NewNullURLFromString creates a new NullURL from a string.
func NewNullURLFromString(s string) NullURL {
	uu, err := url.Parse(s)
	if err != nil {
		return NullURL{}
	}
	return NewNullURL(URL(*uu))
}

// Scan implements the sql.Scanner interface.
func (n *NullURL) Scan(value any) error {
	if value == nil {
		n.Valid = false
		return nil
	}

	var u URL
	err := u.Scan(value)
	if err != nil {
		return err
	}
	n.URL, n.Valid = u, true
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullURL) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.URL.Value()
}

// MarshalJSON implements the json.Marshaler interface.
func (n NullURL) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return n.URL.MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *NullURL) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		return nil
	}
	var u URL
	err := u.UnmarshalJSON(data)
	if err == nil {
		n.URL, n.Valid = u, true
	}
	return err
}

// String implements the fmt.Stringer interface.
func (n NullURL) String() string {
	if !n.Valid {
		return ""
	}
	return n.URL.String()
}

// GormDataType returns the GORM data type for NullURL.
func (NullURL) GormDataType() string {
	return "TEXT"
}

// GormDBDataType returns the database-specific data type for NullURL.
func (NullURL) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "url"
}
