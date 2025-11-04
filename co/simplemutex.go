package co

// SimpleMutex 是一个基于通道的互斥锁实现，用于简单的锁定场景。
// 它使用容量为 1 的缓冲通道来提供互斥访问。
// 这种实现适用于需要轻量级 sync.Mutex 替代方案的基本同步需求。
type SimpleMutex chan struct{}

// NewSimpleMutex 创建并返回一个新的 SimpleMutex 对象。
// 互斥锁使用大小为 1 的缓冲通道初始化，准备好用于锁定操作。
func NewSimpleMutex() SimpleMutex {
	return make(SimpleMutex, 1)
}

// Lock 获取互斥锁。
// 这是一个阻塞操作，将等待直到锁变为可用。
// 如果另一个 goroutine 已经持有锁，当前 goroutine 将阻塞。
func (s SimpleMutex) Lock() {
	s <- struct{}{}
}

// TryLock 尝试获取互斥锁而不阻塞。
// 如果成功获取锁返回 true，否则返回 false。
// 这是一个非阻塞操作，无论是否获取到锁都会立即返回。
func (s SimpleMutex) TryLock() bool {
	select {
	case s <- struct{}{}:
		return true
	default:
		return false
	}
}

// Unlock 释放互斥锁。
// 只能由持有锁的 goroutine 调用。
// 在未锁定的互斥锁上调用 Unlock 会导致 panic。
func (s SimpleMutex) Unlock() {
	<-s
}
