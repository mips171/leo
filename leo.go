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
	nodes      map[string]*Node
	startNodes []*Node
}

func TaskGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
	}
}

func (g Graph) Print() {
	for _, node := range g.nodes {
		fmt.Printf("%s -> ", node.name)
		for _, child := range node.children {
			fmt.Printf("%s, ", child.name)
		}
		fmt.Println()
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

// Conditionally takes a node and a condition then executes that node in the graph if the condition is true
func (g *Graph) Conditionally(name string, task TaskFunc, condition bool) {
	if condition {
		g.Add(name, task)
	}
}

// Precede adds a directed edge from node `from` to node `to`
func (g *Graph) Precede(from, to string) error {
	fromNode, fromExists := g.nodes[from]
	toNode, toExists := g.nodes[to]

	if !fromExists || !toExists {
		return errors.New("one or both nodes do not exist")
	}

	fromNode.children = append(fromNode.children, toNode)
	toNode.parents = append(toNode.parents, fromNode)

	if g.hasCycle() {
		fromNode.children = fromNode.children[:len(fromNode.children)-1]
		toNode.parents = toNode.parents[:len(toNode.parents)-1]
		return errors.New("adding this edge would create a cycle")
	}

	return nil
}

// Succeed sets up a "succeeds" relationship, indicating that `to` should succeed `from`.
func (g *Graph) Succeed(from, to string) error {
	return g.Precede(to, from)
}

type Executor struct {
	graph *Graph
}

func NewExecutor(graph *Graph) *Executor {
	parentCounts := make(map[*Node]int)
	for _, node := range graph.nodes {
		for _, child := range node.children {
			parentCounts[child]++
		}
	}

	for node, count := range parentCounts {
		node.parents = make([]*Node, 0, count)
	}

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
	ready := make(chan *Node, len(e.graph.nodes))
	errors := make(chan error, 1)
	finished := make(chan struct{})

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
		close(finished)
	}()

	go func() {
		for node := range ready {
			go func(n *Node) {
				defer wg.Done()
				if err := n.task(); err != nil {
					select {
					case errors <- fmt.Errorf("error executing node %s: %w", n.name, err):
					default:
						// If an error is already recorded, we ignore subsequent errors
					}
					return
				}

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
		return nil
	case err := <-errors:
		return err
	}
}

func (g *Graph) dfsCheckCycle(node *Node, visited, recStack map[*Node]bool) bool {
	if !visited[node] {

		visited[node] = true
		recStack[node] = true

		for _, child := range node.children {
			if !visited[child] && g.dfsCheckCycle(child, visited, recStack) {
				return true
			} else if recStack[child] {
				return true
			}
		}
	}

	recStack[node] = false
	return false
}

// hasCycle checks if there would be a cycle created by adding an edge from `from` to `to`
func (g *Graph) hasCycle() bool {
	visited := make(map[*Node]bool)
	recStack := make(map[*Node]bool)

	for _, node := range g.nodes {
		if !visited[node] {
			if g.dfsCheckCycle(node, visited, recStack) {
				return true
			}
		}
	}

	return false
}
