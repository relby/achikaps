package opcode

import (
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/match_state"
	"github.com/relby/achikaps/model"
)

type changeUnitTypeReq struct {
	ID uint
	Type uint
}

type changeUnitTypeResp struct {
	Unit *model.Unit
}

func ChangeUnitTypeHandler(dispatcher runtime.MatchDispatcher, msg runtime.MatchData, state *match_state.State) error {
	userID := msg.GetUserId()
	
	var req changeUnitTypeReq
	if err := json.Unmarshal(msg.GetData(), &req); err != nil {
		return sendErrorResp(fmt.Errorf("can't unmarshal data: %w", err), dispatcher, BuildNode, userID, state)
	}

	id, err := model.NewID(req.ID)
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid ID: %w", err), dispatcher, BuildNode, userID, state)
	}

	typ, err := model.NewUnitType(req.Type)	
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid Type: %w", err), dispatcher, BuildNode, userID, state)
	}
	
	
	u, err := state.ChangeUnitType(userID, id, typ)
	if err != nil {
		return sendErrorResp(fmt.Errorf("can't change unit type: %w", err), dispatcher, BuildNode, userID, state)
	}
	
	resp := &changeUnitTypeResp{
		Unit: u,
	}
	
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("can't marshal resp: %w", err)
	}

	if err := dispatcher.BroadcastMessage(int64(ChangeUnitType), respBytes, nil, state.Presences[userID], true); err != nil {
		return fmt.Errorf("can't broadcast message: %w", err)
	}

	return nil
}