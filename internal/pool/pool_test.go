package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSubmit(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() Pool
		task    func()
		wantErr error
	}{
		{
			name: "ok",
			setup: func() Pool {
				return New(WithWorkers(1), WithQueueSize(1))
			},
			task:    func() {},
			wantErr: nil,
		},
		{
			name: "nil_task",
			setup: func() Pool {
				return New(WithWorkers(1))
			},
			task:    nil,
			wantErr: ErrNilTask,
		},
		{
			name: "stopped",
			setup: func() Pool {
				p := New(WithWorkers(0))
				_ = p.Stop()
				return p
			},
			task:    func() {},
			wantErr: ErrStopped,
		},
		{
			name: "queue_full",
			setup: func() Pool {
				return New(WithWorkers(1), WithQueueSize(1))
			},
			task:    func() {},
			wantErr: ErrQueueFull,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
			defer func() { _ = p.Stop() }()

			if tt.name == "queue_full" {
				block := make(chan struct{})
				started := make(chan struct{})

				if err := p.Submit(func() {
					close(started)
					<-block
				}); err != nil {
					t.Fatalf("unexpected first submit err: %v", err)
				}

				<-started

				if err := p.Submit(func() {}); err != nil {
					t.Fatalf("unexpected second submit err: %v", err)
				}

				err := p.Submit(func() {})
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}

				close(block)
				return
			}

			err := p.Submit(tt.task)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestStop(t *testing.T) {
	tests := []struct {
		name string
		test func(_ *testing.T, p Pool)
	}{
		{
			name: "waits_for_tasks",
			test: func(_ *testing.T, p Pool) {
				done := make(chan struct{})
				_ = p.Submit(func() {
					time.Sleep(100 * time.Millisecond)
					close(done)
				})
				start := time.Now()
				_ = p.Stop()
				if time.Since(start) < 100*time.Millisecond {
					t.Fatalf("Stop() returned too early")
				}
				select {
				case <-done:
				default:
					t.Fatalf("task did not finish")
				}
			},
		},
		{
			name: "idempotent",
			test: func(_ *testing.T, p Pool) {
				_ = p.Stop()
				_ = p.Stop()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			p := New(WithWorkers(1))
			tt.test(t, p)
		})
	}
}

func TestAfterTask(t *testing.T) {
	tests := []struct {
		name    string
		afterFn func(counter *int32) func()
		submits int
		want    int32
	}{
		{
			name: "called_correct_times",
			afterFn: func(c *int32) func() {
				return func() { atomic.AddInt32(c, 1) }
			},
			submits: 2,
			want:    2,
		},
		{
			name: "panic_safe",
			afterFn: func(c *int32) func() {
				return func() {
					atomic.AddInt32(c, 1)
					panic("boom")
				}
			},
			submits: 2,
			want:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var count int32
			wg := sync.WaitGroup{}
			after := func() {
				defer wg.Done()
				tt.afterFn(&count)()
			}

			wg.Add(tt.submits)
			p := New(WithWorkers(1), WithQueueSize(tt.submits), WithAfterTask(after))
			for i := 0; i < tt.submits; i++ {
				if err := p.Submit(func() {}); err != nil {
					t.Fatalf("submit err: %v", err)
				}
			}
			_ = p.Stop()

			done := make(chan struct{})
			go func() { wg.Wait(); close(done) }()

			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("timeout waiting for afterTask")
			}

			if got := atomic.LoadInt32(&count); got != tt.want {
				t.Fatalf("afterTask count = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPanicHandler(t *testing.T) {
	tests := []struct {
		name     string
		task     func()
		after    func()
		expected []string
	}{
		{
			name:     "task_panic",
			task:     func() { panic("task") },
			after:    nil,
			expected: []string{"task"},
		},
		{
			name:     "afterTask_panic",
			task:     func() {},
			after:    func() { panic("after") },
			expected: []string{"afterTask"},
		},
		{
			name:     "both_panics",
			task:     func() { panic("task") },
			after:    func() { panic("after") },
			expected: []string{"task", "afterTask"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			events := []string{}
			handler := func(where string, _ any) {
				mu.Lock()
				events = append(events, where)
				mu.Unlock()
			}

			wg := sync.WaitGroup{}
			if tt.after != nil {
				wg.Add(1)
				after := func() {
					defer wg.Done()
					tt.after()
				}
				p := New(
					WithWorkers(1),
					WithQueueSize(1),
					WithPanicHandler(handler),
					WithAfterTask(after),
				)
				_ = p.Submit(tt.task)
				_ = p.Stop()

				done := make(chan struct{})
				go func() { wg.Wait(); close(done) }()

				select {
				case <-done:
				case <-time.After(time.Second):
					t.Fatal("timeout waiting for afterTask panic")
				}
			} else {
				p := New(
					WithWorkers(1),
					WithQueueSize(1),
					WithPanicHandler(handler),
				)
				_ = p.Submit(tt.task)
				_ = p.Stop()
			}

			mu.Lock()
			defer mu.Unlock()
			for _, want := range tt.expected {
				found := false
				for _, got := range events {
					if got == want {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected event %q not found, got %v", want, events)
				}
			}
		})
	}
}
