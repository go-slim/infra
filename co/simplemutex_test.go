package co

import (
	"sync"
	"testing"
	"time"
)

func todo() {
	// 空操作
}

func TestNewSimpleMutex(t *testing.T) {
	sm := NewSimpleMutex()
	if sm == nil {
		t.Fatal("NewSimpleMutex() 返回 nil")
	}
	if cap(sm) != 1 {
		t.Errorf("SimpleMutex 容量 = %d，期望 1", cap(sm))
	}
}

func TestSimpleMutex_LockUnlock(t *testing.T) {
	sm := NewSimpleMutex()

	// Lock 应该成功
	sm.Lock()

	todo()

	// Unlock 应该成功
	sm.Unlock()
}

func TestSimpleMutex_TryLock(t *testing.T) {
	sm := NewSimpleMutex()

	// 第一次 TryLock 应该成功
	if !sm.TryLock() {
		t.Error("第一次 TryLock() 失败，期望成功")
	}

	// 锁定状态下第二次 TryLock 应该失败
	if sm.TryLock() {
		t.Error("第二次 TryLock() 成功，期望失败")
	}

	// 解锁后再试
	sm.Unlock()
	if !sm.TryLock() {
		t.Error("Unlock() 后 TryLock() 失败，期望成功")
	}
	sm.Unlock()
}

func TestSimpleMutex_Concurrency(t *testing.T) {
	sm := NewSimpleMutex()
	counter := 0
	numGoroutines := 100
	numIterations := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range numIterations {
				sm.Lock()
				counter++
				sm.Unlock()
			}
		}()
	}

	wg.Wait()

	expected := numGoroutines * numIterations
	if counter != expected {
		t.Errorf("计数器 = %d，期望 %d（检测到竞态条件）", counter, expected)
	}
}

func TestSimpleMutex_BlockingBehavior(t *testing.T) {
	sm := NewSimpleMutex()

	sm.Lock()

	done := make(chan bool, 1)
	go func() {
		// 这应该阻塞直到锁被释放
		sm.Lock()
		done <- true
		sm.Unlock()
	}()

	// 给 goroutine 时间尝试获取锁
	time.Sleep(50 * time.Millisecond)

	select {
	case <-done:
		t.Error("Lock() 没有按预期阻塞")
	default:
		// 预期行为 - goroutine 被阻塞
	}

	// 释放锁
	sm.Unlock()

	// 现在 goroutine 应该完成
	select {
	case <-done:
		// 预期行为
	case <-time.After(100 * time.Millisecond):
		t.Error("Unlock() 后 Lock() 仍然阻塞")
	}
}

func TestSimpleMutex_MultipleLockUnlock(t *testing.T) {
	sm := NewSimpleMutex()

	for range 10 {
		sm.Lock()
		todo()
		sm.Unlock()
	}
}

func TestSimpleMutex_TryLockUnderContention(t *testing.T) {
	sm := NewSimpleMutex()

	sm.Lock()

	var wg sync.WaitGroup
	successCount := 0
	failCount := 0
	var mu sync.Mutex

	// 启动多个 goroutine 尝试获取锁
	numGoroutines := 10
	wg.Add(numGoroutines)

	for range numGoroutines {
		go func() {
			defer wg.Done()
			if sm.TryLock() {
				mu.Lock()
				successCount++
				mu.Unlock()
				sm.Unlock()
			} else {
				mu.Lock()
				failCount++
				mu.Unlock()
			}
		}()
	}

	// 给 goroutine 时间尝试获取锁
	time.Sleep(50 * time.Millisecond)

	// 释放主锁
	sm.Unlock()

	wg.Wait()

	// 在锁被持有期间，至少有一些 TryLock 调用应该失败
	if failCount == 0 {
		t.Error("期望在竞争下一些 TryLock() 调用失败")
	}

	// 最多一个应该成功（主解锁后）
	if successCount > 1 {
		t.Errorf("多个 TryLock() 成功：%d，期望最多 1 个", successCount)
	}
}

func BenchmarkSimpleMutex_LockUnlock(b *testing.B) {
	sm := NewSimpleMutex()

	for b.Loop() {
		sm.Lock()
		todo()
		sm.Unlock()
	}
}

func BenchmarkSimpleMutex_TryLock(b *testing.B) {
	sm := NewSimpleMutex()

	for b.Loop() {
		if sm.TryLock() {
			sm.Unlock()
		}
	}
}

func BenchmarkSimpleMutex_Contention(b *testing.B) {
	sm := NewSimpleMutex()
	counter := 0

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sm.Lock()
			counter++
			sm.Unlock()
		}
	})
}
