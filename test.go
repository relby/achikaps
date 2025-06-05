package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/davecgh/go-spew/spew"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/assert"
	"github.com/relby/achikaps/graph"
	"github.com/relby/achikaps/match_state"
	"github.com/relby/achikaps/model"
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
		NextNodeIDs: make(map[string]model.ID),

		Units: make(map[string]map[model.ID]*model.Unit),
		NextUnitIDs: make(map[string]model.ID),

		Materials: make(map[string]map[model.ID]*model.Material),
		NextMaterialIDs: make(map[string]model.ID),
		
		ClientUpdates: make(map[string][]*model.UnitAction),
	}

	state.Presences[id] = &MyPresence{username: "test"}

	root := model.NewNode(
		model.ID(1),
		model.SandTransitNodeName,
		vec2.New(0, 0),
	)

	root.BuildFully()
	
	g := graph.New(root)

	state.Graphs[id] = g

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

	state.NextNodeIDs[id] = model.ID(4)
	
	state.Units[id] = map[model.ID]*model.Unit{
		1: model.NewUnit(1, model.IdleUnitType, root),
		2: model.NewUnit(2, model.ProductionUnitType, root),
		3: model.NewUnit(3, model.BuilderUnitType, root),
		4: model.NewUnit(4, model.TransportUnitType, root),
	}
	
	state.NextUnitIDs[id] = model.ID(5)

	state.Materials[id] = make(map[model.ID]*model.Material, 28)
	c := 1
	for range 20 {
		state.Materials[id][model.ID(c)] = model.NewMaterial(model.ID(c), model.GrassMaterialType, root, false)
		c += 1
	}
	for range 6 {
		state.Materials[id][model.ID(c)] = model.NewMaterial(model.ID(c), model.SandMaterialType, root, false)
		c += 1
	}
	for range 2 {
		state.Materials[id][model.ID(c)] = model.NewMaterial(model.ID(c), model.DewMaterialType, root, false)
		c += 1
	}

	state.NextMaterialIDs[id] = model.ID(c)

	n, err := state.BuildNode(id, root.ID(), model.GrassFieldNodeName, vec2.New(50, 50))
	if err != nil {
		fmt.Println(err)
		return
	}

	state.Materials[id][model.ID(c)] = model.NewMaterial(model.ID(c), model.GrassMaterialType, n, true)
	c += 1
	state.Materials[id][model.ID(c)] = model.NewMaterial(model.ID(c), model.GrassMaterialType, n, true)
	c += 1
	state.Materials[id][model.ID(c)] = model.NewMaterial(model.ID(c), model.GrassMaterialType, n, true)
	c += 1
	// state.Materials[id][model.ID(c)] = model.NewMaterial(model.ID(c), model.GrassMaterialType, n)
	// c += 1

	// if _, err := state.BuildNode(id, root.ID(), model.WellNodeName, vec2.New(-50, -50)); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	
	// if _, err := state.ChangeUnitType(id, 4, model.IdleUnitType); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	
	c = 0
	for {
		state.Tick()
		
		// spew.Dump(state.Units[id][3])
		// if c == 5 {
		// 	break
		// }
		if len(state.Graphs[id].BuildingNodes()) == 0 {
			spew.Dump(n)
			break
		}

		c += 1
	}
}