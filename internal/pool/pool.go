// Package pool provides a simple goroutine pool implementation
package pool

import (
	"sync"
)

// Pool defines the worker pool contract
type Pool interface {
	Submit(task func()) error
	Stop() error
}

type pool struct {
	tasks        chan func()
	wg           sync.WaitGroup
	once         sync.Once
	afterTask    func()
	stopped      chan struct{}
	panicHandler PanicHandler
}

// New creates a new Pool instance with the provided options
func New(opts ...Option) Pool {
	cfg := buildConfig(opts...)
	p := &pool{
		tasks:        make(chan func(), cfg.queueSize),
		stopped:      make(chan struct{}),
		afterTask:    cfg.afterTask,
		panicHandler: cfg.panicHandler,
	}

	for workers := 0; workers < cfg.workers; workers++ {
		go func() {
			for task := range p.tasks {
				func() {
					defer func() {
						if r := recover(); r != nil {
							p.panicHandler("task", r)
						}
						p.wg.Done()
					}()
					task()
				}()
				if p.afterTask != nil {
					func() {
						defer func() {
							if r := recover(); r != nil {
								p.panicHandler("afterTask", r)
							}
						}()
						p.afterTask()
					}()
				}
			}
		}()
	}

	return p
}

// Submit adds a new task to the pool
func (p *pool) Submit(task func()) error {
	if task == nil {
		return ErrNilTask
	}

	select {
	case <-p.stopped:
		return ErrStopped
	default:
	}

	p.wg.Add(1)

	select {
	case p.tasks <- task:
		return nil
	case <-p.stopped:
		p.wg.Done()
		return ErrStopped
	default:
		p.wg.Done()
		return ErrQueueFull
	}
}

// Stop stops the pool and waits for all running tasks to complete
func (p *pool) Stop() error {
	p.once.Do(func() {
		close(p.stopped)
		close(p.tasks)
	})

	p.wg.Wait()

	return nil
}
