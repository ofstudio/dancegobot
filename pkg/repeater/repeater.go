package repeater

import (
	"context"
	"sync"
	"time"
)

// TaskFunc is a task function.
// It accepts a context and a task id.
type TaskFunc func(context.Context, string)

// Repeater is a task repeater.
type Repeater struct {
	mu        sync.Mutex
	tasks     map[string]*task
	intervals []time.Duration
}

type task struct {
	cancel context.CancelFunc
}

// NewRepeater creates a new task repeater.
func NewRepeater(intervals []time.Duration) *Repeater {
	return &Repeater{
		tasks:     make(map[string]*task),
		intervals: intervals,
	}
}

// AddTask adds a new task to the repeater.
// If the task with the same id already exists, it will be canceled,
// and the new task will be added.
func (r *Repeater) AddTask(ctx context.Context, id string, t TaskFunc) {
	r.mu.Lock()

	// Cancel the task if it already exists.
	if current, ok := r.tasks[id]; ok {
		current.cancel()
	}
	if len(r.intervals) == 0 {
		delete(r.tasks, id)
		r.mu.Unlock()
		return
	}

	// Create a new task.
	ctx, cancel := context.WithCancel(ctx)
	current := &task{cancel: cancel}
	r.tasks[id] = current
	intervals := append([]time.Duration(nil), r.intervals...)
	r.mu.Unlock()

	// Run the task.
	var wg sync.WaitGroup
	wg.Add(len(intervals))
	for _, duration := range intervals {
		go func(d time.Duration) {
			defer wg.Done()
			timer := time.NewTimer(d)
			defer timer.Stop()

			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				t(ctx, id)
			}
		}(duration)
	}
	go func() {
		wg.Wait()
		cancel()

		r.mu.Lock()
		defer r.mu.Unlock()
		if r.tasks[id] == current {
			delete(r.tasks, id)
		}
	}()
}
