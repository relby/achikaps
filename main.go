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
	"github.com/relby/achikaps/config"
	"github.com/relby/achikaps/graph"
	"github.com/relby/achikaps/match_state"
	"github.com/relby/achikaps/model"
	"github.com/relby/achikaps/opcode"
	"github.com/relby/achikaps/opcode_handler"
	"github.com/relby/achikaps/vec2"
	"github.com/relby/achikaps/win_condition"
)

type Match struct{}

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
		
		WinCondition: win_condition.New(model.JuiceMaterialType, 100),

		RespsWithOpcode: make(map[string][]*opcode.RespWithOpCode, len(players)),
	}
	
	for i, p := range players {
		sessionID := p.GetPresence().GetSessionId()

		root := model.NewNode(
			model.ID(1),
			sessionID,
			model.SandTransitNodeName,
			onCircle(i, len(players), config.PlayersStartRadius),
		)
		root.BuildFully()

		g := graph.New(root)

		state.Graphs[sessionID] = g

		for i := range 2 {
			var n *model.Node
			for {
				angle := rand.Float64() * 2 * math.Pi
				radius := config.MinNodeDistance + rand.Float64() * (config.MaxNodeDistance - config.MinNodeDistance)
				pos := vec2.New(
					root.Position().X + radius*math.Cos(angle),
					root.Position().Y + radius*math.Sin(angle),
				)
				
				n = model.NewNode(
					model.ID(i + 2),
					sessionID,
					model.SandTransitNodeName,
					pos,
				)
				n.BuildFully()
				
				if !g.NodeIntersectsAny(n) {
					break
				}
			}
			
			err := g.AddNodeFrom(root, n)
			assert.NoError(err)
		}

		state.NextNodeIDs[sessionID] = model.ID(4)
		
		state.Units[sessionID] = make(map[model.ID]*model.Unit)
		c := model.ID(1)
		for _, t := range []model.UnitType{model.IdleUnitType, model.BuilderUnitType, model.ProductionUnitType, model.TransportUnitType} {
			for range 10 {
				state.Units[sessionID][c] = model.NewUnit(c, sessionID, t, root)
				c += 1
			}
		}
		state.NextUnitIDs[sessionID] = c

		state.Materials[sessionID] = make(map[model.ID]*model.Material)
		c = model.ID(1)
		for _, t := range []model.MaterialType{model.GrassMaterialType, model.SandMaterialType, model.DewMaterialType, model.SeedMaterialType, model.SugarMaterialType, model.JuiceMaterialType, model.ChitinMaterialType, model.EggMaterialType, model.PheromoneMaterialType, model.AmberMaterialType} {
			for range 100 {
				state.Materials[sessionID][c] = model.NewMaterial(c, sessionID, t, root, false)
				c += 1
			}
		}
		state.NextMaterialIDs[sessionID] = c

		// state.Materials[sessionID] = make(map[model.ID]*model.Material, 28)
		// c := 1
		// for range 20 {
		// 	state.Materials[sessionID][model.ID(c)] = model.NewMaterial(model.ID(c), sessionID, model.GrassMaterialType, root, false)
		// 	c += 1
		// }
		// for range 6 {
		// 	state.Materials[sessionID][model.ID(c)] = model.NewMaterial(model.ID(c), sessionID, model.SandMaterialType, root, false)
		// 	c += 1
		// }
		// for range 2 {
		// 	state.Materials[sessionID][model.ID(c)] = model.NewMaterial(model.ID(c), sessionID, model.DewMaterialType, root, false)
		// 	c += 1
		// }
		
		// state.NextMaterialIDs[sessionID] = model.ID(c)
	}

	tickRate := config.TickRate // 1 tick per second = 1 MatchLoop func invocations per second
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
		sessionID := p.GetSessionId()
		matchState.Presences[sessionID] = p
	}

	resp := &opcode.InitialStateResp{}

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
	resp.WinCondition = matchState.WinCondition

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
		delete(matchState.Presences, p.GetSessionId())
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

		if err := opcode_handler.Handle(opCode, dispatcher, msg, matchState); err != nil {
			logger.Error(err.Error())
			return nil
		}
	}
	
	matchState.Tick()
	
	for _, p := range matchState.Presences {
		sessionID := p.GetSessionId()
		
		respsWithOpcode := matchState.RespsWithOpcode[sessionID]
		if len(respsWithOpcode) == 0 {
			continue
		}
		
		for _, rwo := range respsWithOpcode {
			resp, opCode := rwo.Resp, rwo.OpCode

			b, err := json.Marshal(resp)
			if err != nil {
				logger.Error("can't marshal resp: %w", err)
				return nil
			}

			if err := dispatcher.BroadcastMessage(int64(opCode), b, nil, p, true); err != nil {
				logger.Error("can't broadcast message: %w", err)
				return nil
			}
		}
		
		matchState.RespsWithOpcode[sessionID] = matchState.RespsWithOpcode[sessionID][:0]
	}

	for sessionID, playerMaterials := range matchState.Materials {
		c := 0
		for _, m := range playerMaterials {
			if m.Type() == matchState.WinCondition.MaterialType {
				c += 1
			}
		}
		
		if c >= matchState.WinCondition.Count {
			b, err := json.Marshal(opcode.NewWinResp(sessionID))
			if err != nil {
				logger.Error("can't unmarshal state: %w", err)
				return nil
			}

			if err := dispatcher.BroadcastMessage(int64(opcode.Win), b, nil, matchState.Presences[sessionID], true); err != nil {
				logger.Error("can't broadcast message: %w", err)
				return nil
			}
			
			// This indicate that the match is over
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
