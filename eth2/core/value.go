package core

type Gwei uint64

func Max(a Gwei, b Gwei) Gwei {
	if a > b {
		return a
	}
	return b
}

func Min(a Gwei, b Gwei) Gwei {
	if a < b {
		return a
	}
	return b
}
