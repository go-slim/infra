package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"go-slim.dev/misc"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// IsIdcardFunc 自定义身份证校验函数，比如使用第三方
// 库 "github.com/guanguans/id-validator" 的 IsLooseValid 函数
var IsIdcardFunc func(string) bool

var idcardRegex = regexp.MustCompile(`^(\d{15}|\d{17}[\dXx])$`)

func isIdcardInternal(s string) bool {
	if IsIdcardFunc != nil {
		return IsIdcardFunc(s)
	}
	return idcardRegex.MatchString(s)
}

// Idcard represents a Chinese national ID card number.
// It implements validation for Chinese ID card numbers and supports
// both sql.Scanner and driver.Valuer interfaces for database operations.
type Idcard string

// NewIdcard creates a new Idcard with validation.
func NewIdcard(s string) (Idcard, error) {
	if s == "" {
		return "", nil
	}
	if !isIdcardInternal(s) {
		return "", fmt.Errorf("invalid ID card number: %s", s)
	}
	return Idcard(s), nil
}

// Scan implements the sql.Scanner interface.
// It scans database values into Idcard with validation.
func (e *Idcard) Scan(value any) error {
	var es string
	switch v := value.(type) {
	case []byte:
		es = string(v)
	case string:
		es = v
	default:
		return fmt.Errorf("failed to parse id_card: %v", v)
	}

	// Skip validation for empty strings
	if es == "" {
		*e = ""
		return nil
	}

	if !isIdcardInternal(es) {
		return fmt.Errorf("invalid ID card number: %s", es)
	}
	*e = Idcard(es)
	return nil
}

// Value implements the driver.Valuer interface.
// It returns the ID card number as a string for database storage.
func (e Idcard) Value() (driver.Value, error) {
	str := strings.TrimSpace(string(e))
	if str == "" {
		// Return nil for empty ID cards
		return nil, nil
	}
	if !isIdcardInternal(str) {
		return nil, fmt.Errorf("invalid ID card number: %s", str)
	}
	return str, nil
}

func (Idcard) GormDataType() string {
	return "idcard"
}

func (Idcard) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "VARCHAR(18)"
}

func (e *Idcard) String() string {
	return string(*e)
}

// MarshalJSON implements the json.Marshaler interface.
func (e Idcard) MarshalJSON() ([]byte, error) {
	str := strings.TrimSpace(string(e))
	if str == "" {
		return json.Marshal("")
	}
	if !isIdcardInternal(str) {
		return nil, fmt.Errorf("invalid ID card number: %s", str)
	}
	return json.Marshal(str)
}

func (e *Idcard) UnmarshalJSON(data []byte) error {
	str := misc.BytesToString(data)
	if str == "null" {
		return nil
	}
	str = strings.Trim(str, `"'`)
	if !isIdcardInternal(str) {
		return fmt.Errorf("failed to parse id_card: %v", str)
	}
	*e = Idcard(str)
	return nil
}

// NullIdcard represents an Idcard that may be null.
// It implements sql.Scanner, driver.Valuer, json.Marshaler/Unmarshaler, and GORM interfaces.
type NullIdcard struct {
	IDCard Idcard
	Valid  bool
}

// NewNullIdcard creates a new NullIdcard with the given string value.
func NewNullIdcard(s string) NullIdcard {
	if s == "" {
		return NullIdcard{Valid: false}
	}
	if !isIdcardInternal(s) {
		return NullIdcard{Valid: false}
	}
	return NullIdcard{IDCard: Idcard(s), Valid: true}
}

// Scan implements the sql.Scanner interface.
func (e *NullIdcard) Scan(value any) error {
	if value == nil {
		e.IDCard, e.Valid = "", false
		return nil
	}
	st := new(sql.NullString)
	err := st.Scan(value)
	if err != nil {
		return err
	}
	if !st.Valid {
		e.IDCard, e.Valid = "", false
		return nil
	}

	str := strings.TrimSpace(st.String)
	if !isIdcardInternal(str) {
		return fmt.Errorf("invalid ID card number: %s", str)
	}

	e.IDCard = Idcard(str)
	e.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (e NullIdcard) Value() (driver.Value, error) {
	if !e.Valid || e.IDCard == "" {
		return nil, nil
	}
	str := strings.TrimSpace(e.IDCard.String())
	if !isIdcardInternal(str) {
		return nil, fmt.Errorf("invalid ID card number: %s", str)
	}
	return str, nil
}

func (NullIdcard) GormDataType() string {
	return "idcard"
}

func (NullIdcard) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "VARCHAR(18)"
}

// MarshalJSON implements the json.Marshaler interface.
func (e NullIdcard) MarshalJSON() ([]byte, error) {
	if !e.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(e.IDCard)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *NullIdcard) UnmarshalJSON(data []byte) error {
	str := misc.BytesToString(data)
	if str == "null" || str == `""` || str == "" {
		e.IDCard = ""
		e.Valid = false
		return nil
	}

	str = strings.Trim(str, `"'`)
	str = strings.TrimSpace(str)

	if !isIdcardInternal(str) {
		return fmt.Errorf("invalid ID card number: %s", str)
	}

	e.IDCard = Idcard(str)
	e.Valid = true
	return nil
}
