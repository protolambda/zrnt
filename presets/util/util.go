package util

import "math/big"

// MustBigInt is a helper function for config initialization.
// DO NOT USE for untrusted config data. Panics if invalid int.
func MustBigInt(v string) *big.Int {
	var x big.Int
	if err := x.UnmarshalText([]byte(v)); err != nil {
		panic(err)
	}
	return &x
}
