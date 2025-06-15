package opcode_handler

import (
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/assert"
	"github.com/relby/achikaps/match_state"
	"github.com/relby/achikaps/opcode"
)

type Handler func(runtime.MatchDispatcher, runtime.MatchData, *match_state.State) error

var Handlers = map[opcode.OpCode]Handler{
	opcode.BuildNode: BuildNodeHandler,
	opcode.ChangeUnitType: ChangeUnitTypeHandler,
}

type okResp struct{}

type errorResp struct {
	Error string `json:"error"`
}

func sendOkResp(dispatcher runtime.MatchDispatcher, opCode opcode.OpCode, sessionID string, state *match_state.State) error {
	resp, err := json.Marshal(okResp{})
	assert.NoError(err)

	if err := dispatcher.BroadcastMessage(int64(opCode), resp, nil, state.Presences[sessionID], true); err != nil {
		return fmt.Errorf("can't broadcast message: %w", err)
	}

	return nil
}

func sendErrorResp(err error, dispatcher runtime.MatchDispatcher, opCode opcode.OpCode, sessionID string, state *match_state.State) error {
	resp, err := json.Marshal(errorResp{Error: err.Error()})
	assert.NoError(err)

	if err := dispatcher.BroadcastMessage(int64(opCode), resp, []runtime.Presence{state.Presences[sessionID]}, state.Presences[sessionID], true); err != nil {
		return fmt.Errorf("can't broadcast message: %w", err)
	}

	return nil
}

func Handle(opCode opcode.OpCode, dispatcher runtime.MatchDispatcher, msg runtime.MatchData, state *match_state.State) error {
	handler, ok := Handlers[opCode]
	if !ok {
		return fmt.Errorf("unknown op code: %d", opCode)
	}

	return handler(dispatcher, msg, state)
}
