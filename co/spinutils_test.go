package co

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitFunc_ImmediateCompletion(t *testing.T) {
	called := false

	WaitFunc(func() bool {
		if !called {
			called = true
			return true
		}
		return false
	})

	if !called {
		t.Error("WaitFunc 没有调用 busy 函数")
	}
}

func TestWaitFunc_AlreadyComplete(t *testing.T) {
	callCount := 0

	WaitFunc(func() bool {
		callCount++
		return false // 已经完成
	})

	if callCount != 1 {
		t.Errorf("busy 函数被调用 %d 次，期望 1 次", callCount)
	}
}

func TestWaitFunc_EventualCompletion(t *testing.T) {
	counter := 0
	maxCount := 5

	WaitFunc(func() bool {
		counter++
		return counter < maxCount
	})

	if counter != maxCount {
		t.Errorf("计数器 = %d，期望 %d", counter, maxCount)
	}
}

func TestWaitFunc_BackoffBehavior(t *testing.T) {
	callCount := 0

	start := time.Now()
	WaitFunc(func() bool {
		callCount++
		return callCount < 100
	})
	elapsed := time.Since(start)

	// 使用指数退避，应该需要一些时间
	// 但不要太长（合理性检查）
	if elapsed < time.Microsecond {
		t.Error("WaitFunc 完成太快，退避可能不工作")
	}
	if elapsed > 5*time.Second {
		t.Error("WaitFunc 耗时太长，可能挂起")
	}
}

func TestWaitFunc_Concurrent(t *testing.T) {
	var counter atomic.Int32
	numGoroutines := 10
	done := make(chan struct{}, numGoroutines)

	for range numGoroutines {
		go func() {
			WaitFunc(func() bool {
				current := counter.Add(1)
				return current < 5
			})
			done <- struct{}{}
		}()
	}

	// 等待所有 goroutine
	for range numGoroutines {
		select {
		case <-done:
			// 成功
		case <-time.After(2 * time.Second):
			t.Fatal("WaitFunc goroutine 超时")
		}
	}
}

func TestWaitFor_Zero(t *testing.T) {
	var counter atomic.Int32
	counter.Store(0)

	// 应该立即返回
	done := make(chan struct{})
	go func() {
		WaitFor(&counter)
		close(done)
	}()

	select {
	case <-done:
		// 成功
	case <-time.After(100 * time.Millisecond):
		t.Error("WaitFor 在零计数器上阻塞")
	}
}

func TestWaitFor_Negative(t *testing.T) {
	var counter atomic.Int32
	counter.Store(-5)

	// 对于负值应该立即返回
	done := make(chan struct{})
	go func() {
		WaitFor(&counter)
		close(done)
	}()

	select {
	case <-done:
		// 成功
	case <-time.After(100 * time.Millisecond):
		t.Error("WaitFor 在负计数器上阻塞")
	}
}

func TestWaitFor_Countdown(t *testing.T) {
	var counter atomic.Int32
	counter.Store(5)

	done := make(chan struct{})

	go func() {
		WaitFor(&counter)
		close(done)
	}()

	// 逐渐减少计数器
	for i := 5; i > 0; i-- {
		time.Sleep(10 * time.Millisecond)
		counter.Store(int32(i - 1))
	}

	select {
	case <-done:
		// 成功
	case <-time.After(500 * time.Millisecond):
		t.Error("计数器达到零后 WaitFor 没有完成")
	}
}

func TestWaitFor_MultipleWaiters(t *testing.T) {
	var counter atomic.Int32
	counter.Store(10)

	numWaiters := 5
	done := make(chan struct{}, numWaiters)

	// 启动多个等待者
	for range numWaiters {
		go func() {
			WaitFor(&counter)
			done <- struct{}{}
		}()
	}

	// 在后台减少计数器
	go func() {
		for counter.Load() > 0 {
			time.Sleep(5 * time.Millisecond)
			counter.Add(-1)
		}
	}()

	// 所有等待者应该完成
	for i := range numWaiters {
		select {
		case <-done:
			// 成功
		case <-time.After(2 * time.Second):
			t.Fatalf("等待者 %d 超时", i)
		}
	}
}

func TestWaitFor_RapidChange(t *testing.T) {
	var counter atomic.Int32
	counter.Store(100)

	done := make(chan struct{})

	go func() {
		WaitFor(&counter)
		close(done)
	}()

	// 快速减少计数器
	go func() {
		for counter.Load() > 0 {
			counter.Add(-1)
		}
	}()

	select {
	case <-done:
		// 成功
	case <-time.After(2 * time.Second):
		t.Errorf("WaitFor 超时，计数器 = %d", counter.Load())
	}
}

func TestWaitFor_IncrementThenDecrement(t *testing.T) {
	var counter atomic.Int32
	counter.Store(1)

	done := make(chan struct{})

	go func() {
		WaitFor(&counter)
		close(done)
	}()

	// 先增加（应该继续等待）
	time.Sleep(10 * time.Millisecond)
	counter.Add(5)

	// 然后减少到零
	time.Sleep(10 * time.Millisecond)
	counter.Store(0)

	select {
	case <-done:
		// 成功
	case <-time.After(500 * time.Millisecond):
		t.Error("计数器达到零后 WaitFor 没有完成")
	}
}

func TestWaitFunc_WithTimeout(t *testing.T) {
	timeout := 100 * time.Millisecond
	done := make(chan struct{})

	go func() {
		WaitFunc(func() bool {
			return true // 总是忙碌
		})
		close(done)
	}()

	select {
	case <-done:
		t.Error("WaitFunc 在应该无限等待时完成了")
	case <-time.After(timeout):
		// 预期行为 - 函数应该仍在等待
	}
}

func TestWaitFor_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("在短模式下跳过压力测试")
	}

	var counter atomic.Int32
	counter.Store(1000)

	numWaiters := 50
	done := make(chan struct{}, numWaiters)

	// 启动许多等待者
	for range numWaiters {
		go func() {
			WaitFor(&counter)
			done <- struct{}{}
		}()
	}

	// 减少计数器
	go func() {
		for counter.Load() > 0 {
			counter.Add(-1)
		}
	}()

	// 等待所有
	for i := range numWaiters {
		select {
		case <-done:
			// 成功
		case <-time.After(5 * time.Second):
			t.Fatalf("压力测试在等待者 %d 处超时", i)
		}
	}
}

func BenchmarkWaitFunc_ShortWait(b *testing.B) {
	for b.Loop() {
		counter := 0
		WaitFunc(func() bool {
			counter++
			return counter < 5
		})
	}
}

func BenchmarkWaitFunc_MediumWait(b *testing.B) {
	for b.Loop() {
		counter := 0
		WaitFunc(func() bool {
			counter++
			return counter < 50
		})
	}
}

func BenchmarkWaitFor_ImmediateReturn(b *testing.B) {
	var counter atomic.Int32
	counter.Store(0)

	for b.Loop() {
		WaitFor(&counter)
	}
}

func BenchmarkWaitFor_ShortWait(b *testing.B) {

	for b.Loop() {
		var counter atomic.Int32
		counter.Store(10)

		go func() {
			for counter.Load() > 0 {
				counter.Add(-1)
			}
		}()

		WaitFor(&counter)
	}
}
