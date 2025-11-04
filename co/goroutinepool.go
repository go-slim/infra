package co

// Task 表示要由 goroutine 池执行的工作任务。
// 它是一个没有参数和返回值的函数类型。
type Task func()

// GoRoutinePool 是一个带有信号量机制的 goroutine 池，
// 用于控制并发。它通过重用 goroutine 和限制最大并发工作线程数量
// 来提供高效的任务调度。
type GoRoutinePool struct {
	// work 是任务队列通道，待处理的任务在这里排队。
	work chan Task
	// sem 是用于控制活跃 goroutine 数量的信号量通道。
	// 通道容量定义了最大并发工作线程数量。
	sem chan struct{}
	// stop 是用于优雅关闭工作 goroutine 的信号通道。
	stop chan struct{}
}

// NewGoRoutinePool 分配并初始化一个指定工作线程数量的新 goroutine 池。
// numWorkers 参数设置可以并发运行的 goroutine 最大数量。
func NewGoRoutinePool(numWorkers int) *GoRoutinePool {
	return &GoRoutinePool{
		work: make(chan Task),
		sem:  make(chan struct{}, numWorkers),
		stop: make(chan struct{}, numWorkers),
	}
}

// Schedule 将任务排队到 goroutine 池执行。
// 它首先尝试将任务发送给现有的空闲工作线程。如果没有空闲工作线程
// 且池尚未达到工作线程限制，它会生成一个新的工作线程 goroutine 来处理任务。
func (p *GoRoutinePool) Schedule(task Task) {
	select {
	case p.work <- task:
		// 有空闲工作线程可用，将任务发送给它
	case p.sem <- struct{}{}:
		// 没有空闲工作线程但还有容量，生成新的工作线程
		go p.worker(task)
	}
}

// Stop 向池中所有运行的 goroutine 发送停止信号。
// 通过向每个工作线程发送退出信号来启动优雅关闭，
// 工作线程在完成当前任务后退出。
func (p *GoRoutinePool) Stop() {
	numWorkers := cap(p.sem)
	for range numWorkers {
		p.stop <- struct{}{}
	}
}

// worker 是连续处理任务的主要 goroutine 函数。
// 每个工作线程在循环中执行任务，直到收到停止信号。
// 工作线程退出时释放其信号量槽位。
func (p *GoRoutinePool) worker(task Task) {
	// 工作线程退出时释放信号量槽位
	defer func() { <-p.sem }()
	for {
		// 执行当前任务
		task()
		// 等待下一个任务或停止信号
		select {
		case task = <-p.work:
			// 收到新任务，继续循环
		case <-p.stop:
			// 收到停止信号，退出工作线程
			return
		}
	}
}
