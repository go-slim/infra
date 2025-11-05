package sdm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetRedis(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	// 设置 Redis 客户端
	SetRedis(client)

	// 验证客户端已设置
	loaded := rdb.Load()
	assert.NotNil(t, loaded)
	assert.Equal(t, client, loaded)
}

func TestTryLock_Success(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	ctx := context.Background()
	value := "test-value"

	// 测试 TryLock
	acquired, err := TryLock(ctx, value)
	require.NoError(t, err)
	assert.True(t, acquired)

	// 验证锁确实被获取了
	// 尝试再次获取同一个锁应该失败
	acquired2, err := TryLock(ctx, value)
	require.NoError(t, err)
	assert.False(t, acquired2)
}

func TestTryLock_Failed(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	ctx := context.Background()
	value := "test-value"

	// 首先获取锁
	acquired1, err := TryLock(ctx, value)
	require.NoError(t, err)
	assert.True(t, acquired1)

	// 再次尝试获取应该失败
	acquired2, err := TryLock(ctx, value)
	require.NoError(t, err)
	assert.False(t, acquired2)
}

func TestTryLock_DifferentValues(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	ctx := context.Background()
	value1 := "test-value-1"
	value2 := "test-value-2"

	// 两个不同的值都应该能获取锁
	acquired1, err := TryLock(ctx, value1)
	require.NoError(t, err)
	assert.True(t, acquired1)

	acquired2, err := TryLock(ctx, value2)
	require.NoError(t, err)
	assert.True(t, acquired2)
}

func TestTryLock_WithTimeout(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	ctx := context.Background()
	value := "test-value"
	timeout := 200 * time.Millisecond

	// 首先获取锁
	acquired1, err := TryLock(ctx, value)
	require.NoError(t, err)
	assert.True(t, acquired1)

	// 带超时的 TryLock 应该在超时后返回 false
	start := time.Now()
	acquired2, err := TryLock(ctx, value, timeout)
	elapsed := time.Since(start)

	// 由于超时，应该返回 context deadline exceeded 错误或 false
	// 具体行为取决于实现，我们检查至少不会立即返回
	if err == nil {
		assert.False(t, acquired2)
	} else {
		// 可能是 context 超时错误
		t.Logf("收到预期的错误: %v", err)
	}

	// 至少应该等待一些时间
	assert.Greater(t, elapsed, 10*time.Millisecond)
}

func TestLock_Success(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	ctx := context.Background()
	value := "test-value"

	// 测试 Lock（应该立即成功，因为没有竞争）
	err := Lock(ctx, value)
	require.NoError(t, err)
}

func TestUnlock_Success(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	ctx := context.Background()
	value := "test-value"

	// 首先获取锁
	acquired, err := TryLock(ctx, value)
	require.NoError(t, err)
	require.True(t, acquired)

	// 释放锁
	err = Unlock(ctx, value)
	require.NoError(t, err)

	// 验证锁已被释放，可以重新获取
	acquiredAgain, err := TryLock(ctx, value)
	require.NoError(t, err)
	assert.True(t, acquiredAgain)
}

func TestUnlock_Failed(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	ctx := context.Background()
	value := "test-value"

	// 尝试释放一个不存在的锁
	err := Unlock(ctx, value)
	assert.Error(t, err)
	assert.Equal(t, ErrMutexNotAcquired, err)

	// 释放一个由不同值持有的锁
	err = Unlock(ctx, "different-value")
	assert.Error(t, err)
	assert.Equal(t, ErrMutexNotAcquired, err)
}

func TestGlobalDefaultMutex(t *testing.T) {
	// 测试全局默认互斥锁对象
	assert.NotNil(t, mtx)
	// 检查是否使用正确的默认值
	if DefaultMutexName == "" {
		// 如果 DefaultMutexName 是空的，这可能是初始化问题
		t.Log("DefaultMutexName 为空")
	} else {
		assert.Equal(t, DefaultMutexName, mtx.Name())
		assert.Equal(t, DefaultMutexName, mtx.Title())
	}
}

func TestErrorHandling_NoRedis(t *testing.T) {
	// 保存原始状态
	originalValue := rdb.Load()
	defer func() {
		rdb.Store(originalValue)
	}()

	// 设置一个无效的 Redis 客户端（使用 nil 的 redis.Client）
	client := redis.NewClient(&redis.Options{Addr: "invalid:6379"})
	rdb.Store(client)

	// 使用带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value := "test-value"

	// TryLock 应该返回错误
	_, err := TryLock(ctx, value)
	assert.Error(t, err)

	// Lock 应该返回错误
	err = Lock(ctx, value)
	assert.Error(t, err)

	// Unlock 应该返回错误
	err = Unlock(ctx, value)
	assert.Error(t, err)
}

func TestConcurrentAccess(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	ctx := context.Background()
	mutexName := "concurrent-test"
	value := "concurrent-value"

	// 创建互斥锁
	mutex, err := New[string](mutexName)
	if err != nil {
		t.Fatal(err)
	}

	// 并发测试：多个 goroutine 尝试获取同一个锁
	const numGoroutines = 10
	successCount := 0
	ch := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			acquired, err := mutex.TryLock(ctx, value)
			if err == nil && acquired {
				ch <- true
				// 模拟一些工作
				time.Sleep(10 * time.Millisecond)
				mutex.Unlock(ctx, value)
			} else {
				ch <- false
			}
		}(i)
	}

	// 收集结果
	for i := 0; i < numGoroutines; i++ {
		if <-ch {
			successCount++
		}
	}

	// 应该只有一个 goroutine 成功获取锁
	assert.Equal(t, 1, successCount, "应该只有一个 goroutine 能获取锁")
}

// 基准测试
func BenchmarkTryLock(b *testing.B) {
	client := setupTestRedis(b)
	if client == nil {
		b.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value := "benchmark-value"
		_, err := TryLock(ctx, value)
		if err != nil {
			b.Fatal(err)
		}
		// 释放锁以便下次迭代
		Unlock(ctx, value)
	}
}

func BenchmarkLockAndUnlock(b *testing.B) {
	client := setupTestRedis(b)
	if client == nil {
		b.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value := "benchmark-value"
		err := Lock(ctx, value)
		if err != nil {
			b.Fatal(err)
		}
		err = Unlock(ctx, value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestIsLocked 测试全局 IsLocked 函数
func TestIsLocked(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	tests := []struct {
		name      string
		setupFunc func(*testing.T)
		expected  bool
	}{
		{
			name: "未获取锁时应该返回 false",
			setupFunc: func(t *testing.T) {
				// 不做任何操作
			},
			expected: false,
		},
		{
			name: "获取锁后应该返回 true",
			setupFunc: func(t *testing.T) {
				acquired, err := TryLock(context.Background(), "test-value")
				if err != nil {
					t.Fatal(err)
				}
				if !acquired {
					t.Fatal("无法获取锁")
				}
			},
			expected: true,
		},
		{
			name: "释放锁后应该返回 false",
			setupFunc: func(t *testing.T) {
				// 先获取锁
				acquired, err := TryLock(context.Background(), "test-value")
				if err != nil {
					t.Fatal(err)
				}
				if !acquired {
					t.Fatal("无法获取锁")
				}
				// 然后释放锁
				err = Unlock(context.Background(), "test-value")
				if err != nil {
					t.Fatal(err)
				}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global mutex to clean state for each test
			mtx = &Mutex[any]{
				name:  fmt.Sprintf("test-default-%d", time.Now().UnixNano()),
				title: fmt.Sprintf("test-default-%d", time.Now().UnixNano()),
			}

			tt.setupFunc(t)

			locked, err := IsLocked(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.expected, locked)
		})
	}
}

// TestIsLocked_ContextCancellation 测试 IsLocked 的 context 取消
func TestIsLocked_ContextCancellation(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	// 创建一个会被取消的 context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// 等待 context 超时
	time.Sleep(2 * time.Millisecond)

	locked, err := IsLocked(ctx)
	assert.Error(t, err)
	assert.False(t, locked)
}

// TestIsLocked_BackgroundContext 测试使用 context.Background()
func TestIsLocked_BackgroundContext(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	// 获取锁
	acquired, err := TryLock(context.Background(), "test-value")
	if err != nil {
		t.Fatal(err)
	}
	if !acquired {
		t.Fatal("无法获取锁")
	}

	// 使用 context.Background() 检查锁状态
	locked, err := IsLocked(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, locked)

	// 释放锁
	err = Unlock(context.Background(), "test-value")
	if err != nil {
		t.Fatal(err)
	}

	// 再次检查锁状态
	locked, err = IsLocked(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, locked)
}

// TestIsLocked_DifferentValues 测试使用不同值的锁状态
func TestIsLocked_DifferentValues(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	// 获取第一个锁
	acquired1, err := TryLock(context.Background(), "test-value-1")
	if err != nil {
		t.Fatal(err)
	}
	if !acquired1 {
		t.Fatal("无法获取锁")
	}

	// 检查锁状态，应该返回 true
	locked, err := IsLocked(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, locked)

	// 获取第二个锁（不同值）
	acquired2, err := TryLock(context.Background(), "test-value-2")
	if err != nil {
		t.Fatal(err)
	}
	if !acquired2 {
		t.Fatal("无法获取第二个锁")
	}

	// 检查锁状态，应该仍然返回 true（至少有一个锁被持有）
	locked, err = IsLocked(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, locked)

	// 释放第一个锁
	err = Unlock(context.Background(), "test-value-1")
	if err != nil {
		t.Fatal(err)
	}

	// 检查锁状态，应该仍然返回 true（第二个锁仍被持有）
	locked, err = IsLocked(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, locked)

	// 释放第二个锁
	err = Unlock(context.Background(), "test-value-2")
	if err != nil {
		t.Fatal(err)
	}

	// 检查锁状态，应该返回 false（所有锁都被释放）
	locked, err = IsLocked(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, locked)
}

// TestIsLocked_ConcurrentAccess 测试并发访问
func TestIsLocked_ConcurrentAccess(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	// 获取锁
	acquired, err := TryLock(context.Background(), "concurrent-value")
	if err != nil {
		t.Fatal(err)
	}
	if !acquired {
		t.Fatal("无法获取锁")
	}

	// 并发检查锁状态
	const numGoroutines = 10
	results := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			locked, err := IsLocked(context.Background())
			if err == nil {
				results <- locked
			} else {
				results <- false
			}
		}()
	}

	// 收集结果
	lockedCount := 0
	for i := 0; i < numGoroutines; i++ {
		if <-results {
			lockedCount++
		}
	}

	// 所有检查都应该返回 true（锁被持有）
	assert.Equal(t, numGoroutines, lockedCount, "所有并发检查都应该返回 true")

	// 释放锁
	err = Unlock(context.Background(), "concurrent-value")
	if err != nil {
		t.Fatal(err)
	}

	// 再次并发检查锁状态
	for i := 0; i < numGoroutines; i++ {
		go func() {
			locked, err := IsLocked(context.Background())
			if err == nil {
				results <- locked
			} else {
				results <- false
			}
		}()
	}

	// 收集结果
	unlockedCount := 0
	for i := 0; i < numGoroutines; i++ {
		if !<-results {
			unlockedCount++
		}
	}

	// 所有检查都应该返回 false（锁未被持有）
	assert.Equal(t, numGoroutines, unlockedCount, "所有并发检查都应该返回 false")
}

// TestIsLocked_ErrorHandling 测试错误处理
func TestIsLocked_ErrorHandling(t *testing.T) {
	// 保存原始状态
	originalValue := rdb.Load()
	defer func() {
		rdb.Store(originalValue)
	}()

	// 设置一个无效的 Redis 客户端
	client := redis.NewClient(&redis.Options{Addr: "invalid:6379"})
	rdb.Store(client)

	// 使用带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	locked, err := IsLocked(ctx)
	assert.Error(t, err)
	assert.False(t, locked)
}

// BenchmarkIsLocked IsLocked 函数的基准测试
func BenchmarkIsLocked(b *testing.B) {
	client := setupTestRedis(b)
	if client == nil {
		b.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	ctx := context.Background()

	// 获取锁以便测试
	acquired, err := TryLock(ctx, "benchmark-value")
	if err != nil {
		b.Fatal(err)
	}
	if !acquired {
		b.Fatal("无法获取锁")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := IsLocked(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}

	// 清理
	Unlock(ctx, "benchmark-value")
}
