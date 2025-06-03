package unit

import (
	"encoding/json"

	"github.com/gammazero/deque"
	"github.com/relby/achikaps/material"
	"github.com/relby/achikaps/node"
	"github.com/relby/achikaps/unit_action"
)

const (
	DefaultSpeed = 2
)

type Type uint

const (
	IdleType Type = iota + 1
	ProductionType
	BuilderType
	TransportType
)

type Unit struct {
	Type Type
	Node *node.Node
	Data any
	Actions deque.Deque[*unit_action.UnitAction]
}

func new(typ Type, n *node.Node, data any) *Unit {
	return &Unit{
		typ,
		n,
		data,
		deque.Deque[*unit_action.UnitAction]{},
	}
}

func NewIdle(n *node.Node) *Unit {
	return new(IdleType, n, nil)
}

func NewProduction(n *node.Node) *Unit {
	return new(ProductionType, n, nil)
}

func NewBuilder(n *node.Node) *Unit {
	return new(BuilderType, n, nil)
}

type TransportData struct {
	Material *material.Material
}

func NewTransportData(m *material.Material) *TransportData {
	return &TransportData{m}
}

func NewTranport(n *node.Node, data *TransportData) *Unit {
	return new(TransportType, n, data)
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