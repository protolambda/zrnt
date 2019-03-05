package math

import "github.com/protolambda/go-beacon-transition/eth2"

// Typed wrappers for min/max, may want to unify or move.

func Max(a eth2.Gwei, b eth2.Gwei) eth2.Gwei {
	return eth2.Gwei(MaxU64(uint64(a), uint64(b)))
}
func Min(a eth2.Gwei, b eth2.Gwei) eth2.Gwei {
	return eth2.Gwei(MinU64(uint64(a), uint64(b)))
}


func MaxU64(a uint64, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func MinU64(a uint64, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// The largest integer x such that x**2 is less than or equal to n.
func Integer_squareroot(n uint64) uint64 {
	x := n
	y := (x + 1) >> 1
	for y < x {
		x = y
		y = (x + n/x) >> 1
	}
	return x
}

func Is_power_of_two(n uint64) bool {
	return (n > 0) && (n&(n-1) == 0)
}

func NextPowerOfTwo(in uint64) uint64 {
	v := in
	v--
	v |= v >> (1 << 0)
	v |= v >> (1 << 1)
	v |= v >> (1 << 2)
	v |= v >> (1 << 3)
	v |= v >> (1 << 4)
	v |= v >> (1 << 5)
	v |= v >> (1 << 6)
	v |= v >> (1 << 7)
	v |= v >> (1 << 8)
	v |= v >> (1 << 9)
	v++
	return v
}
