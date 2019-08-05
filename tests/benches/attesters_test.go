package benches

/* TODO: finish attester statuses benchmark.
import (
	"encoding/binary"
	"github.com/protolambda/zrnt/eth2/beacon/attestations"
	. "github.com/protolambda/zrnt/eth2/core"
	"testing"
)


func BenchmarkLoadAttesterStatuses(b *testing.B) {
	validatorCount := uint64(10000)
	meta := CreateTestState(validatorCount, MAX_EFFECTIVE_BALANCE)
	meta.Slot = (GENESIS_EPOCH + 2).GetStartSlot() - 1
	f := meta.AttesterStatusFeature
	for i := Slot(0); i < SLOTS_PER_EPOCH * 2; i++ {
		root := Root{}
		binary.LittleEndian.PutUint64(root[:], uint64(i))
		meta.BlockRoots[i] = root
	}
	att := func(commSize uint64, attested Slot, source Epoch, target Epoch) *attestations.PendingAttestation {
		aggBits := make(attestations.CommitteeBits, (commSize+8)/8)
		for i := uint64(0); i <= commSize; i++ { // also set bit at length index, this delimits the bitlist.
			aggBits.SetBit(i, true)
		}
		return &attestations.PendingAttestation{
			AggregationBits: aggBits,
			Data:            attestations.AttestationData{
				BeaconBlockRoot: meta.GetBlockRootAtSlot(attested),
				Source: Checkpoint{
					Epoch: source,
					Root:  meta.GetBlockRoot(source),
				},
				Target: Checkpoint{
					Epoch: target,
					Root:  meta.GetBlockRoot(target),
				},
			},
			InclusionDelay:  MIN_ATTESTATION_INCLUSION_DELAY,
			ProposerIndex:   0,
		}
	}
	{
		committeeCount := meta.GetCommitteeCount(GENESIS_EPOCH)
		perSlot := committeeCount / uint64(SLOTS_PER_EPOCH)
		for i := uint64(0); i < committeeCount; i++ {
			comm := meta.PreviousShuffling.Committees[(meta.GetStartShard(GENESIS_EPOCH)+Shard(i))%SHARD_COUNT]
			meta.PreviousEpochAttestations = append(meta.PreviousEpochAttestations,
				att(uint64(len(comm)), GENESIS_EPOCH.GetStartSlot() + Slot(i / perSlot), GENESIS_EPOCH, GENESIS_EPOCH),
			)
		}
	}
	{
		committeeCount := meta.GetCommitteeCount(GENESIS_EPOCH+1)
		perSlot := committeeCount / uint64(SLOTS_PER_EPOCH)
		for i := uint64(0); i < committeeCount; i++ {
			comm := meta.CurrentShuffling.Committees[(meta.GetStartShard(GENESIS_EPOCH+1)+Shard(i))%SHARD_COUNT]
			meta.CurrentEpochAttestations = append(meta.CurrentEpochAttestations,
				att(uint64(len(comm)), GENESIS_EPOCH.GetStartSlot() + SLOTS_PER_EPOCH + Slot(i / perSlot), GENESIS_EPOCH+1, GENESIS_EPOCH+1),
			)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.LoadAttesterStatuses()
	}
}
*/
