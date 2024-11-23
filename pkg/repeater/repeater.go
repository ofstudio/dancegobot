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
	tasks     map[string]context.CancelFunc
	intervals []time.Duration
}

// NewRepeater creates a new task repeater.
func NewRepeater(intervals []time.Duration) *Repeater {
	return &Repeater{
		tasks:     make(map[string]context.CancelFunc),
		intervals: intervals,
	}
}

// AddTask adds a new task to the repeater.
// If the task with the same id already exists, it will be canceled,
// and the new task will be added.
func (r *Repeater) AddTask(ctx context.Context, id string, t TaskFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Cancel the task if it already exists.
	if cancel, ok := r.tasks[id]; ok {
		cancel()
	}

	// Create a new task.
	ctx, cancel := context.WithCancel(ctx)
	r.tasks[id] = cancel

	// Run the task.
	for _, duration := range r.intervals {
		go func(d time.Duration) {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(d):
					t(ctx, id)
					return
				}
			}
		}(duration)
	}
}
