// Package sdm provides utility functions for the distributed mutex implementation.
// This file contains helper functions for Redis key generation, value serialization,
// and other common operations used throughout the package.
package sdm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

// Error definitions
var (
	// ErrRedisNotInitialized is returned when the Redis client is not initialized
	ErrRedisNotInitialized = errors.New("sdm: redis client not initialized")
)

var tryLockScript = redis.NewScript(`
	-- Attempt to acquire distributed lock
	-- Uses Set data structure where key is the lock name and member is the lock value
	-- KEYS[1]: Lock key name
	-- ARGV[1]: Lock value
	-- Returns: 1 for successful acquisition, 0 for lock already occupied

	local key = KEYS[1]
	local value = ARGV[1]

	-- Use SADD to try adding the value to the set
	-- If value already exists, returns 0; if addition succeeds, returns 1
	local added = redis.call("SADD", key, value)

	-- If value already exists in set, lock is occupied
	if added == 0 then
		return 0
	end

	-- Successfully acquired lock
	return 1
`)

var unlockScript = redis.NewScript(`
	-- Release distributed lock
	-- KEYS[1]: Lock key name
	-- ARGV[1]: Expected lock value
	-- Returns: 1 for successful release, 0 for failed release (lock doesn't exist or value mismatch)

	local key = KEYS[1]
	local expected_value = ARGV[1]

	-- Check if value exists in set
	local is_member = redis.call("SISMEMBER", key, expected_value)

	-- If value not in set, return failure
	if is_member == 0 then
		return 0
	end

	-- Remove value from set
	redis.call("SREM", key, expected_value)

	-- Delete key if set becomes empty
	if redis.call("SCARD", key) == 0 then
		redis.call("DEL", key)
	end

	return 1
`)

func db() (redis.Scripter, error) {
	v := rdb.Load()
	if v == nil || v == (*redis.Client)(nil) {
		return nil, ErrRedisNotInitialized
	}
	return v.(redis.Scripter), nil
}

// getRedisKey generates a Redis key for the given name using the global RedisKeyPrefix.
// This is a convenience wrapper around getRedisKeyWithPrefix that uses the global prefix.
//
// DESIGN NOTE: This function uses implicit dependency on the global RedisKeyPrefix variable.
// For production code, prefer getRedisKeyWithPrefix to avoid global variable races and
// make dependencies explicit. This function is primarily intended for testing and
// backward compatibility scenarios.
//
// The generated key follows the pattern "prefix:name" where:
//   - prefix is the global RedisKeyPrefix (default: "mutex")
//   - name is the provided resource name
//
// If name is empty, it will use DefaultMutexName ("default").
// If both prefix and name are empty, it returns an error.
//
// Example:
//   - With RedisKeyPrefix = "mutex" and name = "resource": returns "mutex:resource"
//   - With RedisKeyPrefix = "" and name = "resource": returns "resource"
//   - With RedisKeyPrefix = "app" and name = "": returns "app:default"
func getRedisKey(name string) (string, error) {
	return getRedisKeyWithPrefix(RedisKeyPrefix, name)
}

// getRedisKeyWithPrefix generates a Redis key using the specified prefix and name.
// This function follows the principle of "explicit dependencies over implicit dependencies"
// by requiring the prefix to be passed as a parameter rather than relying on global state.
//
// DESIGN ADVANTAGES:
//   - PREDICTABLE: No hidden dependencies on global variables that might change at runtime
//   - TESTABLE: Can inject different prefixes for testing scenarios
//   - THREAD-SAFE: Avoids race conditions with global variable access
//   - EXPLICIT: All dependencies are clearly visible in the function signature
//
// This is the preferred function for production code over getRedisKey.
//
// Parameters:
//   - prefix: The key prefix to use (can be empty)
//   - name: The resource name (will be trimmed of whitespace)
//
// Returns:
//   - The generated Redis key as a string
//   - An error if both prefix and name are empty after trimming
//
// The function handles various edge cases:
//   - Trims whitespace from the name
//   - Uses DefaultMutexName ("default") if name is empty after trimming
//   - Omits the separator if prefix is empty
//   - Returns an error if both prefix and name are empty
//
// Example:
//   - getRedisKeyWithPrefix("mutex", "resource") → "mutex:resource"
//   - getRedisKeyWithPrefix("", "resource") → "resource"
//   - getRedisKeyWithPrefix("app", "") → "app:default"
//   - getRedisKeyWithPrefix("", "") → error
func getRedisKeyWithPrefix(prefix, name string) (string, error) {
	if name = strings.TrimSpace(name); name == "" {
		name = DefaultMutexName
	}

	if name == "" {
		return "", ErrMutexNameEmpty
	}

	// If prefix is empty, return just the name without a separator
	if prefix == "" {
		return name, nil
	}
	return fmt.Sprintf("%s:%s", prefix, name), nil
}

// serializeValue converts a value to a string representation for storage in Redis.
// It supports basic types and any value that can be serialized to JSON.
//
// Supported types:
//   - string, []byte, int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64, bool
//   - Any type that implements encoding.TextMarshaler
//   - Any type that can be serialized to JSON (using json.Marshal)
//
// Returns:
//   - The string representation of the value
//   - An error if the value cannot be serialized
//
// Example:
//   - serializeValue(42) → "42"
//   - serializeValue("hello") → "hello"
//   - serializeValue(struct{Name string}{"test"}) → "{"Name":"test"}"
//
// Performs non-empty validation only for non-string types
func serializeValue[T any](v T) (string, error) {
	switch val := any(v).(type) {
	case string:
		// String types are returned directly, allowing empty strings and whitespace
		return val, nil
	case *string:
		if val == nil {
			return "", fmt.Errorf("sdm: nil string pointer")
		}
		return *val, nil
	default:
		// Use JSON serialization for other types
		data, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("sdm: failed to marshal value: %w", err)
		}
		result := string(data)

		// For non-string types, check if serialized result is empty
		if result == "" {
			return "", fmt.Errorf("sdm: invalid mutex value")
		}
		return result, nil
	}
}
