package game

import (
	"math"
	"math/rand"

	"github.com/relby/achikaps/assert"
	"github.com/relby/achikaps/graph"
	"github.com/relby/achikaps/material"
	"github.com/relby/achikaps/node"
	"github.com/relby/achikaps/unit"
	"github.com/relby/achikaps/vec2"
)

type Manager struct {
	Graph *graph.Graph
	Units []*unit.Unit
	Materials []*material.Material
}

func NewManager(g *graph.Graph, units []*unit.Unit) *Manager {
	return &Manager{g, units, nil}
}

func (m *Manager) executeUnitAction(u *unit.Unit, action *unit.Action) bool {
	switch action.Type {
	case unit.MovingActionType:
		data, ok := action.Data.(*unit.MovingActionData)
		assert.True(ok)

		dist := vec2.Distance(u.Node.Position, data.Node.Position)
		data.Progress += data.Speed / dist
		
		if data.Progress >= 1.0 {
			u.Node = data.Node
			return true
		}
		
		return false
	// TODO: Maybe I should change this logic according to building logic
	case unit.ProductionActionType:
		data, ok := action.Data.(*unit.MovingActionData)
		assert.True(ok)

		const progressIncrement = 0.1
		data.Progress += progressIncrement
		
		if data.Progress >= 1.0 {
			assert.True(u.Node.Type == node.ProductionType)
			data, ok := u.Node.Data.(node.ProductionTypeData)
			assert.True(ok)
			
			switch data {
			case node.UnitProductionTypeData:
				m.Units = append(m.Units, unit.New(unit.IdleType, u.Node))
			case node.TODOMaterialProductionTypeData:
				m.Materials = append(m.Materials, material.New(material.TODOType, u.Node))
			default:
				panic("unreachable")
			}

			return true
		}
		
		return false
	case unit.BuildingActionType:
		const progressIncrement = 0.1
		u.Node.BuildProgress += progressIncrement
		
		if u.Node.BuildProgress >= 1.0 {
			u.Node.BuildProgress = 1.0
			return true
		}
		
		return false
	default:
		panic("unreachable")
	}
}

// pollActions tries to add action to a unit
// If action was added returns true, otherwise false
func (m *Manager) pollActions(a *unit.Unit) bool {
	am := m.Graph.AdjacencyMap()
	adjacentNodeMap, ok := am[a.Node.ID]
	assert.True(ok)
	

	switch a.Type {
	case unit.IdleType:
		adjacentNodes := make([]node.ID, 0, len(adjacentNodeMap))
		for nodeID := range adjacentNodeMap {
			adjacentNodes = append(adjacentNodes, nodeID)
		}
		
		// Filter out nodes that are not built yet
		filteredNodes := make([]*node.Node, 0, len(adjacentNodes))
		for _, nID := range adjacentNodes {
			n, err := m.Graph.Node(nID)
			assert.NoError(err)

			if n.IsBuilt() {
				filteredNodes = append(filteredNodes, n)
			}
		}
		
		// If there are no nodes do nothing
		if len(filteredNodes) == 0 {
			return false
		}

		randomNodeID := adjacentNodes[rand.Intn(len(adjacentNodes))]
		randomNode, _ := m.Graph.Node(randomNodeID)

		a.Actions.PushBack(unit.NewMovingAction(unit.DefaultSpeed, randomNode))

		return true
	case unit.ProductionType:
		// If unit is not in the production node find the node with the least amount of units
		if a.Node.Type != node.ProductionType {
			prodNodes := m.Graph.NodesByType(node.ProductionType)
			
			if len(prodNodes) == 0 {
				// TODO: Maybe move in a random direction like IdleType units, just to be dynamic
				return false
			}
			
			// Find the production node with the least amount of units
			var leastPopulatedNode *node.Node
			minUnitCount := math.MaxInt
			
			for _, n := range prodNodes {
				unitCount := 0
				for _, a := range m.Units {
					if a.Node.ID == n.ID {
						unitCount++
					}
				}
				
				if unitCount < minUnitCount {
					minUnitCount = unitCount
					leastPopulatedNode = n
				}
			}

			ns := m.Graph.FindShortestPath(a.Node, leastPopulatedNode)

			for _, n := range ns {
				a.Actions.PushBack(unit.NewMovingAction(unit.DefaultSpeed, n))
			}
		}

		a.Actions.PushBack(unit.NewProductionAction())

		return true
		
	case unit.BuilderType:
		if a.Node.IsBuilt() {
			buildingNodes := m.Graph.BuildingNodes()
			
			if len(buildingNodes) == 0 {
				return false
			}
			
			var shortestPath []*node.Node
			shortestPathDist := math.MaxFloat64
			for _, n := range buildingNodes {
				ns := m.Graph.FindShortestPath(a.Node, n)
				
				// Calculate the total path length
				totalDist := 0.0
				for i := range len(ns) - 1 {
					// Calculate distance between consecutive nodes in the path
					dist := ns[i].DistanceTo(ns[i+1])
					totalDist += dist
				}

				if totalDist < shortestPathDist {
					shortestPath = ns
					shortestPathDist = totalDist
				}
			}
			
			for _, n := range shortestPath {
				a.Actions.PushBack(unit.NewMovingAction(unit.DefaultSpeed, n))
			}
			
		}
		
		a.Actions.PushBack(unit.NewBuildingAction())

		return true
	default:
		panic("unreachable")
	}
}

func (m *Manager) Tick() {
	for _, unit := range m.Units {
		if unit.Actions.Len() == 0 {
			ok := m.pollActions(unit)
			if !ok {
				continue
			}
		}
		
		// Assert that there's is at least one action
		assert.True(unit.Actions.Len() != 0)
		
		action := unit.Actions.Front()

		done := m.executeUnitAction(unit, action)
		if done {
			unit.Actions.PopFront()
		}
	}
}