package main

import (
	"context"
	"database/sql"
	"math"

	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/graph"
	"github.com/relby/achikaps/match_state"
	"github.com/relby/achikaps/material"
	"github.com/relby/achikaps/node"
	"github.com/relby/achikaps/opcode"
	"github.com/relby/achikaps/unit"
	"github.com/relby/achikaps/vec2"
)

type Match struct{}

const StartRadius = 100.0
func onCircle(i, n int, r float64) vec2.Vec2 {
	angle := float64(i) * 2.0 * math.Pi / float64(n)

	return vec2.New(
		r * math.Cos(angle),
		r * math.Sin(angle),
	)
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	// TODO: handle errors
	players := params["players"].([]runtime.MatchmakerEntry)

	state := &match_state.State{
		Presences:   make(map[string]runtime.Presence, len(players)),
		Graphs: make(map[string]*graph.Graph, len(players)),
		NextNodeIDs: make(map[string]node.ID, len(players)),
		Units: make(map[string][]*unit.Unit),
		Materials: make(map[string][]*material.Material),
	}
	
	for i, p := range players {
		userID := p.GetPresence().GetUserId()

		root := node.NewTransit(
			node.ID(1),
			onCircle(i, len(players), StartRadius),
		)
		// TODO: It's a hack to make the root node built, we need to find a better way to do this
		root.BuildProgress = 1

		g := graph.New(root)

		state.Graphs[userID] = g

		state.NextNodeIDs[userID] = node.ID(2)
		
		state.Units[userID] = []*unit.Unit{
			unit.NewIdle(root),
			unit.NewProduction(root),
			unit.NewBuilder(root),
			unit.NewTranport(root, unit.NewTransportData(nil)),
		}
		
		state.Materials[userID] = nil
	}

	tickRate := 5 // 1 tick per second = 1 MatchLoop func invocations per second
	label := "achikaps"
	return state, tickRate, label
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	ok := true

	return state, ok, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	matchState, ok := state.(*match_state.State)
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
	matchState, ok := state.(*match_state.State)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	// TODO: research if we need to delete only presences or all data
	for _, p := range presences {
		delete(matchState.Presences, p.GetUserId())
	}

	return matchState
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	matchState, ok := state.(*match_state.State)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	for _, msg := range messages {
		logger.Info("got message: %s", string(msg.GetData()))
		opCode, err := opcode.NewOpCode(msg.GetOpCode())
		if err != nil {
			logger.Error("invalid op code: %v", err)
			return nil
		}

		if err := opcode.Handle(opCode, dispatcher, msg, matchState); err != nil {
			logger.Error(err.Error())
			return nil
		}
	}
	
	matchState.Tick()
	
	// TODO: Finish this
	// type resp struct{
	// 	Units map[string][]*unit.Unit
	// }
	// r := resp{
	// 	Units: make(map[string][]*unit.Unit),
	// }
	
	// for _, presence := range matchState.Presences {
	// 	userID := presence.GetUserId()
	// 	for _, u := range matchState.Units {
	// 		r.Units[userID] = append(r.Units[userID], u)
	// 	}
	// }

	// b, err := json.Marshal(r)
	// if err != nil {
	// 	return fmt.Errorf("can't unmarshal state: %w", err)
	// }

	// x := slices.Collect(maps.Values(matchState.Presences))

	// if err := dispatcher.BroadcastMessage(3, b, x, nil, true); err != nil {
	// 	return fmt.Errorf("can't broadcast message state: %w", err)
	// }

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
			return nil, runtime.NewError("internal server error", 13)
		}

		message.MatchmakerAdd.Query = "*"
		message.MatchmakerAdd.MinCount = 2
		message.MatchmakerAdd.MaxCount = 6

		return in, nil
	})

	if err := initializer.RegisterMatchmakerMatched(func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {
		matchID, err := nk.MatchCreate(ctx, "achikaps", map[string]interface{}{"players": entries})
		if err != nil {
			return "", runtime.NewError("unable to create match", 13)
		}

		return matchID, nil
	}); err != nil {
		logger.Error("unable to register matchmaker matched hook: %v", err)
		return err
	}

	return nil
}
