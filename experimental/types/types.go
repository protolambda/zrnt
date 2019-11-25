package types

import (
	"github.com/protolambda/zrnt/experimental/tree"
	. "github.com/protolambda/zrnt/experimental/views"
)

const Bool = BoolType

const (
	Uint8  = Uint8Type
	Uint16 = Uint16Type
	Uint32 = Uint32Type
	Uint64 = Uint64Type
)

const Byte = ByteType

const Root = tree.RootType

type Container = ContainerType

func Vector(elemType TypeDef, length uint64) *VectorType {
	return &VectorType{
		ElementType: elemType,
		Length:      length,
	}
}

func List(elemType TypeDef, limit uint64) *ListType {
	return &ListType{
		ElementType: elemType,
		Limit:       limit,
	}
}

func BasicVector(elemType BasicTypeDef, length uint64) *BasicVectorType {
	return &BasicVectorType{
		ElementType: elemType,
		Length:      length,
	}
}

func BasicList(elemType BasicTypeDef, limit uint64) *BasicListType {
	return &BasicListType{
		ElementType: elemType,
		Limit:       limit,
	}
}

func Bitvector(length uint64) *BitVectorType {
	return &BitVectorType{
		BitLength: length,
	}
}

func Bitlist(limit uint64) *BitListType {
	return &BitListType{
		BitLimit: limit,
	}
}

