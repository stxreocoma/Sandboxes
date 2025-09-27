package pool

import "errors"

var (
	// ErrQueueFull is returned when the task queue has no free capacity
	ErrQueueFull = errors.New("task queue is full")
	// ErrStopped is returned when submitting a task to a stopped pool
	ErrStopped = errors.New("pool is stopped")
	// ErrNilTask is returned when a nil task is submitted
	ErrNilTask = errors.New("task is nil")
)
