// Package rsp provides structured problem handling for HTTP responses.
// This file implements a problem details system that allows for rich, structured
// error reporting with field-specific validation errors and nested problem hierarchies.
//
// The problem system is designed to work with validation libraries and provides
// detailed error information to clients while maintaining a consistent response format.
//
// Key Features:
// - Field-specific error reporting with labels and codes
// - Hierarchical problem organization with nested problems
// - Integration with validation error types from go-slim.dev/v
// - Support for "some" validation logic (at least one of multiple constraints)
// - JSON serialization for API responses
//
// Problem Structure:
//
//	{
//		"code": "VALIDATION_ERROR",
//		"msg": "Validation failed",
//		"problems": {
//			"email": [{"code": "INVALID_EMAIL", "msg": "Invalid email format"}],
//			"age": [{"code": "TOO_YOUNG", "msg": "Must be at least 18 years old"}]
//		}
//	}
package rsp

import (
	"errors"

	"go-slim.dev/v"
)

// Problem represents a single validation or business logic error.
// It provides structured error information with field identification,
// error codes, human-readable messages, and optional nested problems.
//
// Fields:
//   - Label: The field name or identifier this problem is associated with.
//     This field is not included in JSON serialization (json:"-" tag).
//   - Code: A machine-readable error code that can be used for programmatic
//     error handling and internationalization.
//   - Message: A human-readable error message describing the problem.
//   - Problems: Optional nested problems for complex validation scenarios
//     (e.g., "some" validation where at least one condition must be met).
type Problem struct {
	Label    string   `json:"-"`                  // Field identifier (not serialized)
	Code     string   `json:"code"`               // Machine-readable error code
	Message  string   `json:"msg"`                // Human-readable error message
	Problems Problems `json:"problems,omitempty"` // Nested problems (optional)
}

// Problems represents a collection of validation errors organized by field name.
// It provides a structured way to group multiple problems for different fields
// or aspects of a validation operation.
//
// The map key represents the field name or identifier, and the slice contains
// all problems associated with that field. This allows multiple validation errors
// per field, which is useful for complex validation rules.
//
// Example:
//
//	Problems{
//		"email": {
//			{Code: "INVALID_FORMAT", Message: "Invalid email format"},
//			{Code: "BLACKLISTED", Message: "Email domain is not allowed"},
//		},
//		"password": {
//			{Code: "TOO_SHORT", Message: "Password must be at least 8 characters"},
//		},
//	}
type Problems map[string][]*Problem

// Add adds a problem to the Problems collection, organizing it by field name.
// If the field doesn't exist yet in the collection, it creates a new slice.
// Multiple problems can be added to the same field, which is useful for
// complex validation scenarios with multiple error conditions.
//
// Parameters:
//   - problem: The Problem instance to add to the collection
//
// Example:
//
//	p := make(Problems)
//	p.Add(&Problem{Label: "email", Code: "INVALID", Message: "Invalid email"})
//	p.Add(&Problem{Label: "email", Code: "BLACKLISTED", Message: "Domain not allowed"})
//	// Results in: {"email": [problem1, problem2]}
func (p Problems) Add(problem *Problem) {
	if _, ok := p[problem.Label]; !ok {
		p[problem.Label] = make([]*Problem, 0)
	}
	p[problem.Label] = append(p[problem.Label], problem)
}

// AddError converts a validation error from go-slim.dev/v to a Problem
// and adds it to the Problems collection. This provides a bridge between
// the validation library error types and the response problem system.
//
// The method extracts the field name, error code, and message from the
// validation error and creates a new Problem with that information.
//
// Parameters:
//   - err: A validation error from the go-slim.dev/v package
//
// Example:
//
//	p := make(Problems)
//	validationErr := v.NewError("email", "INVALID_FORMAT", "Invalid email format")
//	p.AddError(validationErr)
func (p Problems) AddError(err *v.Error) {
	if err == nil {
		return
	}
	p.Add(&Problem{
		Label:    err.Field(),
		Code:     err.Code(),
		Message:  err.Error(),
		Problems: nil,
	})
}

// collectProblem recursively processes validation errors and converts them
// into a structured problem hierarchy. This function handles complex validation
// scenarios including nested errors and "some" validation conditions.
//
// The function examines the internal structure of validation errors and
// creates appropriate Problem instances with proper nesting for complex
// validation rules.
//
// Parameters:
//   - ps: The Problems collection to add the converted problems to
//   - err: The validation error to process and convert
//
// Supported Error Types:
//   - Simple v.Error: Converted directly to a Problem
//   - v.Errors collection: Each error is processed recursively
//   - "Some" validation: Creates a nested problem where at least one
//     sub-condition must be satisfied
//   - Nested errors: Processed recursively to maintain hierarchy
//
// Note: This function contains TODOs indicating future improvements
// to the go-slim.dev/v library integration.
func collectProblem(ps Problems, err *v.Error) {
	inter := err.Internal()
	if inter == nil {
		// TODO make problem
		ps.AddError(err)
		return
	}

	// TODO Refactor go-slim.dev/v, integrate problem design
	var verr *v.Errors
	if errors.As(inter, &verr) {
		if verr.IsSomeError() {
			tmp := make(Problems)
			for _, e := range verr.All() {
				collectProblem(tmp, e)
			}
			if len(tmp) > 0 {
				ps.Add(&Problem{
					Code:     "some",
					Label:    "$some",
					Message:  "At least one of these conditions must be satisfied",
					Problems: tmp,
				})
			}
			return
		}
		for _, e := range verr.All() {
			collectProblem(ps, e)
		}
		return
	}

	var xerr *v.Error
	if errors.As(inter, &xerr) {
		ps.AddError(xerr)
	} else {
		ps.AddError(err)
	}
}
