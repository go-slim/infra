// Package sdm provides a simple distributed mutex (lock) implementation using Redis.
// It allows multiple processes to coordinate access to shared resources in a distributed environment.
//
// Features:
//   - Distributed locking with Redis as the coordination service
//   - Automatic lock expiration to prevent deadlocks
//   - Support for both blocking and non-blocking lock acquisition
//   - Thread-safe implementation with proper error handling
//   - Configurable timeouts and retry strategies
//
// Example usage:
//
//	// Initialize Redis client
//	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
//
//	// Set the Redis client for the package
//	sdm.SetRedis(rdb)
//
//	// Try to acquire a lock
//	locked, err := sdm.TryLock(context.Background(), "resource-name", 5*time.Second)
//
//	// Make sure to release the lock when done
//	defer sdm.Unlock(context.Background(), "resource-name")
//
// For more advanced usage, create a Mutex instance directly:
//
//	m, _ := sdm.NewMutex("resource-name")
//	err := m.Lock(context.Background(), "owner-id")
//	// ... use the resource ...
//	m.Unlock(context.Background(), "owner-id")
package sdm

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// Common errors returned by the package
var (
	// ErrMutexNameEmpty is returned when attempting to create a mutex with an empty name
	ErrMutexNameEmpty = errors.New("sdm: mutex name cannot be empty")
	// ErrInvalidMutexValue is returned when the mutex value is invalid (empty or serialization failed)
	ErrInvalidMutexValue = errors.New("sdm: invalid mutex value")
	// ErrMutexNotAcquired is returned when the lock cannot be acquired within the specified timeout
	ErrMutexNotAcquired = errors.New("sdm: failed to acquire mutex")

	// RedisKeyPrefix storage prefix, should only be specified during initialization
	RedisKeyPrefix = "mutex"
	// DefaultMutexName global mutex name, should only be specified during initialization
	DefaultMutexName = "default"

	// Global default mutex object
	mtx *Mutex[any]

	rdb atomic.Value // redis.Scripter
	sfg singleflight.Group
)

// init initializes the default mutex instance with default values.
// This function is automatically called when the package is imported.
// It creates a default mutex instance that can be used with the package-level
// Lock/Unlock functions without explicitly creating a Mutex instance.
//
// The default mutex uses DefaultMutexName ("default") as both its name and title.
func init() {
	// Initialize default mutex
	mtx = &Mutex[any]{
		name:  DefaultMutexName,
		title: DefaultMutexName,
	}
}

// SetRedis sets the Redis client to be used by the package for distributed locking.
// This function must be called before any lock operations are performed.
//
// The provided client must implement the redis.Scripter interface, which is satisfied
// by both *redis.Client and *redis.ClusterClient from github.com/redis/go-redis/v9.
//
// Example:
//
//	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
//	sdm.SetRedis(rdb)
//
// Note: This function is safe to call concurrently.
func SetRedis(v redis.Scripter) {
	rdb.Store(v)
}

// TryLock attempts to acquire the default mutex lock with an optional timeout.
// This is a convenience function that uses the default mutex instance.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts (must not be nil)
//   - value: A value that identifies the lock owner (must be JSON-serializable)
//   - timeout: Optional timeout duration. If not provided or zero, the call will not block.
//
// Returns:
//   - bool: true if the lock was acquired, false if not
//   - error: non-nil if an error occurred while trying to acquire the lock
//
// Example:
//
//	locked, err := sdm.TryLock(ctx, "process-1", 5*time.Second)
//	if err != nil {
//	    return fmt.Errorf("failed to acquire lock: %w", err)
//	}
//	if !locked {
//	    return errors.New("could not acquire lock within timeout")
//	}
//	defer sdm.Unlock(ctx, "process-1")
//
// Note: The default mutex uses DefaultMutexName as its name.
func TryLock(ctx context.Context, value any, timeout ...time.Duration) (bool, error) {
	return mtx.TryLock(ctx, value, timeout...)
}

// Lock acquires the default mutex lock, blocking until it is available or the context is cancelled.
// This is a convenience function that uses the default mutex instance.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts (must not be nil)
//   - value: A value that identifies the lock owner (must be JSON-serializable)
//
// Returns:
//   - error: non-nil if the lock could not be acquired or the context was cancelled
//
// Example:
//
//	err := sdm.Lock(ctx, "process-1")
//	if err != nil {
//	    return fmt.Errorf("failed to acquire lock: %w", err)
//	}
//	defer sdm.Unlock(ctx, "process-1")
//	// ... critical section ...
//
// Note: The default mutex uses DefaultMutexName as its name.
func Lock(ctx context.Context, value any) error {
	return mtx.Lock(ctx, value)
}

// Unlock releases the default mutex lock. It is safe to call this even if the lock is not held.
// This is a convenience function that uses the default mutex instance.
//
// Parameters:
//   - ctx: Context for cancellation (must not be nil)
//   - value: The same value that was used to acquire the lock
//
// Returns:
//   - error: non-nil if an error occurred while releasing the lock
//
// Example:
//
//	err := sdm.Lock(ctx, "process-1")
//	if err != nil {
//	    return err
//	}
//	defer sdm.Unlock(ctx, "process-1")
//	// ... critical section ...
//
// Note: The default mutex uses DefaultMutexName as its name.
func Unlock(ctx context.Context, value any) error {
	return mtx.Unlock(ctx, value)
}

// IsLocked checks whether the default mutex is currently locked by any process.
// This is a convenience function that uses the default mutex instance.
// This method does not require knowledge of the lock value and can be used
// to check the lock status without acquiring or releasing it.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts (must not be nil)
//
// Returns:
//   - bool: true if the lock is currently held, false if not
//   - error: non-nil if an error occurred while checking the lock status
//
// Example:
//
//	locked, err := sdm.IsLocked(ctx)
//	if err != nil {
//	    return fmt.Errorf("failed to check lock status: %w", err)
//	}
//	if locked {
//	    fmt.Println("Default mutex is currently locked")
//	}
//
// Note: The default mutex uses DefaultMutexName as its name.
func IsLocked(ctx context.Context) (bool, error) {
	return mtx.IsLocked(ctx)
}
