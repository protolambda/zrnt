package beacon

import (
	"github.com/protolambda/zrnt/eth2/beacon/finality"
	"github.com/protolambda/zrnt/eth2/beacon/finalupdates"
	"github.com/protolambda/zrnt/eth2/beacon/registry"
	"github.com/protolambda/zrnt/eth2/beacon/rewardpenalty"
	"github.com/protolambda/zrnt/eth2/beacon/slashings"
	"github.com/protolambda/zrnt/eth2/meta"

	"github.com/protolambda/zrnt/eth2/util/math"
	"sort"
)

type EpochStakeSummary struct {
	SourceStake Gwei
	TargetStake Gwei
	HeadStake Gwei
}

type EpochProcessState struct {
	Statuses []AttesterStatus
	TotalActiveStake Gwei
	PrevEpoch EpochStakeSummary
	CurrEpoch EpochStakeSummary
	ActiveValidators uint64
	IndicesToSlash []ValidatorIndex
	IndicesToActivate []ValidatorIndex
	ExitQueueEnd Epoch
	ChurnLimit uint64
}

func (state *EpochProcessState) ProcessEpoch(proc EpochProcessors) error {
	if err := proc.ProcessEpochJustification(state); err != nil {
		return err
	}
	if err := proc.ProcessEpochRewardsAndPenalties(state); err != nil {
		return err
	}
	if err := proc.ProcessEpochRegistryUpdates(state); err != nil {
		return err
	}
	if err := proc.ProcessEpochSlashings(state); err != nil {
		return err
	}
	if err := proc.ProcessEpochFinalUpdates(state); err != nil {
		return err
	}
	return nil
}


func GetChurnLimit(activeValidatorCount uint64) uint64 {
	return math.MaxU64(MIN_PER_EPOCH_CHURN_LIMIT, activeValidatorCount/CHURN_LIMIT_QUOTIENT)
}

func PrepareEpochProcessState(input EpochPreparationInput) (out *EpochProcessState, err error) {
	count, err := input.ValidatorCount()
	if err != nil {
		return nil, err
	}

	currentEpoch, err := input.CurrentEpoch()
	if err != nil {
		return nil, err
	}
	prevEpoch, err := input.PreviousEpoch()
	if err != nil {
		return nil, err
	}

	out = &EpochProcessState{
		Statuses: make([]AttesterStatus, count, count),
	}

	finality, err := input.Finalized()
	if err != nil {
		return nil, err
	}
	activationEpoch := finality.Epoch.ComputeActivationExitEpoch()

	withdrawableEpoch := currentEpoch + (EPOCHS_PER_SLASHINGS_VECTOR / 2)
	exitQueueEnd := currentEpoch.ComputeActivationExitEpoch()

	activeCount := uint64(0)
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		status := &out.Statuses[i]
		slashed, err := input.IsSlashed(i)
		if err != nil {
			return nil, err
		}
		if !slashed {
			status.Flags |= UnslashedAttester
		}
		if effBal, err := input.EffectiveBalance(i); err != nil {
			return nil, err
		} else {
			status.EffectiveBalance = effBal
		}
		if active, err := input.IsActive(i, currentEpoch); err != nil {
			return nil, err
		} else if active {
			status.Flags |= EligibleAttester
			status.Active = true
			out.TotalActiveStake += status.EffectiveBalance
			activeCount++
		} else if slashed {
			valWithdrawableEpoch, err := input.WithdrawableEpoch(i)
			if err != nil {
				return nil, err
			} else if prevEpoch+1 < valWithdrawableEpoch {
				status.Flags |= EligibleAttester
			}
			if withdrawableEpoch == valWithdrawableEpoch {
				out.IndicesToSlash = append(out.IndicesToSlash, i)
			}
		}
		valExit, err := input.ExitEpoch(i)
		if err != nil {
			return nil, err
		}
		status.ExitEpoch = valExit
		if valExit != FAR_FUTURE_EPOCH && valExit > exitQueueEnd {
			exitQueueEnd = valExit
		}
		status.AttestedProposer = ValidatorIndexMarker

		valActivationEligibilityEpoch, err := input.ActivationEligibilityEpoch(i)
		if err != nil {
			return nil, err
		}
		status.ActivationEligibilityEpoch = valActivationEligibilityEpoch
		valActivationEpoch, err := input.ActivationEpoch(i)
		if err != nil {
			return nil, err
		}
		status.ActivationEpoch = valActivationEpoch
		if valActivationEligibilityEpoch != FAR_FUTURE_EPOCH && valActivationEpoch >= activationEpoch {
			out.IndicesToActivate = append(out.IndicesToActivate, i)
		}
	}

	// Order by the sequence of activation_eligibility_epoch setting and then index
	sort.Slice(out.IndicesToActivate, func(i int, j int) bool {
		return out.Statuses[out.IndicesToActivate[i]].activationEligibility < out.Statuses[out.IndicesToActivate[j]].activationEligibility
	})

	exitQueueChurn := uint64(0)
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		status := &out.Statuses[i]
		if status.ExitEpoch == exitQueueEnd {
			exitQueueChurn++
		}
	}
	churnLimit := GetChurnLimit(activeCount)
	if exitQueueChurn >= churnLimit {
		exitQueueEnd++
	}
	out.ExitQueueEnd = exitQueueEnd
	out.ChurnLimit = churnLimit

	processEpoch := func(
		attestations []*PendingAttestation, epochSum *EpochStakeSummary, epoch Epoch,
		sourceFlag, targetFlag, headFlag AttesterFlag) error {

		targetBlockRoot, err := input.GetBlockRootAtSlot(epoch.GetStartSlot())
		if err != nil {
			return err
		}
		participants := make([]ValidatorIndex, 0, MAX_VALIDATORS_PER_COMMITTEE)
		for _, att := range attestations {
			attBlockRoot, err := input.GetBlockRootAtSlot(att.Data.Slot)
			if err != nil {
				return err
			}

			// attestation-target is already known to be this epoch, get it from the pre-computed shuffling directly.
			committee, err := input.GetBeaconCommittee(att.Data.Slot, att.Data.Index)
			if err != nil {
				return err
			}

			participants = participants[:0]                   // reset old slice (re-used in for loop)
			participants = append(participants, committee...) // add committee indices

			if epoch == prevEpoch {
				for _, p := range participants {
					status := &out.Statuses[p]

					// If the attestation is the earliest, i.e. has the smallest delay
					if status.AttestedProposer == ValidatorIndexMarker || status.InclusionDelay > att.InclusionDelay {
						status.InclusionDelay = att.InclusionDelay
						status.AttestedProposer = att.ProposerIndex
					}
				}
			}

			participants = att.AggregationBits.FilterParticipants(participants) // only keep the participants
			for _, p := range participants {
				status := &out.Statuses[p]

				// remember the participant as one of the good validators
				status.Flags |= sourceFlag
				epochSum.SourceStake += status.EffectiveBalance

				// If the attestation is for the boundary:
				if att.Data.Target.Root == targetBlockRoot {
					status.Flags |= targetFlag
					epochSum.TargetStake += status.EffectiveBalance

					// If the attestation is for the head (att the time of attestation):
					if att.Data.BeaconBlockRoot == attBlockRoot {
						status.Flags |= headFlag
						epochSum.HeadStake += status.EffectiveBalance
					}
				}
			}
		}
		return nil
	}
	prevPendingRawAtts, err := input.PreviousEpochPendingAttestations()
	if err != nil {
		return nil, err
	}
	if err := processEpoch(prevPendingRawAtts, &out.PrevEpoch, prevEpoch,
		PrevSourceAttester, PrevTargetAttester, PrevHeadAttester); err != nil {
		return nil, err
	}
	currPendingRawAtts, err := input.CurrentEpochPendingAttestations()
	if err != nil {
		return nil, err
	}
	if err := processEpoch(currPendingRawAtts, &out.PrevEpoch, currentEpoch,
		CurrSourceAttester, CurrTargetAttester, CurrHeadAttester); err != nil {
		return nil, err
	}
	return
}
