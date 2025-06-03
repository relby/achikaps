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
	
	if err := state.BuildNode(userID, fromID, typ, req.Position, req.Data); err != nil {
		return sendErrorResp(fmt.Errorf("can't build node: %w", err), dispatcher, BuildNode, userID, state)
	}
	
	return sendOkResp(dispatcher, BuildNode, userID, state)
}