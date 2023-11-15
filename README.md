# Leo - Go Concurrency in 4 easy pieces
[![Go Report Card](https://goreportcard.com/badge/github.com/mips171/leo)](https://goreportcard.com/report/github.com/mips171/leo)

API inspired by [TaskFlow](https://github.com/taskflow/taskflow)

![Leo Cheers](https://imgflip.com/s/meme/Leonardo-Dicaprio-Cheers.jpg)

## Install
Run `go get github.com/mips171/leo`

Or just import it and run `go mod tidy`

## Example usage
See ./examples for more.
```go
import "github.com/mips171/leo"

// Step 1: Initialise the graph to put your tasks/functions in
tasks := leo.TaskGraph()

// Helpers for our example that simulates work by sleeping for a random duration
// but this can be any interface.
taskFunc := func(name string) leo.TaskFunc {
    return func() error {
        fmt.Printf("Executing task %s\n", name)
        // Simulate work with a random sleep
        time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
        fmt.Printf("Completed task %s\n", name)
        return nil
    }
}

// Step 2: Add tasks to the graph. You need to give it a name because I don't know a good way in Go to get __func__
// Each task is just a function that prints its name and sleeps.
tasks.Add("Task A", taskFunc("Task A"))
tasks.Add("Task B", taskFunc("Task B"))
tasks.Add("Task C", taskFunc("Task C"))
tasks.Add("Task D", taskFunc("Task D"))

// Step 3: Establish dependencies: Task A must precede Task B and Task C. Task D succeeds task C. Cycles are an error.
// This means that Task B and Task C will run concurrently after Task A completes,
// and Task D will run after both Task B and Task C complete.
tasks.Precede("Task A", "Task B") // A runs before B
tasks.Precede("Task A", "Task C") // A also runs before C
tasks.Succeed("Task D", "Task B") // D runs after C
tasks.Succeed("Task D", "Task C") // D also runs after C

// Step 4: Create an executor to run the tasks
executor := leo.NewExecutor(tasks)

fmt.Println("Executing graph in a loop...")
for i := 0; i < 3; i++ {
    // Execute the graph. 
    // This will run Task A first, then Task B and Task C concurrently, then Task D once C completes, even if B has not yet finished.
    if err := executor.Execute(); err != nil {
        fmt.Printf("Execution failed: %v\n", err)
    } else {
        fmt.Println("All tasks executed successfully.")
    }
}
```

### Output
You'll see something like this 3 times. Each time the order of B and C could be different because you left it up to the CPU.
```
Executing task Task A
Completed task Task A
Executing task Task C
Executing task Task B
Completed task Task B
Completed task Task C
Executing task Task D
Completed task Task D
All tasks executed successfully.
```

## Lore
I have dealt with dependency resolution problems in the past, and at the time wanted an easy way to just set up my tasks and run them indefinitely as a service, letting the software handle scheduling as defined by me but interleaving tasks when possible for maximum concurrency. For example, updating firmware on live systems normally requires things to be done in a certain order. I saw that TaskFlow could do that easily, and immediately after watching their CppCon talk and demo I knew I had to implement it in Go. I have not looked, there may already be something else out there that does this, but Leo was designed to be simple and what I need. If it's useful for you too, please consider giving it a star, PR or a mention!

## License
Leo is licensed under the Apache 2.0 license as found in the LICENSE file.
