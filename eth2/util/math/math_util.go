package math

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
func IntegerSquareroot(n uint64) uint64 {
	x := n
	y := (x + 1) >> 1
	for y < x {
		x = y
		y = (x + n/x) >> 1
	}
	return x
}

func IsPowerOfTwo(n uint64) bool {
	return (n > 0) && (n&(n-1) == 0)
}

// Returns the next power of two for the given number. (returns the number itself if it's a power of two)
func NextPowerOfTwo(in uint64) uint64 {
	v := in
	v--
	v |= v >> (1 << 0)
	v |= v >> (1 << 1)
	v |= v >> (1 << 2)
	v |= v >> (1 << 3)
	v |= v >> (1 << 4)
	v |= v >> (1 << 5)
	v++
	return v
}
