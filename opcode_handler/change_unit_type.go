package opcode_handler

import (
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/match_state"
	"github.com/relby/achikaps/model"
	"github.com/relby/achikaps/opcode"
)

type changeUnitTypeReq struct {
	ID uint
	Type uint
}

type changeUnitTypeResp struct {
	Unit *model.Unit
}

func ChangeUnitTypeHandler(dispatcher runtime.MatchDispatcher, msg runtime.MatchData, state *match_state.State) error {
	sessionID := msg.GetSessionId()
	
	var req changeUnitTypeReq
	if err := json.Unmarshal(msg.GetData(), &req); err != nil {
		return sendErrorResp(fmt.Errorf("can't unmarshal data: %w", err), dispatcher, opcode.ChangeUnitType, sessionID, state)
	}

	id, err := model.NewID(req.ID)
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid ID: %w", err), dispatcher, opcode.ChangeUnitType, sessionID, state)
	}

	typ, err := model.NewUnitType(req.Type)	
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid Type: %w", err), dispatcher, opcode.ChangeUnitType, sessionID, state)
	}
	
	
	u, err := state.ChangeUnitType(sessionID, id, typ)
	if err != nil {
		return sendErrorResp(fmt.Errorf("can't change unit type: %w", err), dispatcher, opcode.ChangeUnitType, sessionID, state)
	}
	
	resp := &changeUnitTypeResp{
		Unit: u,
	}
	
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("can't marshal resp: %w", err)
	}

	if err := dispatcher.BroadcastMessage(int64(opcode.ChangeUnitType), respBytes, nil, state.Presences[sessionID], true); err != nil {
		return fmt.Errorf("can't broadcast message: %w", err)
	}

	return nil
}