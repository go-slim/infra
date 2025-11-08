package rsp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-slim.dev/slim"
	"go-slim.dev/v"
)

type fundamental struct {
	cause  error  // 原始错误对象
	status int    // HTTP 状态码
	code   string // 请求错误码
	text   string // 响应提示消息
	data   any    // 错误携带的响应数据
}

func (e *fundamental) Status() int   { return e.status }
func (e *fundamental) Code() string  { return e.code }
func (e *fundamental) Text() string  { return e.text }
func (e *fundamental) Data() any     { return e.data }
func (e *fundamental) Cause() error  { return e.cause }
func (e *fundamental) Error() string { return e.text }

var (
	ErrOK         = &fundamental{status: 200, code: "OK", text: "ok"}
	ErrBadRequest = &fundamental{status: 400, code: "BadRequest", text: "请求无效"}
	ErrInternal   = &fundamental{status: 500, code: "InternalError", text: "系统内部错误"}
)

// createContext creates a real slim.Context for testing
func createContext() (slim.Context, *httptest.ResponseRecorder) {
	s := slim.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	ctx := s.NewContext(recorder, request)
	return ctx, recorder
}

// createContextWithAccept creates a slim.Context with specific Accept header
func createContextWithAccept(acceptType string) (slim.Context, *httptest.ResponseRecorder) {
	s := slim.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("Accept", acceptType)
	ctx := s.NewContext(recorder, request)
	return ctx, recorder
}

// createContextWithMethod creates a slim.Context with specific HTTP method
func createContextWithMethod(method string) (slim.Context, *httptest.ResponseRecorder) {
	s := slim.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, "/", nil)
	ctx := s.NewContext(recorder, request)
	return ctx, recorder
}

// createContextWithQuery creates a slim.Context with query parameters
func createContextWithQuery(query string) (slim.Context, *httptest.ResponseRecorder) {
	s := slim.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/?"+query, nil)
	ctx := s.NewContext(recorder, request)
	return ctx, recorder
}

// createContextWithDebug creates a slim.Context with debug mode enabled/disabled
func createContextWithDebug(debug bool) (slim.Context, *httptest.ResponseRecorder) {
	s := slim.New()
	s.Debug = debug
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	ctx := s.NewContext(recorder, request)
	return ctx, recorder
}

// Test data structures
type TestData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestOk(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		accept   string
		wantCode int
	}{
		{
			name:     "Ok without data",
			data:     nil,
			accept:   "application/json",
			wantCode: http.StatusOK,
		},
		{
			name:     "Ok with data",
			data:     TestData{ID: 1, Name: "test"},
			accept:   "application/json",
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, recorder := createContextWithAccept(tt.accept)

			err := Ok(ctx, tt.data)

			if err != nil {
				t.Errorf("Ok() error = %v", err)
				return
			}

			if recorder.Code != tt.wantCode {
				t.Errorf("Ok() status = %v, want %v", recorder.Code, tt.wantCode)
			}

			// Verify response structure
			var response map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Errorf("Ok() invalid JSON response = %v", err)
				return
			}

			if response["ok"] != true {
				t.Errorf("Ok() ok field = %v, want true", response["ok"])
			}

			if tt.data != nil && response["data"] == nil {
				t.Error("Ok() expected data field but got nil")
			}
		})
	}
}

func TestCreated(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		wantCode int
	}{
		{
			name:     "Created without data",
			data:     nil,
			wantCode: http.StatusCreated,
		},
		{
			name:     "Created with data",
			data:     TestData{ID: 2, Name: "created"},
			wantCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, recorder := createContext()

			err := Created(ctx, tt.data)

			if err != nil {
				t.Errorf("Created() error = %v", err)
				return
			}

			if recorder.Code != tt.wantCode {
				t.Errorf("Created() status = %v, want %v", recorder.Code, tt.wantCode)
			}

			var response map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Errorf("Created() invalid JSON response = %v", err)
				return
			}

			if response["ok"] != true {
				t.Errorf("Created() ok field = %v, want true", response["ok"])
			}
		})
	}
}

func TestDeleted(t *testing.T) {
	ctx, recorder := createContext()
	data := TestData{ID: 3, Name: "deleted"}

	err := Deleted(ctx, data)

	if err != nil {
		t.Errorf("Deleted() error = %v", err)
		return
	}

	if recorder.Code != http.StatusOK {
		t.Errorf("Deleted() status = %v, want %v", recorder.Code, http.StatusOK)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Deleted() invalid JSON response = %v", err)
		return
	}

	if response["ok"] != true {
		t.Errorf("Deleted() ok field = %v, want true", response["ok"])
	}
}

func TestAccepted(t *testing.T) {
	ctx, recorder := createContext()

	err := Accepted(ctx)

	if err != nil {
		t.Errorf("Accepted() error = %v", err)
		return
	}

	if recorder.Code != http.StatusAccepted {
		t.Errorf("Accepted() status = %v, want %v", recorder.Code, http.StatusAccepted)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Accepted() invalid JSON response = %v", err)
		return
	}

	if response["ok"] != true {
		t.Errorf("Accepted() ok field = %v, want true", response["ok"])
	}
}

func TestRespondWithDifferentContentTypes(t *testing.T) {
	data := TestData{ID: 4, Name: "test"}
	tests := []struct {
		name     string
		accept   string
		wantCode int
		wantType string
	}{
		{
			name:     "JSON response",
			accept:   "application/json",
			wantCode: http.StatusOK,
			wantType: "application/json",
		},
		{
			name:     "HTML response",
			accept:   "text/html",
			wantCode: http.StatusOK,
			wantType: "text/html",
		},
		{
			name:     "Text response",
			accept:   "text/plain",
			wantCode: http.StatusOK,
			wantType: "text/plain",
		},
		{
			name:     "XML response",
			accept:   "application/xml",
			wantCode: http.StatusOK,
			wantType: "application/json", // XML falls back to JSON for now
		},
		{
			name:     "JSONP response",
			accept:   "application/javascript",
			wantCode: http.StatusOK,
			wantType: "application/javascript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, recorder := createContextWithAccept(tt.accept)

			if tt.accept == "application/javascript" {
				// For JSONP, need both Accept header and callback query parameter
				s := slim.New()
				recorder = httptest.NewRecorder()
				request := httptest.NewRequest("GET", "/?callback=testCallback", nil)
				request.Header.Set("Accept", "application/javascript")
				ctx = s.NewContext(recorder, request)
			}

			err := Respond(ctx, Data(data))

			if err != nil {
				t.Errorf("Respond() error = %v", err)
				return
			}

			contentType := recorder.Header().Get("Content-Type")
			if !strings.HasPrefix(contentType, tt.wantType) {
				t.Errorf("Respond() content-type = %v, want prefix %v", contentType, tt.wantType)
			}
		})
	}
}

func TestRespondWithHeaders(t *testing.T) {
	ctx, recorder := createContext()

	err := Respond(ctx,
		Header("X-Custom-Header", "custom-value"),
		Header("Cache-Control", "no-cache"),
		Data(TestData{ID: 5, Name: "test"}),
	)

	if err != nil {
		t.Errorf("Respond() error = %v", err)
		return
	}

	if recorder.Header().Get("X-Custom-Header") != "custom-value" {
		t.Error("Expected custom header not found")
	}

	if recorder.Header().Get("Cache-Control") != "no-cache" {
		t.Error("Expected cache control header not found")
	}
}

func TestRespondWithCookies(t *testing.T) {
	ctx, recorder := createContext()
	cookie := &http.Cookie{
		Name:  "test-cookie",
		Value: "test-value",
		Path:  "/",
	}

	err := Respond(ctx,
		Cookie(cookie),
		Data(TestData{ID: 6, Name: "test"}),
	)

	if err != nil {
		t.Errorf("Respond() error = %v", err)
		return
	}

	// Check if cookie was set
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Errorf("Expected 1 cookie, got %d", len(cookies))
		return
	}

	if cookies[0].Name != "test-cookie" {
		t.Errorf("Expected cookie name 'test-cookie', got '%s'", cookies[0].Name)
	}

	if cookies[0].Value != "test-value" {
		t.Errorf("Expected cookie value 'test-value', got '%s'", cookies[0].Value)
	}
}

func TestRespondWithCustomMessage(t *testing.T) {
	customMsg := "Custom success message"
	ctx, recorder := createContext()

	err := Respond(ctx,
		Message(customMsg),
		Data(TestData{ID: 7, Name: "test"}),
	)

	if err != nil {
		t.Errorf("Respond() error = %v", err)
		return
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Respond() invalid JSON response = %v", err)
		return
	}

	if response["msg"] != customMsg {
		t.Errorf("Respond() message = %v, want %v", response["msg"], customMsg)
	}
}

func TestRespondWithError(t *testing.T) {
	ctx, recorder := createContext()

	err := Respond(ctx,
		StatusCode(http.StatusBadRequest),
		Data(map[string]string{"error": "Something went wrong"}),
	)

	if err != nil {
		t.Errorf("Respond() error = %v", err)
		return
	}

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Respond() status = %v, want %v", recorder.Code, http.StatusBadRequest)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Respond() invalid JSON response = %v", err)
		return
	}

	if response["ok"] != false {
		t.Errorf("Respond() ok field = %v, want false", response["ok"])
	}
}

func TestRespondWithValidationError(t *testing.T) {
	ctx, recorder := createContext()

	// Create validation errors using v.Errors which will be converted to problems
	valuer := v.Value("invalid-email", "email", "Email")
	valuer.Custom("INVALID_FORMAT", func(val any) any {
		return false
	}, v.ErrorFormat("Invalid email format"))

	validationErr := valuer.Validate()

	err := Respond(ctx,
		StatusCode(http.StatusBadRequest),
		Error(validationErr),
	)

	if err != nil {
		t.Errorf("Respond() error = %v", err)
		return
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Respond() invalid JSON response = %v", err)
		return
	}

	if response["ok"] != false {
		t.Errorf("Respond() ok field = %v, want false", response["ok"])
	}

	// Check if problems field exists
	if _, exists := response["problems"]; !exists {
		t.Error("Expected problems field in response")
	}
}

func TestToText(t *testing.T) {
	data := map[string]any{
		"message": "test message",
		"code":    "TEST_CODE",
	}

	result, err := toText(data)

	if err != nil {
		t.Errorf("toText() error = %v", err)
		return
	}

	// Should be valid JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Errorf("toText() result is not valid JSON = %v", err)
	}

	if parsed["message"] != data["message"] {
		t.Errorf("toText() message = %v, want %v", parsed["message"], data["message"])
	}
}

func TestRespondWithFundamentalError(t *testing.T) {
	ctx, recorder := createContext()

	fundamentalErr := ErrBadRequest

	err := Respond(ctx,
		StatusCode(http.StatusBadRequest),
		Message("Bad request"),
		Error(fundamentalErr),
	)

	if err != nil {
		t.Errorf("Respond() error = %v", err)
		return
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Respond() invalid JSON response = %v", err)
		return
	}

	if response["ok"] != false {
		t.Errorf("Respond() ok field = %v, want false", response["ok"])
	}

	if response["code"] != fundamentalErr.Code() {
		t.Errorf("Respond() code = %v, want %v", response["code"], fundamentalErr.Code())
	}
}

func TestJSONPResponse(t *testing.T) {
	// JSONP requires both callback query parameter AND Accept header
	s := slim.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/?callback=myCallback", nil)
	request.Header.Set("Accept", "application/javascript")
	ctx := s.NewContext(recorder, request)

	data := TestData{ID: 8, Name: "jsonp test"}

	err := Respond(ctx, Data(data))

	if err != nil {
		t.Errorf("Respond() error = %v", err)
		return
	}

	responseBody := recorder.Body.String()

	// Should contain the JSONP callback
	if !contains(responseBody, "myCallback(") {
		t.Errorf("JSONP response should contain callback function, got: %s", responseBody)
	}

	// Response contains full structure, not just raw data
	// Check for the data fields (accounting for JSON formatting with spaces)
	if !contains(responseBody, `"id"`) || !contains(responseBody, `"name"`) ||
		!contains(responseBody, `"jsonp test"`) {
		t.Errorf("JSONP response should contain the data, got: %s", responseBody)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}

func TestHeadRequest(t *testing.T) {
	ctx, recorder := createContextWithMethod(http.MethodHead)

	err := Respond(ctx, Data(TestData{ID: 9, Name: "head test"}))

	if err != nil {
		t.Errorf("Respond() error = %v", err)
		return
	}

	// HEAD requests should have no body
	if recorder.Body.Len() > 0 {
		t.Error("HEAD request should have no response body")
	}

	if recorder.Code != http.StatusOK {
		t.Errorf("HEAD request status = %v, want %v", recorder.Code, http.StatusOK)
	}
}

func TestResponseStructure(t *testing.T) {
	ctx, recorder := createContext()
	data := TestData{ID: 10, Name: "structure test"}

	err := Respond(ctx, Data(data))

	if err != nil {
		t.Errorf("Respond() error = %v", err)
		return
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Respond() invalid JSON response = %v", err)
		return
	}

	// Check required fields
	requiredFields := []string{"code", "ok", "msg"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Required field '%s' missing from response", field)
		}
	}

	// Check data field
	if response["data"] == nil {
		t.Error("Data field should not be nil when data is provided")
	}

	// Verify data structure
	dataMap, ok := response["data"].(map[string]any)
	if !ok {
		t.Error("Data field should be a map")
		return
	}

	if dataMap["id"] != float64(10) {
		t.Errorf("Data ID = %v, want %v", dataMap["id"], 10)
	}

	if dataMap["name"] != "structure test" {
		t.Errorf("Data name = %v, want %v", dataMap["name"], "structure test")
	}
}

func TestDebugMode(t *testing.T) {
	tests := []struct {
		name    string
		debug   bool
		wantErr bool
	}{
		{
			name:    "Debug mode enabled",
			debug:   true,
			wantErr: true,
		},
		{
			name:    "Debug mode disabled",
			debug:   false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, recorder := createContextWithDebug(tt.debug)

			// Respond with an error that will show debug info
			err := Respond(ctx,
				StatusCode(http.StatusInternalServerError),
				Message("Internal server error"),
				Error(ErrInternal),
			)

			if err != nil {
				t.Errorf("Respond() error = %v", err)
				return
			}

			var response map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Errorf("Invalid JSON response = %v", err)
				return
			}

			hasErrorField := response["error"] != nil
			if tt.wantErr && !hasErrorField {
				t.Error("Expected error field in debug mode")
			}

			if !tt.wantErr && hasErrorField {
				t.Error("Unexpected error field in production mode")
			}
		})
	}
}

// Benchmark data structures for rsp package
type BenchmarkUser struct {
	ID       int              `json:"id"`
	Username string           `json:"username"`
	Email    string           `json:"email"`
	Profile  BenchmarkProfile `json:"profile"`
}

type BenchmarkProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
}

// BenchmarkContext creates a real slim.Context for benchmarking
type BenchmarkContext struct {
	ctx      slim.Context
	recorder *httptest.ResponseRecorder
	app      *slim.Slim
}

func NewBenchmarkContext() *BenchmarkContext {
	s := slim.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	ctx := s.NewContext(recorder, request)

	return &BenchmarkContext{
		ctx:      ctx,
		recorder: recorder,
		app:      s,
	}
}

func (bc *BenchmarkContext) GetContext() slim.Context {
	return bc.ctx
}

func (bc *BenchmarkContext) GetRecorder() *httptest.ResponseRecorder {
	return bc.recorder
}

func (bc *BenchmarkContext) SetAccept(acceptType string) {
	bc.ctx.Request().Header.Set("Accept", acceptType)
}

// Test data for benchmarks
var (
	benchmarkUser = BenchmarkUser{
		ID:       123,
		Username: "john_doe",
		Email:    "john@example.com",
		Profile: BenchmarkProfile{
			FirstName: "John",
			LastName:  "Doe",
			Age:       30,
		},
	}

	benchmarkData = map[string]any{
		"users": []BenchmarkUser{benchmarkUser},
		"total": 1,
		"page":  1,
		"limit": 10,
	}

	benchmarkProblems = func() Problems {
		problems := make(Problems)
		problems.Add(&Problem{
			Label:   "email",
			Code:    "INVALID_FORMAT",
			Message: "Invalid email format",
		})
		problems.Add(&Problem{
			Label:   "password",
			Code:    "TOO_SHORT",
			Message: "Password must be at least 8 characters",
		})
		return problems
	}()
)

// Benchmarks for core response functions
func BenchmarkOk(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")

	for b.Loop() {
		_ = Ok(bc.GetContext(), benchmarkUser)
	}
}

func BenchmarkCreated(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")

	for b.Loop() {
		_ = Created(bc.GetContext(), benchmarkUser)
	}
}

func BenchmarkDeleted(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")

	for b.Loop() {
		_ = Deleted(bc.GetContext(), benchmarkUser)
	}
}

func BenchmarkAccepted(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")

	for b.Loop() {
		_ = Accepted(bc.GetContext(), map[string]string{"task_id": "12345"})
	}
}

// Benchmarks for Respond function with different configurations
func BenchmarkRespondSimple(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")

	for b.Loop() {
		_ = Respond(bc.GetContext(), Data(benchmarkUser))
	}
}

func BenchmarkRespondWithMultipleOptions(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")
	options := []Option{
		StatusCode(http.StatusCreated),
		Message("Operation completed successfully"),
		Header("X-API-Version", "1.0"),
		Header("X-Request-ID", "req-12345"),
		Data(benchmarkUser),
		Cookie(&http.Cookie{
			Name:  "session",
			Value: "sess-abc123",
			Path:  "/",
		}),
	}

	for b.Loop() {
		_ = Respond(bc.GetContext(), options...)
	}
}

func BenchmarkRespondWithLargeData(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")
	largeData := make([]BenchmarkUser, 100) // 100 users
	for i := range largeData {
		largeData[i] = benchmarkUser
		largeData[i].ID = i
	}

	for b.Loop() {
		_ = Respond(bc.GetContext(), Data(largeData))
	}
}

func BenchmarkRespondWithProblems(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")

	for b.Loop() {
		_ = Respond(bc.GetContext(),
			StatusCode(http.StatusBadRequest),
			Message("Validation failed"),
			Data(benchmarkProblems),
		)
	}
}

// Memory allocation benchmarks
func BenchmarkMemoryAllocationSimpleResponse(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")

	b.ReportAllocs()

	for b.Loop() {
		_ = Respond(bc.GetContext(), Data(map[string]string{"message": "hello"}))
	}
}

func BenchmarkMemoryAllocationComplexResponse(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")
	options := []Option{
		StatusCode(http.StatusCreated),
		Message("Operation completed successfully"),
		Header("X-API-Version", "1.0"),
		Data(benchmarkData),
		Cookie(&http.Cookie{
			Name:     "session",
			Value:    "sess-abc123",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		}),
	}

	b.ReportAllocs()

	for b.Loop() {
		_ = Respond(bc.GetContext(), options...)
	}
}

func BenchmarkMemoryAllocationWithProblems(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")

	b.ReportAllocs()

	for b.Loop() {
		_ = Respond(bc.GetContext(),
			StatusCode(http.StatusBadRequest),
			Message("Validation failed"),
			Data(benchmarkProblems),
		)
	}
}

// Comparison benchmarks
func BenchmarkRespondVsOk(b *testing.B) {
	bc1 := NewBenchmarkContext()
	bc1.SetAccept("application/json")

	bc2 := NewBenchmarkContext()
	bc2.SetAccept("application/json")

	b.Run("Ok", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_ = Ok(bc1.GetContext(), benchmarkUser)
		}
	})

	b.Run("Respond", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_ = Respond(bc2.GetContext(), Data(benchmarkUser))
		}
	})
}

func BenchmarkSmallVsLargeData(b *testing.B) {
	bc1 := NewBenchmarkContext()
	bc1.SetAccept("application/json")
	smallData := map[string]string{"message": "hello"}

	bc2 := NewBenchmarkContext()
	bc2.SetAccept("application/json")
	largeData := make([]BenchmarkUser, 100)
	for i := range largeData {
		largeData[i] = benchmarkUser
		largeData[i].ID = i
	}

	b.Run("SmallData", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_ = Respond(bc1.GetContext(), Data(smallData))
		}
	})

	b.Run("LargeData", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_ = Respond(bc2.GetContext(), Data(largeData))
		}
	})
}

func BenchmarkFewVsManyOptions(b *testing.B) {
	bc1 := NewBenchmarkContext()
	bc1.SetAccept("application/json")
	fewOptions := []Option{
		Data(benchmarkUser),
	}

	bc2 := NewBenchmarkContext()
	bc2.SetAccept("application/json")
	manyOptions := []Option{
		StatusCode(http.StatusCreated),
		Message("Operation completed successfully"),
		Header("X-API-Version", "1.0"),
		Header("X-Request-ID", "req-12345"),
		Header("Cache-Control", "no-cache"),
		Header("X-Rate-Limit", "1000"),
		Data(benchmarkUser),
		Cookie(&http.Cookie{Name: "session", Value: "sess-abc123", Path: "/"}),
		Cookie(&http.Cookie{Name: "user_pref", Value: "dark-mode", Path: "/"}),
	}

	b.Run("FewOptions", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_ = Respond(bc1.GetContext(), fewOptions...)
		}
	})

	b.Run("ManyOptions", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_ = Respond(bc2.GetContext(), manyOptions...)
		}
	})
}

// Content type negotiation benchmarks
func BenchmarkContentNegotiation(b *testing.B) {
	contentTypes := []string{
		"application/json",
		"text/html",
		"text/plain",
		"application/xml",
		"application/javascript",
	}

	for _, contentType := range contentTypes {
		b.Run(contentType, func(b *testing.B) {
			bc := NewBenchmarkContext()
			bc.SetAccept(contentType)

			if contentType == "application/javascript" {
				// Set up JSONP
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest("GET", "/?callback=test", nil)
				request.Header.Set("Accept", contentType)
				ctx := bc.app.NewContext(recorder, request)
				bc.ctx = ctx
			}

			b.ResetTimer()
			for b.Loop() {
				_ = Respond(bc.GetContext(), Data(benchmarkUser))
			}
		})
	}
}

// Parallel benchmarks
func BenchmarkParallelRespond(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bc := NewBenchmarkContext()
			bc.SetAccept("application/json")
			_ = Respond(bc.GetContext(), Data(benchmarkUser))
		}
	})
}

// Context creation benchmarks
func BenchmarkContextCreation(b *testing.B) {
	for b.Loop() {
		bc := NewBenchmarkContext()
		_ = bc.GetContext()
	}
}

func BenchmarkContextWithDifferentAcceptHeaders(b *testing.B) {
	acceptHeaders := []string{
		"application/json",
		"text/html",
		"text/plain",
		"application/xml",
	}

	for _, header := range acceptHeaders {
		b.Run(header, func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				bc := NewBenchmarkContext()
				bc.SetAccept(header)
				_ = bc.GetContext()
			}
		})
	}
}

// Error handling benchmarks
func BenchmarkErrorResponses(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")

	errorOptions := []Option{
		StatusCode(http.StatusBadRequest),
		Message("Bad request"),
		Data(benchmarkProblems),
	}

	for b.Loop() {
		_ = Respond(bc.GetContext(), errorOptions...)
	}
}

func BenchmarkFundamentalError(b *testing.B) {
	bc := NewBenchmarkContext()
	bc.SetAccept("application/json")

	for b.Loop() {
		_ = Respond(bc.GetContext(),
			StatusCode(http.StatusInternalServerError),
			Message("Internal error"),
			Error(ErrInternal),
		)
	}
}

// Real-world scenario benchmarks
func BenchmarkAPIResponse(b *testing.B) {
	// Simulate a typical API response with user data, headers, and cookies

	for b.Loop() {
		bc := NewBenchmarkContext()
		bc.SetAccept("application/json")

		userData := map[string]any{
			"id":          123,
			"username":    "john_doe",
			"email":       "john@example.com",
			"created_at":  "2023-01-01T00:00:00Z",
			"last_login":  "2023-12-01T12:30:00Z",
			"permissions": []string{"read", "write", "admin"},
		}

		_ = Respond(bc.GetContext(),
			StatusCode(http.StatusOK),
			Message("User retrieved successfully"),
			Header("X-API-Version", "1.0"),
			Header("X-Rate-Limit-Remaining", "999"),
			Header("Cache-Control", "private, max-age=300"),
			Data(userData),
			Cookie(&http.Cookie{
				Name:     "session",
				Value:    "sess-abc123",
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				MaxAge:   3600,
			}),
		)
	}
}

func BenchmarkValidationErrorResponse(b *testing.B) {
	// Simulate a validation error response with multiple field errors
	for b.Loop() {
		bc := NewBenchmarkContext()
		bc.SetAccept("application/json")

		problems := make(Problems)
		problems.Add(&Problem{
			Label:   "username",
			Code:    "REQUIRED",
			Message: "Username is required",
		})
		problems.Add(&Problem{
			Label:   "email",
			Code:    "INVALID_FORMAT",
			Message: "Invalid email format",
		})
		problems.Add(&Problem{
			Label:   "password",
			Code:    "TOO_SHORT",
			Message: "Password must be at least 8 characters",
		})
		problems.Add(&Problem{
			Label:   "password",
			Code:    "NO_UPPERCASE",
			Message: "Password must contain uppercase letters",
		})
		problems.Add(&Problem{
			Label:   "age",
			Code:    "UNDERAGE",
			Message: "Must be at least 18 years old",
		})

		_ = Respond(bc.GetContext(),
			StatusCode(http.StatusUnprocessableEntity),
			Message("Validation failed"),
			Data(problems),
		)
	}
}

// Debug tests from debug_test.go
func TestDebugStatus(t *testing.T) {
	ctx, recorder := createContext()

	// Test with direct StatusCode option
	err := Respond(ctx, StatusCode(http.StatusCreated))

	if err != nil {
		t.Errorf("Respond error = %v", err)
		return
	}

	if recorder.Code != http.StatusCreated {
		t.Errorf("Status = %v, want %v", recorder.Code, http.StatusCreated)
	}
}

// Trace tests from trace_test.go
func TestTraceStatusFlow(t *testing.T) {
	// Test options struct directly
	o := &options{}
	StatusCode(http.StatusCreated)(o)

	t.Logf("Options after StatusCode(201): status=%d, err=%v", o.status, o.err)

	// Test result function
	ctx, recorder := createContext()
	status, result := result(ctx, o)
	t.Logf("result() returned: status=%d, result=%+v", status, result)

	// Test actual respond
	err := respond(ctx, o)
	t.Logf("respond() returned: err=%v", err)

	// Check what actually gets written
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", recorder.Code)

		// Let's see what the response body contains
		var response map[string]any
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err == nil {
			t.Logf("Response JSON: %+v", response)
		}
	}
}

// Verify result tests from verify_result_test.go
func TestVerifyResultStatusZero(t *testing.T) {
	// Test what happens when status is 0
	ctx, _ := createContext()

	o := options{
		err:    ErrOK,
		status: 0, // Explicitly set status to 0
	}

	t.Logf("Options with status 0: status=%d, err=%v", o.status, o.err)

	status, result := result(ctx, &o)
	t.Logf("result() with status 0 returned: status=%d, result=%+v", status, result)

	if status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", status)
	}
}

func TestVerifyResultStatus201(t *testing.T) {
	// Test what happens when status is 201
	ctx, _ := createContext()

	o := options{
		err:    ErrOK,
		status: http.StatusCreated, // Explicitly set status to 201
	}

	t.Logf("Options with status 201: status=%d, err=%v", o.status, o.err)

	status, result := result(ctx, &o)
	t.Logf("result() with status 201 returned: status=%d, result=%+v", status, result)

	if status != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", status)
	}
}

func TestVerifyStepByStep(t *testing.T) {
	ctx, recorder := createContext()

	// Step 1: Create options
	o := options{
		err: ErrOK,
	}

	t.Logf("Step 1: Initial options: status=%d, err=%v", o.status, o.err)

	// Step 2: Apply StatusCode option
	StatusCode(http.StatusCreated)(&o)
	t.Logf("Step 2: After StatusCode(201): status=%d, err=%v", o.status, o.err)

	// Step 3: Call result
	status, result := result(ctx, &o)
	t.Logf("Step 3: After result: status=%d, result=%+v", status, result)

	// Step 4: Call respond
	err := respond(ctx, &o)
	t.Logf("Step 4: After respond: err=%v, recorder code=%d", err, recorder.Code)

	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", recorder.Code)
	}
}

// Integration tests from integration_test.go

// IntegrationTestContext creates a real slim.Context with additional testing utilities
type IntegrationTestContext struct {
	ctx      slim.Context
	recorder *httptest.ResponseRecorder
	s        *slim.Slim
}

func NewIntegrationTestContext() *IntegrationTestContext {
	s := slim.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	ctx := s.NewContext(recorder, request)

	return &IntegrationTestContext{
		ctx:      ctx,
		recorder: recorder,
		s:        s,
	}
}

func (itc *IntegrationTestContext) SetAccept(acceptType string) {
	itc.ctx.Request().Header.Set("Accept", acceptType)
}

func (itc *IntegrationTestContext) SetMethod(method string) {
	request := httptest.NewRequest(method, "/", nil)
	// Copy headers from the old request to preserve Accept and other headers
	for name, values := range itc.ctx.Request().Header {
		for _, value := range values {
			request.Header.Add(name, value)
		}
	}
	itc.ctx = itc.s.NewContext(itc.recorder, request)
}

func (itc *IntegrationTestContext) SetQuery(query string) {
	request := httptest.NewRequest("GET", "/?"+query, nil)
	// Copy headers from the old request to preserve Accept and other headers
	for name, values := range itc.ctx.Request().Header {
		for _, value := range values {
			request.Header.Add(name, value)
		}
	}
	itc.ctx = itc.s.NewContext(itc.recorder, request)
}

func (itc *IntegrationTestContext) SetDebug(debug bool) {
	itc.s.Debug = debug
}

func (itc *IntegrationTestContext) RequestFunc(f func(slim.Context) error) error {
	return f(itc.ctx)
}

func (itc *IntegrationTestContext) GetResponse() *httptest.ResponseRecorder {
	return itc.recorder
}

func TestIntegrationSuccessResponse(t *testing.T) {
	tests := []struct {
		name           string
		requestFunc    func(slim.Context) error
		expectedStatus int
		expectedOK     bool
		expectData     bool
	}{
		{
			name: "Ok with simple data",
			requestFunc: func(ctx slim.Context) error {
				ctx.Request().Header.Set("Accept", "application/json")
				return Ok(ctx, map[string]string{"message": "success"})
			},
			expectedStatus: http.StatusOK,
			expectedOK:     true,
			expectData:     true,
		},
		{
			name: "Created with complex data",
			requestFunc: func(ctx slim.Context) error {
				ctx.Request().Header.Set("Accept", "application/json")
				type User struct {
					ID    int    `json:"id"`
					Name  string `json:"name"`
					Email string `json:"email"`
				}
				user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}
				return Created(ctx, user)
			},
			expectedStatus: http.StatusCreated,
			expectedOK:     true,
			expectData:     true,
		},
		{
			name: "Deleted with message",
			requestFunc: func(ctx slim.Context) error {
				ctx.Request().Header.Set("Accept", "application/json")
				return Deleted(ctx, map[string]string{"message": "Resource deleted"})
			},
			expectedStatus: http.StatusOK,
			expectedOK:     true,
			expectData:     true,
		},
		{
			name: "Accepted with task info",
			requestFunc: func(ctx slim.Context) error {
				ctx.Request().Header.Set("Accept", "application/json")
				return Accepted(ctx, map[string]string{"task_id": "12345", "status": "processing"})
			},
			expectedStatus: http.StatusAccepted,
			expectedOK:     true,
			expectData:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itc := NewIntegrationTestContext()

			err := itc.RequestFunc(tt.requestFunc)
			if err != nil {
				t.Errorf("Request function error = %v", err)
				return
			}

			recorder := itc.GetResponse()
			if recorder.Code != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", recorder.Code, tt.expectedStatus)
			}

			var response map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Errorf("Invalid JSON response = %v", err)
				return
			}

			ok := response["ok"]
			if tt.expectedOK && ok != true {
				t.Error("Expected ok field to be true")
			}
			if !tt.expectedOK && ok != false {
				t.Error("Expected ok field to be false")
			}

			if tt.expectData && response["data"] == nil {
				t.Error("Expected data field")
			}
		})
	}
}

func TestIntegrationErrorResponses(t *testing.T) {
	tests := []struct {
		name           string
		requestFunc    func(slim.Context) error
		expectedStatus int
		expectedOK     bool
		expectProblems bool
	}{
		{
			name: "Bad request with problems",
			requestFunc: func(ctx slim.Context) error {
				ctx.Request().Header.Set("Accept", "application/json")
				problems := make(Problems)
				problems.Add(&Problem{
					Label:   "email",
					Code:    "INVALID_FORMAT",
					Message: "Invalid email format",
				})
				return Respond(ctx,
					StatusCode(http.StatusBadRequest),
					Message("Validation failed"),
					Data(problems),
				)
			},
			expectedStatus: http.StatusBadRequest,
			expectedOK:     false,
			expectProblems: true,
		},
		{
			name: "Internal server error",
			requestFunc: func(ctx slim.Context) error {
				ctx.Request().Header.Set("Accept", "application/json")
				return Respond(ctx,
					StatusCode(http.StatusInternalServerError),
					Message("Internal server error"),
					Error(ErrInternal),
				)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedOK:     false,
			expectProblems: false,
		},
		{
			name: "Not found",
			requestFunc: func(ctx slim.Context) error {
				ctx.Request().Header.Set("Accept", "application/json")
				return Respond(ctx,
					StatusCode(http.StatusNotFound),
					Message("Resource not found"),
				)
			},
			expectedStatus: http.StatusNotFound,
			expectedOK:     false,
			expectProblems: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itc := NewIntegrationTestContext()

			err := itc.RequestFunc(tt.requestFunc)
			if err != nil {
				t.Errorf("Request function error = %v", err)
				return
			}

			recorder := itc.GetResponse()
			if recorder.Code != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", recorder.Code, tt.expectedStatus)
			}

			var response map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Errorf("Invalid JSON response = %v", err)
				return
			}

			ok := response["ok"]
			if tt.expectedOK && ok != true {
				t.Error("Expected ok field to be true")
			}
			if !tt.expectedOK && ok != false {
				t.Error("Expected ok field to be false")
			}

			if tt.expectProblems && response["data"] == nil {
				t.Error("Expected data field with problems")
			}
		})
	}
}

func TestIntegrationContentNegotiation(t *testing.T) {
	tests := []struct {
		name           string
		acceptHeader   string
		expectedType   string
		expectedStatus int
	}{
		{
			name:           "JSON response",
			acceptHeader:   "application/json",
			expectedType:   "application/json",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "HTML response",
			acceptHeader:   "text/html",
			expectedType:   "text/html",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Plain text response",
			acceptHeader:   "text/plain",
			expectedType:   "text/plain",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "XML response",
			acceptHeader:   "application/xml",
			expectedType:   "application/json", // XML falls back to JSON for now
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itc := NewIntegrationTestContext()
			itc.SetAccept(tt.acceptHeader)

			err := Ok(itc.ctx, map[string]string{"message": "test"})
			if err != nil {
				t.Errorf("Ok() error = %v", err)
				return
			}

			recorder := itc.GetResponse()
			if recorder.Code != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", recorder.Code, tt.expectedStatus)
			}

			contentType := recorder.Header().Get("Content-Type")
			if !strings.Contains(contentType, tt.expectedType) {
				t.Errorf("Content-Type = %v, want to contain %v", contentType, tt.expectedType)
			}
		})
	}
}

func TestIntegrationJSONPResponse(t *testing.T) {
	tests := []struct {
		name           string
		callbackParam  string
		callbackValue  string
		expectCallback bool
	}{
		{
			name:           "JSONP with callback",
			callbackParam:  "callback",
			callbackValue:  "testCallback",
			expectCallback: true,
		},
		{
			name:           "JSONP with different param name",
			callbackParam:  "jsonp",
			callbackValue:  "myCallback",
			expectCallback: true,
		},
		{
			name:           "No JSONP parameter",
			callbackParam:  "",
			callbackValue:  "",
			expectCallback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itc := NewIntegrationTestContext()
			itc.SetAccept("application/javascript")

			if tt.callbackParam != "" {
				itc.SetQuery(tt.callbackParam + "=" + tt.callbackValue)
			}

			err := Ok(itc.ctx, map[string]string{"message": "test"})
			if err != nil {
				t.Errorf("Ok() error = %v", err)
				return
			}

			recorder := itc.GetResponse()
			body := recorder.Body.String()

			if tt.expectCallback {
				if !strings.Contains(body, tt.callbackValue) {
					t.Errorf("Response body should contain callback %q, got: %s", tt.callbackValue, body)
				}
				if !strings.Contains(body, "(") || !strings.Contains(body, ")") {
					t.Errorf("Response body should be wrapped in callback function, got: %s", body)
				}
			} else {
				if strings.Contains(body, "(") && strings.Contains(body, ")") {
					t.Errorf("Response body should not be wrapped in callback function, got: %s", body)
				}
			}
		})
	}
}
