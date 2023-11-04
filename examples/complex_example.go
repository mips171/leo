package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"

	"github.com/mips171/leo"
)

func complex_example() {
	cpuProfile, err := os.Create("cpu_profile.prof")
	if err != nil {
		fmt.Printf("Could not create CPU profile: %v\n", err)
		return
	}
	defer cpuProfile.Close()

	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		fmt.Printf("Could not start CPU profile: %v\n", err)
		return
	}
	defer pprof.StopCPUProfile()

	graph := leo.TaskGraph()

	// Create a more complex graph with additional nodes and edges
	nodes := []string{"A", "B", "C", "D"}

	// Create a task function that logs with timestamps
	taskFunc := func(name string) leo.TaskFunc {
		return func() error {
			startTime := time.Now()
			fmt.Printf("[%s] Starting execution of %s\n", startTime.Format(time.RFC3339Nano), name)
			// Random sleep to simulate work
			sleepTime := rand.Intn(1000)
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)
			endTime := time.Now()
			fmt.Printf("[%s] Finished executing %s (took %s)\n", endTime.Format(time.RFC3339Nano), name, endTime.Sub(startTime))
			return nil
		}
	}


	// Add nodes to the graph
	for _, nodeName := range nodes {
		graph.Add(nodeName, taskFunc(nodeName))
	}

	// Simulate more complex dependencies
	graph.Precede("A", "B")
	graph.Precede("A", "C")
	graph.Succeed("D", "B")
	graph.Succeed("D", "C")


	// Create an executor
	executor := leo.NewExecutor(graph)

	// Execute the graph
	fmt.Println("Executing graph with time stamps...")
	if err := executor.Execute(); err != nil {
		fmt.Printf("Execution failed: %v\n", err)
	}
}
