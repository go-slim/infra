# Backoff Package

[中文文档](README.zh-CN.md)

A Go package that provides exponential backoff with jitter for retry operations. This package helps you implement robust retry logic for network operations, database connections, API calls, and other potentially failing operations that may succeed on retry.

## Features

- **Exponential Backoff**: Automatically increases wait times between retries
- **Jitter Support**: Adds randomness to prevent thundering herd problems
- **Context Integration**: Respects context cancellation and timeouts
- **Thread-Safe**: Safe for concurrent use from multiple goroutines
- **Configurable**: Flexible configuration for different use cases
- **Convenience Functions**: Simple APIs for common scenarios

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Performance](#performance)
- [Contributing](#contributing)

## Installation

```bash
go get go-slim.dev/infra/pkg/backoff
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "go-slim.dev/infra/pkg/backoff"
)

func main() {
    // Simple retry with default configuration
    err := backoff.Retry(context.Background(), backoff.DefaultConfig, func(ctx context.Context) error {
        return callExternalAPI()
    })

    if err != nil {
        log.Fatal("Operation failed after retries:", err)
    }

    fmt.Println("Operation succeeded!")
}

func callExternalAPI() error {
    // Your potentially failing operation here
    return nil // or error
}
```

### Convenience Function

```go
// Use RetryDefault for quick implementation with sensible defaults
err := backoff.RetryDefault(ctx, func(ctx context.Context) error {
    return database.Connect()
})
```

### Custom Configuration

```go
config := backoff.Config{
    InitialInterval:     100 * time.Millisecond,
    MaxInterval:         5 * time.Second,
    Multiplier:          2.0,
    MaxRetries:          5,
    RandomizationFactor: 0.1,
}

err := backoff.Retry(ctx, config, func(ctx context.Context) error {
    return httpClient.Get(url)
})
```

## Configuration

The `Config` struct provides full control over backoff behavior:

| Field                 | Type            | Description                          | Default |
| --------------------- | --------------- | ------------------------------------ | ------- |
| `InitialInterval`     | `time.Duration` | Initial wait time before first retry | 500ms   |
| `MaxInterval`         | `time.Duration` | Maximum wait time between retries    | 30s     |
| `Multiplier`          | `float64`       | Backoff multiplier (must be > 1.0)   | 1.5     |
| `MaxRetries`          | `int`           | Maximum number of retry attempts     | 10      |
| `RandomizationFactor` | `float64`       | Jitter factor (0.0-1.0)              | 0.1     |

### Configuration Guidelines

#### InitialInterval

- **Fast local operations**: 10-100ms
- **Network requests**: 100ms-1s
- **Database connections**: 1-5s

#### MaxInterval

- **Interactive applications**: 1-10s
- **Batch processing**: 30s-5min
- **Background services**: 1-10min

#### Multiplier

- **1.5**: Gentle backoff (recommended for most cases)
- **2.0**: Standard exponential backoff
- **3.0**: Aggressive backoff (reaches max interval quickly)

#### RandomizationFactor

- **0.0**: No jitter (deterministic)
- **0.1**: Light jitter (recommended for most cases)
- **0.5**: Heavy jitter (good for large distributed systems)

## API Reference

### Functions

#### `Retry(ctx, config, fn) error`

Executes a function with exponential backoff retry logic.

**Parameters:**

- `ctx context.Context`: Context for cancellation and timeout
- `config Config`: Backoff configuration
- `fn func(context.Context) error`: Function to retry

**Returns:**

- `error`: The last error if all retries fail, nil on success

#### `RetryDefault(ctx, fn) error`

Convenience function using `DefaultConfig`.

### Type: Backoff

#### `New(config Config) *Backoff`

Creates a new Backoff instance with the given configuration.

#### `(*Backoff) Next() time.Duration`

Calculates and returns the next backoff delay.

#### `(*Backoff) Reset()`

Resets the attempt counter to zero.

#### `(*Backoff) Attempt() int`

Returns the current attempt count.

#### `(*Backoff) Do(ctx, fn) error`

Executes a function with retry logic using this Backoff instance.

## Examples

### HTTP Client with Retry

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "go-slim.dev/infra/pkg/backoff"
)

func main() {
    client := &http.Client{Timeout: 5 * time.Second}

    config := backoff.Config{
        InitialInterval:     100 * time.Millisecond,
        MaxInterval:         2 * time.Second,
        Multiplier:          2.0,
        MaxRetries:          3,
        RandomizationFactor: 0.1,
    }

    err := backoff.Retry(context.Background(), config, func(ctx context.Context) error {
        req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
        if err != nil {
            return err
        }

        resp, err := client.Do(req)
        if err != nil {
            return err
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 500 {
            return fmt.Errorf("server error: %d", resp.StatusCode)
        }

        return nil
    })

    if err != nil {
        fmt.Printf("Request failed: %v\n", err)
        return
    }

    fmt.Println("Request succeeded!")
}
```

### Database Connection with Context Timeout

```go
func connectWithRetry(ctx context.Context) error {
    config := backoff.Config{
        InitialInterval:     1 * time.Second,
        MaxInterval:         30 * time.Second,
        Multiplier:          1.5,
        MaxRetries:          5,
        RandomizationFactor: 0.2,
    }

    return backoff.Retry(ctx, config, func(ctx context.Context) error {
        db, err := sql.Open("postgres", connectionString)
        if err != nil {
            return err
        }

        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()

        return db.PingContext(ctx)
    })
}
```

### Reusable Backoff Instance

```go
type APIClient struct {
    backoff *backoff.Backoff
    client  *http.Client
}

func NewAPIClient() *APIClient {
    config := backoff.Config{
        InitialInterval:     200 * time.Millisecond,
        MaxInterval:         5 * time.Second,
        Multiplier:          1.8,
        MaxRetries:          4,
        RandomizationFactor: 0.15,
    }

    return &APIClient{
        backoff: backoff.New(config),
        client:  &http.Client{Timeout: 10 * time.Second},
    }
}

func (c *APIClient) Call(ctx context.Context, endpoint string) error {
    return c.backoff.Do(ctx, func(ctx context.Context) error {
        req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
        if err != nil {
            return err
        }

        resp, err := c.client.Do(req)
        if err != nil {
            return err
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 500 {
            return fmt.Errorf("server error: %d", resp.StatusCode)
        }

        return nil
    })
}
```

### Custom Retry Condition

```go
func retryWithCustomCondition(ctx context.Context) error {
    var lastError error

    err := backoff.Retry(ctx, backoff.DefaultConfig, func(ctx context.Context) error {
        result, err := someOperation()
        if err != nil {
            lastError = err

            // Only retry on specific errors
            if isRetryableError(err) {
                return err
            }
            return nil // Don't retry on non-retryable errors
        }

        // Check if result meets criteria
        if !isAcceptableResult(result) {
            return fmt.Errorf("unacceptable result")
        }

        return nil
    })

    if err != nil {
        return fmt.Errorf("operation failed: %w (last error: %v)", err, lastError)
    }

    return nil
}

func isRetryableError(err error) bool {
    // Define what errors are retryable
    return true // Your logic here
}

func isAcceptableResult(result interface{}) bool {
    // Define what results are acceptable
    return true // Your logic here
}
```

## Best Practices

### 1. Choose Appropriate Timeouts

```go
// Good: Use context timeouts to prevent hanging
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := backoff.Retry(ctx, config, operation)

// Bad: No timeout protection
err := backoff.Retry(context.Background(), config, operation)
```

### 2. Handle Specific Error Types

```go
// Good: Only retry on transient errors
err := backoff.Retry(ctx, config, func(ctx context.Context) error {
    err := someOperation()
    if isTransientError(err) {
        return err // Retry
    }
    return err // Don't retry permanent errors
})

// Bad: Retry all errors
err := backoff.Retry(ctx, config, someOperation)
```

### 3. Use Jitter in Distributed Systems

```go
// Good: Add jitter to prevent thundering herd
config := backoff.Config{
    RandomizationFactor: 0.1, // 10% jitter
    // ... other fields
}

// Bad: No jitter in distributed systems
config := backoff.Config{
    RandomizationFactor: 0.0, // All clients retry simultaneously
    // ... other fields
}
```

### 4. Monitor Retry Behavior

```go
func instrumentedRetry(ctx context.Context, operation string) error {
    attempts := 0
    start := time.Now()

    err := backoff.Retry(ctx, backoff.DefaultConfig, func(ctx context.Context) error {
        attempts++
        metrics.IncrementRetryAttempts(operation)
        return someOperation()
    })

    duration := time.Since(start)

    if err != nil {
        metrics.RecordRetryFailure(operation, attempts, duration)
    } else {
        metrics.RecordRetrySuccess(operation, attempts, duration)
    }

    return err
}
```

### 5. Configuring for Different Environments

```go
func getConfigForEnvironment(env string) backoff.Config {
    switch env {
    case "development":
        return backoff.Config{
            InitialInterval: 50 * time.Millisecond,
            MaxInterval:     1 * time.Second,
            MaxRetries:      3,
            Multiplier:      2.0,
        }
    case "staging":
        return backoff.Config{
            InitialInterval: 100 * time.Millisecond,
            MaxInterval:     5 * time.Second,
            MaxRetries:      5,
            Multiplier:      1.5,
        }
    case "production":
        return backoff.DefaultConfig
    default:
        return backoff.DefaultConfig
    }
}
```

## Performance Considerations

### Memory Usage

- Each `Backoff` instance maintains minimal state (just the attempt counter)
- The package uses atomic operations for thread safety
- Memory overhead is negligible for most applications

### CPU Usage

- Backoff calculation is O(1) and very fast
- Most CPU time is spent waiting between retries
- Jitter calculation adds minimal overhead

### Goroutine Usage

- The package doesn't create additional goroutines
- All operations are synchronous and non-blocking
- Context cancellation is properly respected

## Testing

Run tests with:

```bash
# Run all tests
go test -v ./...

# Run with race detector
go test -race -v ./...

# Run benchmarks
go test -bench=. -benchmem

# Run with coverage
go test -cover ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

### Development Guidelines

- Keep the API simple and intuitive
- Maintain backward compatibility
- Add comprehensive tests for new features
- Update documentation for API changes
- Follow Go conventions and best practices

## License

This package is part of the goapp project.

## Related Packages

- [golang.org/x/net/context](https://pkg.go.dev/golang.org/x/net/context) - Context support
- [github.com/cenkalti/backoff](https://github.com/cenkalti/backoff) - Alternative backoff implementation
- [github.com/sethvargo/go-retry](https://github.com/sethvargo/go-retry) - Modern retry library with more features
