# Response Handler Package (rsp)

A comprehensive HTTP response handling system for Go web applications that provides structured responses with support for multiple content types and automatic content negotiation.

## Features

- **Automatic Content Negotiation**: Support for JSON, JSONP, HTML, XML, and plain text responses based on Accept headers
- **Structured Error Handling**: Rich error reporting with problem details and field-specific validation errors
- **RESTful Helpers**: Pre-built functions for common HTTP status responses (OK, Created, Deleted, Accepted)
- **Functional Options Pattern**: Flexible response configuration using composable options
- **Validation Integration**: Seamless integration with go-slim.dev/v validation library
- **Standardized Response Format**: Consistent JSON structure across all response types

## Installation

```go
import "go-slim.dev/infra/rsp"
```

## Quick Start

### Basic Usage

```go
// Simple success response
rsp.Ok(c, userData)

// Created response with data
rsp.Created(c, newResource)

// Error response with custom options
rsp.Respond(c,
    rsp.StatusCode(400),
    rsp.Message("Invalid input"),
    rsp.Data(validationErrors),
)
```

### Response Format

All responses follow this standardized structure:

```json
{
  "code": "SUCCESS",
  "ok": true,
  "msg": "OK",
  "data": {...},           // optional
  "problems": {...},       // optional, for validation errors
  "error": "..."           // optional, only in debug mode
}
```

## API Reference

### HTTP Response Helpers

#### `Ok(c slim.Context, data ...any) error`

Responds with HTTP 200 status for successful requests.

#### `Created(c slim.Context, data ...any) error`

Responds with HTTP 201 status for successful resource creation.

#### `Deleted(c slim.Context, data ...any) error`

Responds with appropriate HTTP status for successful resource deletion.

- **With data**: Returns HTTP 200 (OK) status with data in response body
- **Without data**: Returns HTTP 204 (No Content) status with empty response body

Useful for deletion confirmation or returning deleted resource details.

#### `Accepted(c slim.Context, data ...any) error`

Responds with HTTP 202 status for accepted asynchronous operations.

### Configuration Options

#### `StatusCode(status int) Option`

Sets the HTTP status code for the response.

#### `Header(key, value string) Option`

Adds a custom HTTP header to the response.

#### `Cookie(cookie *http.Cookie) Option`

Sets an HTTP cookie in the response.

#### `Message(msg string) Option`

Sets a custom message for the response.

#### `Data(data any) Option`

Sets the data payload for the response.

### Error Handling

The package provides structured error reporting through the Problem system:

```go
type Problem struct {
    Code     string   `json:"code"`
    Message  string   `json:"msg"`
    Problems Problems `json:"problems,omitempty"`
}

type Problems map[string][]*Problem
```

## Examples

### Custom Response with Multiple Options

```go
rsp.Respond(c,
    rsp.StatusCode(http.StatusCreated),
    rsp.Header("Location", "/api/users/123"),
    rsp.Header("X-API-Version", "1.0"),
    rsp.Message("User created successfully"),
    rsp.Data(map[string]interface{}{
        "id": 123,
        "username": "john_doe",
    }),
)
```

### Error Response with Validation Problems

```go
problems := make(rsp.Problems)
problems.Add(&rsp.Problem{
    Label:   "email",
    Code:    "INVALID_FORMAT",
    Message: "Invalid email format",
})
problems.Add(&rsp.Problem{
    Label:   "password",
    Code:    "TOO_SHORT",
    Message: "Password must be at least 8 characters",
})

rsp.Respond(c,
    rsp.StatusCode(http.StatusBadRequest),
    rsp.Message("Validation failed"),
    rsp.Data(problems),
)
```

### Setting Cookies

```go
sessionCookie := &http.Cookie{
    Name:     "session_id",
    Value:    "abc123def456",
    Path:     "/",
    HttpOnly: true,
    Secure:   true,
    MaxAge:   3600,
}

rsp.Ok(c, userData)
// Or with custom cookie
rsp.Respond(c,
    rsp.Data(userData),
    rsp.Cookie(sessionCookie),
)
```

## Content Types Supported

The package automatically negotiates response content based on the `Accept` header:

- **JSON**: `application/json`
- **JSONP**: `application/javascript` (with callback)
- **HTML**: `text/html`
- **XML**: `application/xml`
- **Text**: `text/plain`, `text/*`

## Configuration

### Custom Marshalling

You can customize the HTML and text marshalling:

```go
rsp.HTMLMarshaller = func(data map[string]any) (string, error) {
    // Custom HTML rendering logic
    return renderTemplate("response", data), nil
}

rsp.TextMarshaller = func(data map[string]any) (string, error) {
    // Custom text formatting logic
    return formatAsText(data), nil
}
```

### JSONP Configuration

Configure JSONP callback parameter names:

```go
rsp.JsonpCallbacks = []string{"callback", "cb", "jsonp"}
rsp.DefaultJsonpCallback = "callback"
```

## Integration with Validation

The package integrates seamlessly with the `go-slim.dev/v` validation library:

```go
// Convert validation errors to problems
problems := make(rsp.Problems)
for _, err := range validationErrors.All() {
    problems.AddError(err)
}

rsp.Respond(c,
    rsp.StatusCode(http.StatusBadRequest),
    rsp.Data(problems),
)
```

## License

This package is part of the go-slim/infra project.
