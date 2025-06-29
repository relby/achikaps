package model

import (
	"encoding/json"

	"github.com/relby/achikaps/assert"
)

type MaterialType uint

const (
	GrassMaterialType MaterialType = iota + 1
	SandMaterialType
	DewMaterialType
	SeedMaterialType
	SugarMaterialType
	JuiceMaterialType
	ChitinMaterialType
	EggMaterialType
	PheromoneMaterialType
	AmberMaterialType
)

type NodeData struct {
	Node *Node
	IsInput bool
}

func newNodeData(node *Node, isInput bool) *NodeData {
	return &NodeData{node, isInput}
}

type Material struct {
	id ID
	sessionID string
	typ MaterialType
	nodeData *NodeData
	isReserved bool
}

func NewMaterial(id ID, sessionID string, typ MaterialType, n *Node, isInput bool) *Material {
	m := &Material{
		id,
		sessionID,
		typ,
		nil,
		false,
	}
	
	if isInput {
		n.AddInputMaterial(m)
	} else {
		n.AddOutputMaterial(m)
	}
	
	return m
}

func (m *Material) ID() ID {
	return m.id
}

func (m *Material) Type() MaterialType {
	return m.typ
}

func (m *Material) NodeData() *NodeData {
	return m.nodeData
}

func (m *Material) IsReserved() bool {
	return m.isReserved
}

func (m *Material) Reserve() {
	assert.False(m.isReserved)
	m.isReserved = true
}

func (m *Material) UnReserve() {
	assert.True(m.isReserved)
	m.isReserved = false
}

func (m *Material) MarshalJSON() ([]byte, error) {
	type materialJSON struct {
		ID      ID
		SessionID string
		Type    MaterialType
		NodeData    *NodeData
		IsReserved bool
	}

	materialData := materialJSON{
		m.id,
		m.sessionID,
		m.typ,
		m.nodeData,
		m.isReserved,
	}

	return json.Marshal(materialData)
}