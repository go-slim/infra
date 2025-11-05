package sdm

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutex_New(t *testing.T) {
	tests := []struct {
		name        string
		mutexName   string
		title       string
		expectError bool
		expectedErr error
	}{
		{
			name:        "成功创建互斥锁",
			mutexName:   "test-mutex",
			title:       "Test Mutex",
			expectError: false,
		},
		{
			name:        "成功创建互斥锁-无标题",
			mutexName:   "test-mutex-2",
			expectError: false,
		},
		{
			name:        "空名称应该返回错误",
			mutexName:   "",
			expectError: true,
			expectedErr: ErrMutexNameEmpty,
		},
		{
			name:        "只有空格的名称应该返回错误",
			mutexName:   "   ",
			expectError: true,
			expectedErr: ErrMutexNameEmpty,
		},
		{
			name:        "带前后空格的名称应该被修剪",
			mutexName:   "  test-mutex  ",
			title:       "  Test Mutex  ",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mutex Mutex[string]
			var err error

			if tt.title != "" {
				mutex, err = New[string](tt.mutexName, tt.title)
			} else {
				mutex, err = New[string](tt.mutexName)
			}

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				expectedName := "test-mutex"
				if tt.mutexName == "  test-mutex  " {
					expectedName = "test-mutex"
				} else {
					expectedName = tt.mutexName
				}
				assert.Equal(t, expectedName, mutex.Name())

				if tt.title != "" {
					expectedTitle := "Test Mutex"
					if tt.title == "  Test Mutex  " {
						expectedTitle = "Test Mutex"
					} else if tt.title == "" {
						expectedTitle = expectedName
					}
					assert.Equal(t, expectedTitle, mutex.Title())
				} else {
					assert.Equal(t, expectedName, mutex.Title())
				}
			}
		})
	}
}

func TestMutex_TryLock(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	tests := []struct {
		name     string
		value    string
		timeout  time.Duration
		expected bool
	}{
		{
			name:     "成功获取锁",
			value:    "test-value-1",
			timeout:  0,
			expected: true,
		},
		{
			name:     "重复获取锁应该失败",
			value:    "test-value-1", // 同一个值
			timeout:  0,
			expected: false,
		},
		{
			name:     "不同值可以获取锁",
			value:    "test-value-2",
			timeout:  0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mutex, err := New[string]("test-try-lock")
			if err != nil {
				t.Fatal(err)
			}

			var acquired bool

			if tt.timeout > 0 {
				acquired, err = mutex.TryLock(context.Background(), tt.value, tt.timeout)
			} else {
				acquired, err = mutex.TryLock(context.Background(), tt.value)
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, acquired)
		})
	}
}

func TestMutex_TryLock_WithTimeout(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	mutex, err := New[string]("test-timeout")
	if err != nil {
		t.Fatal(err)
	}

	// 测试使用相同值的情况（应该失败）
	acquired1, err := mutex.TryLock(context.Background(), "same-value")
	if err != nil {
		t.Fatal(err)
	}
	require.True(t, acquired1)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 带超时尝试使用相同值获取锁
	start := time.Now()
	acquired2, err := mutex.TryLock(ctx, "same-value", 100*time.Millisecond)
	elapsed := time.Since(start)

	// 使用相同值应该返回 false（已被占用）
	if err == nil {
		assert.False(t, acquired2)
	} else {
		// 可能是 context 超时错误或其他错误
		t.Logf("收到预期的错误: %v", err)
	}

	// 由于使用相同值，应该立即返回，不需要等待
	// 但由于我们设置了100ms超时，它会等待完整的超时时间
	// 检查等待时间在合理范围内（考虑Redis操作延迟）
	assert.Greater(t, elapsed, 90*time.Millisecond)
	assert.Less(t, elapsed, 150*time.Millisecond)
}

func TestMutex_Lock(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	mutex, err := New[string]("test-lock")
	if err != nil {
		t.Fatal(err)
	}

	err = mutex.Lock(context.Background(), "test-value")
	if err != nil {
		t.Fatal(err)
	}
}

func TestMutex_Unlock(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	// 首先获取锁
	mutex, err := New[string]("test-unlock")
	if err != nil {
		t.Fatal(err)
	}
	acquired, err := mutex.TryLock(context.Background(), "test-value")
	if err != nil {
		t.Fatal(err)
	}
	require.True(t, acquired)

	// 释放锁
	err = mutex.Unlock(context.Background(), "test-value")
	if err != nil {
		t.Fatal(err)
	}

	// 验证锁已被释放
	acquiredAgain, err := mutex.TryLock(context.Background(), "test-value")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, acquiredAgain)
}

func TestMutex_Unlock_NotOwned(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	mutex, err := New[string]("test-unlock-fail")
	if err != nil {
		t.Fatal(err)
	}

	// 尝试释放一个不存在的锁
	err = mutex.Unlock(context.Background(), "non-existent-value")
	assert.Error(t, err)
	assert.Equal(t, ErrMutexNotAcquired, err)
}

func TestMutex_ContextCancellation(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	mutex, err := New[string]("test-cancel")
	if err != nil {
		t.Fatal(err)
	}

	// 创建一个会被取消的 context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// 等待 context 超时
	time.Sleep(11 * time.Millisecond)

	acquired, err := mutex.TryLock(ctx, "test-value", 100*time.Millisecond)
	assert.Error(t, err)
	assert.False(t, acquired)
}

func TestMutex_ContextBackground(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	mutex, err := New[string]("test-background-context")
	if err != nil {
		t.Fatal(err)
	}

	// 测试 TryLock with context.Background()
	acquired, err := mutex.TryLock(context.Background(), "test-value")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, acquired)

	// 测试 Unlock with context.Background()
	err = mutex.Unlock(context.Background(), "test-value")
	if err != nil {
		t.Fatal(err)
	}
}

// 测试不同类型的值
func TestMutex_DifferentTypes(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	t.Run("字符串类型", func(t *testing.T) {
		mutex, err := New[string]("test-string")
		if err != nil {
			t.Fatal(err)
		}
		acquired, err := mutex.TryLock(context.Background(), "string-value")
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, acquired)
	})

	t.Run("整数类型", func(t *testing.T) {
		mutex, err := New[int]("test-int")
		if err != nil {
			t.Fatal(err)
		}
		acquired, err := mutex.TryLock(context.Background(), 123)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, acquired)
	})

	t.Run("结构体类型", func(t *testing.T) {
		type TestStruct struct {
			ID   int
			Name string
		}
		mutex, err := New[TestStruct]("test-struct")
		if err != nil {
			t.Fatal(err)
		}
		value := TestStruct{ID: 1, Name: "test"}
		acquired, err := mutex.TryLock(context.Background(), value)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, acquired)
	})

	t.Run("指针类型", func(t *testing.T) {
		type TestStruct struct {
			ID int
		}
		mutex, err := New[*TestStruct]("test-pointer")
		if err != nil {
			t.Fatal(err)
		}
		value := &TestStruct{ID: 1}
		acquired, err := mutex.TryLock(context.Background(), value)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, acquired)
	})
}

func TestMutex_ConcurrentAccess(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	mutex, err := New[string]("concurrent-test")
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			acquired, err := mutex.TryLock(context.Background(), "concurrent-value")
			if err != nil {
				t.Errorf("Goroutine %d 获取锁时出错: %v", id, err)
				return
			}

			if acquired {
				mu.Lock()
				successCount++
				mu.Unlock()

				// 模拟一些工作
				time.Sleep(10 * time.Millisecond)

				err = mutex.Unlock(context.Background(), "concurrent-value")
				if err != nil {
					t.Errorf("Goroutine %d 释放锁时出错: %v", id, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// 应该只有一个 goroutine 成功获取锁
	assert.Equal(t, 1, successCount, "应该只有一个 goroutine 能获取锁")
}

// 基准测试
func BenchmarkMutex_TryLock(b *testing.B) {
	client := setupTestRedis(b)
	if client == nil {
		b.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	mutex, err := New[string]("benchmark-try-lock")
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	value := "benchmark-value"

	for b.Loop() {
		_, err := mutex.TryLock(ctx, value)
		if err != nil {
			b.Fatal(err)
		}
		// 释放锁以便下次迭代
		err = mutex.Unlock(ctx, value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMutex_New(b *testing.B) {
	for b.Loop() {
		_, err := New[string]("benchmark-mutex")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestMutex_IsLocked 测试 Mutex.IsLocked 方法
func TestMutex_IsLocked(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	tests := []struct {
		name      string
		setupFunc func(*testing.T, Mutex[string])
		expected  bool
	}{
		{
			name: "未获取锁时应该返回 false",
			setupFunc: func(t *testing.T, m Mutex[string]) {
				// 不做任何操作
			},
			expected: false,
		},
		{
			name: "获取锁后应该返回 true",
			setupFunc: func(t *testing.T, m Mutex[string]) {
				acquired, err := m.TryLock(context.Background(), "test-value")
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
			setupFunc: func(t *testing.T, m Mutex[string]) {
				// 先获取锁
				acquired, err := m.TryLock(context.Background(), "test-value")
				if err != nil {
					t.Fatal(err)
				}
				if !acquired {
					t.Fatal("无法获取锁")
				}
				// 然后释放锁
				err = m.Unlock(context.Background(), "test-value")
				if err != nil {
					t.Fatal(err)
				}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use unique mutex name for each test to avoid interference
			mutexName := fmt.Sprintf("test-is-locked-%d", time.Now().UnixNano())
			mutex, err := New[string](mutexName)
			if err != nil {
				t.Fatal(err)
			}

			tt.setupFunc(t, mutex)

			locked, err := mutex.IsLocked(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.expected, locked)
		})
	}
}

// TestMutex_IsLocked_ContextCancellation 测试 IsLocked 的 context 取消
func TestMutex_IsLocked_ContextCancellation(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	mutex, err := New[string]("test-is-locked-cancel")
	if err != nil {
		t.Fatal(err)
	}

	// 创建一个会被取消的 context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// 等待 context 超时
	time.Sleep(2 * time.Millisecond)

	locked, err := mutex.IsLocked(ctx)
	assert.Error(t, err)
	assert.False(t, locked)
}

// TestMutex_IsLocked_BackgroundContext 测试使用 context.Background()
func TestMutex_IsLocked_BackgroundContext(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	mutex, err := New[string]("test-is-locked-background")
	if err != nil {
		t.Fatal(err)
	}

	// 获取锁
	acquired, err := mutex.TryLock(context.Background(), "test-value")
	if err != nil {
		t.Fatal(err)
	}
	if !acquired {
		t.Fatal("无法获取锁")
	}

	// 使用 context.Background() 检查锁状态
	locked, err := mutex.IsLocked(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, locked)
}

// TestMutex_IsLocked_DifferentMutexes 测试不同的互斥锁实例
func TestMutex_IsLocked_DifferentMutexes(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	mutex1, err := New[string]("test-is-locked-1")
	if err != nil {
		t.Fatal(err)
	}

	mutex2, err := New[string]("test-is-locked-2")
	if err != nil {
		t.Fatal(err)
	}

	// 只获取 mutex1 的锁
	acquired, err := mutex1.TryLock(context.Background(), "test-value")
	if err != nil {
		t.Fatal(err)
	}
	if !acquired {
		t.Fatal("无法获取锁")
	}

	// 检查 mutex1 应该被锁定
	locked1, err := mutex1.IsLocked(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, locked1)

	// 检查 mutex2 应该未被锁定
	locked2, err := mutex2.IsLocked(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, locked2)
}

// TestMutex_IsLocked_ConcurrentAccess 测试并发访问
func TestMutex_IsLocked_ConcurrentAccess(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	SetRedis(client)

	mutex, err := New[string]("test-is-locked-concurrent")
	if err != nil {
		t.Fatal(err)
	}

	// 获取锁
	acquired, err := mutex.TryLock(context.Background(), "concurrent-value")
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
			locked, err := mutex.IsLocked(context.Background())
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
	err = mutex.Unlock(context.Background(), "concurrent-value")
	if err != nil {
		t.Fatal(err)
	}

	// 再次并发检查锁状态
	for i := 0; i < numGoroutines; i++ {
		go func() {
			locked, err := mutex.IsLocked(context.Background())
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
