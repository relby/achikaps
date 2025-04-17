package match_state

import (
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/game"
	"github.com/relby/achikaps/graph"
	"github.com/relby/achikaps/node"
)

type State struct {
	Presences   map[string]runtime.Presence
	Graphs      map[string]*graph.Graph
	NextNodeIDs map[string]node.ID
	// TODO: I don't like that there are multiple game managers, this sounds weird
	GameManagers map[string]*game.Manager
}