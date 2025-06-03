package material

import "github.com/relby/achikaps/node"

type Type uint

const (
	GrassType Type = iota + 1
	SandType
	DewType
	SeedType
)

type Material struct {
	Type Type
	Node *node.Node
	IsReserved bool
}

func New(typ Type, n *node.Node) *Material {
	return &Material{typ, n, false}
}