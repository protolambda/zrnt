package benches

import (
	. "github.com/protolambda/zrnt/eth2/beacon/eth1"
	. "github.com/protolambda/zrnt/eth2/core"
	"testing"
)

// 1 op = 1 vote applied. Resets votes every SLOTS_PER_ETH1_VOTING_PERIOD ops
func BenchmarkEth1Voting(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	state := &Eth1State{}
	for i, end := 0, b.N/int(SLOTS_PER_ETH1_VOTING_PERIOD); i < end; i++ {
		state.ResetEth1Votes()
		for j := Slot(0); j < SLOTS_PER_ETH1_VOTING_PERIOD; j++ {
			// flip back and forth between 2 different votes
			_ = state.ProcessEth1Vote(Eth1Data{
				DepositRoot:  Root{1, byte(i & 1), byte(j & 1)},
				DepositCount: DepositIndex(i),
				BlockHash:    Root{2, byte(i & 1), byte(j & 1)},
			})
		}
	}
}
