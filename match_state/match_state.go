package match_state

import (
	"errors"
	"fmt"
	"math"
	"math/rand"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/assert"
	"github.com/relby/achikaps/config"
	"github.com/relby/achikaps/graph"
	"github.com/relby/achikaps/model"
	"github.com/relby/achikaps/opcode"
	"github.com/relby/achikaps/vec2"
	"github.com/relby/achikaps/win_condition"
)


type State struct {
	Presences   map[string]runtime.Presence

	Graphs map[string]*graph.Graph
	NextNodeIDs map[string]model.ID

	Units map[string]map[model.ID]*model.Unit
	NextUnitIDs map[string]model.ID

	Materials map[string]map[model.ID]*model.Material
	NextMaterialIDs map[string]model.ID
	
	WinCondition *win_condition.WinCondition
	
	UnitActionExecuteResps map[string][]*opcode.UnitActionExecuteResp
	NodeBuiltResps map[string][]*opcode.NodeBuiltResp
}

func (s *State) BuildNode(sessionID string, fromID model.ID, name model.NodeName, pos vec2.Vec2) (*model.Node, error) {
	playerGraph, ok := s.Graphs[sessionID]
	assert.True(ok)
	
	fromNode, err := playerGraph.Node(fromID)
	if errors.Is(err, graph.ErrEdgeNotFound) {
		return nil, fmt.Errorf("node not found: %w", err)
	}
	assert.NoError(err)
	
	toNodeID, ok := s.NextNodeIDs[sessionID]
	assert.True(ok)

	toNode := model.NewNode(toNodeID, sessionID, name, pos)
	
	if fromNode.DistanceTo(toNode) > 10 * config.NodeRadius {
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

	if err := playerGraph.AddNodeFrom(fromNode, toNode); err != nil {
		return nil, fmt.Errorf("can't add node: %w", err)
	}

	s.NextNodeIDs[sessionID] += 1

	return toNode, nil
}

func (s *State) ChangeUnitType(sessionID string, id model.ID, typ model.UnitType) (*model.Unit, error) {
	playerUnits, ok := s.Units[sessionID]
	assert.True(ok)
	
	u, exists := playerUnits[id]
	if !exists {
		return nil, fmt.Errorf("unit not found")
	}
	
	u.SetType(typ)
	
	return u, nil
}

func (s *State) Tick() {
	for sessionID, units := range s.Units {
		for _, u := range units {
			if u.Actions().Len() == 0 {
				s.pollActions(sessionID, u)
			}
		}
	}
		
	for sessionID, units := range s.Units {
		for _, u := range units {
			if u.Actions().Len() == 0 {
				continue
			}
			
			action := u.Actions().Front()
			
			// Action is about to start, add client updates
			if !action.IsStarted {
				s.UnitActionExecuteResps[sessionID] = append(
					s.UnitActionExecuteResps[sessionID],
					opcode.NewUnitActionExecuteResp(u, action),
				)
			}

			done := s.executeUnitAction(sessionID, u, action)
			if done {
				u.Actions().PopFront()
			}
		}
	}
}

// pollActions tries to add action to a unit
func (s *State) pollActions(sessionID string, u *model.Unit) {
	playerGraph, ok := s.Graphs[sessionID]
	assert.True(ok)

	playerMaterials, ok := s.Materials[sessionID]
	assert.True(ok)

	// Unit should always have a node when polling for actions
	assert.NotNil(u.Node())

	getRandomAdjacentNode := func() (*model.Node, bool) {
		am := playerGraph.AdjacencyMap()
		adjacentNodeMap, ok := am[u.Node().ID()]
		assert.True(ok)

		adjacentNodes := make([]model.ID, 0, len(adjacentNodeMap))
		for nodeID := range adjacentNodeMap {
			adjacentNodes = append(adjacentNodes, nodeID)
		}
		
		// Filter out nodes that are not built yet
		builtNodes := make([]*model.Node, 0, len(adjacentNodes))
		for _, nID := range adjacentNodes {
			n, err := playerGraph.Node(nID)
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
			ns := playerGraph.FindShortestPath(u.Node(), n)
			
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
			return
		}

		u.Actions().PushBack(model.NewMovingUnitAction(config.UnitSpeed, u.Node(), n))
	case model.ProductionUnitType:
		finalNode := u.Node()
		// If unit is not in the production node find the node with the least amount of units
		if u.Node().Type() != model.ProductionNodeType {
			prodNodes := playerGraph.NodesByType(model.ProductionNodeType, true)
			
			if len(prodNodes) == 0 {
				// Move in a random direction like IdleType units, just to be dynamic
				n, ok := getRandomAdjacentNode()
				if !ok {
					return
				}
				
				u.Actions().PushBack(model.NewMovingUnitAction(config.UnitSpeed, u.Node(), n))
				return
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

			ns := playerGraph.FindShortestPath(u.Node(), leastPopulatedNode)

			for i := range len(ns) - 1 {
				n1, n2 := ns[i], ns[i + 1]
				u.Actions().PushBack(model.NewMovingUnitAction(config.UnitSpeed, n1, n2))
			}

			finalNode = leastPopulatedNode
		}

		data, ok := finalNode.ProductionData()
		assert.True(ok)
		
		inputMaterials := make([]*model.Material, 0, len(data.InputMaterials))

		enoughMaterials := false
		if len(data.InputMaterials) >= 0 {
			for _, m := range finalNode.InputMaterials() {
				c, exists := data.InputMaterials[m.Type()];
				if !exists {
					continue
				}
				// Ensure that there's at least 1 input material to build a node
				assert.NotEquals(c, 0)

				c -= 1
				if c == 0 {
					delete(data.InputMaterials, m.Type())
					if len(data.InputMaterials) == 0 {
						enoughMaterials = true
						break
					}
				}

				data.InputMaterials[m.Type()] = c

				inputMaterials = append(inputMaterials, m)
			}
		}
		
		if !enoughMaterials {
			return 
		}
		
		for _, m := range inputMaterials {
			m.Reserve()
		}

		u.Actions().PushBack(model.NewProductionUnitAction(inputMaterials))
	case model.BuilderUnitType:
		buildingNodes := playerGraph.BuildingNodes()
		if len(buildingNodes) == 0 {
			// Move in a random direction like IdleType units, just to be dynamic
			n, ok := getRandomAdjacentNode()
			if !ok {
				return
			}
			
			u.Actions().PushBack(model.NewMovingUnitAction(config.UnitSpeed, u.Node(), n))
			return
		}
		

		validBuildingNodes := make([]*model.Node, 0, len(buildingNodes))
		for _, n := range buildingNodes {
			data := n.BuildingData()
			
			// Every building should require some materials to build
			assert.NotEquals(len(data.Materials), 0)

			enoughMaterials := false
			for _, m := range n.InputMaterials() {
				c, exists := data.Materials[m.Type()];
				if !exists {
					continue
				}
				// Ensure that there's at least 1 input material to build a node
				assert.NotEquals(c, 0)

				c -= 1
				if c == 0 {
					delete(data.Materials, m.Type())
					if len(data.Materials) == 0 {
						enoughMaterials = true
						break
					}
				}

				data.Materials[m.Type()] = c
			}
			
			if enoughMaterials {
				validBuildingNodes = append(validBuildingNodes, n)
			}
		}
		
		if len(validBuildingNodes) == 0 {
			// Move in a random direction like IdleType units, just to be dynamic
			n, ok := getRandomAdjacentNode()
			if !ok {
				return
			}
			
			u.Actions().PushBack(model.NewMovingUnitAction(config.UnitSpeed, u.Node(), n))
			return
		}

		shortestPath, finalNode := findShortestPathOfMultiple(validBuildingNodes)

		if u.Node().ID() != finalNode.ID() {
			for i := range len(shortestPath) - 1 {
				n1, n2 := shortestPath[i], shortestPath[i + 1]
				u.Actions().PushBack(model.NewMovingUnitAction(config.UnitSpeed, n1, n2))
			}
		}
		
		u.Actions().PushBack(model.NewBuildingUnitAction())
	case model.TransportUnitType:
		// TODO:
		neededMaterials := make(
			map[model.MaterialType]struct{Node *model.Node; Count uint},
			playerGraph.NodeCount(),
		)
		
		for _, n := range playerGraph.BuildingNodes() {
			enoughMaterials := false
			data := n.BuildingData()
			for _, m := range n.InputMaterials() {
				c, exists := data.Materials[m.Type()];
				if !exists {
					continue
				}
				// Ensure that there's at least 1 input material to build a node
				assert.NotEquals(c, 0)

				c -= 1
				if c == 0 {
					delete(data.Materials, m.Type())
					if len(data.Materials) == 0 {
						enoughMaterials = true
						break
					}
				}

				data.Materials[m.Type()] = c
			}
			
			if enoughMaterials {
				continue
			}
			
			for matType, count := range data.Materials {
				neededMaterials[matType] = struct{Node *model.Node; Count uint}{n, count}
			}
		}

		for _, n := range playerGraph.NodesByType(model.ProductionNodeType, true) {
			enoughMaterials := false
			data, ok := n.ProductionData()
			assert.True(ok)

			for _, m := range n.InputMaterials() {
				c, exists := data.InputMaterials[m.Type()];
				if !exists {
					continue
				}
				// Ensure that there's at least 1 input material to build a node
				assert.NotEquals(c, 0)

				c -= 1
				if c == 0 {
					delete(data.InputMaterials, m.Type())
					if len(data.InputMaterials) == 0 {
						enoughMaterials = true
						break
					}
				}

				data.InputMaterials[m.Type()] = c
			}
			
			if enoughMaterials {
				continue
			}
			
			for matType, count := range data.InputMaterials {
				neededMaterials[matType] = struct{Node *model.Node; Count uint}{n, count}
			}
		}
		
		if len(neededMaterials) == 0 {
			// Move in a random direction like IdleType units, just to be dynamic
			n, ok := getRandomAdjacentNode()
			if !ok {
				return
			}
			
			u.Actions().PushBack(model.NewMovingUnitAction(config.UnitSpeed, u.Node(), n))
		}
		
		for _, m := range playerMaterials {
			if !m.IsReserved() && !m.NodeData().IsInput {
				matData, ok := neededMaterials[m.Type()]
				if !ok {
					continue
				}
				
				m.Reserve()
				
				if u.Node() != m.NodeData().Node {
					shortestPath := playerGraph.FindShortestPath(u.Node(), m.NodeData().Node)
					for i := range len(shortestPath) - 1 {
						n1, n2 := shortestPath[i], shortestPath[i + 1]
						u.Actions().PushBack(model.NewMovingUnitAction(config.UnitSpeed, n1, n2))
					}
				}
				
				u.Actions().PushBack(model.NewTakeMaterialUnitAction(m))

				if u.Node() != matData.Node {
					shortestPath := playerGraph.FindShortestPath(u.Node(), matData.Node)
					for i := range len(shortestPath) - 1 {
						n1, n2 := shortestPath[i], shortestPath[i + 1]
						u.Actions().PushBack(model.NewMovingUnitAction(config.UnitSpeed, n1, n2))
					}
				}

				u.Actions().PushBack(model.NewDropMaterialUnitAction())
				break
			}
		}
	default:
		panic("unreachable")
	}
}

func (s *State) executeUnitAction(sessionID string, u *model.Unit, action *model.UnitAction) bool {
	playerUnits, ok := s.Units[sessionID]
	assert.True(ok)

	playerMaterials, ok := s.Materials[sessionID]
	assert.True(ok)

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
	case model.ProductionUnitActionType:
		uaData, ok := action.Data.(*model.ProductionUnitActionData)
		assert.True(ok)

		const progressIncrement = 0.1
		uaData.Progress += progressIncrement
		
		if uaData.Progress >= 1.0 {
			uaData.Progress = 1.0

			assert.Equals(u.Node().Type(), model.ProductionNodeType)
			prodData, ok := u.Node().ProductionData()
			assert.True(ok)
			
			for _, m := range uaData.InputMaterials {
				m.NodeData().Node.RemoveInputMaterial(m)
				delete(playerMaterials, m.ID())
			}

			if prodData.OutputMaterials != nil {
				for typ, count := range prodData.OutputMaterials {
					for range count {
						materialID, ok := s.NextMaterialIDs[sessionID]
						assert.True(ok)

						playerMaterials[materialID] = model.NewMaterial(materialID, sessionID, typ, u.Node(), false)
						
						s.NextMaterialIDs[sessionID] += 1
					}
				}
			}

			if prodData.OutputUnits > 0 {
				unitID, ok := s.NextUnitIDs[sessionID]
				assert.True(ok)
				playerUnits[unitID] = model.NewUnit(unitID, sessionID, model.IdleUnitType, u.Node())

				s.NextUnitIDs[sessionID] += 1
			}
			return true
		}
		
		return false
	case model.BuildingUnitActionType:
		const progressIncrement = 0.1
		u.Node().Build(progressIncrement)
		
		if u.Node().IsBuilt() {
			for _, m := range u.Node().InputMaterials() {
				m.NodeData().Node.RemoveInputMaterial(m)
				delete(playerMaterials, m.ID())
			}
			
			s.NodeBuiltResps[sessionID] = append(
				s.NodeBuiltResps[sessionID],
				opcode.NewNodeBuiltResp(u.Node()),
			)

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