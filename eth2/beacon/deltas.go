package beacon

import (
	"errors"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)


type GweiList []Gwei

func (_ *GweiList) Limit() uint64 {
	return VALIDATOR_REGISTRY_LIMIT
}

type Deltas struct {
	Rewards   GweiList
	Penalties GweiList
}

var DeltasSSZ = zssz.GetSSZ((*Deltas)(nil))

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


func (state *BeaconStateView) AttestationRewardsAndPenalties(epc *EpochsContext, process *EpochProcess) (*RewardsAndPenalties, error) {
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
	totalBalance /= EFFECTIVE_BALANCE_INCREMENT
	prevEpochSourceStake /= EFFECTIVE_BALANCE_INCREMENT
	prevEpochTargetStake /= EFFECTIVE_BALANCE_INCREMENT
	prevEpochHeadStake /= EFFECTIVE_BALANCE_INCREMENT

	isInactivityLeak := finalityDelay > MIN_EPOCHS_TO_INACTIVITY_PENALTY

	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := attesterStatuses[i]

		effBalance := status.Validator.EffectiveBalance
		baseReward := effBalance * BASE_REWARD_FACTOR /
			balanceSqRoot / BASE_REWARDS_PER_EPOCH

		// Inclusion delay
		if status.Flags.HasMarkers(PrevSourceAttester | UnslashedAttester) {
			// Inclusion speed bonus
			proposerReward := baseReward / PROPOSER_REWARD_QUOTIENT
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
				proposerReward := baseReward / PROPOSER_REWARD_QUOTIENT
				res.Inactivity.Penalties[i] += BASE_REWARDS_PER_EPOCH * baseReward - proposerReward
				if !status.Flags.HasMarkers(PrevTargetAttester | UnslashedAttester) {
					res.Inactivity.Penalties[i] += effBalance * Gwei(finalityDelay) / INACTIVITY_PENALTY_QUOTIENT
				}
			}
		}
	}

	return res, nil
}

func (state *BeaconStateView) ProcessEpochRewardsAndPenalties(epc *EpochsContext, process *EpochProcess) error {
	currentEpoch := epc.CurrentEpoch.Epoch
	if currentEpoch == GENESIS_EPOCH {
		return nil
	}
	valCount := uint64(len(process.Statuses))
	sum := NewDeltas(valCount)
	rewAndPenalties, err := state.AttestationRewardsAndPenalties(epc, process)
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

	newBalancesTree, err := RegistryBalancesType.FromElements(balancesElements...)
	if err != nil {
		return err
	}
	return balances.SetBacking(newBalancesTree.Backing())
}
