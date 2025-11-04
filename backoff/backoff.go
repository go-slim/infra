// Package backoff 提供带抖动的指数退避重试机制。
//
// 退避算法实现了指数退避和可选的抖动机制，用于防止多个客户端同时重试时产生的惊群效应。
// 通常用于网络操作、数据库连接和其他可能在重试时成功的潜在失败操作。
//
// 基本用法：
//
//	err := Retry(ctx, DefaultConfig, func(ctx context.Context) error {
//	    return connectToDatabase()
//	})
//	if err != nil {
//	    log.Fatal("重试后连接失败:", err)
//	}
//
// 本包是线程安全的，可以从多个 goroutine 中并发使用。
package backoff

import (
	"context"
	"math"
	"math/rand/v2"
	"sync/atomic"
	"time"
)

// Backoff 实现了带抖动的指数退避算法。
// 它维护关于当前尝试次数的状态，并根据配置计算重试之间的适当等待时间。
//
// 该结构体对于并发使用是安全的，因为它使用原子操作来管理尝试计数器。
type Backoff struct {
	config  *Config
	attempt int32
}

// New 使用给定配置创建一个新的 Backoff 实例。
// 配置中的零值或负值将被替换为 DefaultConfig 中的默认值。
//
// 示例：
//
//	config := Config{
//	    InitialInterval: 100 * time.Millisecond,
//	    MaxInterval:     5 * time.Second,
//	    Multiplier:      2.0,
//	    MaxRetries:      5,
//	    RandomizationFactor: 0.1,
//	}
//	backoff := New(config)
func New(config Config) *Backoff {
	if config.InitialInterval <= 0 {
		config.InitialInterval = DefaultConfig.InitialInterval
	}
	if config.MaxInterval <= 0 {
		config.MaxInterval = DefaultConfig.MaxInterval
	}
	if config.Multiplier <= 1 {
		config.Multiplier = DefaultConfig.Multiplier
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = DefaultConfig.MaxRetries
	}
	if config.RandomizationFactor <= 0 {
		config.RandomizationFactor = DefaultConfig.RandomizationFactor
	}

	return &Backoff{
		config:  &config,
		attempt: 0,
	}
}

// Next 使用指数退避和抖动计算下一个退避延迟。
// 它增加尝试计数器并返回等待周期的 time.Duration。
//
// 计算遵循以下模式：
//
//	base_delay = initial_interval * (multiplier ^ (attempt - 1))
//	final_delay = min(base_delay, max_interval)
//	with_jitter = final_delay ± (final_delay * randomization_factor)
//
// 如果尝试次数超过 MaxRetries，则返回最大间隔。
func (b *Backoff) Next() time.Duration {
	attempt := atomic.AddInt32(&b.attempt, 1)

	// 如果超过重试限制，返回最大间隔
	if int(attempt) > b.config.MaxRetries {
		return b.config.MaxInterval
	}

	// 使用指数退避公式计算基础延迟
	// delay = initial_interval * (multiplier ^ (attempt - 1))
	delay := float64(b.config.InitialInterval) * math.Pow(b.config.Multiplier, float64(attempt-1))

	// 将延迟限制在最大间隔内
	if delay > float64(b.config.MaxInterval) {
		delay = float64(b.config.MaxInterval)
	}

	// 添加抖动以防止惊群问题
	if b.config.RandomizationFactor > 0 {
		delta := delay * b.config.RandomizationFactor
		min := delay - delta
		max := delay + delta
		delay = min + rand.Float64()*(max-min)
	}

	return time.Duration(delay)
}

// Reset 将退避尝试计数器重置为零。
// 当您想要重用同一个 Backoff 实例进行新的重试尝试系列时，应该调用此方法。
func (b *Backoff) Reset() {
	atomic.StoreInt32(&b.attempt, 0)
}

// Attempt 返回当前尝试次数。
// 计数从 0 开始，每次调用 Next() 时递增。
// 此方法是线程安全的。
func (b *Backoff) Attempt() int {
	return int(atomic.LoadInt32(&b.attempt))
}

// Do 使用指数退避重试逻辑执行给定函数。
// 它将重试函数最多 MaxRetries 次，延迟递增。
//
// 函数将在以下情况下停止重试并立即返回：
//   - 函数返回 nil（成功）
//   - 上下文被取消（返回上下文错误）
//   - 进行了 MaxRetries 次尝试（返回最后一个错误）
//
// 示例：
//
//	err := backoff.Do(ctx, func(ctx context.Context) error {
//	    return apiCall(ctx)
//	})
//	if err != nil {
//	    return fmt.Errorf("API 调用重试后失败: %w", err)
//	}
func (b *Backoff) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	for i := 0; i < b.config.MaxRetries; i++ {
		err := fn(ctx)
		if err == nil {
			return nil // 成功，无需更多重试
		}

		// 如果不是最后一次尝试，在重试前等待
		if i < b.config.MaxRetries-1 {
			// 计算下次重试的等待时间
			waitTime := b.Next()

			// 等待计算的时间或直到上下文被取消
			select {
			case <-time.After(waitTime):
				continue // 等待完成，继续下次重试
			case <-ctx.Done():
				return ctx.Err() // 上下文被取消
			}
		} else {
			// 这是最后一次尝试，返回错误
			return err
		}
	}

	// 理论上，代码不会执行到此处，
	// 但需要返回一个值保证语法正确。
	return nil
}

// Retry 使用提供的配置执行带指数退避的给定函数。
// 这是一个便利函数，创建新的 Backoff 实例并调用 Do()。
//
// 示例：
//
//	err := Retry(ctx, DefaultConfig, func(ctx context.Context) error {
//	    return http.Get(url)
//	})
//
// 这等同于：
//
//	backoff := New(config)
//	err := backoff.Do(ctx, fn)
func Retry(ctx context.Context, config Config, fn func(ctx context.Context) error) error {
	return New(config).Do(ctx, fn)
}

// RetryDefault 使用 DefaultConfig 执行带指数退避的给定函数。
// 这是一个便利函数，用于想要使用默认退避配置的常见情况。
//
// 示例：
//
//	err := RetryDefault(ctx, func(ctx context.Context) error {
//	    return database.Connect()
//	})
func RetryDefault(ctx context.Context, fn func(ctx context.Context) error) error {
	return Retry(ctx, DefaultConfig, fn)
}
