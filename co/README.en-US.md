# co - Concurrency Utilities

[中文文档](README.zh-CN.md)

The `co` package provides a collection of lightweight concurrency utilities for Go applications. It includes a simple channel-based mutex, a goroutine pool for efficient task scheduling, and spin-wait utilities with exponential backoff.

## Table of Contents

- [Installation](#installation)
- [Components](#components)
  - [SimpleMutex](#simplemutex)
  - [GoRoutinePool](#goroutinepool)
  - [Spin-Wait Utilities](#spin-wait-utilities)
- [Examples](#examples)
- [Performance](#performance)
- [Testing](#testing)
- [License](#license)

## Installation

```bash
go get go-slim.dev/infra/pkg/co
```

## Components

### SimpleMutex

`SimpleMutex` is a channel-based mutex implementation that provides a lightweight alternative to `sync.Mutex` for simple locking scenarios.

#### Features

- Channel-based mutual exclusion using a buffered channel with capacity 1
- Non-blocking `TryLock()` operation
- Simple and easy to understand implementation
- Suitable for basic synchronization needs

#### API

```go
// Create a new SimpleMutex
sm := co.NewSimpleMutex()

// Lock (blocking)
sm.Lock()

// Try to lock (non-blocking)
if sm.TryLock() {
    // Lock acquired
    defer sm.Unlock()
    // ... critical section ...
}

// Unlock
sm.Unlock()
```

#### Usage Example

```go
package main

import (
    "fmt"
    "go-slim.dev/infra/pkg/co"
)

func main() {
    sm := co.NewSimpleMutex()
    counter := 0

    // Spawn multiple goroutines
    for i := 0; i < 100; i++ {
        go func() {
            sm.Lock()
            counter++
            sm.Unlock()
        }()
    }

    // Using TryLock for non-blocking acquisition
    if sm.TryLock() {
        fmt.Println("Lock acquired without blocking")
        sm.Unlock()
    } else {
        fmt.Println("Lock is currently held by another goroutine")
    }
}
```

#### When to Use

- **Use SimpleMutex when:**
  - You need a simple, lightweight mutex
  - You want non-blocking lock attempts with `TryLock()`
  - Your use case doesn't require advanced features like RWMutex

- **Use sync.Mutex when:**
  - You need battle-tested, production-grade synchronization
  - You require advanced features (RWMutex, etc.)
  - Performance is critical in high-contention scenarios

### GoRoutinePool

`GoRoutinePool` provides efficient goroutine pooling with automatic worker management and concurrency control.

#### Features

- Efficient goroutine reuse reduces allocation overhead
- Configurable maximum concurrent workers
- Automatic worker spawning and management
- Graceful shutdown support
- Non-blocking task scheduling

#### API

```go
// Create a pool with maximum 10 concurrent workers
pool := co.NewGoRoutinePool(10)

// Schedule tasks
pool.Schedule(func() {
    // Your task here
})

// Gracefully stop the pool
pool.Stop()
```

#### Usage Example

```go
package main

import (
    "fmt"
    "sync"
    "go-slim.dev/infra/pkg/co"
)

func main() {
    // Create a pool with 5 workers
    pool := co.NewGoRoutinePool(5)
    defer pool.Stop()

    var wg sync.WaitGroup

    // Schedule 20 tasks
    for i := 0; i < 20; i++ {
        wg.Add(1)
        taskNum := i
        pool.Schedule(func() {
            defer wg.Done()
            fmt.Printf("Executing task %d\n", taskNum)
            // Simulate work
        })
    }

    wg.Wait()
    fmt.Println("All tasks completed")
}
```

#### How It Works

1. **Worker Spawning**: When a task is scheduled:
   - If an idle worker is available, the task is sent to it immediately
   - If no idle worker exists and the pool hasn't reached capacity, a new worker is spawned
   - Workers automatically handle multiple tasks sequentially

2. **Worker Lifecycle**: Each worker runs in a loop, processing tasks until:
   - A stop signal is received
   - The worker exits and releases its semaphore slot

3. **Concurrency Control**: A semaphore channel limits the maximum number of concurrent workers

#### When to Use

- **Use GoRoutinePool when:**
  - You need to limit concurrent goroutine execution
  - You have many short-lived tasks
  - You want to reduce goroutine allocation overhead
  - You need predictable resource usage

- **Use raw goroutines when:**
  - You have a small, fixed number of tasks
  - Tasks are long-lived
  - You don't need concurrency limits

### Spin-Wait Utilities

Spin-wait utilities provide CPU-efficient busy-waiting mechanisms with exponential backoff.

#### Features

- Exponential backoff strategy (1, 2, 4, 8, 16, 32)
- CPU-friendly yielding with `runtime.Gosched()`
- Generic `WaitFunc` for custom conditions
- Specialized `WaitFor` for atomic counters

#### API

```go
// Wait until a condition becomes false
co.WaitFunc(func() bool {
    return !conditionMet()
})

// Wait for an atomic counter to reach zero
var counter atomic.Int32
counter.Store(5)
co.WaitFor(&counter)
```

#### Usage Example

```go
package main

import (
    "fmt"
    "sync/atomic"
    "time"
    "go-slim.dev/infra/pkg/co"
)

func main() {
    // Example 1: Wait for a condition
    ready := false
    go func() {
        time.Sleep(100 * time.Millisecond)
        ready = true
    }()

    co.WaitFunc(func() bool {
        return !ready
    })
    fmt.Println("Condition met!")

    // Example 2: Wait for a counter
    var counter atomic.Int32
    counter.Store(10)

    // Spawn goroutines that decrement the counter
    for i := 0; i < 10; i++ {
        go func() {
            time.Sleep(10 * time.Millisecond)
            counter.Add(-1)
        }()
    }

    co.WaitFor(&counter)
    fmt.Println("All goroutines completed!")
}
```

#### How Backoff Works

The backoff mechanism reduces CPU usage while waiting:

```
Iteration 1: Yield 1 time
Iteration 2: Yield 2 times
Iteration 3: Yield 4 times
Iteration 4: Yield 8 times
Iteration 5: Yield 16 times
Iteration 6+: Yield 32 times (capped)
```

Each yield calls `runtime.Gosched()`, which allows other goroutines to run on the same OS thread.

#### When to Use

- **Use spin-wait utilities when:**
  - You're waiting for very short durations (microseconds to milliseconds)
  - You need fine-grained control over busy-waiting
  - You're coordinating with atomic operations
  - Lock-free algorithms require waiting

- **Use channels/sync primitives when:**
  - Wait times are unpredictable or potentially long
  - You need to wait for multiple conditions
  - Standard synchronization patterns fit your use case

## Examples

### Complete Example: HTTP Request Processor

```go
package main

import (
    "fmt"
    "sync/atomic"
    "time"
    "go-slim.dev/infra/pkg/co"
)

type RequestProcessor struct {
    pool       *co.GoRoutinePool
    inFlight   atomic.Int32
    mutex      co.SimpleMutex
    processed  int
}

func NewRequestProcessor(workers int) *RequestProcessor {
    return &RequestProcessor{
        pool:  co.NewGoRoutinePool(workers),
        mutex: co.NewSimpleMutex(),
    }
}

func (rp *RequestProcessor) ProcessRequest(id int) {
    rp.inFlight.Add(1)

    rp.pool.Schedule(func() {
        defer rp.inFlight.Add(-1)

        // Simulate processing
        time.Sleep(10 * time.Millisecond)

        // Update counter with lock
        rp.mutex.Lock()
        rp.processed++
        fmt.Printf("Processed request %d (total: %d)\n", id, rp.processed)
        rp.mutex.Unlock()
    })
}

func (rp *RequestProcessor) WaitForCompletion() {
    co.WaitFor(&rp.inFlight)
}

func (rp *RequestProcessor) Shutdown() {
    rp.pool.Stop()
}

func main() {
    processor := NewRequestProcessor(5)
    defer processor.Shutdown()

    // Process 20 requests
    for i := 1; i <= 20; i++ {
        processor.ProcessRequest(i)
    }

    // Wait for all requests to complete
    processor.WaitForCompletion()
    fmt.Println("All requests processed!")
}
```

## Performance

### SimpleMutex vs sync.Mutex

```
BenchmarkSimpleMutex_LockUnlock-8       20000000    75.2 ns/op
BenchmarkSyncMutex_LockUnlock-8         50000000    35.1 ns/op

BenchmarkSimpleMutex_Contention-8       5000000     312 ns/op
BenchmarkSyncMutex_Contention-8         10000000    198 ns/op
```

SimpleMutex has higher overhead than sync.Mutex but offers:

- Simpler implementation
- Non-blocking TryLock
- Educational value

### GoRoutinePool Benefits

- **Reduced Allocations**: Reuses goroutines instead of spawning new ones
- **Controlled Concurrency**: Prevents resource exhaustion with worker limits
- **Predictable Performance**: Consistent behavior under load

### Spin-Wait Performance

Spin-wait is efficient for short waits (< 1ms) but CPU-intensive. For longer waits, prefer channels or sync.Cond.

## Testing

Run tests with:

```bash
# Run all tests
go test -v ./...

# Run with race detector
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem

# Run with coverage
go test -cover ./...
```

### Test Coverage

The package includes comprehensive tests for:

- **SimpleMutex**: Lock/unlock, TryLock, concurrency, contention
- **GoRoutinePool**: Task scheduling, worker limits, stop behavior, stress tests
- **Spin-Wait**: Immediate/eventual completion, backoff, multiple waiters

## Best Practices

1. **SimpleMutex**
   - Always unlock in defer to prevent deadlocks
   - Use TryLock when you can skip locked sections
   - Consider sync.Mutex for production-critical code

2. **GoRoutinePool**
   - Always call Stop() to prevent goroutine leaks (use defer)
   - Size the pool based on workload and resource constraints
   - Use sync.WaitGroup if you need to wait for task completion

3. **Spin-Wait**
   - Only use for very short waits (microseconds to low milliseconds)
   - Prefer channels/sync primitives for longer or unpredictable waits
   - Monitor CPU usage to ensure backoff is working

## License

This package is part of the goapp project.
