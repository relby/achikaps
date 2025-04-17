package unit

import (
	"encoding/json"

	"github.com/gammazero/deque"
	"github.com/relby/achikaps/node"
)

const (
	DefaultSpeed = 2
)

type Type uint

const (
	IdleType Type = iota + 1
	ProductionType
	BuilderType
)

type Unit struct {
	Type Type
	Node *node.Node
	Actions deque.Deque[*Action]
}

func New(typ Type, n *node.Node) *Unit {
	return &Unit{
		typ,
		n,
		deque.Deque[*Action]{},
	}
}

func (u *Unit) MarshalJSON() ([]byte, error) {
	type unitJSON struct {
		Type    Type
		Node    *node.Node
		Action  any
	}

	// Get only the first action for JSON serialization
	var action any
	if u.Actions.Len() > 0 {
		action = u.Actions.At(0)
	}

	// Create a custom representation for JSON serialization
	unitData := unitJSON{
		Type:    u.Type,
		Node:    u.Node,
		Action:  action,
	}

	return json.Marshal(unitData)
}