package opcode

import (
	"errors"

	"github.com/relby/achikaps/model"
	"github.com/relby/achikaps/win_condition"
)

type OpCode int64

func NewOpCode(v int64) (OpCode, error) {
	switch v := OpCode(v); v {
	case InitialState, BuildNode, UnitActionExecute:
		return v, nil
	}

	return 0, errors.New("invalid op code")
}

const (
	InitialState OpCode = iota + 1
	BuildNode
	UnitActionExecute
	ChangeUnitType
	Win
	NodeBuilt
)

// TODO: create a constructor
type InitialStateResp struct {
	Nodes map[string]map[model.ID]*model.Node
	Connections map[string]map[model.ID][]model.ID
	Units map[string]map[model.ID]*model.Unit
	Materials map[string]map[model.ID]*model.Material
	WinCondition *win_condition.WinCondition
}

type UnitActionExecuteResp struct {
	Unit *model.Unit
	UnitAction *model.UnitAction
}

func NewUnitActionExecuteResp(u *model.Unit, a *model.UnitAction) *UnitActionExecuteResp {
	return &UnitActionExecuteResp{u, a}
}

type WinResp struct {
	SessionID string
}

func NewWinResp(sID string) *WinResp {
	return &WinResp{sID}
}

type NodeBuiltResp struct {
	Node *model.Node
}

func NewNodeBuiltResp(n *model.Node) *NodeBuiltResp {
	return &NodeBuiltResp{n}
}