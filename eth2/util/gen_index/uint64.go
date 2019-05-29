package gen_index

const (
	mask0 = ^GenIndexUint64((1 << (1 << iota)) - 1)
	mask1
	mask2
	mask3
	mask4
	mask5
)
const (
	bit0 = uint64(1 << iota)
	bit1
	bit2
	bit3
	bit4
	bit5
)

// small generalized index, for efficiency
type GenIndexUint64 uint64

func (g GenIndexUint64) IsRoot() bool {
	return g <= 1
}

func (g GenIndexUint64) GetDepth() (out uint64) {
	if g == 0 {
		// technically invalid, but deal with it as the root node as default.
		return 0
	}
	// bitmagic: binary search through the uint64,
	// and set the corresponding index bit (1 of 5, 1<<5 = 32, -> 0x11111 = 63 is max.) in the output.
	if g&mask5 != 0 {
		g >>= bit5
		out |= bit5
	}
	if g&mask4 != 0 {
		g >>= bit4
		out |= bit4
	}
	if g&mask3 != 0 {
		g >>= bit3
		out |= bit3
	}
	if g&mask2 != 0 {
		g >>= bit2
		out |= bit2
	}
	if g&mask1 != 0 {
		g >>= bit1
		out |= bit1
	}
	if g&mask0 != 0 {
		out |= bit0
	}
	return
}
