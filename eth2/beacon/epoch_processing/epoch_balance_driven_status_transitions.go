package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
)

func ProcessBalanceDrivenStatusTransitions(state *BeaconState) {
	// Iterate through the validator registry
	//    and deposit or eject active validators with sufficiently high or low balances
	currentEpoch := state.Epoch()
	for _, vIndex := range state.ValidatorRegistry.GetActiveValidatorIndices(state.Epoch()) {
		v := state.ValidatorRegistry[vIndex]
		balance := state.GetBalance(vIndex)
		if v.ActivationEligibilityEpoch == FAR_FUTURE_EPOCH && balance >= MAX_DEPOSIT_AMOUNT {
			v.ActivationEligibilityEpoch = currentEpoch
		}
		if v.IsActive(currentEpoch) && balance < EJECTION_BALANCE {
			state.InitiateValidatorExit(vIndex)
		}
	}
}
