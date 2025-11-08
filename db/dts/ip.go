package dts

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// IP represents an IP address as a string.
// It supports both IPv4 and IPv6 addresses with validation.
type IP string

// Set sets the IP address with validation.
func (ip *IP) Set(s string) bool {
	if s == "" {
		*ip = IP(s)
		return true
	}
	if net.ParseIP(s) != nil {
		*ip = IP(s)
		return true
	}
	return false
}

// Scan implements the sql.Scanner interface.
func (ip *IP) Scan(value any) error {
	if value == nil {
		*ip = ""
		return nil
	}

	var s string
	switch v := value.(type) {
	case []byte:
		s = string(v)
	case string:
		s = v
	default:
		return fmt.Errorf("failed to scan IP: unsupported type %T", v)
	}

	if !ip.Set(s) {
		return fmt.Errorf("failed to scan IP: invalid IP address %s", s)
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (ip IP) Value() (driver.Value, error) {
	if ip == "" {
		return nil, nil
	}
	return string(ip), nil
}

// MarshalJSON implements the json.Marshaler interface.
func (ip IP) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ip))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ip *IP) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return fmt.Errorf("failed to unmarshal IP: %w", err)
	}
	if !ip.Set(s) {
		return fmt.Errorf("failed to unmarshal IP: invalid IP address %s", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (ip IP) String() string {
	return string(ip)
}

// GormDataType returns the GORM data type for IP.
func (IP) GormDataType() string {
	return "VARCHAR(45)"
}

// GormDBDataType returns the database-specific data type for IP.
func (IP) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "VARCHAR(45)"
}

// IsValid checks if the IP address is valid.
func (ip IP) IsValid() bool {
	return net.ParseIP(string(ip)) != nil
}

// IsIPv4 checks if the IP address is IPv4.
func (ip IP) IsIPv4() bool {
	parsedIP := net.ParseIP(string(ip))
	if parsedIP == nil {
		return false
	}
	return parsedIP.To4() != nil
}

// IsIPv6 checks if the IP address is IPv6.
func (ip IP) IsIPv6() bool {
	parsedIP := net.ParseIP(string(ip))
	if parsedIP == nil {
		return false
	}
	return parsedIP.To4() == nil
}

// NullIP represents an IP that may be null.
// It implements sql.Scanner, driver.Valuer, json.Marshaler/Unmarshaler, and GORM interfaces.
type NullIP struct {
	IP    IP
	Valid bool
}

// NewNullIP creates a new NullIP with the given IP value.
func NewNullIP(ip IP) NullIP {
	return NullIP{
		IP:    ip,
		Valid: true,
	}
}

// NewNullIPFromString creates a new NullIP from a string.
func NewNullIPFromString(s string) NullIP {
	var ip IP
	if ip.Set(s) {
		return NullIP{
			IP:    ip,
			Valid: true,
		}
	}
	return NullIP{}
}

// Scan implements the sql.Scanner interface.
func (n *NullIP) Scan(value any) error {
	if value == nil {
		n.IP, n.Valid = "", false
		return nil
	}

	var s string
	switch v := value.(type) {
	case []byte:
		s = string(v)
	case string:
		s = v
	default:
		return fmt.Errorf("failed to scan NullIP: unsupported type %T", v)
	}

	var ip IP
	if ip.Set(s) {
		n.IP, n.Valid = ip, true
	} else {
		n.IP, n.Valid = "", false
		return fmt.Errorf("failed to scan NullIP: invalid IP address %s", s)
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullIP) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.IP.Value()
}

// MarshalJSON implements the json.Marshaler interface.
func (n NullIP) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return n.IP.MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *NullIP) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		n.Valid = false
		return nil
	}
	var ip IP
	err := ip.UnmarshalJSON(b)
	if err == nil {
		n.IP, n.Valid = ip, true
	}
	return err
}

// String implements the fmt.Stringer interface.
func (n NullIP) String() string {
	if !n.Valid {
		return ""
	}
	return n.IP.String()
}

// GormDataType returns the GORM data type for NullIP.
func (NullIP) GormDataType() string {
	return "VARCHAR(45)"
}

// GormDBDataType returns the database-specific data type for NullIP.
func (NullIP) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "VARCHAR(45)"
}
