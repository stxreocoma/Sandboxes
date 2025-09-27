package pool

import "runtime"

// Option configures the Pool at construction time
type Option func(*config)

type config struct {
	workers      int
	queueSize    int
	afterTask    func()
	panicHandler PanicHandler
}

// PanicHandler handles recovered panics from tasks or hooks
type PanicHandler func(where string, recovered any)

func defaultConfig() config {
	return config{
		workers:      runtime.GOMAXPROCS(0),
		queueSize:    64,
		afterTask:    nil,
		panicHandler: func(_ string, _ any) {},
	}
}

func buildConfig(opts ...Option) config {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	if cfg.workers <= 0 {
		cfg.workers = 1
	}
	if cfg.queueSize < 0 {
		cfg.queueSize = 0
	}

	return cfg
}

// WithWorkers sets the number of worker goroutines
func WithWorkers(n int) Option {
	return func(c *config) {
		c.workers = n
	}
}

// WithQueueSize sets the capacity of the task queue
func WithQueueSize(n int) Option {
	return func(c *config) {
		c.queueSize = n
	}
}

// WithAfterTask sets a hook called after each task execution
func WithAfterTask(f func()) Option {
	return func(c *config) {
		c.afterTask = f
	}
}

// WithPanicHandler sets a handler for recovered panics
func WithPanicHandler(h PanicHandler) Option {
	return func(c *config) {
		c.panicHandler = h
	}
}
