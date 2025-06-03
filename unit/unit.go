package unit

import (
	"encoding/json"
	"errors"

	"github.com/gammazero/deque"
	"github.com/relby/achikaps/material"
	"github.com/relby/achikaps/node"
	"github.com/relby/achikaps/unit_action"
)

const (
	DefaultSpeed = 2
)

type ID uint

func NewID(v uint) (ID, error) {
	if v == 0 {
		return 0, errors.New("invalid node id")
	}
	return ID(v), nil
}

type Type uint

const (
	IdleType Type = iota + 1
	ProductionType
	BuilderType
	TransportType
)
func NewType(v uint) (Type, error) {
	switch v := Type(v); v {
		case IdleType,
			ProductionType,
			BuilderType,
			TransportType:
			return v, nil
	}
	
	return 0, errors.New("invalid unit type")
}

type Unit struct {
	ID ID
	Type Type
	Node *node.Node
	Data any
	Actions deque.Deque[*unit_action.UnitAction]
}

func new(id ID, typ Type, n *node.Node, data any) *Unit {
	return &Unit{
		id,
		typ,
		n,
		data,
		deque.Deque[*unit_action.UnitAction]{},
	}
}


func NewIdle(id ID, n *node.Node) *Unit {
	return new(id, IdleType, n, nil)
}

func NewProduction(id ID, n *node.Node) *Unit {
	return new(id, ProductionType, n, nil)
}

func NewBuilder(id ID, n *node.Node) *Unit {
	return new(id, BuilderType, n, nil)
}

type TransportData struct {
	Material *material.Material
}

func NewTransportData(m *material.Material) *TransportData {
	return &TransportData{m}
}

func NewTranport(id ID, n *node.Node, data *TransportData) *Unit {
	return new(id, TransportType, n, data)
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