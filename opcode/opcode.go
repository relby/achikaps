package opcode

import (
	"errors"

	"github.com/relby/achikaps/model"
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
)

type UnitActionExecuteResp struct {
	Unit *model.Unit
	UnitAction *model.UnitAction
}

func NewUnitActionExecuteResp(unit *model.Unit, unitAction *model.UnitAction) *UnitActionExecuteResp {
	return &UnitActionExecuteResp{unit, unitAction}
}

type WinResp struct {
	SessionID string
}

func NewWinResp(sessionID string) *WinResp {
	return &WinResp{sessionID}
}