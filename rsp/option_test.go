package rsp

import (
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestStatusCodeOption(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		wantCode int
	}{
		{
			name:     "Valid status code 200",
			status:   http.StatusOK,
			wantCode: http.StatusOK,
		},
		{
			name:     "Valid status code 201",
			status:   http.StatusCreated,
			wantCode: http.StatusCreated,
		},
		{
			name:     "Valid status code 400",
			status:   http.StatusBadRequest,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Valid status code 404",
			status:   http.StatusNotFound,
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Valid status code 500",
			status:   http.StatusInternalServerError,
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := options{}
			option := StatusCode(tt.status)
			option(&o)

			if o.status != tt.wantCode {
				t.Errorf("StatusCode() = %v, want %v", o.status, tt.wantCode)
			}
		})
	}
}

func TestStatusCodeOptionWithInvalidValues(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		wantCode int
	}{
		{
			name:     "Negative status code",
			status:   -1,
			wantCode: -1,
		},
		{
			name:     "Invalid range status code 250",
			status:   250,
			wantCode: 250,
		},
		{
			name:     "Invalid range status code 300",
			status:   300,
			wantCode: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := options{}
			option := StatusCode(tt.status)
			option(&o)

			if o.status != tt.wantCode {
				t.Errorf("StatusCode() = %v, want %v", o.status, tt.wantCode)
			}
		})
	}
}

func TestHeaderOption(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		expectNil bool
		expectLen int
	}{
		{
			name:      "Single header",
			key:       "Content-Type",
			value:     "application/json",
			expectNil: false,
			expectLen: 1,
		},
		{
			name:      "Custom header",
			key:       "X-Custom-Header",
			value:     "custom-value",
			expectNil: false,
			expectLen: 1,
		},
		{
			name:      "Cache control header",
			key:       "Cache-Control",
			value:     "no-cache",
			expectNil: false,
			expectLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := options{}
			option := Header(tt.key, tt.value)
			option(&o)

			if tt.expectNil && o.headers != nil {
				t.Error("Header() expected nil headers, got not nil")
			}

			if !tt.expectNil && o.headers == nil {
				t.Error("Header() expected not nil headers, got nil")
			}

			if o.headers != nil {
				if len(o.headers) != tt.expectLen {
					t.Errorf("Header() headers length = %v, want %v", len(o.headers), tt.expectLen)
				}

				if o.headers[tt.key] != tt.value {
					t.Errorf("Header() header value = %v, want %v", o.headers[tt.key], tt.value)
				}
			}
		})
	}
}

func TestMultipleHeaderOptions(t *testing.T) {
	o := options{}

	// Add multiple headers
	Header("Header1", "Value1")(&o)
	Header("Header2", "Value2")(&o)
	Header("Header3", "Value3")(&o)

	if len(o.headers) != 3 {
		t.Errorf("Multiple headers length = %v, want 3", len(o.headers))
	}

	if o.headers["Header1"] != "Value1" {
		t.Errorf("Header1 value = %v, want Value1", o.headers["Header1"])
	}

	if o.headers["Header2"] != "Value2" {
		t.Errorf("Header2 value = %v, want Value2", o.headers["Header2"])
	}

	if o.headers["Header3"] != "Value3" {
		t.Errorf("Header3 value = %v, want Value3", o.headers["Header3"])
	}
}

func TestHeaderOptionOverwrite(t *testing.T) {
	o := options{}

	// Add header
	Header("Same-Header", "First-Value")(&o)

	// Overwrite header with same key
	Header("Same-Header", "Second-Value")(&o)

	if len(o.headers) != 1 {
		t.Errorf("Header() expected 1 header after overwrite, got %v", len(o.headers))
	}

	if o.headers["Same-Header"] != "Second-Value" {
		t.Errorf("Header() value = %v, want Second-Value", o.headers["Same-Header"])
	}
}

func TestCookieOption(t *testing.T) {
	tests := []struct {
		name      string
		cookie    *http.Cookie
		expectNil bool
		expectLen int
	}{
		{
			name: "Basic cookie",
			cookie: &http.Cookie{
				Name:  "session",
				Value: "abc123",
			},
			expectNil: false,
			expectLen: 1,
		},
		{
			name: "Cookie with all attributes",
			cookie: &http.Cookie{
				Name:     "auth",
				Value:    "token123",
				Path:     "/",
				Domain:   "example.com",
				MaxAge:   3600,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Expires:  time.Now().Add(24 * time.Hour),
			},
			expectNil: false,
			expectLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := options{}
			option := Cookie(tt.cookie)
			option(&o)

			if tt.expectNil && o.cookies != nil {
				t.Error("Cookie() expected nil cookies, got not nil")
			}

			if !tt.expectNil && o.cookies == nil {
				t.Error("Cookie() expected not nil cookies, got nil")
			}

			if o.cookies != nil {
				if len(o.cookies) != tt.expectLen {
					t.Errorf("Cookie() cookies length = %v, want %v", len(o.cookies), tt.expectLen)
				}

				if o.cookies[0].Name != tt.cookie.Name {
					t.Errorf("Cookie() name = %v, want %v", o.cookies[0].Name, tt.cookie.Name)
				}

				if o.cookies[0].Value != tt.cookie.Value {
					t.Errorf("Cookie() value = %v, want %v", o.cookies[0].Value, tt.cookie.Value)
				}
			}
		})
	}
}

func TestMultipleCookieOptions(t *testing.T) {
	o := options{}

	cookie1 := &http.Cookie{Name: "cookie1", Value: "value1"}
	cookie2 := &http.Cookie{Name: "cookie2", Value: "value2"}
	cookie3 := &http.Cookie{Name: "cookie3", Value: "value3"}

	// Add multiple cookies
	Cookie(cookie1)(&o)
	Cookie(cookie2)(&o)
	Cookie(cookie3)(&o)

	if len(o.cookies) != 3 {
		t.Errorf("Multiple cookies length = %v, want 3", len(o.cookies))
	}

	// Check all cookies
	cookieNames := make(map[string]bool)
	for _, cookie := range o.cookies {
		cookieNames[cookie.Name] = true
	}

	expectedNames := []string{"cookie1", "cookie2", "cookie3"}
	for _, name := range expectedNames {
		if !cookieNames[name] {
			t.Errorf("Cookie() expected cookie %v not found", name)
		}
	}
}

func TestCookieOptionOverwrite(t *testing.T) {
	o := options{}

	// Add cookie
	cookie1 := &http.Cookie{Name: "same-cookie", Value: "first-value"}
	Cookie(cookie1)(&o)

	// Overwrite cookie with same name
	cookie2 := &http.Cookie{Name: "same-cookie", Value: "second-value", Path: "/new-path"}
	Cookie(cookie2)(&o)

	if len(o.cookies) != 1 {
		t.Errorf("Cookie() expected 1 cookie after overwrite, got %v", len(o.cookies))
	}

	if o.cookies[0].Value != "second-value" {
		t.Errorf("Cookie() value = %v, want second-value", o.cookies[0].Value)
	}

	if o.cookies[0].Path != "/new-path" {
		t.Errorf("Cookie() path = %v, want /new-path", o.cookies[0].Path)
	}
}

func TestMessageOption(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantMsg string
	}{
		{
			name:    "Simple message",
			message: "Operation successful",
			wantMsg: "Operation successful",
		},
		{
			name:    "Empty message",
			message: "",
			wantMsg: "",
		},
		{
			name:    "Long message",
			message: "This is a very long error message with lots of details about what went wrong during the processing of the request",
			wantMsg: "This is a very long error message with lots of details about what went wrong during the processing of the request",
		},
		{
			name:    "Unicode message",
			message: "Message with 中文 and 🚀 emoji",
			wantMsg: "Message with 中文 and 🚀 emoji",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := options{}
			option := Message(tt.message)
			option(&o)

			if o.message != tt.wantMsg {
				t.Errorf("Message() = %v, want %v", o.message, tt.wantMsg)
			}
		})
	}
}

func TestDataOption(t *testing.T) {
	type TestData struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Items []int  `json:"items"`
	}

	tests := []struct {
		name    string
		data    any
		wantNil bool
	}{
		{
			name:    "Nil data",
			data:    nil,
			wantNil: true,
		},
		{
			name:    "String data",
			data:    "simple string",
			wantNil: false,
		},
		{
			name:    "Number data",
			data:    42,
			wantNil: false,
		},
		{
			name:    "Boolean data",
			data:    true,
			wantNil: false,
		},
		{
			name:    "Map data",
			data:    map[string]interface{}{"key": "value", "number": 123},
			wantNil: false,
		},
		{
			name:    "Slice data",
			data:    []string{"item1", "item2", "item3"},
			wantNil: false,
		},
		{
			name: "Struct data",
			data: TestData{
				ID:    1,
				Name:  "test",
				Items: []int{1, 2, 3},
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := options{}
			option := Data(tt.data)
			option(&o)

			if tt.wantNil && o.data != nil {
				t.Error("Data() expected nil data, got not nil")
			}

			if !tt.wantNil && o.data == nil {
				t.Error("Data() expected not nil data, got nil")
			}

			if !tt.wantNil {
				// Use DeepEqual for all types to avoid comparison of uncomparable types
				if !reflect.DeepEqual(o.data, tt.data) {
					t.Errorf("Data() = %v, want %v", o.data, tt.data)
				}
			}
		})
	}
}

func TestDataOptionOverwrite(t *testing.T) {
	o := options{}

	// Set initial data
	Data("first data")(&o)

	// Overwrite with different data
	Data(map[string]interface{}{"key": "value"})(&o)

	if o.data == nil {
		t.Error("Data() expected not nil data, got nil")
	}

	// Check if data was overwritten
	dataMap, ok := o.data.(map[string]interface{})
	if !ok {
		t.Error("Data() expected map[string]interface{}, got different type")
		return
	}

	if dataMap["key"] != "value" {
		t.Errorf("Data() key value = %v, want value", dataMap["key"])
	}
}

func TestMultipleOptionsComposition(t *testing.T) {
	o := options{}

	// Apply multiple options
	StatusCode(http.StatusCreated)(&o)
	Header("X-Test", "test-value")(&o)
	Message("Test message")(&o)
	Data(map[string]interface{}{"test": true})(&o)
	Cookie(&http.Cookie{Name: "test-cookie", Value: "test-value"})(&o)

	// Verify all options were applied
	if o.status != http.StatusCreated {
		t.Errorf("Status code = %v, want %v", o.status, http.StatusCreated)
	}

	if o.headers["X-Test"] != "test-value" {
		t.Errorf("Header value = %v, want test-value", o.headers["X-Test"])
	}

	if o.message != "Test message" {
		t.Errorf("Message = %v, want Test message", o.message)
	}

	if o.data == nil {
		t.Error("Data should not be nil")
	}

	if len(o.cookies) != 1 {
		t.Errorf("Cookies length = %v, want 1", len(o.cookies))
	}

	if o.cookies[0].Name != "test-cookie" {
		t.Errorf("Cookie name = %v, want test-cookie", o.cookies[0].Name)
	}
}

func TestOptionOrderIndependence(t *testing.T) {
	o1 := options{}
	o2 := options{}

	// Apply options in different order
	StatusCode(http.StatusBadRequest)(&o1)
	Message("Error message")(&o1)
	Data(map[string]interface{}{"error": true})(&o1)

	// Reverse order
	Data(map[string]interface{}{"error": true})(&o2)
	Message("Error message")(&o2)
	StatusCode(http.StatusBadRequest)(&o2)

	// Both should have the same result
	if o1.status != o2.status {
		t.Errorf("Status codes differ: %v vs %v", o1.status, o2.status)
	}

	if o1.message != o2.message {
		t.Errorf("Messages differ: %v vs %v", o1.message, o2.message)
	}

	// Compare data (both should be maps)
	data1, ok1 := o1.data.(map[string]interface{})
	data2, ok2 := o2.data.(map[string]interface{})

	if !ok1 || !ok2 {
		t.Error("Both data should be maps")
		return
	}

	if data1["error"] != data2["error"] {
		t.Errorf("Data error values differ: %v vs %v", data1["error"], data2["error"])
	}
}

func TestDefaultOptionsValues(t *testing.T) {
	o := options{}

	// Check default values
	if o.status != 0 {
		t.Errorf("Default status should be 0, got %v", o.status)
	}

	if o.headers != nil {
		t.Error("Default headers should be nil")
	}

	if o.cookies != nil {
		t.Error("Default cookies should be nil")
	}

	if o.message != "" {
		t.Errorf("Default message should be empty, got %v", o.message)
	}

	if o.data != nil {
		t.Error("Default data should be nil")
	}

	if o.err != nil {
		t.Error("Default error should be nil")
	}
}

// Benchmarks for individual options
func BenchmarkOptionStatusCode(b *testing.B) {
	o := &options{}
	option := StatusCode(http.StatusOK)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		option(o)
	}
}

func BenchmarkOptionHeader(b *testing.B) {
	o := &options{}
	option := Header("X-Custom-Header", "custom-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		option(o)
	}
}

func BenchmarkOptionMultipleHeaders(b *testing.B) {
	o := &options{}
	headers := map[string]string{
		"X-API-Version": "1.0",
		"X-Request-ID":  "req-12345",
		"Cache-Control": "no-cache",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k, v := range headers {
			Header(k, v)(o)
		}
	}
}

func BenchmarkOptionCookie(b *testing.B) {
	o := &options{}
	cookie := &http.Cookie{
		Name:  "session",
		Value: "sess-abc123",
		Path:  "/",
	}
	option := Cookie(cookie)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		option(o)
	}
}

func BenchmarkOptionMessage(b *testing.B) {
	o := &options{}
	option := Message("Operation completed successfully")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		option(o)
	}
}

func BenchmarkOptionData(b *testing.B) {
	o := &options{}
	option := Data(map[string]interface{}{
		"id":    123,
		"name":  "test",
		"email": "test@example.com",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		option(o)
	}
}

func BenchmarkParallelOptions(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			o := &options{}
			StatusCode(http.StatusOK)(o)
			Header("X-Test", "value")(o)
			Message("test message")(o)
			Data(map[string]string{"key": "value"})(o)
		}
	})
}

func BenchmarkOptionCombination(b *testing.B) {
	optionList := []Option{
		StatusCode(http.StatusCreated),
		Message("Operation completed successfully"),
		Header("X-API-Version", "1.0"),
		Header("X-Request-ID", "req-12345"),
		Header("Cache-Control", "no-cache"),
		Data(map[string]interface{}{
			"id":   123,
			"name": "test",
		}),
		Cookie(&http.Cookie{
			Name:     "session",
			Value:    "sess-abc123",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o := &options{}
		for _, opt := range optionList {
			opt(o)
		}
	}
}
