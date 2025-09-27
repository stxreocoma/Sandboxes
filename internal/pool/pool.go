package pool

import (
	"log"
	"sync"
)

type Pool interface {
	Submit(task func()) error
	Stop() error
}

type pool struct {
	tasks     chan func()
	wg        sync.WaitGroup
	once      sync.Once
	afterTask func()
	stopped   chan struct{}
}

func New(workerCount, taskQueueSize int, afterTask func()) Pool {
	p := &pool{
		tasks:     make(chan func(), taskQueueSize),
		stopped:   make(chan struct{}),
		afterTask: afterTask,
	}

	for workers := 0; workers < workerCount; workers++ {
		go func() {
			for task := range p.tasks {
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Println("Error while executing task in pool:", r)
						}
						p.wg.Done()
					}()
					task()
				}()
				if p.afterTask != nil {
					func() {
						defer func() {
							if r := recover(); r != nil {
								log.Println("Error while executing afterfunc in pool:", r)
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

func (p *pool) Submit(task func()) error {

}

func (p *pool) Stop() error {

}
