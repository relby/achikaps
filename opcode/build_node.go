package opcode

import (
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/match_state"
	"github.com/relby/achikaps/model"
	"github.com/relby/achikaps/vec2"
)

type buildNodeReq struct {
	FromNodeID uint
	Name       uint
	Position   vec2.Vec2
}

type buildNodeResp struct {
	FromNodeID model.ID
	Node *model.Node
}

func BuildNodeHandler(dispatcher runtime.MatchDispatcher, msg runtime.MatchData, state *match_state.State) error {
	userID := msg.GetUserId()

	var req buildNodeReq
	if err := json.Unmarshal(msg.GetData(), &req); err != nil {
		return sendErrorResp(fmt.Errorf("can't unmarshal data: %w", err), dispatcher, BuildNode, userID, state)
	}

	fromID, err := model.NewID(req.FromNodeID)
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid FromNodeID: %w", err), dispatcher, BuildNode, userID, state)
	}

	name, err := model.NewNodeName(req.Name)	
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid Name: %w", err), dispatcher, BuildNode, userID, state)
	}
	toNode, err := state.BuildNode(userID, fromID, name, req.Position)
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