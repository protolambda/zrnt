package common

import (
	"errors"
	"fmt"
	"math/big"
)

// U256 is for config use only, not a SSZ type.
// stored as big-int internally
type U256 [32]byte

// MarshalText encodes decimal
func (e U256) MarshalText() ([]byte, error) {
	return new(big.Int).SetBytes(e[:]).MarshalText()
}

// UnmarshalText decodes depending on the prefix (decimal by default)
func (e *U256) UnmarshalText(b []byte) error {
	if e == nil {
		return errors.New("cannot decode into nil uint256")
	}
	x, ok := new(big.Int).SetString(string(b), 0)
	if !ok {
		return fmt.Errorf("failed to parse uint256 text: %q", string(b))
	}
	if x.Sign() < 0 {
		return fmt.Errorf("failed to parse uint256: number cannot be negative")
	}
	if l := x.BitLen(); l > 256 {
		return fmt.Errorf("failed to parse uint256: expected no more than 256 bits, got %d bits", l)
	}
	x.FillBytes(e[:])
	return nil
}

func (e U256) String() string {
	return new(big.Int).SetBytes(e[:]).String()
}

func MustU256(v string) (out U256) {
	err := out.UnmarshalText([]byte(v))
	if err != nil {
		panic(err)
	}
	return
}
