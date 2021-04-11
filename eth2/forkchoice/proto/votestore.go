package proto

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	. "github.com/protolambda/zrnt/eth2/forkchoice"
)

type VoteTracker struct {
	Current            NodeRef
	Next               NodeRef
	CurrentTargetEpoch Epoch
	NextTargetEpoch    Epoch
}

type ProtoVoteStore struct {
	spec    *common.Spec
	votes   []VoteTracker
	changed bool
}

var _ VoteStore = (*ProtoVoteStore)(nil)

func NewProtoVoteStore(spec *common.Spec) VoteStore {
	return &ProtoVoteStore{spec: spec, changed: true}
}

// Process an attestation. (Note that the head slot may be for a gap slot after the block root)
func (st *ProtoVoteStore) ProcessAttestation(index ValidatorIndex, blockRoot Root, headSlot Slot) (ok bool) {
	if index >= ValidatorIndex(len(st.votes)) {
		if index < ValidatorIndex(cap(st.votes)) {
			st.votes = st.votes[:index+1]
		} else {
			extension := make([]VoteTracker, index+1-ValidatorIndex(len(st.votes)))
			st.votes = append(st.votes, extension...)
		}
	}
	vote := &st.votes[index]
	targetEpoch := st.spec.SlotToEpoch(headSlot)
	// only update if it's a newer vote, or if it's genesis and no vote has happened yet.
	if targetEpoch > vote.NextTargetEpoch || (targetEpoch == 0 && *vote == (VoteTracker{})) {
		vote.NextTargetEpoch = targetEpoch
		vote.Next = NodeRef{Root: blockRoot, Slot: headSlot}
		st.changed = true
	}
	// TODO: maybe help detect slashable votes on the fly?
	return true
}

func (st *ProtoVoteStore) HasChanges() bool {
	return st.changed
}

// Returns a list of `deltas`, where there is one delta for each of the ProtoArray nodes.
// The deltas are calculated between `oldBalances` and `newBalances`, and/or a change of vote.
// The votestore is updated, the next deltas will be 0 if ProcessAttestation is not changing any vote.
func (st *ProtoVoteStore) ComputeDeltas(indices map[NodeRef]NodeIndex, oldBalances []Gwei, newBalances []Gwei) []SignedGwei {
	deltas := make([]SignedGwei, len(indices), len(indices))
	for i := 0; i < len(st.votes); i++ {
		vote := &st.votes[i]
		// There is no need to create a score change if the validator has never voted (may not be active)
		// or both their votes are for the zero checkpoint (alias to the genesis block).
		if vote.Current == (NodeRef{}) && vote.Next == (NodeRef{}) {
			continue
		}

		// Validator sets may have different sizes (but attesters are not different, activation only under finality)
		oldBal := Gwei(0)
		if i < len(oldBalances) {
			oldBal = oldBalances[i]
		}
		newBal := Gwei(0)
		if i < len(newBalances) {
			newBal = newBalances[i]
		}

		if vote.Current == (NodeRef{}) || vote.CurrentTargetEpoch < vote.NextTargetEpoch || oldBal != newBal {
			// Ignore the current or next vote if it is not known in `indices`.
			// We assume that it is outside of our tree (i.e., pre-finalization) and therefore not interesting.
			if currentIndex, ok := indices[vote.Current]; ok {
				deltas[currentIndex] -= SignedGwei(oldBal)
			}
			if nextIndex, ok := indices[vote.Next]; ok {
				deltas[nextIndex] += SignedGwei(newBal)
				vote.Current = vote.Next
				vote.CurrentTargetEpoch = vote.NextTargetEpoch
			}
		}
	}
	st.changed = false

	return deltas
}
