package model

import (
	"encoding/json"
	"errors"

	"github.com/relby/achikaps/assert"
	"github.com/relby/achikaps/config"
	"github.com/relby/achikaps/vec2"
)

type NodeType uint

const (
	TransitNodeType NodeType = iota + 1
	ProductionNodeType
	DefenseNodeType
)

type NodeName uint

const (
	SandTransitNodeName NodeName = iota + 1
	GrassFieldNodeName
	WellNodeName
	SeedStorageNodeName
	AphidDistillationNodeName
	RawMaterialVatNodeName
	ChitinPressNodeName
	EggFarmNodeName
	PheromoneMineNodeName
	IncubatorNodeName
	GeneticHatcheryNodeName
	GuardOutpostNodeName
	AmberTurretNodeName
)

func NewNodeName(v uint) (NodeName, error) {
	switch v := NodeName(v); v {
	case SandTransitNodeName,
	GrassFieldNodeName,
	WellNodeName,
	SeedStorageNodeName,
	AphidDistillationNodeName,
	RawMaterialVatNodeName,
	ChitinPressNodeName,
	EggFarmNodeName,
	PheromoneMineNodeName,
	IncubatorNodeName,
	GeneticHatcheryNodeName,
	GuardOutpostNodeName,
	AmberTurretNodeName:
		return v, nil
	}

	return 0, errors.New("invalid node name")
}

type Node struct {
	id       ID
	sessionID string
	typ     NodeType
	name     NodeName
	position vec2.Vec2
	radius   float64
	buildProgress float64
	units map[ID]*Unit
	inputMaterials map[ID]*Material
	outputMaterials map[ID]*Material
}

func NewNode(id ID, sessionID string, name NodeName, pos vec2.Vec2) *Node {
	return &Node{
		id,
		sessionID,
		nodeNameToNodeType(name),
		name,
		pos,
		config.NodeRadius,
		0,
		make(map[ID]*Unit),
		make(map[ID]*Material),
		make(map[ID]*Material),
	}
}

func (n *Node) ID() ID {
	return n.id
}

func (n *Node) Type() NodeType {
	return n.typ
}

func (n *Node) Name() NodeName {
	return n.name
}

func (n *Node) Position() vec2.Vec2 {
	return n.position
}

func (n *Node) Radius() float64 {
	return n.radius
}

func (n *Node) Build(inc float64) {
	n.buildProgress += inc
	if n.buildProgress >= 1.0 {
		n.buildProgress = 1.0
	}
}

func (n *Node) BuildFully() {
	n.buildProgress = 1.0
}

func (n *Node) IsBuilt() bool {
	return n.buildProgress >= 1.0
}

func (n1 *Node) DistanceTo(n2 *Node) float64 {
	return vec2.Distance(n1.position, (n2.position))
}

func (n1 *Node) Intersects(n2 *Node) bool {
	distance := n1.DistanceTo(n2)
	sumOfRadii := n1.radius + n2.radius
	return distance < sumOfRadii
}

func (n *Node) Units() map[ID]*Unit {
	return n.units
}

func (n *Node) AddUnit(u *Unit) {
	assert.Nil(u.node)
	u.node = n
	n.units[u.id] = u
}

func (n *Node) RemoveUnit(u *Unit) {
	assert.NotNil(u.node)
	assert.Equals(u.node.id, n.id)
	_, exists := n.units[u.id]
	assert.True(exists)

	u.node = nil
	delete(n.units, u.id)
}

func (n *Node) InputMaterials() map[ID]*Material {
	return n.inputMaterials
}

func (n *Node) AddInputMaterial(m *Material) {
	assert.Nil(m.nodeData)
	m.nodeData = newNodeData(n, true)
	n.inputMaterials[m.id] = m
}

func (n *Node) RemoveInputMaterial(m *Material) {
	assert.NotNil(m.nodeData)
	assert.True(m.nodeData.IsInput)
	assert.Equals(m.nodeData.Node.id, n.id)
	_, exists := n.inputMaterials[m.id]
	assert.True(exists)

	m.nodeData = nil
	delete(n.inputMaterials, m.id)
}

func (n *Node) OutputMaterials() map[ID]*Material {
	return n.outputMaterials
}

func (n *Node) AddOutputMaterial(m *Material) {
	assert.Nil(m.nodeData)
	m.nodeData = newNodeData(n, false)
	n.outputMaterials[m.id] = m
}

func (n *Node) RemoveOutputMaterial(m *Material) {
	assert.NotNil(m.nodeData)
	assert.False(m.nodeData.IsInput)
	assert.Equals(m.nodeData.Node.id, n.id)
	_, exists := n.outputMaterials[m.id]
	assert.True(exists)

	m.nodeData = nil
	delete(n.outputMaterials, m.id)
}

func (n *Node) MarshalJSON() ([]byte, error) {
	type nodeJSON struct {
		ID      ID
		SessionID string
		Type    NodeType
		Name    NodeName
		Position vec2.Vec2
		Radius float64
		BuildProgress float64
	}

	nodeData := nodeJSON{
		n.id,
		n.sessionID,
		n.typ,
		n.name,
		n.position,
		n.radius,
		n.buildProgress,
	}

	return json.Marshal(nodeData)
}

type ProductionNodeData struct {
	Speed float64
	InputMaterials map[MaterialType]uint
	OutputMaterials map[MaterialType]uint
	OutputUnits uint
}

func newProductionNodeData(speed float64, inputMaterials, outputMaterials map[MaterialType]uint, outputUnits uint) *ProductionNodeData {
	return &ProductionNodeData{
		speed,
		inputMaterials,
		outputMaterials,
		outputUnits,
	}
}

func (n *Node) ProductionData() (*ProductionNodeData, bool) {
	if n.typ != ProductionNodeType {
		return nil, false
	}
	
	// TODO: Change this
	defaultSpeed := 1.0
	switch n.name {
	case GrassFieldNodeName:
		return newProductionNodeData(
			defaultSpeed,
			nil,
			map[MaterialType]uint{
				GrassMaterialType: 1,
			},
			0,
		), true
    case WellNodeName:
		return newProductionNodeData(
			defaultSpeed,
			nil,
			map[MaterialType]uint{
				DewMaterialType: 1,
			},
			0,
		), true
    case SeedStorageNodeName:
		return newProductionNodeData(
			defaultSpeed,
			nil,
			map[MaterialType]uint{
				SeedMaterialType: 1,
			},
			0,
		), true
    case AphidDistillationNodeName:
		return newProductionNodeData(
			defaultSpeed,
			map[MaterialType]uint{
				DewMaterialType: 1,
			},
			map[MaterialType]uint{
				SugarMaterialType: 1,
			},
			0,
		), true
    case RawMaterialVatNodeName:
		return newProductionNodeData(
			defaultSpeed,
			map[MaterialType]uint{
				DewMaterialType: 1,
				SeedMaterialType: 1,
			},
			map[MaterialType]uint{
				JuiceMaterialType: 1,
			},
			0,
		), true
    case ChitinPressNodeName:
		return newProductionNodeData(
			defaultSpeed,
			map[MaterialType]uint{
				SandMaterialType: 1,
			},
			map[MaterialType]uint{
				ChitinMaterialType: 1,
			},
			0,
		), true
    case EggFarmNodeName:
		return newProductionNodeData(
			defaultSpeed,
			map[MaterialType]uint{
				SeedMaterialType: 1,
				SugarMaterialType: 1,
			},
			map[MaterialType]uint{
				EggMaterialType: 1,
			},
			0,
		), true
    case PheromoneMineNodeName:
		return newProductionNodeData(
			defaultSpeed,
			map[MaterialType]uint{
				JuiceMaterialType: 1,
				ChitinMaterialType: 1,
			},
			map[MaterialType]uint{
				PheromoneMaterialType: 1,
			},
			0,
		), true
    case IncubatorNodeName:
		return newProductionNodeData(
			defaultSpeed,
			map[MaterialType]uint{
				EggMaterialType: 1,
				GrassMaterialType: 1,
			},
			nil,
			1,
		), true
    case GeneticHatcheryNodeName:
		return newProductionNodeData(
			defaultSpeed,
			map[MaterialType]uint{
				EggMaterialType: 1,
				JuiceMaterialType: 1,
				PheromoneMaterialType: 1,
			},
			nil,
			1,
		), true
	default:
		panic("unreachable")
	}
}

type BuildingNodeData struct {
	Materials map[MaterialType]uint
}

func newBuildingNodeData(materials map[MaterialType]uint) *BuildingNodeData {
	return &BuildingNodeData{materials}
}

func (n *Node) BuildingData() *BuildingNodeData {
	switch n.name {
	case SandTransitNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			GrassMaterialType: 2,
			SandMaterialType: 1,
		})
	case GrassFieldNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			GrassMaterialType: 3,
		})
	case WellNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			GrassMaterialType: 2,
			SandMaterialType: 1,
		})
	case SeedStorageNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			GrassMaterialType: 3,
			DewMaterialType: 1,
		})
	case AphidDistillationNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			GrassMaterialType: 4,
			DewMaterialType: 2,
		})
	case RawMaterialVatNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			GrassMaterialType: 3,
			DewMaterialType: 2,
			SeedMaterialType: 1,
		})
	case ChitinPressNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			GrassMaterialType: 2,
			SandMaterialType: 2,
		})
	case EggFarmNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			SeedMaterialType: 3,
			SugarMaterialType: 2,
		})
	case PheromoneMineNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			DewMaterialType: 2,
			JuiceMaterialType: 2,
			ChitinMaterialType: 2,
		})
	case IncubatorNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			EggMaterialType: 5,
			GrassMaterialType: 3,
		})
	case GeneticHatcheryNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			EggMaterialType: 7,
			JuiceMaterialType: 5,
		})
	case GuardOutpostNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			ChitinMaterialType: 5,
			PheromoneMaterialType: 3,
		})
	case AmberTurretNodeName:
		return newBuildingNodeData(map[MaterialType]uint{
			JuiceMaterialType: 5,
			AmberMaterialType: 3,
		})
	default: 
		panic("unreachable")
	}
}

func nodeNameToNodeType(name NodeName) NodeType {
	switch name {
	case SandTransitNodeName:
		return TransitNodeType
	case GrassFieldNodeName,
		WellNodeName,
		SeedStorageNodeName,
		AphidDistillationNodeName,
		RawMaterialVatNodeName,
		ChitinPressNodeName,
		EggFarmNodeName,
		PheromoneMineNodeName,
		IncubatorNodeName,
		GeneticHatcheryNodeName:
		return ProductionNodeType
	case GuardOutpostNodeName,
		AmberTurretNodeName:
		return DefenseNodeType
	default:
		panic("unreachable")
	}
}