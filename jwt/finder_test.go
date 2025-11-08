package jwt

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go-slim.dev/slim"
)

// mockResponseWriter implements http.ResponseWriter for testing
type mockResponseWriter struct {
	headers http.Header
	status  int
	body    []byte
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		headers: make(http.Header),
	}
}

func (m *mockResponseWriter) Header() http.Header {
	return m.headers
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.body = append(m.body, data...)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

func TestFromHeader(t *testing.T) {
	// Test FromHeader with real slim.Context
	tests := []struct {
		name          string
		authHeader    string
		expectedToken string
	}{
		{"valid Bearer token", "Bearer token123", "token123"},
		{"valid Bearer token uppercase", "BEARER uppercase", "uppercase"},
		{"valid Bearer token lowercase", "bearer lowercase", "lowercase"},
		{"valid Bearer token mixed case", "BeArEr mixed", "mixed"},
		{"too short", "Bear", ""},
		{"wrong prefix", "Basic dGVzdDp0ZXN0", ""},
		{"empty header", "", ""},
		{"empty token", "Bearer ", ""},
		{"token with spaces", "Bearer   token   ", "  token   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a real HTTP request
			req := &http.Request{
				Header: make(http.Header),
				Method: "GET",
				URL:    &url.URL{Path: "/test"},
			}

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create slim app and context
			s := slim.New()
			ctx := s.NewContext(&mockResponseWriter{}, req)

			// Test the actual FromHeader function
			result := FromHeader(ctx)
			assert.Equal(t, tt.expectedToken, result)
		})
	}
}

func TestFromWebSocket(t *testing.T) {
	// Test FromWebSocket with real slim.Context
	tests := []struct {
		name         string
		subprotocols []string
		expected     string
	}{
		{
			name: "valid Bearer token in subprotocol",
			subprotocols: []string{
				"Bearer.websocket-token-123",
				"other.protocol",
			},
			expected: "websocket-token-123",
		},
		{
			name: "valid Bearer token with uppercase in subprotocol",
			subprotocols: []string{
				"BEARER uppercase-websocket-token",
			},
			expected: "uppercase-websocket-token",
		},
		{
			name: "multiple subprotocols with Bearer tokens",
			subprotocols: []string{
				"other.protocol",
				"Bearer.second-token",
			},
			expected: "second-token",
		},
		{
			name: "subprotocol too short",
			subprotocols: []string{
				"Bea",
				"other.protocol",
			},
			expected: "",
		},
		{
			name: "subprotocol with wrong prefix",
			subprotocols: []string{
				"Basic.websocket-token",
				"other.protocol",
			},
			expected: "",
		},
		{
			name:         "no subprotocols",
			subprotocols: []string{},
			expected:     "",
		},
		{
			name: "subprotocol with empty token",
			subprotocols: []string{
				"Bearer ",
				"other.protocol",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a real HTTP request
			req := &http.Request{
				Header: make(http.Header),
				Method: "GET",
				URL:    &url.URL{Path: "/test"},
			}

			// Set Sec-Websocket-Protocol header with the subprotocols
			if len(tt.subprotocols) > 0 {
				req.Header.Set("Sec-Websocket-Protocol", strings.Join(tt.subprotocols, ", "))
			}

			// Create slim app and context
			s := slim.New()
			ctx := s.NewContext(newMockResponseWriter(), req)

			// Test FromWebSocket function
			result := FromWebSocket(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromQuery(t *testing.T) {
	// Test FromQuery with real slim.Context
	tests := []struct {
		name          string
		queryParams   map[string]string
		keys          []string
		expectedToken string
	}{
		{
			name: "get token from default jwt query param",
			queryParams: map[string]string{
				"jwt": "default-jwt-token",
			},
			keys:          nil,
			expectedToken: "default-jwt-token",
		},
		{
			name: "get token from specified query param",
			queryParams: map[string]string{
				"token": "query-token-value",
				"jwt":   "default-jwt-token",
			},
			keys:          []string{"token"},
			expectedToken: "query-token-value",
		},
		{
			name: "try multiple query params in order",
			queryParams: map[string]string{
				"param3": "param3-value",
				"param1": "param1-value",
				"jwt":    "default-jwt-token",
			},
			keys:          []string{"param1", "param2", "param3"},
			expectedToken: "param1-value",
		},
		{
			name: "fallback to default jwt query param",
			queryParams: map[string]string{
				"jwt": "default-jwt-token",
			},
			keys:          []string{"nonexistent"},
			expectedToken: "default-jwt-token",
		},
		{
			name:          "return empty when no query params found",
			queryParams:   map[string]string{},
			keys:          nil,
			expectedToken: "",
		},
		{
			name: "handle empty query param value",
			queryParams: map[string]string{
				"jwt": "",
			},
			keys:          nil,
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a real HTTP request with query parameters
			values := url.Values{}
			for key, value := range tt.queryParams {
				values.Set(key, value)
			}

			req := &http.Request{
				Header: make(http.Header),
				Method: "GET",
				URL:    &url.URL{Path: "/test", RawQuery: values.Encode()},
			}

			// Create slim app and context
			s := slim.New()
			ctx := s.NewContext(newMockResponseWriter(), req)

			// Test the actual FromQuery function
			result := FromQuery(ctx, tt.keys...)
			assert.Equal(t, tt.expectedToken, result)
		})
	}
}

func TestFromCookie(t *testing.T) {
	// Test FromCookie with real slim.Context
	tests := []struct {
		name     string
		cookies  map[string]string
		keys     []string
		expected string
	}{
		{
			name: "get token from default jwt cookie",
			cookies: map[string]string{
				"jwt": "default-jwt-token",
			},
			keys:     nil,
			expected: "default-jwt-token",
		},
		{
			name: "get token from specified cookie",
			cookies: map[string]string{
				"auth_token": "auth-token-value",
				"jwt":        "default-jwt-token",
			},
			keys:     []string{"auth_token"},
			expected: "auth-token-value",
		},
		{
			name: "try multiple cookies in order",
			cookies: map[string]string{
				"token3": "token3-value",
				"token1": "token1-value",
				"jwt":    "default-jwt-token",
			},
			keys:     []string{"token1", "token2", "token3"},
			expected: "token1-value",
		},
		{
			name: "fallback to default jwt cookie",
			cookies: map[string]string{
				"jwt": "default-jwt-token",
			},
			keys:     []string{"nonexistent"},
			expected: "default-jwt-token",
		},
		{
			name:     "return empty when no cookies found",
			cookies:  map[string]string{},
			keys:     nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a real HTTP request with cookies
			req := &http.Request{
				Header: make(http.Header),
				Method: "GET",
				URL:    &url.URL{Path: "/test"},
			}

			// Add cookies to the request
			for name, value := range tt.cookies {
				req.AddCookie(&http.Cookie{Name: name, Value: value})
			}

			// Create slim app and context
			s := slim.New()
			ctx := s.NewContext(newMockResponseWriter(), req)

			// Test the actual FromCookie function
			result := FromCookie(ctx, tt.keys...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultFinder(t *testing.T) {
	// Test DefaultFinder with real slim.Context - tests the priority logic
	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedToken  string
		expectedSource string
	}{
		{
			name: "websocket token has highest priority",
			setupRequest: func() *http.Request {
				req := &http.Request{
					Header: make(http.Header),
					Method: "GET",
					URL:    &url.URL{Path: "/test"},
				}
				// Set up WebSocket headers
				req.Header.Set("Connection", "upgrade")
				req.Header.Set("Upgrade", "websocket")
				req.Header.Set("Sec-WebSocket-Protocol", "Bearer.websocket-token")
				req.Header.Set("Authorization", "Bearer header-token")
				return req
			},
			expectedToken:  "websocket-token", // Would come from WebSocket subprotocol
			expectedSource: "websocket",
		},
		{
			name: "query token has second priority",
			setupRequest: func() *http.Request {
				req := &http.Request{
					Header: make(http.Header),
					Method: "GET",
					URL:    &url.URL{Path: "/test", RawQuery: "jwt=query-token"},
				}
				req.Header.Set("Authorization", "Bearer header-token")
				req.AddCookie(&http.Cookie{Name: "jwt", Value: "cookie-token"})
				return req
			},
			expectedToken:  "query-token",
			expectedSource: "query",
		},
		{
			name: "header token has third priority",
			setupRequest: func() *http.Request {
				req := &http.Request{
					Header: make(http.Header),
					Method: "GET",
					URL:    &url.URL{Path: "/test"},
				}
				req.Header.Set("Authorization", "Bearer header-token")
				req.AddCookie(&http.Cookie{Name: "jwt", Value: "cookie-token"})
				return req
			},
			expectedToken:  "header-token",
			expectedSource: "header",
		},
		{
			name: "cookie token has lowest priority",
			setupRequest: func() *http.Request {
				req := &http.Request{
					Header: make(http.Header),
					Method: "GET",
					URL:    &url.URL{Path: "/test"},
				}
				req.AddCookie(&http.Cookie{Name: "jwt", Value: "cookie-token"})
				return req
			},
			expectedToken:  "cookie-token",
			expectedSource: "cookie",
		},
		{
			name: "no tokens found",
			setupRequest: func() *http.Request {
				req := &http.Request{
					Header: make(http.Header),
					Method: "GET",
					URL:    &url.URL{Path: "/test"},
				}
				return req
			},
			expectedToken:  "",
			expectedSource: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()

			// Create slim app and context
			s := slim.New()
			ctx := s.NewContext(newMockResponseWriter(), req)

			// Test the actual DefaultFinder function
			result := DefaultFinder(ctx)
			assert.Equal(t, tt.expectedToken, result)
		})
	}
}
