package beacon

import (
	"github.com/protolambda/zrnt/eth2/util/math"
	"sort"
)

type EpochStakeSummary struct {
	SourceStake Gwei
	TargetStake Gwei
	HeadStake   Gwei
}

type EpochProcess struct {
	PrevEpoch Epoch
	CurrEpoch Epoch

	Statuses []AttesterStatus

	TotalActiveStake Gwei

	PrevEpochUnslashedStake EpochStakeSummary
	PrevEpochTargetStake    Gwei
	CurrEpochTargetStake    Gwei

	// Thanks to exit delay, this does not change within the epoch processing.
	ActiveValidators uint64

	IndicesToSlash                    []ValidatorIndex
	IndicesToSetActivationEligibility []ValidatorIndex
	// Ignores churn. Apply churn-limit manually.
	// Maybe, because finality affects it still.
	IndicesToMaybeActivate []ValidatorIndex

	IndicesToEject []ValidatorIndex

	ExitQueueEnd      Epoch
	ExitQueueEndChurn uint64
	ChurnLimit        uint64
}

func GetChurnLimit(activeValidatorCount uint64) uint64 {
	return math.MaxU64(MIN_PER_EPOCH_CHURN_LIMIT, activeValidatorCount/CHURN_LIMIT_QUOTIENT)
}

func (state *BeaconStateView) PrepareEpochProcess(epc *EpochsContext) (out *EpochProcess, err error) {
	validators, err := state.Validators()
	if err != nil {
		return nil, err
	}
	count, err := validators.Length()
	if err != nil {
		return nil, err
	}

	prevEpoch := epc.PreviousEpoch.Epoch
	currentEpoch := epc.CurrentEpoch.Epoch

	out = &EpochProcess{
		Statuses:  make([]AttesterStatus, count, count),
		PrevEpoch: prevEpoch,
		CurrEpoch: currentEpoch,
	}

	slashingsEpoch := currentEpoch + (EPOCHS_PER_SLASHINGS_VECTOR / 2)
	exitQueueEnd := currentEpoch.ComputeActivationExitEpoch()

	activeCount := uint64(0)
	valIter := validators.ReadonlyIter()
	for i := ValidatorIndex(0); true; i++ {
		valContainer, ok, err := valIter.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		val, err := AsValidator(valContainer, nil)
		if err != nil {
			return nil, err
		}
		flat, err := ToFlatValidator(val)
		if err != nil {
			return nil, err
		}

		status := &out.Statuses[i]
		status.Validator = flat
		status.AttestedProposer = ValidatorIndexMarker

		if flat.Slashed {
			if slashingsEpoch == flat.WithdrawableEpoch {
				out.IndicesToSlash = append(out.IndicesToSlash, i)
			}
		} else {
			status.Flags |= UnslashedAttester
		}

		if flat.IsActive(prevEpoch) || (flat.Slashed && (prevEpoch+1 < flat.WithdrawableEpoch)) {
			status.Flags |= EligibleAttester
		}

		status.Active = flat.IsActive(currentEpoch)
		if status.Active {
			activeCount++
			out.TotalActiveStake += flat.EffectiveBalance
		}

		if flat.ActivationEligibilityEpoch == FAR_FUTURE_EPOCH && flat.EffectiveBalance == MAX_EFFECTIVE_BALANCE {
			out.IndicesToSetActivationEligibility = append(out.IndicesToSetActivationEligibility, i)
		}

		if flat.ActivationEpoch == FAR_FUTURE_EPOCH && flat.ActivationEligibilityEpoch <= currentEpoch {
			out.IndicesToMaybeActivate = append(out.IndicesToMaybeActivate, i)
		}

		if status.Active && flat.EffectiveBalance <= EJECTION_BALANCE && flat.ExitEpoch == FAR_FUTURE_EPOCH {
			out.IndicesToEject = append(out.IndicesToEject, i)
		}
	}

	// Order by the sequence of activation_eligibility_epoch setting and then index
	sort.Slice(out.IndicesToMaybeActivate, func(i int, j int) bool {
		valIndexA := out.IndicesToMaybeActivate[i]
		valIndexB := out.IndicesToMaybeActivate[j]
		a := out.Statuses[valIndexA].Validator.ActivationEligibilityEpoch
		b := out.Statuses[valIndexB].Validator.ActivationEligibilityEpoch
		if a == b { // Order by the sequence of activation_eligibility_epoch setting and then index
			return valIndexA < valIndexB
		}
		return a < b
	})

	exitQueueEndChurn := uint64(0)
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		status := &out.Statuses[i]
		if status.Validator.ExitEpoch == exitQueueEnd {
			exitQueueEndChurn++
		}
	}
	churnLimit := GetChurnLimit(activeCount)
	if exitQueueEndChurn >= churnLimit {
		exitQueueEnd++
		exitQueueEndChurn = 0
	}
	out.ExitQueueEndChurn = exitQueueEndChurn
	out.ExitQueueEnd = exitQueueEnd
	out.ChurnLimit = churnLimit

	processEpoch := func(
		attestations *PendingAttestationsView,
		epoch Epoch,
		sourceFlag, targetFlag, headFlag AttesterFlag) error {

		actualTargetBlockRoot, err := state.GetBlockRootAtSlot(epoch.GetStartSlot())
		if err != nil {
			return err
		}
		participants := make([]ValidatorIndex, 0, MAX_VALIDATORS_PER_COMMITTEE)
		attIter := attestations.ReadonlyIter()
		for {
			el, ok, err := attIter.Next()
			if err != nil {
				return err
			}
			if !ok {
				break
			}
			attView, err := AsPendingAttestation(el, nil)
			if err != nil {
				return err
			}
			att, err := attView.Raw()
			if err != nil {
				return err
			}

			attBlockRoot, err := state.GetBlockRootAtSlot(att.Data.Slot)
			if err != nil {
				return err
			}

			// attestation-target is already known to be this epoch, get it from the pre-computed shuffling directly.
			committee, err := epc.GetBeaconCommittee(att.Data.Slot, att.Data.Index)
			if err != nil {
				return err
			}

			participants = participants[:0]                                     // reset old slice (re-used in for loop)
			participants = append(participants, committee...)                   // add committee indices
			participants = att.AggregationBits.FilterParticipants(participants) // only keep the participants

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

			for _, p := range participants {
				status := &out.Statuses[p]

				// remember the participant as one of the good validators
				status.Flags |= sourceFlag

				// If the attestation is for the boundary:
				if att.Data.Target.Root == actualTargetBlockRoot {
					status.Flags |= targetFlag

					// If the attestation is for the head (att the time of attestation):
					if att.Data.BeaconBlockRoot == attBlockRoot {
						status.Flags |= headFlag
					}
				}
			}
		}
		return nil
	}
	prevAtts, err := state.PreviousEpochAttestations()
	if err != nil {
		return nil, err
	}
	if err := processEpoch(prevAtts, prevEpoch,
		PrevSourceAttester, PrevTargetAttester, PrevHeadAttester); err != nil {
		return nil, err
	}
	currAtts, err := state.CurrentEpochAttestations()
	if err != nil {
		return nil, err
	}
	if err := processEpoch(currAtts, currentEpoch,
		CurrSourceAttester, CurrTargetAttester, CurrHeadAttester); err != nil {
		return nil, err
	}

	for i := 0; i < len(out.Statuses); i++ {
		status := out.Statuses[i]
		// nested, since they are subsets anyway
		if status.Flags.HasMarkers(PrevSourceAttester | UnslashedAttester) {
			out.PrevEpochUnslashedStake.SourceStake += status.Validator.EffectiveBalance
			// already know it's unslashed, just look if attesting target, then head
			if status.Flags.HasMarkers(PrevTargetAttester) {
				out.PrevEpochUnslashedStake.TargetStake += status.Validator.EffectiveBalance
				if status.Flags.HasMarkers(PrevHeadAttester) {
					out.PrevEpochUnslashedStake.HeadStake += status.Validator.EffectiveBalance
				}
			}
		}
		if status.Flags.HasMarkers(PrevTargetAttester) {
			out.PrevEpochTargetStake += status.Validator.EffectiveBalance
		}
		if status.Flags.HasMarkers(CurrTargetAttester) {
			out.CurrEpochTargetStake += status.Validator.EffectiveBalance
		}
	}
	if out.TotalActiveStake < EFFECTIVE_BALANCE_INCREMENT {
		out.TotalActiveStake = EFFECTIVE_BALANCE_INCREMENT
	}
	if out.PrevEpochUnslashedStake.SourceStake < EFFECTIVE_BALANCE_INCREMENT {
		out.PrevEpochUnslashedStake.SourceStake = EFFECTIVE_BALANCE_INCREMENT
	}
	if out.PrevEpochUnslashedStake.TargetStake < EFFECTIVE_BALANCE_INCREMENT {
		out.PrevEpochUnslashedStake.TargetStake = EFFECTIVE_BALANCE_INCREMENT
	}
	if out.PrevEpochUnslashedStake.HeadStake < EFFECTIVE_BALANCE_INCREMENT {
		out.PrevEpochUnslashedStake.HeadStake = EFFECTIVE_BALANCE_INCREMENT
	}
	if out.PrevEpochTargetStake < EFFECTIVE_BALANCE_INCREMENT {
		out.PrevEpochTargetStake = EFFECTIVE_BALANCE_INCREMENT
	}
	if out.CurrEpochTargetStake < EFFECTIVE_BALANCE_INCREMENT {
		out.CurrEpochTargetStake = EFFECTIVE_BALANCE_INCREMENT
	}

	return
}
