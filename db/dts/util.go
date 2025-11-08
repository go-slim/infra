// Package dts provides database-specific data types for Go applications.
// This file contains utility functions that support the data type implementations.
package dts

import (
	"database/sql"
	"database/sql/driver"
	"reflect"
)

var valuerReflectType = reflect.TypeFor[driver.Valuer]()

// callValuerValue safely calls the Value() method on a driver.Valuer.
// It handles the special case where a nil pointer implements driver.Valuer
// and would panic if called directly.
//
// This is based on Go's database/sql package implementation (Issue 8415).
// It allows developers to implement driver.Valuer on value types and
// still use nil pointers to those types to mean nil/NULL, just like string/*string.
//
// Parameters:
//
//	vr: The driver.Valuer to call Value() on
//
// Returns:
//
//	The value from vr.Value(), or nil if vr is a nil pointer implementing driver.Valuer
//	An error if vr.Value() returns an error
func callValuerValue(vr driver.Valuer) (v driver.Value, err error) {
	if rv := reflect.ValueOf(vr); rv.Kind() == reflect.Pointer &&
		rv.IsNil() &&
		rv.Type().Elem().Implements(valuerReflectType) {
		return nil, nil
	}
	return vr.Value()
}

// toSqlValue converts a value to a driver.Value with proper type handling.
// This function handles the conversion of values to database-compatible types,
// with special handling for driver.Valuer implementations.
//
// Parameters:
//
//	v: The value to convert
//	ok: Boolean flag indicating if the value should be processed (false = nil)
//
// Returns:
//
//	A driver.Value suitable for database operations
//	An error if conversion fails
func toSqlValue(v any, ok bool) (driver.Value, error) {
	if !ok {
		return nil, nil
	}

	// Handle driver.Valuer implementations
	if valuer, ok := v.(driver.Valuer); ok {
		val, err := callValuerValue(valuer)
		if err != nil {
			return val, err
		}
		v = val
	}

	// Use default converter for other types
	return driver.DefaultParameterConverter.ConvertValue(v)
}

// scan is a generic function that scans database values into a Go type.
// It uses sql.Null[T] to handle nullable values and returns both the scanned value
// and a boolean indicating whether the value was valid (not NULL).
//
// Type Parameters:
//
//	T: The target type to scan into
//
// Parameters:
//
//	value: The database value to scan
//
// Returns:
//
//	v: The scanned value of type T
//	ok: Boolean indicating if the value was valid (not NULL)
//	err: Error if scanning failed
func scan[T any](value any) (v T, ok bool, err error) {
	var nt sql.Null[T]
	err = nt.Scan(value)
	ok = nt.Valid
	v = nt.V
	return
}
