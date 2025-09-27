package pool

import (
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

}

func (p *pool) Submit(task func()) error {

}

func (p *pool) Stop() error {

}
