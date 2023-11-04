package leo

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestAddNode(t *testing.T) {
    graph := TaskGraph()

    graph.Add("A", func() error { return nil })

    if _, exists := graph.nodes["A"]; !exists {
        t.Errorf("AddNode failed to add node 'A'")
    }
}

func TestPrecede(t *testing.T) {
    graph := TaskGraph()

    graph.Add("A", func() error { return nil })
    graph.Add("B", func() error { return nil })
    graph.Add("C", func() error { return nil })

    if err := graph.Precede("A", "B"); err != nil {
        t.Errorf("Precede failed to add edge from 'A' to 'B': %v", err)
    }

    if err := graph.Precede("B", "C"); err != nil {
        t.Errorf("Precede failed to add edge from 'B' to 'C': %v", err)
    }

    // This should create a cycle and hence should fail
    if err := graph.Precede("C", "A"); err == nil {
        t.Errorf("%v, Precede should have detected a cycle when adding edge from 'C' to 'A'", err)
    }
}

// TestSucceed checks if edges are added correctly for the Succeed function.
func TestSucceed(t *testing.T) {
    graph := TaskGraph()

    graph.Add("A", func() error { return nil })
    graph.Add("B", func() error { return nil })
    graph.Add("C", func() error { return nil })

    if err := graph.Succeed("B", "A"); err != nil {
        t.Errorf("Succeed failed to add edge from 'B' to 'A': %v", err)
    }

    if err := graph.Succeed("C", "B"); err != nil {
        t.Errorf("Succeed failed to add edge from 'C' to 'B': %v", err)
    }

    // This should create a cycle because it closes the cycle A -> B -> C -> A
    if err := graph.Succeed("A", "C"); err == nil {
        t.Errorf("Succeed should have detected a cycle when adding edge from 'A' to 'C' to close the cycle")
    }

    // This should not create a cycle and should be allowed
    if err := graph.Succeed("C", "A"); err != nil {
        t.Errorf("Succeed should not have detected a cycle when adding edge from 'C' to 'A': %v", err)
    }
}

func TestExecutorExecute(t *testing.T) {
    graph := TaskGraph()

    executedNodes := make(map[string]bool)

    graph.Add("A", func() error {
        executedNodes["A"] = true
        return nil
    })
    graph.Add("B", func() error {
        if !executedNodes["A"] {
            return errors.New("node 'A' should have executed before 'B'")
        }
        executedNodes["B"] = true
        return nil
    })
    graph.Add("C", func() error {
        if !executedNodes["B"] {
            return errors.New("node 'B' should have executed before 'C'")
        }
        executedNodes["C"] = true
        return nil
    })

    graph.Precede("A", "B")
    graph.Precede("B", "C")

    executor := NewExecutor(graph)

    if err := executor.Execute(); err != nil {
        t.Errorf("Execute failed: %v", err)
    }

    for _, node := range []string{"A", "B", "C"} {
        if !executedNodes[node] {
            t.Errorf("node '%s' did not execute", node)
        }
    }
}

func TestDAGExecution(t *testing.T) {
    graph := TaskGraph()

    var executionOrder []string
    var orderLock sync.Mutex

    // Helper function to record execution order
    recordExecution := func(name string) {
        orderLock.Lock()
        defer orderLock.Unlock()
        executionOrder = append(executionOrder, name)
    }

    // Define tasks
    graph.Add("A", func() error {
        recordExecution("A")
        time.Sleep(100 * time.Millisecond)
        return nil
    })
    graph.Add("B", func() error {
        recordExecution("B")
        time.Sleep(100 * time.Millisecond)
        return nil
    })
    graph.Add("C", func() error {
        recordExecution("C")
        time.Sleep(100 * time.Millisecond)
        return nil
    })
    graph.Add("D", func() error {
        recordExecution("D")
        time.Sleep(100 * time.Millisecond)
        return nil
    })

    // Set up dependencies
    err := graph.Precede("A", "B")
    if err != nil {
        t.Fatalf("Failed to add edge: %s", err)
    }
    err = graph.Precede("A", "C")
    if err != nil {
        t.Fatalf("Failed to add edge: %s", err)
    }
    err = graph.Succeed("D", "B")
    if err != nil {
        t.Fatalf("Failed to add edge: %s", err)
    }
    err = graph.Succeed("D", "C")
    if err != nil {
        t.Fatalf("Failed to add edge: %s", err)
    }

    // Execute graph
    executor := NewExecutor(graph)

    if err := executor.Execute(); err != nil {
        t.Fatalf("Execution failed: %v", err)
    }

    fmt.Println("Execution Order:", executionOrder)

    // A must come before B and C, and D must come after B and C
    if !(indexOf(executionOrder, "A") < indexOf(executionOrder, "B") &&
        indexOf(executionOrder, "A") < indexOf(executionOrder, "C") &&
        indexOf(executionOrder, "B") < indexOf(executionOrder, "D") &&
        indexOf(executionOrder, "C") < indexOf(executionOrder, "D")) {
        t.Errorf("Execution order does not match expected dependencies")
    }
}

func indexOf(slice []string, val string) int {
    for i, item := range slice {
        if item == val {
            return i
        }
    }
    return -1
}
