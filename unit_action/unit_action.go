package unit_action

import (
	"github.com/relby/achikaps/material"
	"github.com/relby/achikaps/node"
)

type Type uint

const (
	MovingType Type = iota + 1
	ProductionType
	BuildingType
	TakeMaterialType
	DropMaterialType
)

type UnitAction struct {
	Type Type
	IsStarted bool
	Data any
}

func new(typ Type, data any) *UnitAction {
	return &UnitAction{typ, false, data}
}

type MovingData struct {
	Speed float64
	FromNode *node.Node
	ToNode *node.Node
	Progress float64
}

func NewMoving(speed float64, fromNode, toNode *node.Node) *UnitAction {
	return new(
		MovingType,
		&MovingData{speed, fromNode, toNode, 0},
	)
}

type ProductionData struct {
	Progress float64
}

func NewProduction() *UnitAction {
	return new(
		ProductionType,
		&ProductionData{0},
	)
}

func NewBuilding() *UnitAction {
	return new(
		BuildingType,
		nil,
	)
}

type TakeMaterialData struct {
	Material *material.Material
}

func NewTakeMaterial(m *material.Material) *UnitAction {
	return new(TakeMaterialType, &TakeMaterialData{m})
}

func NewDropMaterial() *UnitAction {
	return new(DropMaterialType, nil)
}