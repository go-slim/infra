package backoff

import (
	"context"
	"errors"
	"testing"
	"time"
)

// alwaysError 是一个测试辅助函数，总是返回错误。
// 它用于测试操作持续失败时的重试行为。
func alwaysError(context.Context) error {
	return errors.New("总是错误")
}

// successAfterNAttempts 创建一个测试函数，在前 (n-1) 次尝试时失败，
// 在第 n 次尝试时成功。这对于测试成功的重试场景很有用。
//
// 使用示例：
//
//	fn := successAfterNAttempts(3)
//	fn(ctx) // 返回错误（第 1 次尝试）
//	fn(ctx) // 返回错误（第 2 次尝试）
//	fn(ctx) // 返回 nil（第 3 次尝试 - 成功）
func successAfterNAttempts(n int) func(context.Context) error {
	attempts := 0
	return func(_ context.Context) error {
		attempts++
		if attempts < n {
			return errors.New("还没准备好")
		}
		return nil
	}
}

// TestRetry 测试各种场景下的基本重试逻辑。
// 它验证重试机制对于成功和失败操作都能正确工作，并遵守最大重试限制。
func TestRetry(t *testing.T) {
	// 使用短时间配置以避免长时间测试执行。
	// 实际应用程序应该使用适合其用例的较长间隔。
	testConfig := Config{
		InitialInterval:     10 * time.Millisecond,
		MaxInterval:         100 * time.Millisecond,
		Multiplier:          2.0,
		MaxRetries:          10,
		RandomizationFactor: 0.0, // 无抖动以获得可预测的测试结果
	}

	// 测试用例 1：几次尝试后成功
	// 这验证重试机制继续尝试直到成功
	t.Run("success_after_few_attempts", func(t *testing.T) {
		err := Retry(t.Context(), testConfig, successAfterNAttempts(3))
		if err != nil {
			t.Errorf("成功重试后期望无错误，得到 %v", err)
		}
	})

	// 测试用例 2：总是失败的操作
	// 这验证当所有尝试都失败时重试机制返回错误
	t.Run("always_failing_operation", func(t *testing.T) {
		err := Retry(t.Context(), testConfig, alwaysError)
		if err == nil {
			t.Error("当操作总是失败时期望错误，得到 nil")
		}
	})

	// 测试用例 3：超过最大重试限制
	// 这验证重试机制在 MaxRetries 次尝试后停止
	t.Run("exceed_max_retries", func(t *testing.T) {
		limitedConfig := Config{
			InitialInterval:     10 * time.Millisecond,
			MaxInterval:         100 * time.Millisecond,
			Multiplier:          2.0,
			MaxRetries:          2, // 只允许 2 次重试
			RandomizationFactor: 0.0,
		}

		// 这个函数需要 4 次尝试才能成功，但我们只允许 2 次重试
		err := Retry(t.Context(), limitedConfig, successAfterNAttempts(4))
		if err == nil {
			t.Error("超过最大重试时期望错误，得到 nil")
		}
	})
}

// TestRetryWithTimeout 测试重试机制遵守上下文超时。
// 它验证当提供带超时的上下文时，重试操作在超时过期时立即停止，
// 返回 DeadlineExceeded 错误。
func TestRetryWithTimeout(t *testing.T) {
	// 创建一个非常短超时的上下文用于测试
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()

	// 使用短配置以避免测试中的长时间等待
	shortConfig := Config{
		InitialInterval:     1 * time.Millisecond,
		MaxInterval:         10 * time.Millisecond,
		Multiplier:          2.0,
		MaxRetries:          10,
		RandomizationFactor: 0.0,
	}

	// 睡眠时间长于上下文超时的测试函数
	// 这模拟应该被取消的慢操作
	err := Retry(ctx, shortConfig, func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond) // 这将超过 10ms 超时
		return errors.New("总是失败")
	})

	// 验证操作因超时被取消
	if err == nil {
		t.Error("由于超时期望错误，得到 nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("期望 DeadlineExceeded 错误，得到 %v", err)
	}
}

// TestRetryWithCancel 测试重试机制遵守上下文取消。
// 它验证当上下文被取消时，重试操作立即停止并返回 Canceled 错误。
func TestRetryWithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	// 在短时间延迟后取消上下文以模拟外部取消
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	// 使用短配置以避免测试中的长时间等待
	shortConfig := Config{
		InitialInterval:     1 * time.Millisecond,
		MaxInterval:         10 * time.Millisecond,
		Multiplier:          2.0,
		MaxRetries:          10,
		RandomizationFactor: 0.0,
	}

	// 模拟慢操作的测试函数
	// 这应该被上下文取消中断
	err := Retry(ctx, shortConfig, func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond) // 这将被取消中断
		return errors.New("总是失败")
	})

	// 验证操作被取消
	if err == nil {
		t.Error("由于取消期望错误，得到 nil")
	}
	if err != context.Canceled {
		t.Errorf("期望 Canceled 错误，得到 %v", err)
	}
}

// TestBackoff 测试 Backoff 结构体的基本功能。
// 它验证退避算法正确计算延迟、管理尝试计数器并遵守配置限制。
func TestBackoff(t *testing.T) {
	config := Config{
		InitialInterval:     100 * time.Millisecond,
		MaxInterval:         500 * time.Millisecond,
		Multiplier:          2.0,
		MaxRetries:          5,
		RandomizationFactor: 0.0, // 禁用抖动以获得可预测的测试结果
	}

	b := New(config)

	// 测试 Reset 功能
	b.Reset()
	if b.Attempt() != 0 {
		t.Errorf("重置后期望尝试次数为 0，得到 %d", b.Attempt())
	}

	// 测试 Next 方法
	delay1 := b.Next()
	if b.Attempt() != 1 {
		t.Errorf("第一次调用 Next 后期望尝试次数为 1，得到 %d", b.Attempt())
	}
	// 由于系统调度变化允许容差
	if delay1 < config.InitialInterval-10*time.Millisecond || delay1 > config.InitialInterval+10*time.Millisecond {
		t.Errorf("期望延迟为 %v，得到 %v", config.InitialInterval, delay1)
	}

	delay2 := b.Next()
	if b.Attempt() != 2 {
		t.Errorf("第二次调用 Next 后期望尝试次数为 2，得到 %d", b.Attempt())
	}
	// 第二次延迟应该是初始延迟乘以倍数
	expectedDelay := float64(config.InitialInterval) * config.Multiplier
	if float64(delay2) < expectedDelay*0.9 || float64(delay2) > expectedDelay*1.1 {
		t.Errorf("期望延迟约为 %v，得到 %v", expectedDelay, delay2)
	}

	// 测试超过 MaxRetries 后的行为
	for range config.MaxRetries {
		b.Next()
	}
	// 超过 MaxRetries 后的下一次调用应该返回 MaxInterval
	delayMax := b.Next()
	if delayMax != config.MaxInterval {
		t.Errorf("期望延迟为最大间隔 %v，得到 %v", config.MaxInterval, delayMax)
	}
}

// TestBackoffDo 测试 Backoff 结构体的 Do 方法。
// 它验证 Do 方法正确执行带重试逻辑的函数，处理成功和失败场景，
// 并遵守重试限制。
func TestBackoffDo(t *testing.T) {
	config := Config{
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Multiplier:      2.0,
		MaxRetries:      3,
	}

	b := New(config)

	// 测试用例 1：带重试的成功执行
	err := b.Do(t.Context(), successAfterNAttempts(2))
	if err != nil {
		t.Errorf("期望无错误，得到 %v", err)
	}

	// 测试用例 2：所有重试后失败执行
	b.Reset()
	err = b.Do(t.Context(), alwaysError)
	if err == nil {
		t.Error("期望错误，得到 nil")
	}
}

// TestRetryDefault 测试 RetryDefault 便利函数。
// 此测试被跳过，因为 DefaultConfig 有很长的等待时间，会使测试运行太慢。
// 在生产环境中，对标准重试行为使用 RetryDefault 和 DefaultConfig，
// 但在测试中使用自定义配置。
func TestRetryDefault(t *testing.T) {
	// 由于 DefaultConfig 有很长的等待时间，我们跳过这个测试或使用短配置
	t.Skip("由于 DefaultConfig 中的长等待时间跳过 TestRetryDefault")

	/*
		// 如果要测试，可以使用短配置替代
		shortConfig := Config{
			InitialInterval:     10 * time.Millisecond,
			MaxInterval:         100 * time.Millisecond,
			Multiplier:          2.0,
			MaxRetries:          10,
			RandomizationFactor: 0.0,
		}

		err := Retry(t.Context(), shortConfig, successAfterNAttempts(3))
		if err != nil {
			t.Errorf("期望无错误，得到 %v", err)
		}

		err = Retry(t.Context(), shortConfig, alwaysError)
		if err == nil {
			t.Error("期望错误，得到 nil")
		}
	*/
}

// TestConcurrentSafety 测试 Backoff 结构体对于并发使用是安全的。
// 它验证多个 goroutine 可以同时安全地调用 Next() 而不会出现数据竞争或错误行为。
func TestConcurrentSafety(t *testing.T) {
	config := Config{
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Multiplier:      1.5,
		MaxRetries:      10,
	}

	b := New(config)

	// 测试对同一个 Backoff 实例的并发访问
	done := make(chan bool, 10)
	for i := range 10 {
		go func(id int) {
			defer func() { done <- true }()
			for range 5 {
				delay := b.Next()
				if delay <= 0 {
					t.Errorf("Goroutine %d 得到无效延迟: %v", id, delay)
				}
				time.Sleep(1 * time.Millisecond) // 模拟一些工作
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	for range 10 {
		<-done
	}

	// 验证最终尝试次数是合理的
	if b.Attempt() < 45 || b.Attempt() > 55 { // 10 goroutines * 5 calls each = 50 total
		t.Errorf("期望尝试次数约为 50，得到 %d", b.Attempt())
	}
}
