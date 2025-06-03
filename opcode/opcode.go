package opcode

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/match_state"
)

type OpCode int64

func NewOpCode(v int64) (OpCode, error) {
	switch v := OpCode(v); v {
	case InitialState, BuildNode, UnitActionExecute:
		return v, nil
	}

	return 0, errors.New("invalid op code")
}

// TODO: Refactor all opcodes
const (
	InitialState OpCode = iota + 1
	BuildNode
	UnitActionExecute
)

type Handler func(runtime.MatchDispatcher, runtime.MatchData, *match_state.State) error

var Handlers = map[OpCode]Handler{
	BuildNode: BuildNodeHandler,
}

type okResp struct{}

type errorResp struct {
	Error error `json:"error"`
}

func sendOkResp(dispatcher runtime.MatchDispatcher, opCode OpCode, userID string, state *match_state.State) error {
	resp, _ := json.Marshal(okResp{})
	if err := dispatcher.BroadcastMessage(int64(opCode), resp, nil, state.Presences[userID], true); err != nil {
		return fmt.Errorf("can't broadcast message: %w", err)
	}

	return nil
}

func sendErrorResp(err error, dispatcher runtime.MatchDispatcher, opCode OpCode, userID string, state *match_state.State) error {
	resp, err := json.Marshal(errorResp{Error: err})
	if err != nil {
		return fmt.Errorf("can't unmarshal state: %w", err)
	}

	if err := dispatcher.BroadcastMessage(int64(opCode), resp, []runtime.Presence{state.Presences[userID]}, state.Presences[userID], true); err != nil {
		return fmt.Errorf("can't broadcast message: %w", err)
	}

	return nil
}

func Handle(opCode OpCode, dispatcher runtime.MatchDispatcher, msg runtime.MatchData, state *match_state.State) error {
	handler, ok := Handlers[opCode]
	if !ok {
		return fmt.Errorf("unknown op code: %d", opCode)
	}

	return handler(dispatcher, msg, state)
}
