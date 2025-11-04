# Backoff 包

[English Documentation](README.md)

一个为重试操作提供指数退避和抖动的 Go 包。该包帮助你为网络操作、数据库连接、API 调用和其他可能失败但重试可能成功的操作实现健壮的重试逻辑。

## 特性

- **指数退避**: 自动增加重试之间的等待时间
- **抖动支持**: 添加随机性以防止惊群效应
- **上下文集成**: 尊重上下文取消和超时
- **线程安全**: 可以安全地从多个 goroutine 并发使用
- **可配置**: 为不同用例提供灵活的配置
- **便利函数**: 为常见场景提供简单的 API

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [配置](#配置)
- [API 参考](#api-参考)
- [示例](#示例)
- [最佳实践](#最佳实践)
- [性能](#性能)
- [贡献](#贡献)

## 安装

```bash
go get go-slim.dev/infra/pkg/backoff
```

## 快速开始

### 基本用法

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
    // 使用默认配置进行简单重试
    err := backoff.Retry(context.Background(), backoff.DefaultConfig, func(ctx context.Context) error {
        return callExternalAPI()
    })

    if err != nil {
        log.Fatal("重试后操作失败:", err)
    }

    fmt.Println("操作成功!")
}

func callExternalAPI() error {
    // 你的可能失败的操作
    return nil // 或错误
}
```

### 便利函数

```go
// 使用 RetryDefault 快速实现，具有合理的默认值
err := backoff.RetryDefault(ctx, func(ctx context.Context) error {
    return database.Connect()
})
```

### 自定义配置

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

## 配置

`Config` 结构体提供对退避行为的完全控制：

| 字段                  | 类型            | 描述                     | 默认值 |
| --------------------- | --------------- | ------------------------ | ------ |
| `InitialInterval`     | `time.Duration` | 首次重试前的初始等待时间 | 500ms  |
| `MaxInterval`         | `time.Duration` | 重试之间的最大等待时间   | 30s    |
| `Multiplier`          | `float64`       | 退避乘数（必须 > 1.0）   | 1.5    |
| `MaxRetries`          | `int`           | 最大重试次数             | 10     |
| `RandomizationFactor` | `float64`       | 抖动因子（0.0-1.0）      | 0.1    |

### 配置指南

#### InitialInterval（初始间隔）

- **快速本地操作**: 10-100ms
- **网络请求**: 100ms-1s
- **数据库连接**: 1-5s

#### MaxInterval（最大间隔）

- **交互式应用**: 1-10s
- **批处理**: 30s-5min
- **后台服务**: 1-10min

#### Multiplier（乘数）

- **1.5**: 温和退避（大多数情况推荐）
- **2.0**: 标准指数退避
- **3.0**: 激进退避（快速达到最大间隔）

#### RandomizationFactor（随机化因子）

- **0.0**: 无抖动（确定性）
- **0.1**: 轻度抖动（大多数情况推荐）
- **0.5**: 重度抖动（大型分布式系统推荐）

## API 参考

### 函数

#### `Retry(ctx, config, fn) error`

使用指数退避重试逻辑执行函数。

**参数:**

- `ctx context.Context`: 用于取消和超时的上下文
- `config Config`: 退避配置
- `fn func(context.Context) error`: 要重试的函数

**返回:**

- `error`: 如果所有重试都失败则返回最后一个错误，成功时返回 nil

#### `RetryDefault(ctx, fn) error`

使用 `DefaultConfig` 的便利函数。

### 类型: Backoff

#### `New(config Config) *Backoff`

使用给定配置创建新的 Backoff 实例。

#### `(*Backoff) Next() time.Duration`

计算并返回下一个退避延迟。

#### `(*Backoff) Reset()`

将尝试计数器重置为零。

#### `(*Backoff) Attempt() int`

返回当前尝试次数。

#### `(*Backoff) Do(ctx, fn) error`

使用此 Backoff 实例执行带有重试逻辑的函数。

## 示例

### 带重试的 HTTP 客户端

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
            return fmt.Errorf("服务器错误: %d", resp.StatusCode)
        }

        return nil
    })

    if err != nil {
        fmt.Printf("请求失败: %v\n", err)
        return
    }

    fmt.Println("请求成功!")
}
```

### 带上下文超时的数据库连接

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

### 可重用的 Backoff 实例

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
            return fmt.Errorf("服务器错误: %d", resp.StatusCode)
        }

        return nil
    })
}
```

### 自定义重试条件

```go
func retryWithCustomCondition(ctx context.Context) error {
    var lastError error

    err := backoff.Retry(ctx, backoff.DefaultConfig, func(ctx context.Context) error {
        result, err := someOperation()
        if err != nil {
            lastError = err

            // 只在特定错误时重试
            if isRetryableError(err) {
                return err
            }
            return nil // 不重试非可重试错误
        }

        // 检查结果是否符合条件
        if !isAcceptableResult(result) {
            return fmt.Errorf("不可接受的结果")
        }

        return nil
    })

    if err != nil {
        return fmt.Errorf("操作失败: %w (最后错误: %v)", err, lastError)
    }

    return nil
}

func isRetryableError(err error) bool {
    // 定义哪些错误是可重试的
    return true // 你的逻辑
}

func isAcceptableResult(result interface{}) bool {
    // 定义哪些结果是可接受的
    return true // 你的逻辑
}
```

## 最佳实践

### 1. 选择适当的超时

```go
// 好: 使用上下文超时防止挂起
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := backoff.Retry(ctx, config, operation)

// 不好: 没有超时保护
err := backoff.Retry(context.Background(), config, operation)
```

### 2. 处理特定错误类型

```go
// 好: 只在瞬时错误时重试
err := backoff.Retry(ctx, config, func(ctx context.Context) error {
    err := someOperation()
    if isTransientError(err) {
        return err // 重试
    }
    return err // 不重试永久性错误
})

// 不好: 重试所有错误
err := backoff.Retry(ctx, config, someOperation)
```

### 3. 在分布式系统中使用抖动

```go
// 好: 添加抖动防止惊群效应
config := backoff.Config{
    RandomizationFactor: 0.1, // 10% 抖动
    // ... 其他字段
}

// 不好: 在分布式系统中无抖动
config := backoff.Config{
    RandomizationFactor: 0.0, // 所有客户端同时重试
    // ... 其他字段
}
```

### 4. 监控重试行为

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

### 5. 为不同环境配置

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

## 性能考虑

### 内存使用

- 每个 `Backoff` 实例维持最小状态（仅尝试计数器）
- 包使用原子操作保证线程安全
- 对大多数应用来说内存开销可以忽略

### CPU 使用

- 退避计算是 O(1) 且非常快
- 大部分 CPU 时间花在重试之间的等待上
- 抖动计算增加的开销很小

### Goroutine 使用

- 包不创建额外的 goroutine
- 所有操作都是同步和非阻塞的
- 正确尊重上下文取消

## 测试

运行测试：

```bash
# 运行所有测试
go test -v ./...

# 使用竞态检测器运行
go test -race -v ./...

# 运行基准测试
go test -bench=. -benchmem

# 运行覆盖率测试
go test -cover ./...
```

## 贡献

1. Fork 仓库
2. 创建功能分支
3. 为新功能添加测试
4. 确保所有测试通过
5. 提交 pull request

### 开发指南

- 保持 API 简单直观
- 维持向后兼容性
- 为新功能添加全面测试
- 更新 API 变更的文档
- 遵循 Go 约定和最佳实践

## 许可证

此包是 goapp 项目的一部分。

## 相关包

- [golang.org/x/net/context](https://pkg.go.dev/golang.org/x/net/context) - 上下文支持
- [github.com/cenkalti/backoff](https://github.com/cenkalti/backoff) - 替代的退避实现
- [github.com/sethvargo/go-retry](https://github.com/sethvargo/go-retry) - 功能更多的现代重试库
