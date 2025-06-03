package opcode

import (
	"encoding/json"
	"fmt"

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

type buildNodeResp struct {
	FromNodeID node.ID
	Node *node.Node
}

func BuildNodeHandler(dispatcher runtime.MatchDispatcher, msg runtime.MatchData, state *match_state.State) error {
	userID := msg.GetUserId()

	var req buildNodeReq
	if err := json.Unmarshal(msg.GetData(), &req); err != nil {
		return sendErrorResp(fmt.Errorf("can't unmarshal data: %w", err), dispatcher, BuildNode, userID, state)
	}

	fromID, err := node.NewID(req.FromNodeID)
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid FromNodeID: %w", err), dispatcher, BuildNode, userID, state)
	}

	typ, err := node.NewType(req.Type)	
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid Type: %w", err), dispatcher, BuildNode, userID, state)
	}
	toNode, err := state.BuildNode(userID, fromID, typ, req.Position, req.Data)
	if err != nil {
		return sendErrorResp(fmt.Errorf("can't build node: %w", err), dispatcher, BuildNode, userID, state)
	}
	
	resp := &buildNodeResp{
		FromNodeID: fromID,
		Node: toNode,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("can't marshal resp: %w", err)
	}

	if err := dispatcher.BroadcastMessage(int64(BuildNode), respBytes, nil, state.Presences[userID], true); err != nil {
		return fmt.Errorf("can't broadcast message: %w", err)
	}

	return nil
}