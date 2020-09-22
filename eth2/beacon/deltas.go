package beacon

import (
	"context"
	"errors"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type GweiList []Gwei

func (a *GweiList) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, Gwei(0))
		return &(*a)[i]
	}, GweiType.TypeByteLength(), spec.VALIDATOR_REGISTRY_LIMIT)
}

func (a GweiList) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, GweiType.TypeByteLength(), uint64(len(a)))
}

func (a GweiList) ByteLength(spec *Spec) (out uint64) {
	return uint64(len(a)) * GweiType.TypeByteLength()
}

func (a *GweiList) FixedLength(spec *Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li GweiList) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	length := uint64(len(li))
	return hFn.Uint64ListHTR(func(i uint64) uint64 {
		return uint64(li[i])
	}, length, spec.VALIDATOR_REGISTRY_LIMIT)
}

type Deltas struct {
	Rewards   GweiList
	Penalties GweiList
}

func (a *Deltas) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&a.Rewards), spec.Wrap(&a.Penalties))
}

func (a *Deltas) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&a.Rewards), spec.Wrap(&a.Penalties))
}

func (a *Deltas) ByteLength(spec *Spec) uint64 {
	return 2*codec.OFFSET_SIZE + a.Rewards.ByteLength(spec) + a.Penalties.ByteLength(spec)
}

func (a *Deltas) FixedLength(*Spec) uint64 {
	return 0
}

func (a *Deltas) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(spec.Wrap(&a.Rewards), spec.Wrap(&a.Penalties))
}

func NewDeltas(validatorCount uint64) *Deltas {
	return &Deltas{
		Rewards:   make(GweiList, validatorCount, validatorCount),
		Penalties: make(GweiList, validatorCount, validatorCount),
	}
}

func (deltas *Deltas) Add(other *Deltas) {
	for i := 0; i < len(deltas.Rewards); i++ {
		deltas.Rewards[i] += other.Rewards[i]
	}
	for i := 0; i < len(deltas.Penalties); i++ {
		deltas.Penalties[i] += other.Penalties[i]
	}
}

type RewardsAndPenalties struct {
	Source         *Deltas
	Target         *Deltas
	Head           *Deltas
	InclusionDelay *Deltas
	Inactivity     *Deltas
}

func NewRewardsAndPenalties(validatorCount uint64) *RewardsAndPenalties {
	return &RewardsAndPenalties{
		Source:         NewDeltas(validatorCount),
		Target:         NewDeltas(validatorCount),
		Head:           NewDeltas(validatorCount),
		InclusionDelay: NewDeltas(validatorCount),
		Inactivity:     NewDeltas(validatorCount),
	}
}

func (spec *Spec) AttestationRewardsAndPenalties(ctx context.Context,
	epc *EpochsContext, process *EpochProcess, state *BeaconStateView) (*RewardsAndPenalties, error) {

	validatorCount := ValidatorIndex(uint64(len(process.Statuses)))
	res := NewRewardsAndPenalties(uint64(validatorCount))

	previousEpoch := epc.PreviousEpoch.Epoch

	attesterStatuses := process.Statuses

	totalBalance := process.TotalActiveStake

	prevEpochStake := &process.PrevEpochUnslashedStake
	prevEpochSourceStake := prevEpochStake.SourceStake
	prevEpochTargetStake := prevEpochStake.TargetStake
	prevEpochHeadStake := prevEpochStake.HeadStake

	balanceSqRoot := Gwei(math.IntegerSquareroot(uint64(totalBalance)))
	finalized, err := state.FinalizedCheckpoint()
	if err != nil {
		return nil, err
	}
	finalizedEpoch, err := finalized.Epoch()
	if err != nil {
		return nil, err
	}
	finalityDelay := previousEpoch - finalizedEpoch

	// All summed effective balances are normalized to effective-balance increments, to avoid overflows.
	totalBalance /= spec.EFFECTIVE_BALANCE_INCREMENT
	prevEpochSourceStake /= spec.EFFECTIVE_BALANCE_INCREMENT
	prevEpochTargetStake /= spec.EFFECTIVE_BALANCE_INCREMENT
	prevEpochHeadStake /= spec.EFFECTIVE_BALANCE_INCREMENT

	isInactivityLeak := finalityDelay > spec.MIN_EPOCHS_TO_INACTIVITY_PENALTY

	for i := ValidatorIndex(0); i < validatorCount; i++ {
		// every 1024 validators, check if the context is done.
		if i&((1<<10)-1) == 0 {
			select {
			case <-ctx.Done():
				return nil, TransitionCancelErr
			default: // Don't block.
				break
			}
		}
		status := attesterStatuses[i]

		effBalance := status.Validator.EffectiveBalance
		baseReward := effBalance * Gwei(spec.BASE_REWARD_FACTOR) /
			balanceSqRoot / BASE_REWARDS_PER_EPOCH

		// Inclusion delay
		if status.Flags.HasMarkers(PrevSourceAttester | UnslashedAttester) {
			// Inclusion speed bonus
			proposerReward := baseReward / Gwei(spec.PROPOSER_REWARD_QUOTIENT)
			res.InclusionDelay.Rewards[status.AttestedProposer] += proposerReward
			maxAttesterReward := baseReward - proposerReward
			res.InclusionDelay.Rewards[i] += maxAttesterReward / Gwei(status.InclusionDelay)
		}

		if status.Flags&EligibleAttester != 0 {
			// Since full base reward will be canceled out by inactivity penalty deltas,
			// optimal participation receives full base reward compensation here.

			// Expected FFG source
			if status.Flags.HasMarkers(PrevSourceAttester | UnslashedAttester) {
				if isInactivityLeak {
					res.Source.Rewards[i] += baseReward
				} else {
					// Justification-participation reward
					res.Source.Rewards[i] += baseReward * prevEpochSourceStake / totalBalance
				}
			} else {
				//Justification-non-participation R-penalty
				res.Source.Penalties[i] += baseReward
			}

			// Expected FFG target
			if status.Flags.HasMarkers(PrevTargetAttester | UnslashedAttester) {
				if isInactivityLeak {
					res.Target.Rewards[i] += baseReward
				} else {
					// Boundary-attestation reward
					res.Target.Rewards[i] += baseReward * prevEpochTargetStake / totalBalance
				}
			} else {
				//Boundary-attestation-non-participation R-penalty
				res.Target.Penalties[i] += baseReward
			}

			// Expected head
			if status.Flags.HasMarkers(PrevHeadAttester | UnslashedAttester) {
				if isInactivityLeak {
					res.Head.Rewards[i] += baseReward
				} else {
					// Canonical-participation reward
					res.Head.Rewards[i] += baseReward * prevEpochHeadStake / totalBalance
				}
			} else {
				// Non-canonical-participation R-penalty
				res.Head.Penalties[i] += baseReward
			}

			// Take away max rewards if we're not finalizing
			if isInactivityLeak {
				// If validator is performing optimally this cancels all rewards for a neutral balance
				proposerReward := baseReward / Gwei(spec.PROPOSER_REWARD_QUOTIENT)
				res.Inactivity.Penalties[i] += BASE_REWARDS_PER_EPOCH*baseReward - proposerReward
				if !status.Flags.HasMarkers(PrevTargetAttester | UnslashedAttester) {
					res.Inactivity.Penalties[i] += effBalance * Gwei(finalityDelay) / Gwei(spec.INACTIVITY_PENALTY_QUOTIENT)
				}
			}
		}
	}

	return res, nil
}

func (spec *Spec) ProcessEpochRewardsAndPenalties(ctx context.Context, epc *EpochsContext, process *EpochProcess, state *BeaconStateView) error {
	select {
	case <-ctx.Done():
		return TransitionCancelErr
	default: // Don't block.
		break
	}
	currentEpoch := epc.CurrentEpoch.Epoch
	if currentEpoch == GENESIS_EPOCH {
		return nil
	}
	valCount := uint64(len(process.Statuses))
	sum := NewDeltas(valCount)
	rewAndPenalties, err := spec.AttestationRewardsAndPenalties(ctx, epc, process, state)
	if err != nil {
		return err
	}
	sum.Add(rewAndPenalties.Source)
	sum.Add(rewAndPenalties.Target)
	sum.Add(rewAndPenalties.Head)
	sum.Add(rewAndPenalties.InclusionDelay)
	sum.Add(rewAndPenalties.Inactivity)

	balances, err := state.Balances()
	if err != nil {
		return err
	}
	balLen, err := balances.Length()
	if err != nil {
		return err
	}
	if uint64(len(sum.Penalties)) != balLen || uint64(len(sum.Rewards)) != balLen {
		return errors.New("cannot apply deltas to balances list with different length")
	}
	balancesElements := make([]BasicView, 0, balLen)
	balIter := balances.ReadonlyIter()
	i := ValidatorIndex(0)
	for {
		el, ok, err := balIter.Next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		bal, err := AsGwei(el, err)
		if err != nil {
			return err
		}
		bal += sum.Rewards[i]
		if penalty := sum.Penalties[i]; bal >= penalty {
			bal -= penalty
		} else {
			bal = 0
		}
		balancesElements = append(balancesElements, Uint64View(bal))
		i++
	}

	newBalancesTree, err := spec.RegistryBalances().FromElements(balancesElements...)
	if err != nil {
		return err
	}
	return balances.SetBacking(newBalancesTree.Backing())
}
