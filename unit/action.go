package unit

import (
	"github.com/relby/achikaps/node"
)

type ActionType uint

const (
	MovingActionType ActionType = iota + 1
	ProductionActionType
	BuildingActionType
)

type Action struct {
	Type ActionType
	Data any
}

func new(typ ActionType, data any) *Action {
	return &Action{typ, data}
}

type MovingActionData struct {
	Speed float64
	Node *node.Node
	Progress float64
}

func NewMovingAction(speed float64, n *node.Node) *Action {
	return new(
		MovingActionType,
		&MovingActionData{speed, n, 0},
	)
}

type ProductionActionData struct {
	Progress float64
}

func NewProductionAction() *Action {
	return new(
		ProductionActionType,
		&ProductionActionData{0},
	)
}

type BuildingActionData struct {}

func NewBuildingAction() *Action {
	return new(
		BuildingActionType,
		&BuildingActionData{},
	)
}