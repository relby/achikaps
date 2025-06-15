package opcode

import (
	"errors"

	"github.com/relby/achikaps/model"
	"github.com/relby/achikaps/win_condition"
)

type OpCode int64

func NewOpCode(v int64) (OpCode, error) {
	switch v := OpCode(v); v {
	case InitialState,
		BuildNode,
		UnitActionExecute,
		ChangeUnitType,
		Win,
		NodeBuilt,
		MaterialDestroyed,
		MaterialCreated,
		UnitCreated:
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
	MaterialDestroyed
	MaterialCreated
	UnitCreated
)

type RespWithOpCode struct {
	Resp any
	OpCode OpCode
}

func NewRespWithOpCode(resp any, opcode OpCode) *RespWithOpCode {
	return &RespWithOpCode{resp, opcode}
}

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

type MaterialDestroyedResp struct {
	Material *model.Material
}

func NewMaterialDestroyedResp(m *model.Material) *MaterialDestroyedResp {
	return &MaterialDestroyedResp{m}
}

type MaterialCreatedResp struct {
	Material *model.Material
}

func NewMaterialCreatedResp(m *model.Material) *MaterialCreatedResp {
	return &MaterialCreatedResp{m}
}

type UnitCreatedResp struct {
	Unit *model.Unit
}

func NewUnitCreatedResp(u *model.Unit) *UnitCreatedResp {
	return &UnitCreatedResp{u}
}