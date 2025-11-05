package sdm

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// setupTestRedis 创建测试用的 Redis 客户端
// 注意：这些测试需要一个运行中的 Redis 实例
func setupTestRedis(t testing.TB) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // 默认 Redis 地址
		DB:   1,                // 使用专用的测试数据库
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		// 使用 t.Skip 或 b.Skip 来跳过测试
		t.Skip("Redis 不可用，跳过测试")
		return nil
	}

	// 清理测试数据
	client.FlushDB(ctx)

	return client
}
