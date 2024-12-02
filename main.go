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
	X, Y, Z float32
}

type NodeID uint

func NewNodeID(v uint) NodeID {
	return NodeID(v)
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
	OpCodeBuildPlatform = 1
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
	label := "TODO"
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
			Position{0, 0, 0},
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

type BuildNodeReq struct {
	FromNodeID uint
	Type       uint
	Position   Position
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	matchState, ok := state.(*MatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	for _, msg := range messages {
		logger.Info("got message: %v", msg, msg.GetOpCode())
		userID := msg.GetUserId()

		if opCode := msg.GetOpCode(); opCode == OpCodeBuildPlatform {
			g, ok := matchState.Graphs[userID]
			if !ok {
				logger.Error("can't find graph for user with id %q", userID)
				return nil
			}

			var req BuildNodeReq
			if err := json.Unmarshal(msg.GetData(), &req); err != nil {
				logger.Warn("can't unmarshal data: %v", err)
				return nil
			}

			fromNodeID := NewNodeID(req.FromNodeID)
			if _, err := g.Vertex(fromNodeID); err != nil {
				if errors.Is(err, graph.ErrEdgeNotFound) {
					if err := dispatcher.BroadcastMessage(OpCodeBuildPlatform, []byte(`{"success":false,"reason":"node not found"}`), nil, matchState.Presences[userID], true); err != nil {
						logger.Error("can't broadcast message: %v", err)
						return nil
					}
					continue
				}
				logger.Error("unkown error happened: %v", err)
			}

			typ, err := NewNodeType(req.Type)
			if err != nil {
				if err := dispatcher.BroadcastMessage(OpCodeBuildPlatform, []byte(`{"success":false,"reason":"invalid node type"}`), nil, matchState.Presences[userID], true); err != nil {
					logger.Error("can't broadcast message: %v", err)
					return nil
				}
				continue
			}

			// TODO: add validation for intersection

			nodeID, ok := matchState.NextNodeIDs[userID]
			if !ok {
				panic("TODO")
			}

			_ = g.AddVertex(NewNode(
				nodeID,
				typ,
				req.Position,
			))
			_ = g.AddEdge(fromNodeID, nodeID)

			if err := dispatcher.BroadcastMessage(OpCodeBuildPlatform, []byte(`{"success":true}`), nil, matchState.Presences[userID], true); err != nil {
				logger.Error("can't broadcast message: %v", err)
				return nil
			}
		} else {
			logger.Warn("invalid opcode: %v", opCode)
		}
	}

	// edgeCount := 0
	// if len(matchState.graphs) > 0 {
	// 	e, _ := slices.Collect(maps.Values(matchState.graphs))[0].Edges()
	// 	edgeCount = len(e)
	// }

	// b, err := json.Marshal(map[string]any{
	// 	"edgeCount": edgeCount,
	// })
	// if err != nil {
	// 	logger.Error("failed to marshal graphs")
	// 	return nil
	// }
	// if err := dispatcher.BroadcastMessage(69, b, nil, nil, true); err != nil {
	// 	logger.Error("failed to broadcast message")
	// 	return nil
	// }

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

	modulename := "match"

	if matchId, err := nk.MatchCreate(ctx, modulename, params); err != nil {
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
