package phase0

import (
	. "github.com/protolambda/zrnt/eth2/beacon/eth1"
	. "github.com/protolambda/zrnt/eth2/beacon/versioning"
	. "github.com/protolambda/zrnt/eth2/core"
)

type KickstartValidatorData struct {
	Pubkey                BLSPubkey
	WithdrawalCredentials Root
	Balance               Gwei
}

// To build a genesis state without Eth 1.0 deposits, i.e. directly from a sequence of minimal validator data.
func KickStartState(eth1BlockHash Root, time Timestamp, validators []KickstartValidatorData) (*FullFeaturedState, error) {
	state := &BeaconState{
		VersioningState: VersioningState{
			GenesisTime: time,
		},
		// Ethereum 1.0 chain data
		Eth1State: Eth1State{
			Eth1Data: Eth1Data{
				DepositRoot:  Root{}, // incrementally overwritten during deposit processing
				DepositCount: DepositIndex(len(validators)),
				BlockHash:    eth1BlockHash,
			},
		},
	}
	for _, v := range validators {
		state.AddNewValidator(v.Pubkey, v.WithdrawalCredentials, v.Balance)
	}

	return InitState(state)
}
