package main

import (
	"context"
	"database/sql"

	"github.com/dominikbraun/graph"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
)

type Match struct{}

type MatchState struct {
	Presences   map[string]runtime.Presence
	Graphs      map[string]graph.Graph[NodeID, Node]
	NextNodeIDs map[string]NodeID
}

const StartRadius = 100.0

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	// TODO: handle errors
	players := params["players"].([]runtime.MatchmakerEntry)

	state := &MatchState{
		Presences:   make(map[string]runtime.Presence, len(players)),
		Graphs:      make(map[string]graph.Graph[NodeID, Node], len(players)),
		NextNodeIDs: make(map[string]NodeID, len(players)),
	}

	for i, p := range players {
		userID := p.GetPresence().GetUserId()

		g := graph.New(func(v Node) NodeID { return v.ID })
		root := NewNode(
			1,
			TransitNodeType,
			getPositionOnCircle(i, len(players), StartRadius), // TODO: refactor radius to constant
		)
		_ = g.AddVertex(root)

		state.Graphs[userID] = g

		state.NextNodeIDs[userID] = root.ID + 1
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

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	matchState, ok := state.(*MatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	for _, msg := range messages {
		logger.Info("got message: %s", string(msg.GetData()))
		opCode := OpCode(msg.GetOpCode()) // TODO: convert properly

		if err := HandleOpCode(opCode, dispatcher, msg, matchState); err != nil {
			logger.Error(err.Error())
			return nil
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

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if err := initializer.RegisterMatch("achikaps", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &Match{}, nil
	}); err != nil {
		logger.Error("unable to register: %v", err)
		return err
	}

	initializer.RegisterBeforeRt("MatchmakerAdd", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *rtapi.Envelope) (*rtapi.Envelope, error) {
		message, ok := in.Message.(*rtapi.Envelope_MatchmakerAdd)
		if !ok {
			return nil, runtime.NewError("internal server error", int(codes.Internal))
		}

		message.MatchmakerAdd.Query = "*"
		message.MatchmakerAdd.MinCount = 2
		message.MatchmakerAdd.MaxCount = 6

		return in, nil
	})

	if err := initializer.RegisterMatchmakerMatched(func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {
		matchID, err := nk.MatchCreate(ctx, "achikaps", map[string]interface{}{"players": entries})
		if err != nil {
			return "", runtime.NewError("unable to create match", int(codes.Internal))
		}

		return matchID, nil
	}); err != nil {
		logger.Error("unable to register matchmaker matched hook: %v", err)
		return err
	}

	return nil
}
