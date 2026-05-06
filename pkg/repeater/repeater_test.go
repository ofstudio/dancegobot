package repeater

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func ExampleRepeater_AddTask() {
	// Create a new repeater with intervals.
	r := NewRepeater([]time.Duration{
		200 * time.Millisecond,
		400 * time.Millisecond,
	})

	// Add a new task.
	r.AddTask(context.Background(), "task1", func(ctx context.Context, id string) {
		fmt.Println("This is task with id:", id)
	})

	// Wait for the task to finish.
	time.Sleep(time.Second * 1)

	// Output:
	// This is task with id: task1
	// This is task with id: task1
}

func ExampleRepeater_AddTask_withDuplicates() {
	// Create a new repeater with intervals.
	r := NewRepeater([]time.Duration{
		200 * time.Millisecond,
		400 * time.Millisecond,
		600 * time.Millisecond,
	})

	// Add a new task.
	r.AddTask(context.Background(), "task1", func(ctx context.Context, id string) {
		fmt.Println("First run of task with id:", id)
	})

	// First run of the task should be executed 2 times.
	time.Sleep(500 * time.Millisecond)

	// Add a new task with the same id.
	r.AddTask(context.Background(), "task1", func(ctx context.Context, id string) {
		fmt.Println("Second run of task with id:", id)
	})

	// Wait for the task to finish.
	time.Sleep(time.Second * 1)

	// Output:
	// First run of task with id: task1
	// First run of task with id: task1
	// Second run of task with id: task1
	// Second run of task with id: task1
	// Second run of task with id: task1
}

func TestRepeaterCleansUpCompletedTask(t *testing.T) {
	r := NewRepeater([]time.Duration{
		5 * time.Millisecond,
		10 * time.Millisecond,
	})

	var calls atomic.Int64
	var wrongID atomic.Bool
	r.AddTask(context.Background(), "task1", func(ctx context.Context, id string) {
		if id != "task1" {
			wrongID.Store(true)
		}
		calls.Add(1)
	})

	require.Eventually(t, func() bool {
		return calls.Load() == 2 && taskCount(r) == 0
	}, time.Second, 10*time.Millisecond)
	require.False(t, wrongID.Load())
}

func TestRepeaterCleansUpEmptyIntervalTask(t *testing.T) {
	r := NewRepeater(nil)

	var called atomic.Bool
	r.AddTask(context.Background(), "task1", func(ctx context.Context, id string) {
		called.Store(true)
	})

	require.Equal(t, 0, taskCount(r))
	require.False(t, called.Load())
}

func TestRepeaterCleansUpCanceledTask(t *testing.T) {
	r := NewRepeater([]time.Duration{time.Hour})
	ctx, cancel := context.WithCancel(context.Background())

	var called atomic.Bool
	r.AddTask(ctx, "task1", func(ctx context.Context, id string) {
		called.Store(true)
	})
	require.Equal(t, 1, taskCount(r))

	cancel()

	require.Eventually(t, func() bool {
		return taskCount(r) == 0
	}, time.Second, 10*time.Millisecond)
	require.False(t, called.Load())
}

func TestRepeaterCanceledTaskDoesNotDeleteReplacement(t *testing.T) {
	r := NewRepeater([]time.Duration{150 * time.Millisecond})

	var calls atomic.Int64
	r.AddTask(context.Background(), "task1", func(ctx context.Context, id string) {
		calls.Add(100)
	})
	r.AddTask(context.Background(), "task1", func(ctx context.Context, id string) {
		calls.Add(1)
	})

	require.Never(t, func() bool {
		return taskCount(r) == 0
	}, 50*time.Millisecond, 5*time.Millisecond)

	require.Eventually(t, func() bool {
		return calls.Load() == 1 && taskCount(r) == 0
	}, time.Second, 10*time.Millisecond)
}

func taskCount(r *Repeater) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.tasks)
}
