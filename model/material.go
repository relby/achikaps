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

type Material struct {
	id ID
	typ MaterialType
	node *Node
	isReserved bool
}

func NewMaterial(id ID, typ MaterialType, n *Node) *Material {
	m := &Material{id, typ, nil, false}
	
	n.AddMaterial(m)
	
	return m
}

func (m *Material) ID() ID {
	return m.id
}

func (m *Material) Type() MaterialType {
	return m.typ
}

func (m *Material) Node() *Node {
	return m.node
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
		Type    MaterialType
		Node    *Node
		IsReserved bool
	}

	materialData := materialJSON{
		ID:      m.id,
		Type:    m.typ,
		Node:    m.node,
		IsReserved: m.isReserved,
	}

	return json.Marshal(materialData)
}