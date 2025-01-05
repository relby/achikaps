package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/dominikbraun/graph"
	"github.com/heroiclabs/nakama-common/runtime"
)

type Position struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type NodeID uint

func NewNodeID(v uint) (NodeID, error) {
	if v == 0 {
		return 0, errors.New("invalid node id")
	}
	return NodeID(v), nil
}

type NodeType uint

const (
	TransitNodeType NodeType = iota + 1
	FactoryNodeType
	StorageNodeType
	DefenseNodeType
)

func NewNodeType(v uint) (NodeType, error) {
	switch v := NodeType(v); v {
	case TransitNodeType,
		FactoryNodeType,
		StorageNodeType,
		DefenseNodeType:
		return v, nil
	}

	return 0, errors.New("invalid node type")
}

var NodeTypeRadiuses = map[NodeType]float32{
	TransitNodeType: 1,
	FactoryNodeType: 2,
	StorageNodeType: 3,
	DefenseNodeType: 1,
}

type Node struct {
	ID       NodeID
	Type     NodeType
	Position Position
	Radius   float32
}

func NewNode(id NodeID, typ NodeType, position Position) Node {
	return Node{
		id,
		typ,
		position,
		NodeTypeRadiuses[typ],
	}
}

type OpCode int

const (
	OpCodeBuildNode = 1
)

type Match struct{}

type MatchState struct {
	Presences   map[string]runtime.Presence
	Graphs      map[string]graph.Graph[NodeID, Node]
	NextNodeIDs map[string]NodeID
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &MatchState{
		Presences:   make(map[string]runtime.Presence),
		Graphs:      make(map[string]graph.Graph[NodeID, Node]),
		NextNodeIDs: make(map[string]NodeID),
	}

	tickRate := 1 // 1 tick per second = 1 MatchLoop func invocations per second
	label := "achikaps"
	return state, tickRate, label
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	ok := true

	return state, ok, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	matchState, ok := state.(*MatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	for _, p := range presences {
		userID := p.GetUserId()
		matchState.Presences[userID] = p

		g := graph.New(func(v Node) NodeID { return v.ID })
		root := NewNode(
			1,
			TransitNodeType,
			Position{0, 0},
		)
		_ = g.AddVertex(root)
		matchState.Graphs[userID] = g

		matchState.NextNodeIDs[userID] = root.ID + 1
	}

	return matchState
}

func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	matchState, ok := state.(*MatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	for _, p := range presences {
		delete(matchState.Presences, p.GetUserId())
		delete(matchState.Graphs, p.GetUserId())
		delete(matchState.NextNodeIDs, p.GetUserId())
	}

	return matchState
}

type buildNodeReq struct {
	FromNodeID uint     `json:"from_node_id"`
	Type       uint     `json:"type"`
	Position   Position `json:"position"`
}

type buildNodeResp struct {
	Error string `json:"error"`
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	matchState, ok := state.(*MatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	for _, msg := range messages {
		logger.Info("got message: %s", string(msg.GetData()))
		userID := msg.GetUserId()

		if opCode := msg.GetOpCode(); opCode == OpCodeBuildNode {
			g, ok := matchState.Graphs[userID]
			if !ok {
				logger.Error("can't find graph for user with id %q", userID)
				return nil
			}

			var req buildNodeReq
			if err := json.Unmarshal(msg.GetData(), &req); err != nil {
				logger.Warn("can't unmarshal data: %v", err)
				continue
			}

			fromNodeID, err := NewNodeID(req.FromNodeID)
			if err != nil {
				logger.Warn("invalid from_node_id: %v", err)
				continue
			}

			if _, err := g.Vertex(fromNodeID); err != nil {
				resp, _ := json.Marshal(buildNodeResp{Error: "node not found"})
				if errors.Is(err, graph.ErrEdgeNotFound) {
					if err := dispatcher.BroadcastMessage(OpCodeBuildNode, resp, nil, matchState.Presences[userID], true); err != nil {
						logger.Error("can't broadcast message: %v", err)
						return nil
					}
					continue
				}
				logger.Error("unkown error happened: %v", err)
			}

			t, err := NewNodeType(req.Type)
			if err != nil {
				resp, _ := json.Marshal(buildNodeResp{Error: "invalid node type"})
				if err := dispatcher.BroadcastMessage(OpCodeBuildNode, resp, nil, matchState.Presences[userID], true); err != nil {
					logger.Error("can't broadcast message: %v", err)
					return nil
				}
				continue
			}

			// TODO: add validation for intersection

			nodeID, ok := matchState.NextNodeIDs[userID]
			if !ok {
				logger.Error("can't get next node id")
			}

			_ = g.AddVertex(NewNode(
				nodeID,
				t,
				req.Position,
			))
			_ = g.AddEdge(fromNodeID, nodeID)

			resp, _ := json.Marshal(buildNodeResp{Error: ""})
			if err := dispatcher.BroadcastMessage(OpCodeBuildNode, resp, nil, matchState.Presences[userID], true); err != nil {
				logger.Error("can't broadcast message: %v", err)
				return nil
			}
		} else {
			logger.Warn("invalid opcode: %v", opCode)
		}
	}

	return matchState
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	return state
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}

func newMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
	return &Match{}, nil
}

func CreateMatchRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	params := make(map[string]interface{})

	if err := json.Unmarshal([]byte(payload), &params); err != nil {
		return "", err
	}

	moduleName := "match"

	if matchId, err := nk.MatchCreate(ctx, moduleName, params); err != nil {
		logger.Error("failed to create a match: %v", err)
		return "", err
	} else {
		logger.Info("matchId: %s", matchId)
		return matchId, nil
	}
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if err := initializer.RegisterMatch("match", newMatch); err != nil {
		logger.Error("unable to register: %v", err)
		return err
	}

	// Register as RPC function, this call should be in InitModule.
	if err := initializer.RegisterRpc("create_match_rpc", CreateMatchRPC); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	return nil
}
