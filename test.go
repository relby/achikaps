package main

import (
	"crypto/rand"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/graph"
	"github.com/relby/achikaps/match_state"
	"github.com/relby/achikaps/material"
	"github.com/relby/achikaps/node"
	"github.com/relby/achikaps/unit"
	"github.com/relby/achikaps/vec2"
)

func generateUUID() string {
    b := make([]byte, 16)
    _, err := rand.Read(b)
    if err != nil {
        panic(err)
    }
    return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type MyPresence struct {
	username string
}

func (mp *MyPresence) GetHidden() bool {return false}
func (mp *MyPresence) GetPersistence() bool {return true}
func (mp *MyPresence) GetUsername() string {return mp.username}
func (mp *MyPresence) GetStatus() string {return "ok"}
func (mp *MyPresence) GetReason() runtime.PresenceReason {return 1}
func (mp *MyPresence) GetUserId() string {return ""}
func (mp *MyPresence) GetSessionId() string {return ""}
func (mp *MyPresence) GetNodeId() string {return ""}

func main() {
	id := generateUUID()
	state := match_state.State{
		Presences: make(map[string]runtime.Presence),
		Graphs: make(map[string]*graph.Graph),
		NextNodeIDs: make(map[string]node.ID),
		Units: make(map[string][]*unit.Unit),
		Materials: make(map[string][]*material.Material),
	}

	state.Presences[id] = &MyPresence{username: "test"}

	root := node.NewTransit(
		node.ID(1),
		vec2.New(0, 0),
	)

	root.BuildProgress = 1
	
	g := graph.New(root)

	state.Graphs[id] = g

	state.NextNodeIDs[id] = node.ID(2)
	
	state.Units[id] = []*unit.Unit{
		unit.NewIdle(root),
		unit.NewProduction(root),
		unit.NewBuilder(root),
		unit.NewTranport(root, unit.NewTransportData(nil)),
	}
	
	state.Materials[id] = nil

	if err := state.BuildNode(id, root.ID, node.TransitType, vec2.New(100, 100), nil); err != nil {
		fmt.Println(err)
		return
	}

	if err := state.BuildNode(id, root.ID, node.ProductionType, vec2.New(-100, -100), node.UnitProductionTypeData); err != nil {
		fmt.Println(err)
		return
	}
	
	c := 0
	for {
		state.Tick()
		builder := state.Units[id][2]
		spew.Dump(builder)
		if len(g.BuildingNodes()) == 0 {
			break
		}
		c += 1
	}
}