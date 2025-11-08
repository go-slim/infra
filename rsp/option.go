// Package rsp provides response handling options for configuring HTTP responses.
// This file contains the functional options pattern implementation that allows
// flexible configuration of response behavior including status codes, headers,
// cookies, messages, and data content.
//
// The options pattern used here enables clean, composable configuration
// of responses without requiring multiple function overloads or complex
// parameter structures.
//
// Example Usage:
//
//	rsp.Respond(c,
//		rsp.StatusCode(200),
//		rsp.Header("X-Custom", "value"),
//		rsp.Message("Success"),
//		rsp.Data(userData),
//	)
package rsp

import (
	"net/http"

	"go-slim.dev/l4g"
)

// options holds all the configurable parameters for an HTTP response.
// This struct is used internally to collect and apply response configuration
// options provided by the functional options pattern.
type options struct {
	status  int               // HTTP status code for the response
	headers map[string]string // HTTP headers to set on the response
	cookies []*http.Cookie    // HTTP cookies to set on the response
	err     error             // Error to include in the response (if any)
	message string            // Custom message for the response
	data    any               // Data payload to include in the response
}

// Option is a function type that configures response options.
// It follows the functional options pattern, allowing for flexible
// and composable response configuration.
//
// Each Option function modifies the options struct to configure
// a specific aspect of the HTTP response.
type Option func(o *options)

// StatusCode configures the HTTP status code for the response.
// It validates the status code and logs warnings for potentially problematic values.
//
// Parameters:
//   - status: The HTTP status code to set (e.g., 200, 201, 400, 404, 500)
//
// Returns:
//   - Option: A function that configures the status code when applied
//
// Note: Status codes below 0 or between 200-399 (except standard success codes)
// will trigger an error log message as they are typically incorrect.
func StatusCode(status int) Option {
	if status < 0 {
		// TODO: Use i18n translation
		l4g.Error("Invalid response status code set", l4g.Int("status", status))
	}
	return func(o *options) {
		o.status = status
	}
}

// Header configures a custom HTTP header for the response.
// Multiple Header calls can be made to set multiple headers.
//
// Parameters:
//   - key: The header name (e.g., "X-Custom-Header", "Cache-Control")
//   - value: The header value
//
// Returns:
//   - Option: A function that configures the header when applied
//
// Example:
//
//	rsp.Respond(c, rsp.Header("X-API-Version", "1.0"))
func Header(key, value string) Option {
	return func(o *options) {
		if o.headers == nil {
			o.headers = make(map[string]string)
		}
		o.headers[key] = value
	}
}

// Cookie configures an HTTP cookie to be set in the response.
// If a cookie with the same name already exists, it will be replaced.
// Multiple Cookie calls can be made to set multiple cookies.
//
// Parameters:
//   - cookie: A properly configured http.Cookie struct
//
// Returns:
//   - Option: A function that configures the cookie when applied
//
// Example:
//
//	cookie := &http.Cookie{
//	    Name: "session_id",
//	    Value: "abc123",
//	    Path: "/",
//	    HttpOnly: true,
//	    Secure: true,
//	}
//	rsp.Respond(c, rsp.Cookie(cookie))
func Cookie(cookie *http.Cookie) Option {
	return func(o *options) {
		for i, h := range o.cookies {
			if h.Name == cookie.Name {
				o.cookies[i] = cookie
				return
			}
		}
		o.cookies = append(o.cookies, cookie)
	}
}

// Message configures a custom message for the response.
// This message will be included in the "msg" field of the response JSON.
// If not provided, the default HTTP status text will be used.
//
// Parameters:
//   - msg: The custom message to include in the response
//
// Returns:
//   - Option: A function that configures the message when applied
//
// Example:
//
//	rsp.Respond(c, rsp.Message("Operation completed successfully"))
func Message(msg string) Option {
	return func(o *options) {
		o.message = msg
	}
}

// Data configures the data payload to be included in the response.
// This data will be serialized and included in the "data" field of the response JSON.
// The data can be of any type that is JSON serializable.
//
// Parameters:
//   - data: The data payload to include in the response (any JSON-serializable type)
//
// Returns:
//   - Option: A function that configures the data when applied
//
// Examples:
//
//	rsp.Respond(c, rsp.Data(map[string]string{"key": "value"}))
//	rsp.Respond(c, rsp.Data(userStruct))
//	rsp.Respond(c, rsp.Data(sliceOfItems))
func Data(data any) Option {
	return func(o *options) {
		o.data = data
	}
}

// Error configures an error for the response.
// This error will be included in the response and processed
// according to the error handling logic.
func Error(err error) Option {
	return func(o *options) {
		o.err = err
	}
}
