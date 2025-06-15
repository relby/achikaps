package model

import (
	"encoding/json"
	"errors"

	"github.com/gammazero/deque"
	"github.com/relby/achikaps/assert"
	"github.com/relby/achikaps/config"
)

type UnitType uint

const (
	IdleUnitType UnitType = iota + 1
	ProductionUnitType
	BuilderUnitType
	TransportUnitType
)

func NewUnitType(v uint) (UnitType, error) {
	switch v := UnitType(v); v {
		case IdleUnitType,
			ProductionUnitType,
			BuilderUnitType,
			TransportUnitType:
			return v, nil
	}
	
	return 0, errors.New("invalid unit type")
}

type Unit struct {
	id ID
	sessionID string
	typ UnitType
	node *Node
	material *Material
	actions *deque.Deque[*UnitAction]
}

func NewUnit(id ID, sessionID string, typ UnitType, n *Node) *Unit {
	u := &Unit{
		id,
		sessionID,
		typ,
		nil,
		nil,
		&deque.Deque[*UnitAction]{},
	}
	
	n.AddUnit(u)
	
	return u
}

func (u *Unit) ID() ID {
	return u.id
}

func (u *Unit) Type() UnitType {
	return u.typ
}

func (u *Unit) SetType(t UnitType) {
	// Do nothing if the type is the same
	if u.typ == t {
		return
	}
	
	// If we change the type of the transport unit
	// we should ensure that material is not lost
	switch u.typ {
		case TransportUnitType:
			if u.material == nil {
				break
			}
			assert.Nil(u.material.NodeData())

			if u.node != nil {
				u.node.AddOutputMaterial(u.material)
			} else {
				// In here unit is moving
				assert.NotEquals(u.actions.Len(), 0)

				movingAction := u.actions.Front()
				assert.Equals(movingAction.Type, MovingUnitActionType)
				
				movingActionData, ok := movingAction.Data.(*MovingUnitActionData)
				assert.True(ok)

				movingActionData.FromNode.AddOutputMaterial(u.material)
			}
	}

	for i := range u.actions.Len() {
		a := u.actions.At(i)
		if a.Type == ProductionUnitActionType {
			uaData, ok := a.Data.(*ProductionUnitActionData)
			assert.True(ok)
			for _, m := range uaData.InputMaterials {
				m.UnReserve()
			}
		}
	}
	
	if u.actions.Len() != 0 {
		a := u.actions.Front()

		u.actions.Clear()
		
		if a.Type == MovingUnitActionType {
			u.actions.PushBack(a)
		}
	}

	u.typ = t
}

func (u *Unit) Node() *Node {
	return u.node
}

func (u *Unit) Material() *Material {
	assert.Equals(u.typ, TransportUnitType)
	return u.material
}

func (u *Unit) AddMaterial(m *Material) {
	assert.Equals(u.typ, TransportUnitType)
	assert.Nil(u.material)
	
	u.node.RemoveOutputMaterial(m)
	
	u.material = m
}

func (u *Unit) RemoveMaterial() {
	assert.Equals(u.typ, TransportUnitType)
	assert.NotNil(u.material)
	
	u.node.AddInputMaterial(u.material)
	
	u.material = nil
}

func (u *Unit) Actions() *deque.Deque[*UnitAction] {
	return u.actions
}

func (u *Unit) MarshalJSON() ([]byte, error) {
	actions := make([]*UnitAction, 0, u.actions.Len())
	for i := range u.actions.Len() {
		actions = append(actions, u.actions.At(i))
	}

	var unitData any
	if u.typ == TransportUnitType {
		unitData = struct {
			ID      ID
			SessionID string
			Type    UnitType
			Node    *Node
			Material *Material
			Actions  []*UnitAction
		}{
			u.id,
			u.sessionID,
			u.typ,
			u.node,
			u.material,
			actions,
		}
	} else {
		unitData = struct {
			ID      ID
			SessionID string
			Type    UnitType
			Node    *Node
			Actions  []*UnitAction
		}{
			u.id,
			u.sessionID,
			u.typ,
			u.node,
			actions,
		}
	}

	return json.Marshal(unitData)
}

type UnitActionType uint

const (
	MovingUnitActionType UnitActionType = iota + 1
	ProductionUnitActionType
	BuildingUnitActionType
	TakeMaterialUnitActionType
	DropMaterialUnitActionType
)

type UnitAction struct {
	Type UnitActionType
	IsStarted bool
	Data any
}

func newUnitAction(typ UnitActionType, data any) *UnitAction {
	return &UnitAction{typ, false, data}
}

type MovingUnitActionData struct {
	Speed float64
	TimeMs float64
	FromNode *Node
	ToNode *Node
	Progress float64
}

func NewMovingUnitAction(speed float64, fromNode, toNode *Node) *UnitAction {
	ticks := 1.0 / (speed / fromNode.DistanceTo(toNode))

	timeMs := ticks * (1000.0 / float64(config.TickRate))
	return newUnitAction(
		MovingUnitActionType,
		&MovingUnitActionData{speed, timeMs, fromNode, toNode, 0},
	)
}

type ProductionUnitActionData struct {
	InputMaterials []*Material
	Progress float64
}

func NewProductionUnitAction(materials []*Material) *UnitAction {
	return newUnitAction(
		ProductionUnitActionType,
		&ProductionUnitActionData{
			materials,
			0,
		},
	)
}

func NewBuildingUnitAction() *UnitAction {
	return newUnitAction(
		BuildingUnitActionType,
		nil,
	)
}

type TakeMaterialUnitActionData struct {
	Material *Material
}

func NewTakeMaterialUnitAction(m *Material) *UnitAction {
	return newUnitAction(TakeMaterialUnitActionType, &TakeMaterialUnitActionData{m})
}

func NewDropMaterialUnitAction() *UnitAction {
	return newUnitAction(DropMaterialUnitActionType, nil)
}