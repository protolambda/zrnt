package views

import . "github.com/protolambda/zrnt/experimental/tree"

type View interface {
	Backing() Node
}

type TypeDef interface {
	DefaultNode() Node
	ViewFromBacking(node Node) View
}

type SubView interface {
	BackingFromBase(base *Root, i uint8) *Root
}

type BasicTypeDef interface {
	TypeDef
	ByteLength() uint64
	SubViewFromBacking(node *Root, i uint8) SubView
}
