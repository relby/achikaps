package material

import "github.com/relby/achikaps/node"

type Type uint

const (
	TODOType Type = iota + 1
)

type Material struct {
	Type Type
	Node *node.Node
}

func New(typ Type, n *node.Node) *Material {
	return &Material{typ, n}
}