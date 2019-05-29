package gen_index

// binary merkle-tree node location identity: depth**2 + index
type GenIndex interface {
	IsRoot() bool
	GetDepth() uint64
}
