package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dominikbraun/graph"
	"github.com/heroiclabs/nakama-common/runtime"
)

type OpCode int

const (
	OpCodeGetState = iota
	OpCodeBuildNode
)

type OpCodeHandler func(runtime.MatchDispatcher, runtime.MatchData, *MatchState) error

var OpCodeHandlers = map[OpCode]OpCodeHandler{
	OpCodeGetState:  OpCodeGetStateHandler,
	OpCodeBuildNode: OpCodeBuildNodeHandler,
}

func OpCodeGetStateHandler(dispatcher runtime.MatchDispatcher, msg runtime.MatchData, state *MatchState) error {
	b, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("can't unmarshal state: %w", err)
	}
	if err := dispatcher.BroadcastMessage(OpCodeGetState, b, []runtime.Presence{state.Presences[msg.GetUserId()]}, nil, true); err != nil {
		return fmt.Errorf("can't broadcast message state: %w", err)
	}

	return nil
}

type buildNodeReq struct {
	FromNodeID uint
	Type       uint
	Position   Position
}

type okResp struct{}

type errorResp struct {
	Error error `json:"error"`
}

func sendErrorResp(err error, dispatcher runtime.MatchDispatcher, opCode OpCode, userID string, state *MatchState) error {
	resp, err := json.Marshal(errorResp{Error: err})
	if err != nil {
		return fmt.Errorf("can't unmarshal state: %w", err)
	}

	if err := dispatcher.BroadcastMessage(OpCodeBuildNode, resp, []runtime.Presence{state.Presences[userID]}, state.Presences[userID], true); err != nil {
		return fmt.Errorf("can't broadcast message: %w", err)
	}

	return nil
}

func OpCodeBuildNodeHandler(dispatcher runtime.MatchDispatcher, msg runtime.MatchData, state *MatchState) error {
	userID := msg.GetUserId()
	g, ok := state.Graphs[userID]
	if !ok {
		return fmt.Errorf("can't find graph for user with id %q", userID)
	}

	var req buildNodeReq
	if err := json.Unmarshal(msg.GetData(), &req); err != nil {
		return sendErrorResp(fmt.Errorf("can't unmarshal data: %w", err), dispatcher, OpCodeBuildNode, userID, state)
	}

	fromNodeID, err := NewNodeID(req.FromNodeID)
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid FromNodeID: %w", err), dispatcher, OpCodeBuildNode, userID, state)
	}

	_, err = g.Vertex(fromNodeID)
	if errors.Is(err, graph.ErrEdgeNotFound) {
		return sendErrorResp(fmt.Errorf("node not found: %w", err), dispatcher, OpCodeBuildNode, userID, state)
	} else if err != nil {
		return fmt.Errorf("can't get vertex: %w", err)
	}

	t, err := NewNodeType(req.Type)
	if err != nil {
		return sendErrorResp(fmt.Errorf("invalid node type: %w", err), dispatcher, OpCodeBuildNode, userID, state)
	}

	// TODO: add validation for intersection

	nodeID, ok := state.NextNodeIDs[userID]
	if !ok {
		return fmt.Errorf("can't get next node id")
	}
	state.NextNodeIDs[userID] += 1

	_ = g.AddVertex(NewNode(
		nodeID,
		t,
		req.Position,
	))
	_ = g.AddEdge(fromNodeID, nodeID)

	resp, _ := json.Marshal(okResp{})
	if err := dispatcher.BroadcastMessage(OpCodeBuildNode, resp, nil, state.Presences[userID], true); err != nil {
		return fmt.Errorf("can't broadcast message: %w", err)
	}

	return nil
}

func HandleOpCode(opCode OpCode, dispatcher runtime.MatchDispatcher, msg runtime.MatchData, state *MatchState) error {
	return OpCodeHandlers[opCode](dispatcher, msg, state)
}
