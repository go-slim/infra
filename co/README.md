# co - 并发工具包

[English Documentation](README.md)

`co` 包提供了一组轻量级的 Go 并发工具。包括基于 channel 的简单互斥锁、高效的 goroutine 池以及带指数退避的自旋等待工具。

## 目录

- [安装](#安装)
- [组件](#组件)
  - [SimpleMutex](#simplemutex)
  - [GoRoutinePool](#goroutinepool)
  - [自旋等待工具](#自旋等待工具)
- [示例](#示例)
- [性能](#性能)
- [测试](#测试)
- [许可证](#许可证)

## 安装

```bash
go get go-slim.dev/infra/pkg/co
```

## 组件

### SimpleMutex

`SimpleMutex` 是一个基于 channel 的互斥锁实现,为简单的锁场景提供了轻量级的替代方案。

#### 特性

- 使用容量为 1 的缓冲 channel 实现互斥
- 提供非阻塞的 `TryLock()` 操作
- 实现简单易懂
- 适用于基本的同步需求

#### API

```go
// 创建新的 SimpleMutex
sm := co.NewSimpleMutex()

// 加锁(阻塞)
sm.Lock()

// 尝试加锁(非阻塞)
if sm.TryLock() {
    // 获得锁
    defer sm.Unlock()
    // ... 临界区代码 ...
}

// 解锁
sm.Unlock()
```

#### 使用示例

```go
package main

import (
    "fmt"
    "go-slim.dev/infra/pkg/co"
)

func main() {
    sm := co.NewSimpleMutex()
    counter := 0

    // 启动多个 goroutine
    for i := 0; i < 100; i++ {
        go func() {
            sm.Lock()
            counter++
            sm.Unlock()
        }()
    }

    // 使用 TryLock 进行非阻塞获取
    if sm.TryLock() {
        fmt.Println("无阻塞获得锁")
        sm.Unlock()
    } else {
        fmt.Println("锁当前被其他 goroutine 持有")
    }
}
```

#### 使用场景

- **使用 SimpleMutex 的情况:**
  - 需要简单、轻量级的互斥锁
  - 需要通过 `TryLock()` 进行非阻塞锁尝试
  - 使用场景不需要高级特性如 RWMutex

- **使用 sync.Mutex 的情况:**
  - 需要经过实战检验的生产级同步原语
  - 需要高级特性(RWMutex 等)
  - 在高竞争场景下性能至关重要

### GoRoutinePool

`GoRoutinePool` 提供高效的 goroutine 池化,具有自动 worker 管理和并发控制功能。

#### 特性

- 高效的 goroutine 复用减少分配开销
- 可配置最大并发 worker 数量
- 自动 worker 创建和管理
- 支持优雅关闭
- 非阻塞任务调度

#### API

```go
// 创建最多 10 个并发 worker 的池
pool := co.NewGoRoutinePool(10)

// 调度任务
pool.Schedule(func() {
    // 你的任务代码
})

// 优雅停止池
pool.Stop()
```

#### 使用示例

```go
package main

import (
    "fmt"
    "sync"
    "go-slim.dev/infra/pkg/co"
)

func main() {
    // 创建有 5 个 worker 的池
    pool := co.NewGoRoutinePool(5)
    defer pool.Stop()

    var wg sync.WaitGroup

    // 调度 20 个任务
    for i := 0; i < 20; i++ {
        wg.Add(1)
        taskNum := i
        pool.Schedule(func() {
            defer wg.Done()
            fmt.Printf("执行任务 %d\n", taskNum)
            // 模拟工作
        })
    }

    wg.Wait()
    fmt.Println("所有任务完成")
}
```

#### 工作原理

1. **Worker 创建**: 当调度任务时:
   - 如果有空闲 worker,任务立即发送给它
   - 如果没有空闲 worker 且池未达到容量,则创建新 worker
   - Worker 自动按顺序处理多个任务

2. **Worker 生命周期**: 每个 worker 循环运行,处理任务直到:
   - 收到停止信号
   - Worker 退出并释放其信号量槽位

3. **并发控制**: 信号量 channel 限制最大并发 worker 数量

#### 使用场景

- **使用 GoRoutinePool 的情况:**
  - 需要限制并发 goroutine 执行数量
  - 有大量短期任务
  - 想要减少 goroutine 分配开销
  - 需要可预测的资源使用

- **使用原生 goroutine 的情况:**
  - 只有少量固定数量的任务
  - 任务是长期运行的
  - 不需要并发限制

### 自旋等待工具

自旋等待工具提供带指数退避的 CPU 高效忙等待机制。

#### 特性

- 指数退避策略(1, 2, 4, 8, 16, 32)
- 通过 `runtime.Gosched()` 实现 CPU 友好的让渡
- 通用的 `WaitFunc` 用于自定义条件
- 专用的 `WaitFor` 用于原子计数器

#### API

```go
// 等待直到条件变为 false
co.WaitFunc(func() bool {
    return !conditionMet()
})

// 等待原子计数器归零
var counter atomic.Int32
counter.Store(5)
co.WaitFor(&counter)
```

#### 使用示例

```go
package main

import (
    "fmt"
    "sync/atomic"
    "time"
    "go-slim.dev/infra/pkg/co"
)

func main() {
    // 示例 1: 等待条件满足
    ready := false
    go func() {
        time.Sleep(100 * time.Millisecond)
        ready = true
    }()

    co.WaitFunc(func() bool {
        return !ready
    })
    fmt.Println("条件满足!")

    // 示例 2: 等待计数器
    var counter atomic.Int32
    counter.Store(10)

    // 启动多个 goroutine 递减计数器
    for i := 0; i < 10; i++ {
        go func() {
            time.Sleep(10 * time.Millisecond)
            counter.Add(-1)
        }()
    }

    co.WaitFor(&counter)
    fmt.Println("所有 goroutine 完成!")
}
```

#### 退避机制工作原理

退避机制在等待时减少 CPU 使用:

```
迭代 1: 让渡 1 次
迭代 2: 让渡 2 次
迭代 3: 让渡 4 次
迭代 4: 让渡 8 次
迭代 5: 让渡 16 次
迭代 6+: 让渡 32 次(上限)
```

每次让渡调用 `runtime.Gosched()`,允许其他 goroutine 在同一 OS 线程上运行。

#### 使用场景

- **使用自旋等待工具的情况:**
  - 等待时间非常短(微秒到毫秒级)
  - 需要对忙等待进行细粒度控制
  - 与原子操作配合使用
  - 无锁算法需要等待

- **使用 channel/sync 原语的情况:**
  - 等待时间不可预测或可能很长
  - 需要等待多个条件
  - 标准同步模式适合你的使用场景

## 示例

### 完整示例: HTTP 请求处理器

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

        // 模拟处理
        time.Sleep(10 * time.Millisecond)

        // 使用锁更新计数器
        rp.mutex.Lock()
        rp.processed++
        fmt.Printf("已处理请求 %d (总计: %d)\n", id, rp.processed)
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

    // 处理 20 个请求
    for i := 1; i <= 20; i++ {
        processor.ProcessRequest(i)
    }

    // 等待所有请求完成
    processor.WaitForCompletion()
    fmt.Println("所有请求已处理!")
}
```

## 性能

### SimpleMutex vs sync.Mutex

```
BenchmarkSimpleMutex_LockUnlock-8       20000000    75.2 ns/op
BenchmarkSyncMutex_LockUnlock-8         50000000    35.1 ns/op

BenchmarkSimpleMutex_Contention-8       5000000     312 ns/op
BenchmarkSyncMutex_Contention-8         10000000    198 ns/op
```

SimpleMutex 相比 sync.Mutex 有更高的开销,但提供:

- 更简单的实现
- 非阻塞的 TryLock
- 教育价值

### GoRoutinePool 优势

- **减少分配**: 复用 goroutine 而不是创建新的
- **受控并发**: 通过 worker 限制防止资源耗尽
- **可预测性能**: 负载下行为一致

### 自旋等待性能

自旋等待对短时间等待(< 1ms)高效,但会占用 CPU。对于较长等待,建议使用 channel 或 sync.Cond。

## 测试

运行测试:

```bash
# 运行所有测试
go test -v ./...

# 使用竞态检测器运行
go test -race ./...

# 运行基准测试
go test -bench=. -benchmem

# 运行覆盖率测试
go test -cover ./...
```

### 测试覆盖

该包包含全面的测试:

- **SimpleMutex**: 加锁/解锁、TryLock、并发、竞争
- **GoRoutinePool**: 任务调度、worker 限制、停止行为、压力测试
- **自旋等待**: 立即/最终完成、退避、多等待者

## 最佳实践

1. **SimpleMutex**
   - 始终在 defer 中解锁以防止死锁
   - 当可以跳过已锁定的部分时使用 TryLock
   - 对于生产关键代码考虑使用 sync.Mutex

2. **GoRoutinePool**
   - 始终调用 Stop() 防止 goroutine 泄漏(使用 defer)
   - 根据工作负载和资源约束调整池大小
   - 如需等待任务完成,使用 sync.WaitGroup

3. **自旋等待**
   - 仅用于非常短的等待(微秒到低毫秒级)
   - 对于较长或不可预测的等待,优先使用 channel/sync 原语
   - 监控 CPU 使用率以确保退避机制正常工作

## 许可证

此包是 goapp 项目的一部分。
