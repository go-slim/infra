package dts

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"testing"

	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var (
	_ driver.Valuer                  = (*Email)(nil)
	_ sql.Scanner                    = (*Email)(nil)
	_ fmt.Stringer                   = (*Email)(nil)
	_ json.Marshaler                 = (*Email)(nil)
	_ json.Unmarshaler               = (*Email)(nil)
	_ schema.GormDataTypeInterface   = (*Email)(nil)
	_ migrator.GormDataTypeInterface = (*Email)(nil)

	_ driver.Valuer                  = (*NullEmail)(nil)
	_ sql.Scanner                    = (*NullEmail)(nil)
	_ json.Marshaler                 = (*NullEmail)(nil)
	_ json.Unmarshaler               = (*NullEmail)(nil)
	_ schema.GormDataTypeInterface   = (*NullEmail)(nil)
	_ migrator.GormDataTypeInterface = (*NullEmail)(nil)
)

// TestNullEmailEdgeCases tests the edge cases that were fixed
func TestNullEmailEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected NullEmail
		wantErr  bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: NullEmail{Valid: false},
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: NullEmail{Valid: false},
			wantErr:  false,
		},
		{
			name:     "valid email",
			input:    "test@example.com",
			expected: NullEmail{Email: "test@example.com", Valid: true},
			wantErr:  false,
		},
		{
			name:     "invalid email",
			input:    "invalid-email",
			expected: NullEmail{Valid: false},
			wantErr:  true,
		},
		{
			name:     "byte slice valid email",
			input:    []byte("test@example.com"),
			expected: NullEmail{Email: "test@example.com", Valid: true},
			wantErr:  false,
		},
		{
			name:     "byte slice empty",
			input:    []byte(""),
			expected: NullEmail{Valid: false},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ne NullEmail
			err := ne.Scan(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("NullEmail.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if ne.Valid != tt.expected.Valid {
				t.Errorf("NullEmail.Scan() Valid = %v, want %v", ne.Valid, tt.expected.Valid)
			}

			if ne.Valid && string(ne.Email) != string(tt.expected.Email) {
				t.Errorf("NullEmail.Scan() Email = %v, want %v", ne.Email, tt.expected.Email)
			}
		})
	}
}

// TestNullEmailJSONHandling tests JSON marshaling/unmarshaling edge cases
func TestNullEmailJSONHandling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData []byte
		expected NullEmail
		wantErr  bool
	}{
		{
			name:     "null JSON",
			jsonData: []byte("null"),
			expected: NullEmail{Valid: false},
			wantErr:  false,
		},
		{
			name:     "empty JSON",
			jsonData: []byte(""),
			expected: NullEmail{Valid: false},
			wantErr:  false,
		},
		{
			name:     "valid email JSON",
			jsonData: []byte(`"test@example.com"`),
			expected: NullEmail{Email: "test@example.com", Valid: true},
			wantErr:  false,
		},
		{
			name:     "invalid email JSON",
			jsonData: []byte(`"invalid-email"`),
			expected: NullEmail{Valid: false},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ne NullEmail
			err := ne.UnmarshalJSON(tt.jsonData)

			if (err != nil) != tt.wantErr {
				t.Errorf("NullEmail.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if ne.Valid != tt.expected.Valid {
				t.Errorf("NullEmail.UnmarshalJSON() Valid = %v, want %v", ne.Valid, tt.expected.Valid)
			}

			if ne.Valid && string(ne.Email) != string(tt.expected.Email) {
				t.Errorf("NullEmail.UnmarshalJSON() Email = %v, want %v", ne.Email, tt.expected.Email)
			}
		})
	}
}

// TestEmailJSONRoundTrip tests that JSON marshaling/unmarshaling works correctly
func TestEmailJSONRoundTrip(t *testing.T) {
	tests := []string{
		"test@example.com",
		"",
		"user.name+tag@domain.co.uk",
	}

	for _, email := range tests {
		t.Run(email, func(t *testing.T) {
			e := Email(email)

			// Marshal to JSON
			jsonData, err := e.MarshalJSON()
			if err != nil {
				t.Errorf("Email.MarshalJSON() error = %v", err)
				return
			}

			// Unmarshal from JSON
			var e2 Email
			err = e2.UnmarshalJSON(jsonData)
			if err != nil {
				t.Errorf("Email.UnmarshalJSON() error = %v", err)
				return
			}

			if string(e2) != email {
				t.Errorf("Email round trip = %v, want %v", string(e2), email)
			}
		})
	}
}

// TestNullEmailJSONRoundTrip tests that NullEmail JSON marshaling/unmarshaling works correctly
func TestNullEmailJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		email NullEmail
	}{
		{
			name:  "valid email",
			email: NullEmail{Email: "test@example.com", Valid: true},
		},
		{
			name:  "invalid email",
			email: NullEmail{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonData, err := tt.email.MarshalJSON()
			if err != nil {
				t.Errorf("NullEmail.MarshalJSON() error = %v", err)
				return
			}

			// Unmarshal from JSON
			var ne2 NullEmail
			err = ne2.UnmarshalJSON(jsonData)
			if err != nil {
				t.Errorf("NullEmail.UnmarshalJSON() error = %v", err)
				return
			}

			if ne2.Valid != tt.email.Valid {
				t.Errorf("NullEmail round trip Valid = %v, want %v", ne2.Valid, tt.email.Valid)
			}

			if ne2.Valid && string(ne2.Email) != string(tt.email.Email) {
				t.Errorf("NullEmail round trip Email = %v, want %v", ne2.Email, tt.email.Email)
			}
		})
	}
}
