# Simple Distributed Mutex (SDM)

[![Go Reference](https://pkg.go.dev/badge/go-slim.dev/infra/sdm.svg)](https://pkg.go.dev/go-slim.dev/infra/sdm)
[![Go Report Card](https://goreportcard.com/badge/go-slim.dev/infra/sdm)](https://goreportcard.com/report/go-slim.dev/infra/sdm)
[![Test Status](https://github.com/go-slim/sdm/workflows/Test/badge.svg)](https://github.com/go-slim/sdm/actions?query=workflow%3ATest)

A simple and efficient distributed mutex implementation using Redis, designed for coordinating access to shared resources across multiple processes or servers.

## Features

- 🚀 Simple and easy-to-use API
- 🔒 Distributed locking with Redis as the coordination service
- ⏱️ Automatic lock expiration to prevent deadlocks
- 🔄 Support for both blocking and non-blocking lock acquisition
- 🛡️ Thread-safe implementation with proper error handling
- 🧩 Configurable timeouts and retry strategies
- 🔄 Automatic retry with exponential backoff
- 🔍 Lock status checking without acquiring the lock

## Installation

```bash
go get go-slim.dev/infra/sdm
```

## Quick Start

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"go-slim.dev/infra/sdm"
)

func main() {
	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Set Redis client for the package
	sdm.SetRedis(rdb)

	// Try to acquire a lock
	locked, err := sdm.TryLock(context.Background(), "process-1", 5*time.Second)
	if err != nil {
		log.Fatalf("Failed to acquire lock: %v", err)
	}
	if !locked {
		log.Fatal("Could not acquire lock within timeout")
	}

	// Make sure to release the lock when done
	defer sdm.Unlock(context.Background(), "process-1")

	// Critical section here
	log.Println("Lock acquired, doing work...")
	time.Sleep(2 * time.Second)
}
```

## Advanced Usage

### Creating a Named Mutex

```go
m, err := sdm.NewMutex("resource-123", "Resource Update Lock")
if err != nil {
    log.Fatal(err)
}

err = m.Lock(context.Background(), "process-1")
if err != nil {
    log.Fatal(err)
}
defer m.Unlock(context.Background(), "process-1")

// Work with the protected resource
```

### Using Custom Timeout

```go
// Try to acquire lock with 5 second timeout
acquired, err := sdm.TryLock(context.Background(), "process-1", 5*time.Second)
if err != nil {
    log.Fatal(err)
}
if !acquired {
    log.Println("Could not acquire lock within timeout")
    return
}
defer sdm.Unlock(context.Background(), "process-1")
```

### Checking Lock Status

```go
// Check if mutex is currently locked
m, err := sdm.NewMutex("resource-123")
if err != nil {
    log.Fatal(err)
}

locked, err := m.IsLocked(context.Background())
if err != nil {
    log.Fatal(err)
}
if locked {
    log.Println("Resource is currently locked")
} else {
    log.Println("Resource is currently not locked")
}

// Global lock status check
globalLocked, err := sdm.IsLocked(context.Background())
if err != nil {
    log.Fatal(err)
}
if globalLocked {
    log.Println("Global mutex is currently locked")
}
```

## Configuration

### Global Settings

```go
// Change the default Redis key prefix (default: "mutex")
sdm.RedisKeyPrefix = "myapp:mutex"

// Change the default mutex name (default: "default")
sdm.DefaultMutexName = "global"
```

## Error Handling

Common errors you might encounter:

- `sdm.ErrMutexNameEmpty`: When trying to create a mutex with an empty name
- `sdm.ErrInvalidMutexValue`: When the mutex value is invalid (empty or serialization failed)
- `sdm.ErrMutexNotAcquired`: When the lock cannot be acquired within the specified timeout

## Best Practices

1. Always use `defer` to ensure locks are released
2. Set appropriate timeouts to avoid deadlocks
3. Use descriptive lock names to identify resources
4. Handle errors appropriately
5. Keep the critical section as short as possible

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
