package pool

import "errors"

var (
	ErrQueueFull = errors.New("task queue is full!")
	ErrStopped   = errors.New("pool is stopped")
	ErrNilTask   = errors.New("task is nil")
)
