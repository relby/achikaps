package opcode

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dominikbraun/graph"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/match_state"
	"github.com/relby/achikaps/node"
	"github.com/relby/achikaps/vec2"
)

type buildNodeReq struct {
	FromNodeID uint
	Type       uint
	Position   vec2.Vec2
	Data 	   any
}

func BuildNodeHandler(dispatcher runtime.MatchDispatcher, msg runtime.MatchData, state *match_state.State) error {
	userID := msg.GetUserId()
	g, ok := state.Graphs[userID]
	if !ok {
		return fmt.Errorf("can't find graph for user with id %q", userID)
	}

	var req buildNodeReq
	if err := json.Unmarshal(msg.GetData(), &req); err != nil {
		return sendErrorResp(fmt.Errorf("can't unmarshal data: %w", err), dispatcher, BuildNode, userID, state)
	}

	fromNodeID, err := node.NewID(req.FromNodeID)
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid FromNodeID: %w", err), dispatcher, BuildNode, userID, state)
	}

	fromNode, err := g.Node(fromNodeID)
	if errors.Is(err, graph.ErrEdgeNotFound) {
		return sendErrorResp(fmt.Errorf("node not found: %w", err), dispatcher, BuildNode, userID, state)
	} else if err != nil {
		return fmt.Errorf("can't get graph vertex: %w", err)
	}
	
	toNodeID, ok := state.NextNodeIDs[userID]
	if !ok {
		return fmt.Errorf("can't get next node id")
	}
	
	toNodeType, err := node.NewType(req.Type)	
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid Type: %w", err), dispatcher, BuildNode, userID, state)
	}

	var toNode *node.Node
	switch toNodeType {
	case node.TransitType:
		toNode = node.NewTransit(toNodeID, req.Position)
	case node.ProductionType:
		data, ok := req.Data.(node.ProductionTypeData)
		if !ok {
			return sendErrorResp(fmt.Errorf("invalid Data for ProductionType node"), dispatcher, BuildNode, userID, state)
		}
		
		toNode = node.NewProduction(toNodeID, req.Position, data)
	default:
		panic("unreachable")
	}

	// TODO: We also need to check intersections with graphs of enemies
	if g.NodeIntersectsAny(toNode) {
		return sendErrorResp(fmt.Errorf("new node intersects the graph"), dispatcher, BuildNode, userID, state)
	}
	
	if g.EdgeIntersectsAny(fromNode, toNode) {
		return sendErrorResp(fmt.Errorf("new edge intersects the graph"), dispatcher, BuildNode, userID, state)
	}
	
	if err := g.AddNodeFrom(fromNode, toNode); err != nil {
		return sendErrorResp(fmt.Errorf("can't add node: %w", err), dispatcher, BuildNode, userID, state)
	}

	state.NextNodeIDs[userID] += 1

	return sendOkResp(dispatcher, BuildNode, userID, state)
}