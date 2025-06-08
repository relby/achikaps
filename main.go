package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"maps"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/assert"
	"github.com/relby/achikaps/consts"
	"github.com/relby/achikaps/graph"
	"github.com/relby/achikaps/match_state"
	"github.com/relby/achikaps/model"
	"github.com/relby/achikaps/opcode"
	"github.com/relby/achikaps/vec2"
)

type Match struct{}

const StartRadius = model.DefaultNodeRadius * 20
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
		NextNodeIDs: make(map[string]model.ID, len(players)),

		Units: make(map[string]map[model.ID]*model.Unit, len(players)),
		NextUnitIDs: make(map[string]model.ID),

		Materials: make(map[string]map[model.ID]*model.Material, len(players)),
		NextMaterialIDs: make(map[string]model.ID, len(players)),

		ClientUpdates: make(map[string][]*match_state.ClientUpdate, len(players)),
	}
	
	for i, p := range players {
		userID := p.GetPresence().GetUserId()

		root := model.NewNode(
			model.ID(1),
			model.SandTransitNodeName,
			onCircle(i, len(players), StartRadius),
		)
		root.BuildFully()

		g := graph.New(root)

		state.Graphs[userID] = g

		for i := range 2 {
			angle := rand.Float64() * 2 * math.Pi
			radius := model.DefaultNodeRadius * (4 + rand.Float64()*4) // Random radius between 4-8 times DefaultNodeRadius
			pos := vec2.New(
				root.Position().X + radius*math.Cos(angle),
				root.Position().Y + radius*math.Sin(angle),
			)
			
			n := model.NewNode(
				model.ID(i + 2),
				model.SandTransitNodeName,
				pos,
			)
			n.BuildFully()
			
			err := g.AddNodeFrom(root, n)
			assert.NoError(err)
		}

		state.NextNodeIDs[userID] = model.ID(3)
		
		state.Units[userID] = map[model.ID]*model.Unit{
			1: model.NewUnit(1, model.IdleUnitType, root),
			2: model.NewUnit(2, model.ProductionUnitType, root),
			3: model.NewUnit(3, model.BuilderUnitType, root),
			4: model.NewUnit(4, model.TransportUnitType, root),
		}
		
		state.NextUnitIDs[userID] = model.ID(5)
		
		state.Materials[userID] = make(map[model.ID]*model.Material, 28)
		c := 1
		for range 20 {
			state.Materials[userID][model.ID(c)] = model.NewMaterial(model.ID(c), model.GrassMaterialType, root, false)
			c += 1
		}
		for range 6 {
			state.Materials[userID][model.ID(c)] = model.NewMaterial(model.ID(c), model.SandMaterialType, root, false)
			c += 1
		}
		for range 2 {
			state.Materials[userID][model.ID(c)] = model.NewMaterial(model.ID(c), model.DewMaterialType, root, false)
			c += 1
		}
	}

	tickRate := consts.TickRate // 1 tick per second = 1 MatchLoop func invocations per second
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

	type initialStateResp struct {
		Nodes map[string]map[model.ID]*model.Node
		Connections map[string]map[model.ID][]model.ID
		Units map[string]map[model.ID]*model.Unit
		Materials map[string]map[model.ID]*model.Material
	}

	resp := &initialStateResp{}

	resp.Nodes = make(map[string]map[model.ID]*model.Node, len(matchState.Graphs))
	resp.Connections = make(map[string]map[model.ID][]model.ID, len(matchState.Graphs))
	for uID, g := range matchState.Graphs {
		resp.Nodes[uID] = g.Nodes()

		am := g.AdjacencyMap()
		resp.Connections[uID] = make(map[model.ID][]model.ID, len(am))
		for k, v := range am {
			resp.Connections[uID][k] = slices.Collect(maps.Keys(v))
		}
	}

	resp.Units = matchState.Units
	resp.Materials = matchState.Materials

	respBytes, err := json.Marshal(resp)
	if err != nil {
		logger.Error("can't marshal state: %w", err)
		return nil
	}

	if err := dispatcher.BroadcastMessage(int64(opcode.InitialState), respBytes, nil, nil, true); err != nil {
		logger.Error("can't broadcast message state: %w", err)
		return nil
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
	
	for _, p := range matchState.Presences {
		userID := p.GetUserId()
		
		actions := matchState.ClientUpdates[userID]
		if len(actions) == 0 {
			continue
		}
		
		for _, a := range actions {
			b, err := json.Marshal(a)
			if err != nil {
				logger.Error("can't unmarshal state: %w", err)
				return nil
			}

			if err := dispatcher.BroadcastMessage(int64(opcode.UnitActionExecute), b, nil, p, true); err != nil {
				logger.Error("can't broadcast message state: %w", err)
				return nil
			}
		}
		
		matchState.ClientUpdates[userID] = matchState.ClientUpdates[userID][:0]
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
			return nil, runtime.NewError("internal server error", 13)
		}

		message.MatchmakerAdd.Query = "*"
		message.MatchmakerAdd.MinCount = 1
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
