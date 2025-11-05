# 简单分布式互斥锁 (SDM)

[![Go 参考文档](https://pkg.go.dev/badge/go-slim.dev/infra/sdm.svg)](https://pkg.go.dev/go-slim.dev/infra/sdm)
[![Go 代码质量](https://goreportcard.com/badge/go-slim.dev/infra/sdm)](https://goreportcard.com/report/go-slim.dev/infra/sdm)
[![测试状态](https://github.com/go-slim/sdm/workflows/Test/badge.svg)](https://github.com/go-slim/sdm/actions?query=workflow%3ATest)

一个简单高效的基于 Redis 的分布式互斥锁实现，用于在多个进程或服务器之间协调对共享资源的访问。

## 功能特性

- 🚀 简单易用的 API
- 🔒 基于 Redis 的分布式锁实现
- ⏱️ 自动锁过期，防止死锁
- 🔄 支持阻塞和非阻塞的锁获取方式
- 🛡️ 线程安全，完善的错误处理
- 🧩 可配置的超时和重试策略
- 🔄 自动重试和指数退避
- 🔍 锁状态检查功能，无需获取锁即可查询状态

## 安装

```bash
go get go-slim.dev/infra/sdm
```

## 快速开始

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
	// 初始化 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 设置全局 Redis 客户端
	sdm.SetRedis(rdb)

	// 尝试获取锁
	locked, err := sdm.TryLock(context.Background(), "进程-1", 5*time.Second)
	if err != nil {
		log.Fatalf("获取锁失败: %v", err)
	}
	if !locked {
		log.Fatal("未能在超时时间内获取锁")
	}

	// 确保在完成后释放锁
	defer sdm.Unlock(context.Background(), "进程-1")

	// 临界区代码
	log.Println("成功获取锁，执行任务中...")
	time.Sleep(2 * time.Second)
}
```

## 高级用法

### 创建命名的互斥锁

```go
m, err := sdm.NewMutex("资源-123", "资源更新锁")
if err != nil {
    log.Fatal(err)
}

err = m.Lock(context.Background(), "进程-1")
if err != nil {
    log.Fatal(err)
}
defer m.Unlock(context.Background(), "进程-1")

// 操作受保护的资源
```

### 使用自定义超时

```go
// 尝试在5秒内获取锁
acquired, err := sdm.TryLock(context.Background(), "进程-1", 5*time.Second)
if err != nil {
    log.Fatal(err)
}
if !acquired {
    log.Println("未能在超时时间内获取锁")
    return
}
defer sdm.Unlock(context.Background(), "进程-1")
```

### 检查锁状态

```go
// 检查互斥锁是否被持有
m, err := sdm.NewMutex("资源-123")
if err != nil {
    log.Fatal(err)
}

locked, err := m.IsLocked(context.Background())
if err != nil {
    log.Fatal(err)
}
if locked {
    log.Println("资源当前被锁定")
} else {
    log.Println("资源当前未被锁定")
}

// 全局锁状态检查
globalLocked, err := sdm.IsLocked(context.Background())
if err != nil {
    log.Fatal(err)
}
if globalLocked {
    log.Println("全局互斥锁当前被锁定")
}
```

## 配置

### 全局设置

```go
// 修改默认的 Redis 键前缀（默认: "mutex"）
sdm.RedisKeyPrefix = "myapp:mutex"

// 修改默认的互斥锁名称（默认: "default"）
sdm.DefaultMutexName = "全局锁"
```

## 错误处理

常见的错误类型：

- `sdm.ErrMutexNameEmpty`: 尝试创建空名称的互斥锁时返回
- `sdm.ErrInvalidMutexValue`: 互斥锁值无效（空值或序列化失败）
- `sdm.ErrMutexNotAcquired`: 在指定超时时间内无法获取锁

## 最佳实践

1. 始终使用 `defer` 确保锁被释放
2. 设置合理的超时时间，避免死锁
3. 使用描述性的锁名称来标识资源
4. 正确处理错误
5. 尽量缩短临界区代码的执行时间

## 许可证

MIT

## 贡献

欢迎贡献代码！请随时提交 Pull Request。
