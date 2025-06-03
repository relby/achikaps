package graph

import (
	"fmt"
	"math"

	"github.com/dominikbraun/graph"
	"github.com/relby/achikaps/assert"
	"github.com/relby/achikaps/node"
	"github.com/relby/achikaps/vec2"
)

var (
	ErrEdgeNotFound = graph.ErrEdgeNotFound
	ErrEdgeAlreadyExists = graph.ErrEdgeAlreadyExists
)

type Edge = graph.Edge[node.ID]

type Graph struct {
	g graph.Graph[node.ID, *node.Node]
}

func New(root *node.Node) *Graph {
	g := graph.New(func(v *node.Node) node.ID { return v.ID })

	// We can safely assert this error because this is a first node and this method can't throw any error
	err := g.AddVertex(root)
	assert.NoError(err)

	return &Graph{g}
}

func (g *Graph) NodeCount() int {
	count, err := g.g.Order()
	assert.NoError(err)
	
	return count
}

func (g *Graph) Node(id node.ID) (*node.Node, error) {
	n, err := g.g.Vertex(id)
	if err != nil {
		return nil, fmt.Errorf("can't get vertex of a graph: %w", err)
	}
	
	return n, nil
}

func (g *Graph) Nodes() []*node.Node {
	am := g.AdjacencyMap()
	
	out := make([]*node.Node, 0, len(am))
	for nID := range am {
		n, err := g.Node(nID)
		assert.NoError(err)
		
		out = append(out, n)
	}
	
	return out
}

func (g *Graph) AdjacencyMap() map[node.ID]map[node.ID]Edge {
	am, err := g.g.AdjacencyMap()
	assert.NoError(err)

	return am
}

func (g *Graph) AddNodeFrom(n1, n2 *node.Node) (error) {
	if err := g.g.AddVertex(n2); err != nil {
		return fmt.Errorf("can't add vertex to the graph: %w", err)
	}
	if err := g.g.AddEdge(n1.ID, n2.ID); err != nil {
		return fmt.Errorf("can't add edge to the graph: %w", err)
	}
	
	return nil
}

func (g *Graph) NodesByType(typ node.Type) []*node.Node {
	ns := g.Nodes()
	
	out := make([]*node.Node, 0, len(ns))

	for _, n := range ns {
		if n.Type == typ {
			out = append(out, n)
		}
	}
	
	return out
}

func (g *Graph) BuildingNodes() []*node.Node {
	ns := g.Nodes()
	
	out := make([]*node.Node, 0, len(ns))

	for _, n := range ns {
		if !n.IsBuilt() {
			out = append(out, n)
		}
	}
	
	return out
}

// FindShortestPath computes the shortest path from source node to target node
// using Dijkstra's algorithm. It returns a slice of node IDs representing the path
// and the total distance of the path.
func (g *Graph) FindShortestPath(source, target *node.Node) []*node.Node {
	ids, err := graph.ShortestPath(g.g, source.ID, target.ID)
	assert.NoError(err)
	
	out := make([]*node.Node, 0, len(ids))
	for _, id := range ids {
		n, err := g.Node(id)
		assert.NoError(err)
		
		out = append(out, n)
	}
	
	return out
}

// NodeIntersectsAny checks if the given node intersects with any existing nodes or edges in the graph.
// It returns true if an intersection is found, false otherwise.
// The function performs three types of intersection checks:
// 1. Node-to-node intersection with source nodes of edges
// 2. Node-to-node intersection with target nodes of edges
// 3. Node-to-edge intersection by calculating the minimum distance between the node and each edge
func (g *Graph) NodeIntersectsAny(n *node.Node) bool {
	edges, err := g.g.Edges()
	assert.NoError(err)
	
	// Check if the node intersects with any existing nodes or edges in the graph
	for _, edge := range edges {
		// Get the source node of the current edge
		sourceNode, err := g.Node(edge.Source)
		assert.NoError(err)
		
		// Check if the node intersects with the source node (excluding self-intersection)
		if sourceNode.ID != n.ID && sourceNode.Intersects(n) {
			return true
		}

		// Get the target node of the current edge
		targetNode, err := g.Node(edge.Target)
		assert.NoError(err)

		// Check if the node intersects with the target node (excluding self-intersection)
		if targetNode.ID != n.ID && targetNode.Intersects(n) {
			return true
		}
		
		// Check if the node intersects with the edge itself
		// Get positions for calculations
		p := n.Position   // Position of the node we're checking
		p1 := sourceNode.Position  // Start point of the edge
		p2 := targetNode.Position  // End point of the edge
		
		// Calculate vectors for projection
		// Vector from edge start to end
		edgeVec := vec2.New(p2.X - p1.X, p2.Y - p1.Y)
		// Vector from edge start to the node position
		pointVec := vec2.New(p.X - p1.X, p.Y - p1.Y)
		
		// Calculate squared length of the edge for normalization
		edgeLengthSq := edgeVec.X*edgeVec.X + edgeVec.Y*edgeVec.Y
		
		// Calculate normalized projection parameter (clamped between 0 and 1)
		// This gives us the position along the edge that's closest to our node
		t := math.Max(0, math.Min(1, (pointVec.X*edgeVec.X+pointVec.Y*edgeVec.Y)/edgeLengthSq))
		
		// Calculate the closest point on the edge to our node
		closestPoint := vec2.New(
			p1.X + t*edgeVec.X,
			p1.Y + t*edgeVec.Y,
		)
		
		// Calculate the distance from the node to the closest point on the edge
		dx := p.X - closestPoint.X
		dy := p.Y - closestPoint.Y
		distance := math.Sqrt(dx*dx + dy*dy)
		
		// If the distance is less than the node's radius, they intersect
		if distance < n.Radius {
			return true
		}
	}
	
	// No intersections found
	return false
}

func (g *Graph) EdgeIntersectsAny(n1, n2 *node.Node) bool {
	return false
}