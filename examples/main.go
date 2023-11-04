package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mips171/leo"
)

func main() {

	// Step 1. Create a new task tasksGraph
	tasksGraph := leo.TaskGraph()

	// Define a simple task function that simulates work by sleeping for a random duration
	taskFunc := func(name string) leo.TaskFunc {
		return func() error {
			fmt.Printf("Executing task %s\n", name)
			// Simulate work with a random sleep
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			fmt.Printf("Completed task %s\n", name)
			return nil
		}
	}

	// Step 2. Add tasks to the graph. Each task is just a function that prints its name and sleeps.
	tasksGraph.Add("Task A", taskFunc("Task A"))
	tasksGraph.Add("Task B", taskFunc("Task B"))
	tasksGraph.Add("Task C", taskFunc("Task C"))
	tasksGraph.Add("Task D", taskFunc("Task D"))

	// Step 3. Establish dependencies: Task A must precede Task B and Task C. Task D succeeds task B and C.
	// This means that Task B and Task C will run in parallel after Task A completes,
	// and Task D will run after both Task B and Task C complete.
	tasksGraph.Precede("Task A", "Task B") // A runs before B
	tasksGraph.Precede("Task A", "Task C") // A also runs before C
	tasksGraph.Succeed("Task D", "Task B") // D runs after B
	tasksGraph.Succeed("Task D", "Task C") // D also runs after C

	tasksGraph.Print()

	// Step 4. Create an executor to run the graph
	executor := leo.NewExecutor(tasksGraph)
	fmt.Println("Executing graph in a loop...")
	for i := 0; i < 3; i++ {
		// Step 5. Execute the graph. This will run Task A first, then Task B and Task C in parallel, then task D after B and C
		if err := executor.Execute(); err != nil {
			fmt.Printf("Execution failed: %v\n", err)
		} else {
			fmt.Println("All tasks executed successfully.")
		}
	}
}
