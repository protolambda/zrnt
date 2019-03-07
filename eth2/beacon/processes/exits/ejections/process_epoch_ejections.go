package ejections

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
)

func ProcessEpochEjections(state *beacon.BeaconState) {
	// After we are done slashing, eject the validators that don't have enough balance left.
	for _, vIndex := range state.Validator_registry.Get_active_validator_indices(state.Epoch()) {
		if state.Validator_balances[vIndex] < beacon.EJECTION_BALANCE {
			state.Exit_validator(vIndex)
		}
	}
}
