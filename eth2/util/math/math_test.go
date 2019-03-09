package math

import (
	"fmt"
	"testing"
)

var nextPower2Tests = []struct {
	in  uint64
	out uint64
}{
	{0, 0},
	{1, 1},
	{2, 2},
	{3, 4},
	{6, 8},
	{7, 8},
	{8, 8},
	{32, 32},
	{33, 64},
	{63, 64},
	{64, 64},
	{(1 << 63) - 3, 1 << 63},
}

func TestNextPowerOfTwo(t *testing.T) {
	for _, testCase := range nextPower2Tests {
		t.Run(fmt.Sprintf("NextPower2: %d -> %d", testCase.in, testCase.out), func(tt *testing.T) {
			o := NextPowerOfTwo(testCase.in)
			if o != testCase.out {
				tt.Errorf("got %q, want %q", o, testCase.out)
			}
		})
	}
}
