package co

import (
	"runtime"
	"sync/atomic"
)

// WaitFunc 实现带指数退避的自旋等待机制。
// 它重复调用 busy 函数直到返回 false，使用自适应退避策略
// 在等待时减少 CPU 使用率。退避每次迭代加倍（1, 2, 4, 8, 16, 32），
// 以在响应性和 CPU 效率之间取得平衡。
func WaitFunc(busy func() bool) {
	backoff := 1

	// 使用自旋等待方法等待任务完成
	for busy() {
		// 根据 backoff 多次让出 CPU 时间片给其他 goroutine
		for i := 0; i < backoff; i++ {
			// 将 CPU 时间片让给其他 goroutine：
			// 当 goroutine 阻塞时，Go 会自动将同一系统线程上的
			// 其他 goroutine 移动到另一个系统线程，
			// 防止它们被阻塞。
			// 参考：https://juejin.cn/post/7207810396420358181
			runtime.Gosched()
		}

		// 指数增加退避时间，最大为 32
		if backoff < 32 {
			backoff <<= 1
		}
	}
}

// WaitFor 使用自旋等待原子计数器达到零。
// 它连续检查 atomic.Int32 值并阻塞直到变为零或负数。
// 这对于在并发场景中等待引用计数器或完成标志很有用。
func WaitFor(busy *atomic.Int32) {
	WaitFunc(func() bool {
		return busy.Load() > 0
	})
}
