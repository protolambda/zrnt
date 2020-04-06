package beacon

import (
	"errors"
	"github.com/protolambda/zrnt/eth2/util/math"
)

type Deltas struct {
	// element for each validator in registry
	Rewards []Gwei
	// element for each validator in registry
	Penalties []Gwei
}

func NewDeltas(validatorCount uint64) *Deltas {
	return &Deltas{
		Rewards:   make([]Gwei, validatorCount, validatorCount),
		Penalties: make([]Gwei, validatorCount, validatorCount),
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

func (state *BeaconStateView) AttestationDeltas(epc *EpochsContext, process *EpochProcess) (*Deltas, error) {
	deltas := NewDeltas(epc.ValCount())

	previousEpoch := epc.PreviousEpoch.Epoch

	attesterStatuses := process.Statuses

	totalBalance := process.TotalActiveUnslashedStake

	prevEpochStake := &process.PrevEpochStake
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

	validatorCount := ValidatorIndex(epc.ValCount())
	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := attesterStatuses[i]
		if status.Flags&EligibleAttester != 0 {

			effBalance := status.Validator.EffectiveBalance
			baseReward := effBalance * BASE_REWARD_FACTOR /
				balanceSqRoot / BASE_REWARDS_PER_EPOCH

			// Expected FFG source
			if status.Flags.HasMarkers(PrevSourceAttester | UnslashedAttester) {
				// Justification-participation reward
				deltas.Rewards[i] += baseReward * prevEpochSourceStake / totalBalance

				// Inclusion speed bonus
				proposerReward := baseReward / PROPOSER_REWARD_QUOTIENT
				deltas.Rewards[status.AttestedProposer] += proposerReward
				maxAttesterReward := baseReward - proposerReward
				deltas.Rewards[i] += maxAttesterReward / Gwei(status.InclusionDelay)
			} else {
				//Justification-non-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Expected FFG target
			if status.Flags.HasMarkers(PrevTargetAttester | UnslashedAttester) {
				// Boundary-attestation reward
				deltas.Rewards[i] += baseReward * prevEpochTargetStake / totalBalance
			} else {
				//Boundary-attestation-non-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Expected head
			if status.Flags.HasMarkers(PrevHeadAttester | UnslashedAttester) {
				// Canonical-participation reward
				deltas.Rewards[i] += baseReward * prevEpochHeadStake / totalBalance
			} else {
				// Non-canonical-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Take away max rewards if we're not finalizing
			if finalityDelay > MIN_EPOCHS_TO_INACTIVITY_PENALTY {
				deltas.Penalties[i] += baseReward * BASE_REWARDS_PER_EPOCH
				if !status.Flags.HasMarkers(PrevHeadAttester | UnslashedAttester) {
					deltas.Penalties[i] += effBalance * Gwei(finalityDelay) / INACTIVITY_PENALTY_QUOTIENT
				}
			}
		}
	}

	return deltas, nil
}

func (state *BeaconStateView) ProcessEpochRewardsAndPenalties(epc *EpochsContext, process *EpochProcess) error {
	currentEpoch := epc.CurrentEpoch.Epoch
	if currentEpoch == GENESIS_EPOCH {
		return nil
	}
	valCount := epc.ValCount()
	sum := NewDeltas(valCount)
	attDeltas, err := state.AttestationDeltas(epc, process)
	if err != nil {
		return err
	}
	sum.Add(attDeltas)

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
	// TODO: can be optimized a lot, make a new tree in one go
	for i := ValidatorIndex(0); i < ValidatorIndex(balLen); i++ {
		if err := balances.IncreaseBalance(i, sum.Rewards[i]); err != nil {
			return err
		}
		if err := balances.DecreaseBalance(i, sum.Penalties[i]); err != nil {
			return err
		}
	}
	return nil
}

