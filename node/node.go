package node

import (
	"errors"

	"github.com/relby/achikaps/vec2"
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
	TransitType Type = iota + 1
	ProductionType
	StorageType
	DefenseType
)

func NewType(v uint) (Type, error) {
	switch v := Type(v); v {
	case TransitType,
		ProductionType,
		StorageType,
		DefenseType:
		return v, nil
	}

	return 0, errors.New("invalid node type")
}

var TypeRadiuses = map[Type]float64{
	TransitType: 1,
	ProductionType: 2,
	StorageType: 3,
	DefenseType: 1,
}

type ProductionTypeData uint
const (
	UnitProductionTypeData ProductionTypeData = iota + 1
	TODOMaterialProductionTypeData
)

type Node struct {
	ID       ID
	Type     Type
	Data     any
	Position vec2.Vec2
	Radius   float64
	BuildProgress float64
}

func new(id ID, typ Type, pos vec2.Vec2, data any) *Node {
	return &Node{
		id,
		typ,
		data,
		pos,
		TypeRadiuses[typ],
		0,
	}
}

func NewTransit(id ID, pos vec2.Vec2) *Node {
	return new(
		id,
		TransitType,
		pos,
		nil,
	)
}

func NewProduction(id ID, pos vec2.Vec2, data ProductionTypeData) *Node {
	return new(
		id,
		ProductionType,
		pos,
		data,
	)
}

func (n1 *Node) DistanceTo(n2 *Node) float64 {
	return vec2.Distance(n1.Position, (n2.Position))
}

func (n1 *Node) Intersects(n2 *Node) bool {
	distance := n1.DistanceTo(n2)
	sumOfRadii := n1.Radius + n2.Radius
	return distance < sumOfRadii
}

func (n *Node) IsBuilt() bool {
	return n.BuildProgress >= 1
}