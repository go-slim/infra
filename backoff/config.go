package backoff

import "time"

// Config 定义了指数退避行为的配置参数。
// 这些参数控制退避算法计算重试尝试之间的等待时间以及何时停止重试。
//
// 应该根据您的具体用例要求配置字段：
//   - 对于快速失败系统：使用较短的间隔和较少的重试
//   - 对于弹性系统：使用较长的间隔和更多的重试
//   - 对于分布式系统：考虑更高的抖动以防止惊群效应
type Config struct {
	// InitialInterval 是第一次重试前的初始等待时间。
	// 应该根据操作通常恢复的速度来设置。
	//
	// 示例：
	//   - 100ms 用于快速本地操作
	//   - 1s 用于对响应服务的网络请求
	//   - 5s 用于数据库连接
	InitialInterval time.Duration

	// MaxInterval 是重试之间的最大等待时间。
	// 退避算法永远不会等待超过此持续时间，即使是指数增长。
	//
	// 应该根据应用程序的可接受延迟来设置。
	// 常见值范围从几秒到几分钟。
	MaxInterval time.Duration

	// Multiplier 是每次失败后等待时间增加的倍数。
	// 值为 2.0 意味着每次尝试后等待时间翻倍（标准指数退避）。
	//
	// 常见值：
	//   - 1.5 用于更渐进的退避（对资源更温和）
	//   - 2.0 用于标准指数退避
	//   - 3.0 用于激进退避（更快达到最大间隔）
	// 必须大于 1.0 才能有意义的退避。
	Multiplier float64

	// MaxRetries 是放弃前的最大重试尝试次数。
	// 这不包括初始尝试 - 值为 3 意味着在第一次失败后最多
	// 3 次额外尝试（总共 4 次尝试）。
	//
	// 基于以下因素设置：
	//   - 错误通常的瞬态程度
	//   - 重试成本 vs 失败成本
	//   - 用户体验考虑
	MaxRetries int

	// RandomizationFactor 添加抖动以防止多个客户端的同步重试尝试
	//（惊群问题）。实际等待时间将是：base_delay ± (base_delay * RandomizationFactor)。
	//
	// 值应该在 0.0 到 1.0 之间：
	//   - 0.0：无抖动（确定性等待时间）
	//   - 0.1：±10% 抖动（大多数情况推荐）
	//   - 0.5：±50% 抖动（高抖动，适用于大型分布式系统）
	//
	// 抖动帮助将重试尝试分散在时间上，减少当许多客户端同时经历故障时
	// 对下游服务的负载峰值。
	RandomizationFactor float64
}

// DefaultConfig 为常见用例提供合理的默认值。
// 适用于网络操作和 API 的通用重试逻辑。
//
// 默认值代表平衡的方法：
//   - 从适中的 500ms 初始延迟开始
//   - 以 1.5x 倍数指数增长（比 2x 更温和）
//   - 限制在 30 秒以防止过度等待
//   - 允许最多 10 次重试（总共 11 次尝试）
//   - 包含 10% 抖动以防止同步
var DefaultConfig = Config{
	InitialInterval:     500 * time.Millisecond,
	MaxInterval:         30 * time.Second,
	Multiplier:          1.5,
	MaxRetries:          10,
	RandomizationFactor: 0.1,
}

// ConfigBackoffAbility 是类型可以实现以提供自己退避配置的接口。
// 这允许在不同上下文中可配置的退避行为。
//
// 示例：
//
//	type DatabaseClient struct {
//	    backoffConfig Config
//	}
//
//	func (c *DatabaseClient) BackoffConfig() Config {
//	    return c.backoffConfig
//	}
//
// 然后您可以这样使用：
//
//	client := &DatabaseClient{backoffConfig: customConfig}
//	config := client.BackoffConfig()
//	err := Retry(ctx, config, client.Connect)
type ConfigBackoffAbility interface {
	BackoffConfig() Config
}
