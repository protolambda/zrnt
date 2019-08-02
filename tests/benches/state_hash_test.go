package benches

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"testing"
)

// 1 op = 1 vote applied. Resets votes every SLOTS_PER_ETH1_VOTING_PERIOD ops
func BenchmarkStateHash(b *testing.B) {
	full := CreateTestState(100, MAX_EFFECTIVE_BALANCE)
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		root := ssz.HashTreeRoot(full.BeaconState, phase0.BeaconStateSSZ)
		res ^= root[0]
	}
	b.Logf("res: %d", res)
}
