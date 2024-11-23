package repeater

import (
	"context"
	"fmt"
	"time"
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
