package altair

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/util/math"
)

func ProcessAttestations(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, ops []phase0.Attestation) error {
	for i := range ops {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := ProcessAttestation(spec, epc, state, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

func ProcessAttestation(spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, attestation *phase0.Attestation) error {
	data := &attestation.Data

	// Check slot
	currentSlot, err := state.Slot()
	if err != nil {
		return err
	}
	if !(currentSlot <= data.Slot+spec.SLOTS_PER_EPOCH) {
		return errors.New("attestation slot is too old")
	}
	if !(data.Slot+spec.MIN_ATTESTATION_INCLUSION_DELAY <= currentSlot) {
		return errors.New("attestation is too new")
	}

	currentEpoch := spec.SlotToEpoch(currentSlot)
	previousEpoch := currentEpoch.Previous()

	// Check target
	if data.Target.Epoch < previousEpoch {
		return errors.New("attestation data is invalid, target is too far in past")
	} else if data.Target.Epoch > currentEpoch {
		return errors.New("attestation data is invalid, target is in future")
	}
	// And if it matches the slot
	if data.Target.Epoch != spec.SlotToEpoch(data.Slot) {
		return errors.New("attestation data is invalid, slot epoch does not match target epoch")
	}

	// Check committee index
	if commCount, err := epc.GetCommitteeCountAtSlot(data.Slot); err != nil {
		return err
	} else if uint64(data.Index) >= commCount {
		return errors.New("attestation data is invalid, committee index out of range")
	}

	var justifiedCheckpoint common.Checkpoint
	var epochParticipation *ParticipationRegistryView
	// Check source
	if data.Target.Epoch == currentEpoch {
		justifiedCheckpoint, err = state.CurrentJustifiedCheckpoint()
		if err != nil {
			return err
		}
		epochParticipation, err = state.CurrentEpochParticipation()
		if err != nil {
			return err
		}
	} else {
		justifiedCheckpoint, err = state.PreviousJustifiedCheckpoint()
		if err != nil {
			return err
		}
		epochParticipation, err = state.PreviousEpochParticipation()
		if err != nil {
			return err
		}
	}

	// In spec: assert is_matching_source
	if data.Source != justifiedCheckpoint {
		return errors.New("attestation source does not match current justified checkpoint")
	}

	// Check signature and bitfields
	committee, err := epc.GetBeaconCommittee(data.Slot, data.Index)
	if err != nil {
		return err
	}
	indexedAtt, err := attestation.ConvertToIndexed(spec, committee)
	if err != nil {
		return fmt.Errorf("attestation could not be converted to an indexed attestation: %v", err)
	} else if err := phase0.ValidateIndexedAttestation(spec, epc, state, indexedAtt); err != nil {
		return fmt.Errorf("attestation could not be verified in its indexed form: %v", err)
	}

	expectedHead, err := common.GetBlockRootAtSlot(spec, state, data.Slot)
	if err != nil {
		return err
	}
	expectedTarget, err := common.GetBlockRoot(spec, state, data.Target.Epoch)
	if err != nil {
		return err
	}
	matchingHead := expectedHead == data.BeaconBlockRoot
	matchingTaget := expectedTarget == data.Target.Root

	flags := ParticipationFlags(0)
	if currentSlot <= data.Slot+common.Slot(math.IntegerSquareroot(uint64(spec.SLOTS_PER_EPOCH))) {
		flags |= TIMELY_SOURCE_FLAG
	}
	if matchingTaget && currentSlot <= data.Slot+spec.SLOTS_PER_EPOCH {
		flags |= TIMELY_TARGET_FLAG
	}
	if matchingHead && matchingTaget && currentSlot <= data.Slot+spec.MIN_ATTESTATION_INCLUSION_DELAY {
		flags |= TIMELY_HEAD_FLAG
	}

	// TODO: probably better to batch flag changes, needs optimization, tree structure not good for this.
	proposerRewardNumerator := uint64(0)
	for _, vi := range indexedAtt.AttestingIndices {
		baseReward := uint64(epc.EffectiveBalances[vi]) * spec.BASE_REWARD_FACTOR / uint64(epc.TotalActiveStakeSqRoot)
		existingFlags, err := epochParticipation.GetFlags(vi)
		if err != nil {
			return err
		}
		if (flags&TIMELY_SOURCE_FLAG != 0) && (existingFlags&TIMELY_SOURCE_FLAG == 0) {
			proposerRewardNumerator += baseReward * TIMELY_SOURCE_WEIGHT
		}
		if (flags&TIMELY_TARGET_FLAG != 0) && (existingFlags&TIMELY_TARGET_FLAG == 0) {
			proposerRewardNumerator += baseReward * TIMELY_TARGET_WEIGHT
		}
		if (flags&TIMELY_HEAD_FLAG != 0) && (existingFlags&TIMELY_HEAD_FLAG == 0) {
			proposerRewardNumerator += baseReward * TIMELY_HEAD_WEIGHT
		}
		if flags != 0 {
			if err := epochParticipation.SetFlags(vi, existingFlags|flags); err != nil {
				return err
			}
		}
	}

	proposerReward := common.Gwei(proposerRewardNumerator / (WEIGHT_DENOMINATOR * spec.PROPOSER_REWARD_QUOTIENT))
	proposerIndex, err := epc.GetBeaconProposer(currentSlot)
	if err != nil {
		return err
	}
	bals, err := state.Balances()
	if err != nil {
		return err
	}
	if err := common.IncreaseBalance(bals, proposerIndex, proposerReward); err != nil {
		return err
	}

	return nil
}
