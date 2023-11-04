package leo

import (
	"errors"
	"fmt"
	"sync"
)

type TaskFunc func() error

type Node struct {
	task     TaskFunc
	children []*Node
	parents  []*Node
	name     string
}

type Graph struct {
	nodes     map[string]*Node
	startNodes []*Node
}

func TaskGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
	}
}

func (g *Graph) Add(name string, task TaskFunc) {
	if _, exists := g.nodes[name]; !exists {
		g.nodes[name] = &Node{
			task:     task,
			children: make([]*Node, 0),
			parents:  make([]*Node, 0),
			name:     name,
		}
		g.startNodes = append(g.startNodes, g.nodes[name])
	}
}

// Precede adds a directed edge from node `from` to node `to`
func (g *Graph) Precede(from, to string) error {
	fromNode, fromExists := g.nodes[from]
	toNode, toExists := g.nodes[to]

	if !fromExists || !toExists {
		return errors.New("one or both nodes do not exist")
	}

	// Tentatively add the edge
	fromNode.children = append(fromNode.children, toNode)
	toNode.parents = append(toNode.parents, fromNode)

	// Check if adding the edge would create a cycle
	if g.hasCycle() {
		// If there is a cycle, remove the edge
		fromNode.children = fromNode.children[:len(fromNode.children)-1]
		toNode.parents = toNode.parents[:len(toNode.parents)-1]
		return errors.New("adding this edge would create a cycle")
	}

	// If no cycle is found, the edge is successfully added.
	return nil
}

// Succeed sets up a "succeeds" relationship, indicating that `to` should succeed `from`.
func (g *Graph) Succeed(from, to string) error {
    // `to` should succeed `from`, so the edge is from `to` to `from`
    return g.Precede(to, from)
}

// hasCycle checks if there would be a cycle created by adding an edge from `from` to `to`
func (g *Graph) hasCycle() bool {
    visited := make(map[*Node]bool)
    recStack := make(map[*Node]bool)

    // Check each node in the graph
    for _, node := range g.nodes {
        if !visited[node] {
            if g.dfsCheckCycle(node, visited, recStack) {
                return true
            }
        }
    }

    return false
}

func (g *Graph) dfsCheckCycle(node *Node, visited, recStack map[*Node]bool) bool {
    if !visited[node] {
        // Mark the current node as visited and part of recursion stack
        visited[node] = true
        recStack[node] = true

        // Recur for all the children (neighbors)
        for _, child := range node.children {
            if !visited[child] && g.dfsCheckCycle(child, visited, recStack) {
                return true // If a child is visited and in recStack then graph is cyclic
            } else if recStack[child] {
                return true // A recursion stack entry means there's a cycle
            }
        }
    }

    // Remove the vertex from recursion stack
    recStack[node] = false
    return false
}


type Executor struct {
	graph *Graph
}

func NewExecutor(graph *Graph) *Executor {
    // Precompute the number of children for each node to efficiently allocate parents slices.
    parentCounts := make(map[*Node]int)
    for _, node := range graph.nodes {
        for _, child := range node.children {
            parentCounts[child]++
        }
    }

    // Allocate the parents slice for each node with exact capacity to avoid reallocation.
    for node, count := range parentCounts {
        node.parents = make([]*Node, 0, count)
    }

    // Fill the parents slice now that it is preallocated with sufficient capacity.
    for _, node := range graph.nodes {
        for _, child := range node.children {
            child.parents = append(child.parents, node)
        }
    }

    return &Executor{
        graph: graph,
    }
}
func (e *Executor) Execute() error {
	var wg sync.WaitGroup
	inDegree := make(map[*Node]int)
	ready := make(chan *Node, len(e.graph.nodes)) // Buffered channel
	errors := make(chan error, 1) // A single buffered channel is sufficient
	finished := make(chan struct{}) // Signal that execution is done

	// Initialize inDegree map
	for _, node := range e.graph.nodes {
		inDegree[node] = len(node.parents)
		if inDegree[node] == 0 {
			wg.Add(1)
			go func(n *Node) {
				ready <- n
			}(node)
		}
	}

	go func() {
		wg.Wait()
		close(finished) // Close finished when all tasks are done
	}()

	go func() {
		for node := range ready {
			go func(n *Node) {
				defer wg.Done() // Ensure that Done is called when the goroutine finishes
				if err := n.task(); err != nil {
					select {
					case errors <- fmt.Errorf("error executing node %s: %w", n.name, err):
						// Non-blocking send to errors channel
					default:
						// If an error is already recorded, we ignore subsequent errors
					}
					return
				}
				// Signal ready for child nodes if they are ready
				for _, child := range n.children {
					inDegree[child]--
					if inDegree[child] == 0 {
						wg.Add(1)
						ready <- child
					}
				}
			}(node)
		}
	}()

	select {
	case <-finished:
		// Execution finished without error
		return nil
	case err := <-errors:
		// Return the first error that was encountered
		return err
	}
}
