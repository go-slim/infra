// Package rsp provides a comprehensive HTTP response handling system for Go web applications.
// It offers a unified way to create structured responses with support for multiple content types
// including JSON, JSONP, HTML, XML, and plain text. The package follows RESTful conventions
// and provides helper functions for common HTTP status responses.
//
// Key Features:
// - Automatic content negotiation based on Accept headers
// - Structured error reporting with problem details
// - Support for multiple response formats (JSON, JSONP, HTML, XML, Text)
// - Configurable marshaling for custom formats
// - Integration with validation errors and business logic errors
// - Standardized response structure with code, status, message, and data fields
//
// Typical Usage:
//
//	// Simple success response
//	rsp.Ok(c, userData)
//
//	// Created response with data
//	rsp.Created(c, newResource)
//
//	// Error response with custom options
//	rsp.Respond(c, rsp.StatusCode(400), rsp.Message("Invalid input"))
//
// The response format follows this structure:
//
//	{
//		"code": "SUCCESS",
//		"ok": true,
//		"msg": "OK",
//		"data": {...},           // optional
//		"problems": {...},       // optional, for validation errors
//		"error": "..."           // optional, only in debug mode
//	}
package rsp

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go-slim.dev/misc"
	"go-slim.dev/slim"
	"go-slim.dev/v"
)

var (
	// HTMLMarshaller converts response data maps to HTML format for client responses.
	// This function is used when the client accepts HTML content via the Accept header.
	// By default, it uses JSON formatting, but can be customized for proper HTML rendering.
	HTMLMarshaller func(map[string]any) (string, error)

	// TextMarshaller converts response data maps to plain text format for client responses.
	// This function is used when the client accepts text/plain or text/* content types.
	// By default, it uses JSON encoding for text output.
	TextMarshaller func(map[string]any) (string, error)

	// JsonpCallbacks defines the list of query parameter names to check for JSONP callback functions.
	// These parameters are checked in order to find the callback function name for JSONP responses.
	// Common values include "callback", "cb", and "jsonp".
	JsonpCallbacks []string

	// DefaultJsonpCallback specifies the default callback function name to use when
	// no JSONP callback is found in the query parameters. This ensures JSONP responses
	// always have a valid callback function name.
	DefaultJsonpCallback string
)

// init initializes the package with default values for marshalling functions
// and JSONP configuration. This ensures the package works out-of-the-box
// with sensible defaults.
func init() {
	TextMarshaller = toText
	HTMLMarshaller = toText
	JsonpCallbacks = []string{"callback", "cb", "jsonp"}
	DefaultJsonpCallback = "callback"
}

// toText is the default marshaller function that converts a response map to JSON text.
// It's used by both TextMarshaller and HTMLMarshaller by default, providing a simple
// JSON-based text representation of response data.
func toText(m map[string]any) (string, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(m); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Ok responds to a successful request with HTTP 200 status.
// It's the most common response method for successful operations.
// If data is provided, it will be included in the response body.
//
// Parameters:
//   - c: The slim.Context for the current request
//   - data: Optional data to include in the response (0 or 1 parameter)
//
// Returns:
//   - error: Any error that occurred during response writing
func Ok(c slim.Context, data ...any) error {
	return Respond(c, Data(cmp.Or(data...)))
}

// Created responds to a successful resource creation with HTTP 201 status.
// Use this when a new resource has been successfully created on the server.
// If data is provided, it will be included in the response body.
//
// Parameters:
//   - c: The slim.Context for the current request
//   - data: Optional data to include in the response (0 or 1 parameter)
//
// Returns:
//   - error: Any error that occurred during response writing
func Created(c slim.Context, data ...any) error {
	return Respond(c, StatusCode(http.StatusCreated), Data(cmp.Or(data...)))
}

// Deleted responds to a successful resource deletion with appropriate HTTP status.
// Use this when a resource has been successfully deleted from the server.
//
// This function follows HTTP best practices for deletion operations:
//   - If data is provided (such as deletion confirmation or deleted resource info),
//     it responds with HTTP 200 (OK) and includes the data in the response body
//   - If no data is provided, it responds with HTTP 204 (No Content) with an empty body
//
// The 200 status with data is useful when clients need confirmation that the deletion
// was successful or when returning the deleted resource's details for auditing purposes.
// The 204 status is appropriate when the deletion is successful but no meaningful
// response data needs to be returned.
//
// Parameters:
//   - c: The slim.Context for the current request
//   - data: Optional data to include in the response (0 or 1 parameter)
//   - When provided: responds with HTTP 200 + data in response body
//   - When nil or empty: responds with HTTP 204 + empty response body
//
// Returns:
//   - error: Any error that occurred during response writing
//
// Examples:
//
//	// Simple deletion without response data (HTTP 204)
//	err := rsp.Deleted(c)
//
//	// Deletion with confirmation data (HTTP 200)
//	err := rsp.Deleted(c, map[string]string{
//		"message": "User deleted successfully",
//		"deleted_at": "2023-12-01T10:30:00Z",
//	})
//
//	// Deletion returning the deleted resource details (HTTP 200)
//	err := rsp.Deleted(c, deletedUser)
func Deleted(c slim.Context, data ...any) error {
	if len(data) > 0 && data[0] != nil {
		// Data provided: use HTTP 200 (OK) with data in response body
		// This is useful for deletion confirmation or returning deleted resource info
		return Respond(c, StatusCode(http.StatusOK), Data(data[0]))
	}

	// No data provided: use HTTP 204 (No Content) with empty response body
	// This is the standard for successful deletion when no response data is needed
	return Respond(c, StatusCode(http.StatusNoContent))
}

// Accepted responds to an accepted asynchronous operation with HTTP 202 status.
// Use this for long-running operations that have been accepted but not yet completed,
// such as task scheduling, background processing, or batch operations.
//
// Parameters:
//   - c: The slim.Context for the current request
//   - data: Optional data to include in the response (0 or 1 parameter)
//
// Returns:
//   - error: Any error that occurred during response writing
func Accepted(c slim.Context, data ...any) error {
	return Respond(c, StatusCode(http.StatusAccepted), Data(cmp.Or(data...)))
}

// Respond is the core response function that handles all HTTP responses.
// It applies functional options to configure the response and then performs
// content negotiation to determine the appropriate response format.
//
// This function automatically handles:
// - Content type negotiation based on Accept headers
// - Error formatting and problem details
// - Response structure standardization
// - JSONP callback handling
// - Header and cookie setting
//
// Parameters:
//   - c: The slim.Context for the current request
//   - opts: Optional functions to configure the response behavior
//
// Returns:
//   - error: Any error that occurred during response writing
func Respond(c slim.Context, opts ...Option) error {
	o := options{}
	for _, option := range opts {
		option(&o)
	}
	return respond(c, &o)
}

func respond(c slim.Context, o *options) (err error) {
	// Ignore if response has already been written
	if c.Written() {
		return
	}

	for key, value := range o.headers {
		c.SetHeader(key, value)
	}

	for _, cookie := range o.cookies {
		c.SetCookie(cookie)
	}

	status, m := result(c, o)

	// HEAD requests have no response body
	if c.Request().Method == http.MethodHead {
		return c.NoContent(status)
	}

	// Respond with different formats based on Accept header
	switch c.Accepts("html", "json", "jsonp", "xml", "text", "text/*") {
	case "html":
		var html string
		if html, err = HTMLMarshaller(m); err == nil {
			err = c.HTML(status, html)
		}
	case "json":
		err = c.JSON(status, m)
	case "jsonp":
		qs := c.Request().URL.Query()
		for _, name := range JsonpCallbacks {
			if cb := qs.Get(name); cb != "" {
				err = c.JSONP(status, cb, m)
				return
			}
		}
		// No callback parameter found, fall back to JSON instead of using default callback
		err = c.JSON(status, m)
	case "xml":
		// Note: XML support is limited. For now, fall back to JSON
		// since XML marshalling of interface{} types is complex
		err = c.JSON(status, m)
	case "text", "text/*":
		var text string
		if text, err = TextMarshaller(m); err == nil {
			err = c.String(status, text)
		}
	default:
		err = c.JSON(status, m)
	}

	return
}

func result(c slim.Context, o *options) (int, slim.Map) {
	if status, m, ok := inferHTTPError(c, o); ok {
		return status, m
	}
	if status, m, ok := inferValidationError(o); ok {
		return status, m
	}
	if status, m, ok := inferFundamentalErrir(c, o); ok {
		return status, m
	}
	if o.err != nil {
		return inferMistaken(c, o)
	}
	return inferStatusCode(o)
}

func inferHTTPError(c slim.Context, o *options) (int, slim.Map, bool) {
	var he *slim.HTTPError
	if errors.As(o.err, &he) {
		opts := *o
		o.status = cmp.Or(o.status, he.Code)
		status, m := inferStatusCode(&opts)
		if !misc.IsZero(he.Message) && http.StatusText(status) != he.Message {
			m["msg"] = he.Message
		}
		if o.data != nil {
			m["data"] = o.data
		}
		if he.Internal != nil && c.Slim().Debug {
			m["error"] = fmt.Sprintf("%+v", he.Internal)
		}
		return status, m, true
	}
	return 0, nil, false
}

func inferValidationError(o *options) (int, slim.Map, bool) {
	problems := make(Problems)

	// Handle v.Errors (multiple validation errors)
	var verrs *v.Errors
	if errors.As(o.err, &verrs) && !verrs.IsEmpty() {
		for _, e := range verrs.All() {
			collectProblem(problems, e)
		}
	} else {
		// Handle single v.Error
		var verr *v.Error
		if !errors.As(o.err, &verr) {
			return 0, nil, false
		}
		collectProblem(problems, verr)
	}

	if len(problems) == 0 {
		return 0, nil, false
	}

	m := slim.Map{
		"code": "InvalidParams",
		"ok":   false,
		"msg":  cmp.Or(o.message, "Invalid parameters"),
	}
	if o.data != nil {
		m["data"] = o.data
	}
	m["problems"] = problems
	return cmp.Or(o.status, 400), m, true
}

func inferFundamentalErrir(c slim.Context, o *options) (int, slim.Map, bool) {
	var rerr Fundamental
	if o.err != nil && errors.As(o.err, &rerr) {
		status := cmp.Or(o.status, rerr.Status())
		m := slim.Map{
			"code": rerr.Code(),
			"ok":   status >= 200 && status < 300, // Only 2xx status codes indicate success
			"msg":  cmp.Or(o.message, rerr.Text()),
		}
		if o.data != nil {
			m["data"] = o.data
		} else if data := rerr.Data(); data != nil {
			m["data"] = data
		}
		if c.Slim().Debug {
			if err := rerr.Cause(); err != nil {
				m["error"] = fmt.Sprintf("%+v", err)
			} else {
				// Show the fundamental error itself in debug mode
				m["error"] = fmt.Sprintf("%+v", o.err)
			}
		}
		return status, m, true
	}
	return 0, nil, false
}

func inferMistaken(c slim.Context, o *options) (int, slim.Map) {
	status := o.status
	// 1. unset
	// 2. Only 2xx status codes indicate success
	if status <= 0 || (status >= 200 && status < 300) {
		status = http.StatusInternalServerError
	}

	m := slim.Map{
		"code": "InternalError",
		"ok":   false,
		"msg":  cmp.Or(o.message, "An unexpected error occurred"),
	}
	if o.data != nil {
		m["data"] = o.data
	}
	if c.Slim().Debug {
		m["error"] = fmt.Sprintf("%+v", o.err)
	}

	return status, m
}

func inferStatusCode(o *options) (int, slim.Map) {
	m := make(slim.Map)
	status := o.status
	switch {
	case status < 0:
		status = http.StatusInternalServerError
		m["ok"] = false
		m["msg"] = cmp.Or(o.message, "An unexpected error occurred")
		m["code"] = "InternalError"
	case status == 0:
		status = http.StatusOK
		m["ok"] = true
		m["msg"] = cmp.Or(o.message, "ok")
		m["code"] = "OK"
	case o.status < 200:
		m["ok"] = false
		m["msg"] = cmp.Or(o.message, "An unexpected error occurred")
		m["code"] = "InternalError"
	case status < 300:
		m["ok"] = true
		m["msg"] = "ok"
		m["code"] = "OK"
	case status < 400:
		// An error status code was set, we treat it as an internal error
		m["ok"] = false
		m["msg"] = cmp.Or(o.message, "An unexpected error occurred")
		m["code"] = "InternalError"
	case status < 500:
		m["ok"] = false
		m["msg"] = cmp.Or(o.message, "Bad request")
		m["code"] = "BadRequest"
	default:
		status = http.StatusInternalServerError
		m["ok"] = false
		m["msg"] = cmp.Or(o.message, "An unexpected error occurred")
		m["code"] = "InternalError"
	}
	if o.data != nil {
		m["data"] = o.data
	}
	return status, m
}
