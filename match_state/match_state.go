package match_state

import (
	"errors"
	"fmt"
	"math"
	"math/rand"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/assert"
	"github.com/relby/achikaps/graph"
	"github.com/relby/achikaps/material"
	"github.com/relby/achikaps/node"
	"github.com/relby/achikaps/unit"
	"github.com/relby/achikaps/unit_action"
	"github.com/relby/achikaps/vec2"
)

type State struct {
	Presences   map[string]runtime.Presence
	Graphs map[string]*graph.Graph
	NextNodeIDs map[string]node.ID
	Units map[string][]*unit.Unit
	Materials map[string][]*material.Material
}

func (s *State) BuildNode(userID string, fromID node.ID, typ node.Type, pos vec2.Vec2, data any) error {
	userGraph, ok := s.Graphs[userID]
	assert.True(ok)
	
	fromNode, err := userGraph.Node(fromID)
	if errors.Is(err, graph.ErrEdgeNotFound) {
		return fmt.Errorf("node not found: %w", err)
	}
	assert.NoError(err)
	
	toNodeID, ok := s.NextNodeIDs[userID]
	assert.True(ok)

	var toNode *node.Node
	switch typ {
	case node.TransitType:
		toNode = node.NewTransit(toNodeID, pos)
	case node.ProductionType:
		data, ok := data.(node.ProductionTypeData)
		if !ok {
			return fmt.Errorf("invalid Data for ProductionType node")
		}
		
		toNode = node.NewProduction(toNodeID, pos, data)
	default:
		assert.Unreachable()
	}

	for _, g := range s.Graphs {
		if g.NodeIntersectsAny(toNode) {
			return fmt.Errorf("new node intersects the graph")
		}
		
		if g.EdgeIntersectsAny(fromNode, toNode) {
			return fmt.Errorf("new edge intersects the graph")
		}
		
	}

	if err := userGraph.AddNodeFrom(fromNode, toNode); err != nil {
		return fmt.Errorf("can't add node: %w", err)
	}

	s.NextNodeIDs[userID] += 1

	return nil
}

func (s *State) Tick() {
	for userID, units := range s.Units {
		for _, u := range units {
			if u.Actions.Len() == 0 {
				ok := s.pollActions(userID, u)
				if !ok {
					continue
				}
			}
			
			// Assert that there's is at least one action
			assert.NotEquals(u.Actions.Len(), 0)
			
			action := u.Actions.Front()

			done := s.executeUnitAction(userID, u, action)
			if done {
				u.Actions.PopFront()
			}
		}
	}
}

// pollActions tries to add action to a unit
// If action was added returns true, otherwise false
func (s *State) pollActions(userID string, u *unit.Unit) bool {
	userGraph, ok := s.Graphs[userID]
	assert.True(ok)

	userUnits, ok := s.Units[userID]
	assert.True(ok)

	userMaterials, ok := s.Materials[userID]
	assert.True(ok)

	am := userGraph.AdjacencyMap()
	adjacentNodeMap, ok := am[u.Node.ID]
	assert.True(ok)
	
	getRandomAdjacentNode := func () *node.Node {
		adjacentNodes := make([]node.ID, 0, len(adjacentNodeMap))
		for nodeID := range adjacentNodeMap {
			adjacentNodes = append(adjacentNodes, nodeID)
		}
		
		// Filter out nodes that are not built yet
		filteredNodes := make([]*node.Node, 0, len(adjacentNodes))
		for _, nID := range adjacentNodes {
			n, err := userGraph.Node(nID)
			assert.NoError(err)

			if n.IsBuilt() {
				filteredNodes = append(filteredNodes, n)
			}
		}
		
		// If there are no nodes do nothing
		if len(filteredNodes) == 0 {
			return nil
		}

		randomNodeID := adjacentNodes[rand.Intn(len(adjacentNodes))]
		randomNode, err := userGraph.Node(randomNodeID)
		assert.NoError(err)

		return randomNode
	}
	
	findShortestPathOfMultiple := func (ns []*node.Node) ([]*node.Node, *node.Node) {
		var path []*node.Node
		var finalNode *node.Node
		pathDist := math.MaxFloat64
		for _, n := range ns {
			ns := userGraph.FindShortestPath(u.Node, n)
			
			// Calculate the total path length
			totalDist := 0.0
			for i := range len(ns) - 1 {
				// Calculate distance between consecutive nodes in the path
				dist := ns[i].DistanceTo(ns[i+1])
				totalDist += dist
			}

			if totalDist < pathDist {
				path = ns
				finalNode = n
				pathDist = totalDist
			}
		}

		return path, finalNode
	}

	switch u.Type {
	case unit.IdleType:
		n := getRandomAdjacentNode()
		
		// No node found
		if n == nil {
			return false
		}

		u.Actions.PushBack(unit_action.NewMoving(unit.DefaultSpeed, u.Node, n))

		return true
	case unit.ProductionType:
		// If unit is not in the production node find the node with the least amount of units
		if u.Node.Type != node.ProductionType {
			prodNodes := userGraph.NodesByType(node.ProductionType)
			
			if len(prodNodes) == 0 {
				// Move in a random direction like IdleType units, just to be dynamic
				n := getRandomAdjacentNode()
				if n == nil {
					return false
				}
				
				u.Actions.PushBack(unit_action.NewMoving(unit.DefaultSpeed, u.Node, n))
				return true
			}
			
			// Find the production node with the least amount of units
			var leastPopulatedNode *node.Node
			minUnitCount := math.MaxInt
			
			for _, n := range prodNodes {
				unitCount := 0
				for _, u := range userUnits {
					if u.Node.ID == n.ID {
						unitCount++
					}
				}
				
				if unitCount < minUnitCount {
					minUnitCount = unitCount
					leastPopulatedNode = n
				}
			}

			ns := userGraph.FindShortestPath(u.Node, leastPopulatedNode)

			for i := range len(ns) - 1 {
				n1, n2 := ns[i], ns[i + 1]
				u.Actions.PushBack(unit_action.NewMoving(unit.DefaultSpeed, n1, n2))
			}
		}

		u.Actions.PushBack(unit_action.NewProduction())

		return true

	case unit.BuilderType:
		if u.Node.IsBuilt() {
			buildingNodes := userGraph.BuildingNodes()
			
			if len(buildingNodes) == 0 {
				// Move in a random direction like IdleType units, just to be dynamic
				n := getRandomAdjacentNode()
				if n == nil {
					return false
				}
				
				u.Actions.PushBack(unit_action.NewMoving(unit.DefaultSpeed, u.Node, n))
			}
			
			
			shortestPath, _ := findShortestPathOfMultiple(buildingNodes)

			for i := range len(shortestPath) - 1 {
				n1, n2 := shortestPath[i], shortestPath[i + 1]
				u.Actions.PushBack(unit_action.NewMoving(unit.DefaultSpeed, n1, n2))
			}
			
		}
		
		u.Actions.PushBack(unit_action.NewBuilding())

		return true
	case unit.TransportType:
		data, ok := u.Data.(*unit.TransportData)
		assert.True(ok)

		if data.Material == nil {
			type NodeWithMaterials struct {
				Node *node.Node
				Materials []*material.Material
			}
			nMap := make(map[node.ID]NodeWithMaterials, userGraph.NodeCount())

			// Add nodes with non reserved materials
			for _, m := range userMaterials {
				// If material is reserved or node is nil, that means that material is already being transported
				if m.IsReserved {
					continue
				}

				assert.NotNil(m.Node)

				nMap[m.Node.ID] = NodeWithMaterials{
					m.Node,
					append(nMap[m.Node.ID].Materials, m),
				}
			}
			
			ns := make([]*node.Node, 0, len(nMap))
			for _, nwm := range nMap {
				ns = append(ns, nwm.Node)
			}
			
			if len(ns) == 0 {
				// Move in a random direction like IdleType units, just to be dynamic
				n := getRandomAdjacentNode()
				if n == nil {
					return false
				}
				
				u.Actions.PushBack(unit_action.NewMoving(unit.DefaultSpeed, u.Node, n))
				return true
			}


			shortestPath, finalNode := findShortestPathOfMultiple(ns)
			
			ms := nMap[finalNode.ID].Materials
			m := ms[rand.Intn(len(ms))]

			m.IsReserved = true
			
			if u.Node.ID != finalNode.ID {
				for i := range len(shortestPath) - 1 {
					n1, n2 := shortestPath[i], shortestPath[i + 1]
					u.Actions.PushBack(unit_action.NewMoving(unit.DefaultSpeed, n1, n2))
				}
			}
			
			u.Actions.PushBack(unit_action.NewTakeMaterial(m))
		}
		
		// TODO: Move the unit to the end location

		u.Actions.PushBack(unit_action.NewDropMaterial())

		return true
	default:
		panic("unreachable")
	}
}

func (s *State) executeUnitAction(userID string, u *unit.Unit, action *unit_action.UnitAction) bool {
	switch action.Type {
	case unit_action.MovingType:
		data, ok := action.Data.(*unit_action.MovingData)
		assert.True(ok)
		
		u.Node = nil

		dist := vec2.Distance(data.FromNode.Position, data.ToNode.Position)
		data.Progress += data.Speed / dist
		
		if data.Progress >= 1.0 {
			u.Node = data.ToNode
			return true
		}
		
		return false
	// TODO: Maybe I should change this logic according to building logic
	case unit_action.ProductionType:
		data, ok := action.Data.(*unit_action.ProductionData)
		assert.True(ok)

		const progressIncrement = 0.1
		data.Progress += progressIncrement
		
		if data.Progress >= 1.0 {
			assert.Equals(u.Node.Type, node.ProductionType)
			data, ok := u.Node.Data.(node.ProductionTypeData)
			assert.True(ok)
			
			switch data {
			case node.UnitProductionTypeData:
				s.Units[userID] = append(s.Units[userID], unit.NewIdle(u.Node))
			case node.TODOMaterialProductionTypeData:
				s.Materials[userID] = append(s.Materials[userID], material.New(material.TODOType, u.Node))
			default:
				panic("unreachable")
			}

			return true
		}
		
		return false
	case unit_action.BuildingType:
		const progressIncrement = 0.1
		u.Node.BuildProgress += progressIncrement
		
		if u.Node.BuildProgress >= 1.0 {
			u.Node.BuildProgress = 1.0
			return true
		}
		
		return false
	// TODO: right now this action is instant, maybe consider making it non instant
	case unit_action.TakeMaterialType:
		uData, ok := u.Data.(*unit.TransportData)
		assert.True(ok)

		uaData, ok := action.Data.(*unit_action.TakeMaterialData)
		assert.True(ok)

		assert.Nil(uData.Material)

		uData.Material = uaData.Material
		uData.Material.Node = nil
		
		return true
	case unit_action.DropMaterialType:
		uData, ok := u.Data.(*unit.TransportData)
		assert.True(ok)

		assert.NotNil(uData.Material)
		
		uData.Material.IsReserved = false
		uData.Material.Node = u.Node
		uData.Material = nil
		
		return true
	default:
		panic("unreachable")
	}
}