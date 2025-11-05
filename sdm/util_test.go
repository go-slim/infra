package sdm

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/singleflight"
)

// TestUtilDB 测试 db() 函数
func TestUtilDB(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	t.Run("返回错误当没有设置Redis客户端", func(t *testing.T) {
		// 保存原始值
		originalValue := rdb.Load()
		defer func() {
			rdb.Store(originalValue)
		}()

		// 清空 Redis 客户端
		rdb.Store((*redis.Client)(nil))
		sfg = singleflight.Group{}

		_, err := db()
		assert.Error(t, err)
		assert.Equal(t, ErrRedisNotInitialized, err)
	})

	t.Run("使用已设置的Redis客户端", func(t *testing.T) {
		rdb.Store(client)

		retrievedClient, err := db()
		assert.NoError(t, err)
		assert.Equal(t, client, retrievedClient)
	})
}

// getRedisKeyWithConfig 是一个辅助函数，用于在测试中安全地调用 getRedisKey
func getRedisKeyWithConfig(prefix, defaultName, name string) (string, error) {
	// 保存原始值
	originalPrefix := RedisKeyPrefix
	originalDefault := DefaultMutexName

	// 设置测试值
	RedisKeyPrefix = prefix
	DefaultMutexName = defaultName

	// 确保恢复原始值
	defer func() {
		RedisKeyPrefix = originalPrefix
		DefaultMutexName = originalDefault
	}()

	return getRedisKey(name)
}

// TestGetRedisKey 测试 getRedisKey 函数
func TestGetRedisKey(t *testing.T) {
	tests := []struct {
		name        string
		prefix      string
		defaultName string
		input       string
		expectedKey string
		hasError    bool
	}{
		{
			name:        "正常名称",
			prefix:      "mutex",
			defaultName: "default",
			input:       "test-mutex",
			expectedKey: "mutex:test-mutex",
		},
		{
			name:        "空名称使用默认名称",
			prefix:      "mutex",
			defaultName: "default",
			input:       "",
			expectedKey: "mutex:default",
		},
		{
			name:        "只有空格的名称使用默认名称",
			prefix:      "mutex",
			defaultName: "default",
			input:       "   ",
			expectedKey: "mutex:default",
		},
		{
			name:        "带前后空格的名称被修剪",
			prefix:      "mutex",
			defaultName: "default",
			input:       "  test-mutex  ",
			expectedKey: "mutex:test-mutex",
		},
		{
			name:        "默认前缀和名称",
			prefix:      "mutex",
			defaultName: "default",
			input:       "custom",
			expectedKey: "mutex:custom",
		},
		{
			name:        "自定义前缀",
			prefix:      "custom",
			defaultName: "default",
			input:       "key",
			expectedKey: "custom:key",
		},
		{
			name:        "空前缀",
			prefix:      "",
			defaultName: "default",
			input:       "key",
			expectedKey: "key",
		},
		{
			name:        "空前缀和默认名称",
			prefix:      "",
			defaultName: "default",
			input:       "",
			expectedKey: "default",
		},
		{
			name:        "错误情况：前缀为空且名称为空",
			prefix:      "",
			defaultName: "",
			input:       "",
			hasError:    true,
		},
	}

	for _, tt := range tests {
		tt := tt // 创建局部变量副本
		t.Run(tt.name, func(t *testing.T) {
			key, err := getRedisKeyWithConfig(tt.prefix, tt.defaultName, tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedKey, key)
			}
		})
	}
}

// TestSerializeValue 测试 serializeValue 函数
func TestSerializeValue(t *testing.T) {
	tests := []struct {
		name          string
		value         interface{}
		expected      string
		expectedError error
		checkError    bool // 是否检查错误消息而不是直接比较错误
	}{
		{
			name:     "字符串类型",
			value:    "test-string",
			expected: "test-string",
		},
		{
			name:     "带空格的字符串",
			value:    "  test string  ",
			expected: "  test string  ",
		},
		{
			name:     "整数类型",
			value:    123,
			expected: "123",
		},
		{
			name:     "浮点数类型",
			value:    123.45,
			expected: "123.45",
		},
		{
			name:     "布尔类型 true",
			value:    true,
			expected: "true",
		},
		{
			name:     "布尔类型 false",
			value:    false,
			expected: "false",
		},
		{
			name: "结构体类型",
			value: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{
				ID:   1,
				Name: "test",
			},
			expected: `{"id":1,"name":"test"}`,
		},
		{
			name:     "切片类型",
			value:    []string{"a", "b", "c"},
			expected: `["a","b","c"]`,
		},
		{
			name: "map 类型",
			value: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
			expected: `{"key1":"value1","key2":123}`,
		},
		{
			name:     "空字符串",
			value:    "",
			expected: "",
		},
		{
			name:     "只有空格的字符串",
			value:    "   ",
			expected: "   ",
		},
		{
			name:     "零值整数",
			value:    0,
			expected: "0",
		},
		{
			name:     "nil 值",
			value:    nil,
			expected: "null",
		},
		{
			name:     "字节切片",
			value:    []byte("test"),
			expected: `"dGVzdA=="`,
		},
		{
			name:          "无法序列化的值（函数）",
			value:         func() {},
			expectedError: errors.New("sdm: failed to marshal value: json: unsupported type: func()"),
			checkError:    true,
		},
		{
			name:          "无法序列化的值（channel）",
			value:         make(chan int),
			expectedError: errors.New("sdm: failed to marshal value: json: unsupported type: chan int"),
			checkError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用泛型函数测试
			result, err := serializeValue(tt.value)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.checkError {
					assert.Contains(t, err.Error(), tt.expectedError.Error())
				} else {
					assert.Equal(t, tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestSerializeValue_Generic 测试 serializeValue 泛型函数
func TestSerializeValue_Generic(t *testing.T) {
	t.Run("字符串类型泛型", func(t *testing.T) {
		result, err := serializeValue("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("整数类型泛型", func(t *testing.T) {
		result, err := serializeValue(42)
		require.NoError(t, err)
		assert.Equal(t, "42", result)
	})

	t.Run("结构体泛型", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		value := TestStruct{Name: "Alice", Age: 30}
		result, err := serializeValue(value)
		require.NoError(t, err)
		assert.Equal(t, `{"name":"Alice","age":30}`, result)
	})

	t.Run("空值处理泛型", func(t *testing.T) {
		// 测试空字符串
		emptyString := ""
		result, err := serializeValue(emptyString)
		require.NoError(t, err)
		assert.Equal(t, "", result)

		// 测试 nil 值
		var nilString *string
		_, err = serializeValue(nilString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil string pointer")

		// 测试空结构体
		emptyStruct := struct{}{}
		_, err = serializeValue(emptyStruct)
		assert.NoError(t, err)
	})
}

// TestRedisScripts 测试 Redis 脚本
func TestRedisScripts(t *testing.T) {
	t.Run("TryLock 脚本不为空", func(t *testing.T) {
		assert.NotNil(t, tryLockScript)
	})

	t.Run("Unlock 脚本不为空", func(t *testing.T) {
		assert.NotNil(t, unlockScript)
	})

	t.Run("脚本键检查", func(t *testing.T) {
		assert.NotEmpty(t, tryLockScript)
		assert.NotEmpty(t, unlockScript)
	})
}

// 测试 Redis 脚本的实际执行逻辑
func TestRedisScriptExecution(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		t.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	ctx := context.Background()

	// 测试 tryLock 脚本
	result, err := tryLockScript.Run(ctx, client, []string{"test-key"}, "test-value").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), result)

	// 再次尝试同一个值应该失败
	result, err = tryLockScript.Run(ctx, client, []string{"test-key"}, "test-value").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), result)

	// 测试 unlock 脚本
	result, err = unlockScript.Run(ctx, client, []string{"test-key"}, "test-value").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), result)

	// 尝试释放不存在的锁应该失败
	result, err = unlockScript.Run(ctx, client, []string{"test-key"}, "test-value").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

// TestRedisConnectionError 测试 Redis 连接错误处理
func TestRedisConnectionError(t *testing.T) {
	// 保存原始值
	oldRdb := rdb.Load()
	defer func() {
		rdb.Store(oldRdb)
	}()

	t.Run("Redis 未初始化", func(t *testing.T) {
		// 清空 Redis 客户端
		rdb.Store((*redis.Client)(nil))

		_, err := db()
		assert.Error(t, err)
		assert.Equal(t, ErrRedisNotInitialized, err)
	})

	t.Run("Redis 连接错误", func(t *testing.T) {
		// 创建一个真实的 Redis 客户端，但指向一个无效的地址
		// 这样测试会更接近真实场景
		client := redis.NewClient(&redis.Options{
			Addr: "invalid-address:6379",
		})
		defer client.Close()
		rdb.Store(client)

		// 获取客户端应该成功
		scripter, err := db()
		assert.NoError(t, err)
		assert.NotNil(t, scripter)

		// 实际执行命令时会失败
		_, err = scripter.Eval(context.Background(), "return 1", []string{"test"}).Result()
		assert.Error(t, err)
	})
}

// mockRedisClient 是一个模拟的 Redis 客户端，用于测试错误情况
type mockRedisClient struct {
	err error
}

func (m *mockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	cmdArgs := []interface{}{"EVAL", script, len(keys)}
	for _, k := range keys {
		cmdArgs = append(cmdArgs, k)
	}
	for _, a := range args {
		cmdArgs = append(cmdArgs, a)
	}
	cmd := redis.NewCmd(ctx, cmdArgs...)
	cmd.SetErr(m.err)
	return cmd
}

func (m *mockRedisClient) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	cmdArgs := []interface{}{"EVALSHA", sha1, len(keys)}
	for _, k := range keys {
		cmdArgs = append(cmdArgs, k)
	}
	for _, a := range args {
		cmdArgs = append(cmdArgs, a)
	}
	cmd := redis.NewCmd(ctx, cmdArgs...)
	cmd.SetErr(m.err)
	return cmd
}

func (m *mockRedisClient) ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd {
	cmdArgs := []interface{}{"SCRIPT", "EXISTS"}
	for _, h := range hashes {
		cmdArgs = append(cmdArgs, h)
	}
	cmd := redis.NewBoolSliceCmd(ctx, cmdArgs...)
	cmd.SetErr(m.err)
	return cmd
}

func (m *mockRedisClient) ScriptLoad(ctx context.Context, script string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx, "SCRIPT", "LOAD", script)
	cmd.SetErr(m.err)
	return cmd
}

// 测试复杂类型的序列化
func TestComplexTypeSerialization(t *testing.T) {
	t.Run("嵌套结构体", func(t *testing.T) {
		type Address struct {
			Street string `json:"street"`
			City   string `json:"city"`
		}
		type Person struct {
			Name    string  `json:"name"`
			Age     int     `json:"age"`
			Address Address `json:"address"`
		}

		person := Person{
			Name: "John",
			Age:  30,
			Address: Address{
				Street: "123 Main St",
				City:   "New York",
			},
		}

		result, err := serializeValue(person)
		require.NoError(t, err)
		expected := `{"name":"John","age":30,"address":{"street":"123 Main St","city":"New York"}}`
		assert.Equal(t, expected, result)
	})

	t.Run("包含指针的结构体", func(t *testing.T) {
		type Node struct {
			Value string `json:"value"`
			Next  *Node  `json:"next,omitempty"`
		}

		node1 := &Node{Value: "first"}
		node2 := &Node{Value: "second", Next: node1}

		result, err := serializeValue(node2)
		require.NoError(t, err)
		expected := `{"value":"second","next":{"value":"first"}}`
		assert.Equal(t, expected, result)
	})

	t.Run("时间类型", func(t *testing.T) {
		now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		result, err := serializeValue(now)
		require.NoError(t, err)

		// 时间序列化格式
		expected, _ := json.Marshal(now)
		assert.Equal(t, string(expected), result)
	})
}

// TestRedisKeyGeneration 测试 Redis 键生成的各种场景
func TestRedisKeyGeneration(t *testing.T) {
	// 保存原始配置
	originalPrefix := RedisKeyPrefix
	originalDefault := DefaultMutexName
	defer func() {
		RedisKeyPrefix = originalPrefix
		DefaultMutexName = originalDefault
	}()

	// 使用 t.Run 的并行执行控制
	t.Run("不同前缀组合", func(t *testing.T) {
		// 不并行执行，避免修改全局变量导致竞态
		testCases := []struct {
			name     string
			prefix   string
			input    string
			expect   string
			hasError bool
		}{
			{"空前缀和空名称", "", "", "default", false},
			{"空名称使用默认值", "mutex", "", "mutex:default", false},
			{"正常前缀和名称", "mutex", "test", "mutex:test", false},
			{"不同前缀", "lock", "test", "lock:test", false},
			{"带命名空间的键", "myapp", "user:123", "myapp:user:123", false},
		}

		for _, tc := range testCases {
			tc := tc // 创建局部变量副本
			t.Run(tc.name, func(t *testing.T) {
				// 设置测试特定的前缀
				RedisKeyPrefix = tc.prefix

				// 执行测试
				key, err := getRedisKey(tc.input)

				// 验证结果
				if tc.hasError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tc.expect, key)
				}
			})
		}
	})

	t.Run("特殊字符处理", func(t *testing.T) {
		testCases := []struct {
			name   string
			input  string
			expect string
		}{
			{"冒号分隔符", "user:123", "mutex:user:123"},
			{"特殊字符@", "lock@resource", "mutex:lock@resource"},
			{"换行符", "test\nvalue", "mutex:test\nvalue"},
			{"空格", "test value", "mutex:test value"},
		}

		for _, tc := range testCases {
			tc := tc // 创建局部变量副本
			t.Run(tc.name, func(t *testing.T) {
				// 设置测试前缀
				RedisKeyPrefix = "mutex"

				key, err := getRedisKey(tc.input)
				require.NoError(t, err)
				assert.Equal(t, tc.expect, key)
			})
		}
	})
}

// 基准测试
func BenchmarkSerializeValue(b *testing.B) {
	value := struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Data []byte `json:"data"`
	}{
		ID:   123,
		Name: "benchmark test",
		Data: make([]byte, 1000),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := serializeValue(value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetRedisKey(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := getRedisKey("benchmark-key")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRedisScript(b *testing.B) {
	client := setupTestRedis(b)
	if client == nil {
		b.Skip("需要 Redis 服务器")
		return
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark-key"
		value := "benchmark-value"

		// 获取锁
		_, err := tryLockScript.Run(ctx, client, []string{key}, value).Result()
		if err != nil {
			b.Fatal(err)
		}

		// 释放锁
		_, err = unlockScript.Run(ctx, client, []string{key}, value).Result()
		if err != nil {
			b.Fatal(err)
		}
	}
}
