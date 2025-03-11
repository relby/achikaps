package main

import "errors"

type NodeID uint

func NewNodeID(v uint) (NodeID, error) {
	if v == 0 {
		return 0, errors.New("invalid node id")
	}
	return NodeID(v), nil
}

type NodeType uint

const (
	TransitNodeType NodeType = iota + 1
	FactoryNodeType
	StorageNodeType
	DefenseNodeType
)

func NewNodeType(v uint) (NodeType, error) {
	switch v := NodeType(v); v {
	case TransitNodeType,
		FactoryNodeType,
		StorageNodeType,
		DefenseNodeType:
		return v, nil
	}

	return 0, errors.New("invalid node type")
}

var NodeTypeRadiuses = map[NodeType]float32{
	TransitNodeType: 1,
	FactoryNodeType: 2,
	StorageNodeType: 3,
	DefenseNodeType: 1,
}

type Node struct {
	ID       NodeID
	Type     NodeType
	Position Position
	Radius   float32
}

func NewNode(id NodeID, typ NodeType, position Position) Node {
	return Node{
		id,
		typ,
		position,
		NodeTypeRadiuses[typ],
	}
}
