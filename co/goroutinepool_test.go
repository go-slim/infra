package co

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewGoRoutinePool(t *testing.T) {
	numWorkers := 5
	pool := NewGoRoutinePool(numWorkers)

	if pool == nil {
		t.Fatal("NewGoRoutinePool() 返回 nil")
	}
	if pool.work == nil {
		t.Error("work 通道为 nil")
	}
	if pool.sem == nil {
		t.Error("sem 通道为 nil")
	}
	if pool.stop == nil {
		t.Error("stop 通道为 nil")
	}
	if cap(pool.sem) != numWorkers {
		t.Errorf("sem 容量 = %d，期望 %d", cap(pool.sem), numWorkers)
	}
	if cap(pool.stop) != numWorkers {
		t.Errorf("stop 容量 = %d，期望 %d", cap(pool.stop), numWorkers)
	}
}

func TestGoRoutinePool_Schedule(t *testing.T) {
	pool := NewGoRoutinePool(2)
	defer pool.Stop()

	executed := false
	var mu sync.Mutex

	pool.Schedule(func() {
		mu.Lock()
		executed = true
		mu.Unlock()
	})

	// 等待任务执行
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if !executed {
		t.Error("任务未执行")
	}
	mu.Unlock()
}

func TestGoRoutinePool_MultipleSchedule(t *testing.T) {
	pool := NewGoRoutinePool(5)
	defer pool.Stop()

	var counter atomic.Int32
	numTasks := 20

	for range numTasks {
		pool.Schedule(func() {
			counter.Add(1)
		})
	}

	// 等待所有任务完成
	time.Sleep(200 * time.Millisecond)

	if counter.Load() != int32(numTasks) {
		t.Errorf("计数器 = %d，期望 %d", counter.Load(), numTasks)
	}
}

func TestGoRoutinePool_Concurrency(t *testing.T) {
	numWorkers := 3
	pool := NewGoRoutinePool(numWorkers)
	defer pool.Stop()

	var activeCount atomic.Int32
	var maxActive atomic.Int32
	var wg sync.WaitGroup

	numTasks := 10
	wg.Add(numTasks)

	for range numTasks {
		pool.Schedule(func() {
			defer wg.Done()

			// 跟踪活跃的 goroutine
			current := activeCount.Add(1)

			// 必要时更新最大值
			for {
				max := maxActive.Load()
				if current <= max || maxActive.CompareAndSwap(max, current) {
					break
				}
			}

			// 模拟工作
			time.Sleep(50 * time.Millisecond)

			activeCount.Add(-1)
		})
	}

	wg.Wait()

	max := maxActive.Load()
	if max > int32(numWorkers) {
		t.Errorf("最大并发工作线程 = %d，期望 <= %d", max, numWorkers)
	}
	if max < 1 {
		t.Error("未检测到并发执行")
	}
}

func TestGoRoutinePool_Stop(t *testing.T) {
	pool := NewGoRoutinePool(3)

	var counter atomic.Int32

	// 调度一些任务
	for range 5 {
		pool.Schedule(func() {
			counter.Add(1)
			time.Sleep(10 * time.Millisecond)
		})
	}

	// 等待任务开始
	time.Sleep(50 * time.Millisecond)

	// 停止池
	pool.Stop()

	// 验证任务已执行
	if counter.Load() == 0 {
		t.Error("停止前没有任务执行")
	}
}

func TestGoRoutinePool_WorkerReuse(t *testing.T) {
	pool := NewGoRoutinePool(2)
	defer pool.Stop()

	var counter atomic.Int32

	// 顺序调度任务以确保工作线程重用
	for range 10 {
		pool.Schedule(func() {
			counter.Add(1)
		})
		time.Sleep(20 * time.Millisecond)
	}

	if counter.Load() != 10 {
		t.Errorf("计数器 = %d，期望 10", counter.Load())
	}
}

func TestGoRoutinePool_MaxWorkers(t *testing.T) {
	numWorkers := 5
	pool := NewGoRoutinePool(numWorkers)
	defer pool.Stop()

	// 使用原子计数器跟踪并发执行
	var activeCount atomic.Int32
	var maxActive atomic.Int32
	var wg sync.WaitGroup

	// 调度比工作线程更多的任务
	numTasks := numWorkers + 5
	wg.Add(numTasks)

	for range numTasks {
		pool.Schedule(func() {
			defer wg.Done()

			// 跟踪活跃的 goroutine
			current := activeCount.Add(1)

			// 必要时更新最大值
			for {
				max := maxActive.Load()
				if current <= max || maxActive.CompareAndSwap(max, current) {
					break
				}
			}

			// 模拟工作
			time.Sleep(50 * time.Millisecond)

			activeCount.Add(-1)
		})
	}

	wg.Wait()

	max := maxActive.Load()
	// 不应超过 numWorkers
	if max > int32(numWorkers) {
		t.Errorf("最大并发工作线程 = %d，期望 <= %d", max, numWorkers)
	}
}

func TestGoRoutinePool_TaskOrder(t *testing.T) {
	pool := NewGoRoutinePool(1) // 单个工作线程用于顺序执行
	defer pool.Stop()

	results := make([]int, 0)
	var mu sync.Mutex

	for i := range 5 {
		val := i
		pool.Schedule(func() {
			mu.Lock()
			results = append(results, val)
			mu.Unlock()
		})
	}

	// 等待所有任务完成
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if len(results) != 5 {
		t.Errorf("执行了 %d 个任务，期望 5 个", len(results))
	}
	mu.Unlock()
}

func TestGoRoutinePool_EmptyPool(t *testing.T) {
	pool := NewGoRoutinePool(0)
	defer pool.Stop()

	// 使用 0 个工作线程，Schedule 会阻塞，因为：
	// - work 通道是无缓冲的
	// - sem 通道容量为 0
	// 这是预期行为 - 0 个工作线程的池无法执行任务

	executed := false
	scheduleStarted := make(chan bool, 1)

	go func() {
		scheduleStarted <- true
		// 使用 0 个工作线程这会永远阻塞
		pool.Schedule(func() {
			executed = true
		})
	}()

	// 等待 Schedule 被调用
	<-scheduleStarted

	// 给它一点时间可能执行（它不应该）
	time.Sleep(50 * time.Millisecond)

	// 使用 0 个工作线程任务不应该执行
	if executed {
		t.Error("使用 0 个工作线程任务不应该执行")
	}

	// 注意：调用 Schedule 的 goroutine 将保持阻塞，
	// 这是 0 容量池的预期行为
}

func TestGoRoutinePool_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("在短模式下跳过压力测试")
	}

	pool := NewGoRoutinePool(10)
	defer pool.Stop()

	var counter atomic.Int32
	numTasks := 1000
	var wg sync.WaitGroup
	wg.Add(numTasks)

	for range numTasks {
		pool.Schedule(func() {
			defer wg.Done()
			counter.Add(1)
			// 模拟一些工作
			time.Sleep(time.Microsecond)
		})
	}

	wg.Wait()

	if counter.Load() != int32(numTasks) {
		t.Errorf("计数器 = %d，期望 %d", counter.Load(), numTasks)
	}
}

func BenchmarkGoRoutinePool_Schedule(b *testing.B) {
	pool := NewGoRoutinePool(10)
	defer pool.Stop()

	var wg sync.WaitGroup

	for b.Loop() {
		wg.Add(1)
		pool.Schedule(func() {
			wg.Done()
		})
	}

	wg.Wait()
}

func BenchmarkGoRoutinePool_HighContention(b *testing.B) {
	pool := NewGoRoutinePool(4)
	defer pool.Stop()

	var counter atomic.Int32

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			done := make(chan struct{})
			pool.Schedule(func() {
				counter.Add(1)
				close(done)
			})
			<-done
		}
	})
}
