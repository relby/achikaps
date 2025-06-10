package opcode_handler

import (
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/match_state"
	"github.com/relby/achikaps/model"
	"github.com/relby/achikaps/opcode"
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
	sessionID := msg.GetSessionId()

	var req buildNodeReq
	if err := json.Unmarshal(msg.GetData(), &req); err != nil {
		return sendErrorResp(fmt.Errorf("can't unmarshal data: %w", err), dispatcher, opcode.BuildNode, sessionID, state)
	}

	fromID, err := model.NewID(req.FromNodeID)
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid FromNodeID: %w", err), dispatcher, opcode.BuildNode, sessionID, state)
	}

	name, err := model.NewNodeName(req.Name)	
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid Name: %w", err), dispatcher, opcode.BuildNode, sessionID, state)
	}
	toNode, err := state.BuildNode(sessionID, fromID, name, req.Position)
	if err != nil {
		return sendErrorResp(fmt.Errorf("can't build node: %w", err), dispatcher, opcode.BuildNode, sessionID, state)
	}
	
	resp := &buildNodeResp{
		FromNodeID: fromID,
		Node: toNode,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("can't marshal resp: %w", err)
	}

	if err := dispatcher.BroadcastMessage(int64(opcode.BuildNode), respBytes, nil, state.Presences[sessionID], true); err != nil {
		return fmt.Errorf("can't broadcast message: %w", err)
	}

	return nil
}