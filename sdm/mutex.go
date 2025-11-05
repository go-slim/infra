// Package sdm provides a simple distributed mutex (lock) implementation using Redis.
// This file contains the core Mutex type and its methods for distributed locking.
package sdm

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	minBackoff    = 1 * time.Millisecond // Minimum backoff time
	maxBackoff    = 1 * time.Second      // Maximum backoff time
	backoffFactor = 1.5                  // Backoff factor
)

// Mutex represents a distributed mutex that can be used to coordinate access to shared resources
// across multiple processes or servers. Each mutex is identified by a unique name.
//
// The generic type parameter T specifies the type of the value that will be stored in Redis
// to identify the lock owner. This is typically a string or a struct that can be serialized to JSON.
type Mutex[T any] struct {
	name  string // Unique identifier for the lock
	title string // Display title for the lock, used for logging and debugging
}

// New creates a new distributed mutex with the given name and optional title.
// The name must be a non-empty string that uniquely identifies the resource being locked.
// The title is an optional human-readable description of the mutex, used for logging and debugging.
//
// Example:
//
//	// Create a new mutex for a specific user resource
//	m, err := sdm.New[any]("user:123:profile", "user profile update lock")
//	if err != nil {
//	    return err
//	}
//	defer m.Unlock(context.Background(), "process-1")
//
// Returns an error if the name is empty.
func New[T any](name string, title ...string) (Mutex[T], error) {
	if name = strings.TrimSpace(name); name == "" {
		return Mutex[T]{}, ErrMutexNameEmpty
	}

	ttl := strings.TrimSpace(cmp.Or(title...))
	ttl = cmp.Or(ttl, name)

	return Mutex[T]{
		name:  name,
		title: ttl,
	}, nil
}

// Name returns the unique identifier for this mutex.
// This is the name that was passed to New when creating the mutex.
func (m Mutex[T]) Name() string {
	return m.name
}

// Title returns the human-readable title for this mutex.
// If no title was provided when creating the mutex, this returns the same as Name().
func (m Mutex[T]) Title() string {
	return m.title
}

// TryLock attempts to acquire the mutex lock with an optional timeout.
// If the lock is already held by another process, it will either return immediately
// (if no timeout is specified) or wait for the specified duration before giving up.
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
//	locked, err := m.TryLock(ctx, "process-1", 5*time.Second)
//	if err != nil {
//	    return fmt.Errorf("failed to acquire lock: %w", err)
//	}
//	if !locked {
//	    return errors.New("could not acquire lock within timeout")
//	}
//	defer m.Unlock(ctx, "process-1")
func (m Mutex[T]) TryLock(ctx context.Context, value T, timeout ...time.Duration) (bool, error) {
	if len(timeout) == 0 || timeout[0] <= 0 {
		return m.tryLock(ctx, value)
	}
	return m.tryLockWithTimeout(ctx, value, timeout[0])
}

// Lock acquires the mutex lock, blocking until it is available or the context is cancelled.
// This is a convenience method that calls TryLock with a very long timeout.
// The context parameter must not be nil and should be used for cancellation and timeouts.
//
// Example:
//
//	err := m.Lock(ctx, "process-1")
//	if err != nil {
//	    return fmt.Errorf("failed to acquire lock: %w", err)
//	}
//	defer m.Unlock(ctx, "process-1")
//	// ... critical section ...
func (m Mutex[T]) Lock(ctx context.Context, value T) error {
	acquired, err := m.tryLockWithTimeout(ctx, value, -1)
	if err != nil {
		return err
	}
	if !acquired {
		// This should theoretically not be reached, as negative timeout causes infinite retries
		return errors.New("sdm: failed to acquire lock: unknown error")
	}
	return nil
}

func (m Mutex[T]) tryLock(ctx context.Context, value T) (bool, error) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	valstr, err := serializeValue(value)
	if err != nil {
		return false, fmt.Errorf("sdm: failed to serialize value: %w", err)
	}

	rdb, err := db()
	if err != nil {
		return false, err
	}

	key, err := getRedisKeyWithPrefix(RedisKeyPrefix, m.name)
	if err != nil {
		return false, err
	}
	result, err := tryLockScript.Run(ctx, rdb, []string{key}, valstr).Result()
	if err != nil {
		return false, fmt.Errorf("sdm: try lock failed: %w", err)
	}

	return result.(int64) == 1, nil
}

func (m Mutex[T]) tryLockWithTimeout(ctx context.Context, value T, timeout time.Duration) (bool, error) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	// Create context with timeout (if timeout > 0)
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Pre-fetch Redis key and serialize value
	key, err := getRedisKeyWithPrefix(RedisKeyPrefix, m.name)
	if err != nil {
		return false, err
	}

	valstr, err := serializeValue(value)
	if err != nil {
		return false, err
	}

	rdb, err := db()
	if err != nil {
		return false, err
	}

	// Get current time
	startTime := time.Now()
	attempt := 0

	for {
		attempt++

		// Try to acquire lock
		result, err := tryLockScript.Run(ctx, rdb, []string{key}, valstr).Result()
		if err != nil {
			return false, fmt.Errorf("sdm: try lock failed: %w", err)
		}

		// If lock acquired successfully, return
		if result.(int64) == 1 {
			return true, nil
		}

		// Calculate backoff time
		backoff := min(
			time.Duration(math.Pow(float64(backoffFactor), float64(attempt-1))*float64(minBackoff)),
			maxBackoff,
		)

		// Check if timeout is reached
		if time.Since(startTime) >= timeout {
			return false, nil
		}

		// Wait for a while before retrying
		select {
		case <-time.After(backoff):
			continue
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}
}

// Unlock releases the mutex lock. It is safe to call this even if the lock is not held.
// The value parameter must match the value that was used to acquire the lock.
//
// The context parameter must not be nil and should be used for cancellation and timeouts.
//
// It is recommended to use defer to ensure the lock is always released:
//
//	err := m.Lock(ctx, "process-1")
//	if err != nil {
//	    return err
//	}
//	defer m.Unlock(ctx, "process-1")
//
// Note: If the context is cancelled while trying to release the lock, the error from
// the context will be returned, but the lock may still be released in the background.
func (m Mutex[T]) Unlock(ctx context.Context, value T) error {
	valstr, err := serializeValue(value)
	if err != nil {
		return fmt.Errorf("sdm: failed to serialize value: %w", err)
	}

	rdb, err := db()
	if err != nil {
		return err
	}

	key, err := getRedisKeyWithPrefix(RedisKeyPrefix, m.name)
	if err != nil {
		return err
	}
	result, err := unlockScript.Run(ctx, rdb, []string{key}, valstr).Result()
	if err != nil {
		return fmt.Errorf("sdm: unlock failed: %w", err)
	}

	if result.(int64) == 0 {
		return ErrMutexNotAcquired
	}
	return nil
}

// IsLocked checks whether the mutex is currently locked by any process.
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
//	locked, err := m.IsLocked(ctx)
//	if err != nil {
//	    return fmt.Errorf("failed to check lock status: %w", err)
//	}
//	if locked {
//	    fmt.Println("Mutex is currently locked")
//	}
func (m Mutex[T]) IsLocked(ctx context.Context) (bool, error) {
	rdb, err := db()
	if err != nil {
		return false, err
	}

	key, err := getRedisKeyWithPrefix(RedisKeyPrefix, m.name)
	if err != nil {
		return false, err
	}

	// Check if the key exists and has any members using SCARD
	count, err := rdb.(redis.Cmdable).SCard(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("sdm: failed to check lock status: %w", err)
	}

	return count > 0, nil
}
