package math

import "testing"

func TestNextPowerOfTwo(t *testing.T) {
	if x := NextPowerOfTwo(3); x != 4 {
		t.Error("failed: nextpow2(3) -> 4, was: ", x)
	}
}