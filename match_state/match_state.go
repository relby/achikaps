package match_state

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"math/rand"
	"slices"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/assert"
	"github.com/relby/achikaps/graph"
	"github.com/relby/achikaps/model"
	"github.com/relby/achikaps/vec2"
)

type State struct {
	Presences   map[string]runtime.Presence

	Graphs map[string]*graph.Graph
	NextNodeIDs map[string]model.ID

	Units map[string]map[model.ID]*model.Unit
	NextUnitIDs map[string]model.ID

	Materials map[string]map[model.ID]*model.Material
	NextMaterialIDs map[string]model.ID
	
	ClientUpdates map[string][]*model.UnitAction
}

func (s *State) BuildNode(userID string, fromID model.ID, name model.NodeName, pos vec2.Vec2) (*model.Node, error) {
	userGraph, ok := s.Graphs[userID]
	assert.True(ok)
	
	fromNode, err := userGraph.Node(fromID)
	if errors.Is(err, graph.ErrEdgeNotFound) {
		return nil, fmt.Errorf("node not found: %w", err)
	}
	assert.NoError(err)
	
	toNodeID, ok := s.NextNodeIDs[userID]
	assert.True(ok)

	toNode := model.NewNode(toNodeID, name, pos)
	
	if fromNode.DistanceTo(toNode) > 10 * model.DefaultNodeRadius {
		return nil, fmt.Errorf("new node is too far")
	}

	for _, g := range s.Graphs {
		if g.NodeIntersectsAny(toNode) {
			return nil, fmt.Errorf("new node intersects the graph")
		}
		
		if g.EdgeIntersectsAny(fromNode, toNode) {
			return nil, fmt.Errorf("new edge intersects the graph")
		}
	}

	if err := userGraph.AddNodeFrom(fromNode, toNode); err != nil {
		return nil, fmt.Errorf("can't add node: %w", err)
	}

	s.NextNodeIDs[userID] += 1

	return toNode, nil
}

func (s *State) ChangeUnitType(userID string, id model.ID, typ model.UnitType) (*model.Unit, error) {
	userUnits, ok := s.Units[userID]
	assert.True(ok)
	
	u, exists := userUnits[id]
	if !exists {
		return nil, fmt.Errorf("unit not found")
	}
	
	u.SetType(typ)
	
	return u, nil
}

func (s *State) Tick() {
	for userID, units := range s.Units {
		for _, u := range units {
			if u.Actions().Len() == 0 {
				s.pollActions(userID, u)
			}
		}
	}
		
	for userID, units := range s.Units {
		for _, u := range units {
			if u.Actions().Len() == 0 {
				continue
			}
			
			action := u.Actions().Front()
			
			// Action is about to start, add client updates
			if !action.IsStarted {
				s.ClientUpdates[userID] = append(s.ClientUpdates[userID], action)
			}

			done := s.executeUnitAction(userID, u, action)
			if done {
				u.Actions().PopFront()
			}
		}
	}
}

// pollActions tries to add action to a unit
// If action was added returns true, otherwise false
func (s *State) pollActions(userID string, u *model.Unit) bool {
	userGraph, ok := s.Graphs[userID]
	assert.True(ok)

	// Unit should always have a node when polling for actions
	assert.NotNil(u.Node())

	getRandomAdjacentNode := func() (*model.Node, bool) {
		am := userGraph.AdjacencyMap()
		adjacentNodeMap, ok := am[u.Node().ID()]
		assert.True(ok)

		adjacentNodes := make([]model.ID, 0, len(adjacentNodeMap))
		for nodeID := range adjacentNodeMap {
			adjacentNodes = append(adjacentNodes, nodeID)
		}
		
		// Filter out nodes that are not built yet
		builtNodes := make([]*model.Node, 0, len(adjacentNodes))
		for _, nID := range adjacentNodes {
			n, err := userGraph.Node(nID)
			assert.NoError(err)

			if n.IsBuilt() {
				builtNodes = append(builtNodes, n)
			}
		}
		
		if len(builtNodes) == 0 {
			return nil, false
		}

		randomNode := builtNodes[rand.Intn(len(builtNodes))]

		return randomNode, true
	}
	
	findShortestPathOfMultiple := func(ns []*model.Node) ([]*model.Node, *model.Node) {
		var path []*model.Node
		var finalNode *model.Node
		pathDist := math.MaxFloat64
		for _, n := range ns {
			ns := userGraph.FindShortestPath(u.Node(), n)
			
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

	switch u.Type() {
	case model.IdleUnitType:
		n, ok := getRandomAdjacentNode()
		
		if !ok {
			return false
		}

		u.Actions().PushBack(model.NewMovingUnitAction(model.DefaultUnitSpeed, u.Node(), n))
		return true

	case model.ProductionUnitType:
		// If unit is not in the production node find the node with the least amount of units
		if u.Node().Type() != model.ProductionNodeType {
			prodNodes := userGraph.NodesByType(model.ProductionNodeType)
			
			if len(prodNodes) == 0 {
				// Move in a random direction like IdleType units, just to be dynamic
				n, ok := getRandomAdjacentNode()
				if !ok {
					return false
				}
				
				u.Actions().PushBack(model.NewMovingUnitAction(model.DefaultUnitSpeed, u.Node(), n))
				return true
			}
			
			// Find the production node with the least amount of units
			var leastPopulatedNode *model.Node
			minUnitCount := math.MaxInt
			
			for _, n := range prodNodes {
				unitCount := 0
				for _, u := range n.Units() {
					if u.Type() == model.ProductionUnitType {
						unitCount += 1
					}
				}
				
				if unitCount < minUnitCount {
					minUnitCount = unitCount
					leastPopulatedNode = n
				}
			}

			ns := userGraph.FindShortestPath(u.Node(), leastPopulatedNode)

			for i := range len(ns) - 1 {
				n1, n2 := ns[i], ns[i + 1]
				u.Actions().PushBack(model.NewMovingUnitAction(model.DefaultUnitSpeed, n1, n2))
			}
		}

		u.Actions().PushBack(model.NewProductionUnitAction())
		return true

	case model.BuilderUnitType:
		if u.Node().IsBuilt() {
			buildingNodes := userGraph.BuildingNodes()
			
			if len(buildingNodes) == 0 {
				// Move in a random direction like IdleType units, just to be dynamic
				n, ok := getRandomAdjacentNode()
				if !ok {
					return false
				}
				
				u.Actions().PushBack(model.NewMovingUnitAction(model.DefaultUnitSpeed, u.Node(), n))
				return true
			}
			
			
			shortestPath, finalNode := findShortestPathOfMultiple(buildingNodes)

			if u.Node().ID() != finalNode.ID() {
				for i := range len(shortestPath) - 1 {
					n1, n2 := shortestPath[i], shortestPath[i + 1]
					u.Actions().PushBack(model.NewMovingUnitAction(model.DefaultUnitSpeed, n1, n2))
				}
			}
		}
		
		u.Actions().PushBack(model.NewBuildingUnitAction())
		return true

	case model.TransportUnitType:
		// TODO: We should find materials only when needed, so make the logic of need
		if u.Material() == nil {
			nodesInNeed := make([]*model.Node, 0, userGraph.NodeCount())
			_ = nodesInNeed
			

			// Find nodes with non reserved materials
			nonReservedNodes := make([]*model.Node, 0, userGraph.NodeCount())
			for _, n := range userGraph.Nodes() {
				for _, m := range n.Materials() {
					if !m.IsReserved() {
						nonReservedNodes = append(nonReservedNodes, n)
						break
					}
				}
			}
			
			if len(nonReservedNodes) == 0 {
				// Move in a random direction like IdleType units, just to be dynamic
				n, ok := getRandomAdjacentNode()
				if !ok {
					return false
				}
				
				u.Actions().PushBack(model.NewMovingUnitAction(model.DefaultUnitSpeed, u.Node(), n))
				return true
			}


			shortestPath, finalNode := findShortestPathOfMultiple(nonReservedNodes)
			
			ms := slices.Collect(maps.Values(finalNode.Materials()))
			m := ms[rand.Intn(len(ms))]

			m.Reserve()
			
			if u.Node().ID() != finalNode.ID() {
				for i := range len(shortestPath) - 1 {
					n1, n2 := shortestPath[i], shortestPath[i + 1]
					u.Actions().PushBack(model.NewMovingUnitAction(model.DefaultUnitSpeed, n1, n2))
				}
			}
			
			u.Actions().PushBack(model.NewTakeMaterialUnitAction(m))
		}
		
		// TODO: Move the unit to the end location

		u.Actions().PushBack(model.NewDropMaterialUnitAction())

		return true
	default:
		panic("unreachable")
	}
}

func (s *State) executeUnitAction(userID string, u *model.Unit, action *model.UnitAction) bool {
	justStarted := !action.IsStarted

	if !action.IsStarted {
		action.IsStarted = true
	}

	switch action.Type {
	case model.MovingUnitActionType:
		data, ok := action.Data.(*model.MovingUnitActionData)
		assert.True(ok)
		
		if justStarted {
			u.Node().RemoveUnit(u)
		}

		dist := vec2.Distance(data.FromNode.Position(), data.ToNode.Position())
		data.Progress += data.Speed / dist
		
		if data.Progress >= 1.0 {
			data.ToNode.AddUnit(u)
			return true
		}
		
		return false
	// TODO: Maybe I should change this logic according to building logic
	case model.ProductionUnitActionType:
		data, ok := action.Data.(*model.ProductionUnitActionData)
		assert.True(ok)

		const progressIncrement = 0.1
		data.Progress += progressIncrement
		
		if data.Progress >= 1.0 {
			data.Progress = 1.0

			assert.Equals(u.Node().Type(), model.ProductionNodeType)
			data, ok := u.Node().ProductionData()
			assert.True(ok)
			
			if data.InputMaterials != nil {
				// TODO: delete materials in the node
			}

			if data.OutputMaterials != nil {
				for typ, count := range data.OutputMaterials {
					for range count {
						materialID, ok := s.NextMaterialIDs[userID]
						assert.True(ok)

						s.Materials[userID][materialID] = model.NewMaterial(materialID, typ, u.Node())
						
						s.NextMaterialIDs[userID] += 1
					}
				}
			}

			if data.OutputUnits > 0 {
				unitID, ok := s.NextUnitIDs[userID]
				assert.True(ok)
				s.Units[userID][unitID] = model.NewUnit(unitID, model.IdleUnitType, u.Node())

				s.NextUnitIDs[userID] += 1
			}
			return true
		}
		
		return false
	case model.BuildingUnitActionType:
		const progressIncrement = 0.1
		u.Node().Build(progressIncrement)
		
		if u.Node().IsBuilt() {
			buildingData := u.Node().BuildingData()

			_ = buildingData
			// TODO: Remove materials from node

			return true
		}
		
		return false
	// TODO: right now this action is instant, maybe consider making it non instant
	case model.TakeMaterialUnitActionType:
		assert.Nil(u.Material())

		uaData, ok := action.Data.(*model.TakeMaterialUnitActionData)
		assert.True(ok)

		u.AddMaterial(uaData.Material)
		
		return true
	case model.DropMaterialUnitActionType:
		assert.NotNil(u.Material())

		u.Material().UnReserve()
		u.RemoveMaterial()
		
		return true
	default:
		panic("unreachable")
	}
}