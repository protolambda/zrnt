package phase0

import (
	. "github.com/protolambda/zrnt/eth2/beacon/deposits"
	. "github.com/protolambda/zrnt/eth2/beacon/eth1"
	"github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/beacon/versioning"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type DepositRoots []Root

func (_ *DepositRoots) Limit() uint64 {
	return 1 << DEPOSIT_CONTRACT_TREE_DEPTH
}

var DepositRootsSSZ = zssz.GetSSZ((*DepositRoots)(nil))

func GenesisFromEth1(eth1BlockHash Root, time Timestamp, deps []Deposit) (*FullFeaturedState, error) {
	state := &BeaconState{
		VersioningState: VersioningState{
			GenesisTime: time - (time % SECONDS_PER_DAY) + (2 * SECONDS_PER_DAY),
		},
		// Ethereum 1.0 chain data
		Eth1State: Eth1State{
			Eth1Data: Eth1Data{
				DepositRoot:  Root{}, // incrementally overwritten during deposit processing
				DepositCount: DepositIndex(len(deps)),
				BlockHash:    eth1BlockHash,
			},
		},
		BlockHeaderState: header.BlockHeaderState{
			LatestBlockHeader: header.BeaconBlockHeader{
				BodyRoot: ssz.HashTreeRoot(BeaconBlockBody{}, BeaconBlockBodySSZ),
			},
		},
	}
	depProcessor := &DepositFeature{Meta: state}

	depRoots := make(DepositRoots, 0, len(deps))
	// Pre-process deposits: get roots
	for i := range deps {
		depRoots = append(depRoots, ssz.HashTreeRoot(&deps[i].Data, DepositDataSSZ))
	}
	// Process deposits
	for i := range deps {
		roots := DepositRoots(depRoots[:i+1])
		state.Eth1Data.DepositRoot = ssz.HashTreeRoot(&roots, DepositRootsSSZ)
		// in the rare case someone tries to create a genesis block using invalid data, panic.
		if err := depProcessor.ProcessDeposit(&deps[i]); err != nil {
			return nil, err
		}
	}
	return InitState(state), nil
}

// After creating a state and onboarding validators,
// process the new validators as genesis-validators, and initialize other state variables.
func InitState(state *BeaconState) *FullFeaturedState {
	// Process activations
	state.UpdateEffectiveBalances()
	for _, v := range state.Validators {
		if v.EffectiveBalance == MAX_EFFECTIVE_BALANCE {
			v.ActivationEligibilityEpoch = GENESIS_EPOCH
			v.ActivationEpoch = GENESIS_EPOCH
		}
	}
	// Now that validators are activated, we can load the full feature set.
	// Committees will now be pre-computed.
	full := NewFullFeaturedState(state)
	// pre-compute the committee data
	full.LoadPrecomputedData()
	// Populate active_index_roots and compact_committees_roots
	activeIndexRoot := full.ComputeActiveIndexRoot(GENESIS_EPOCH)
	committeeRoot := full.ComputeCompactCommitteesRoot(GENESIS_EPOCH)
	for i := Epoch(0); i < EPOCHS_PER_HISTORICAL_VECTOR; i++ {
		state.ActiveIndexRoots[i] = activeIndexRoot
		state.CompactCommitteesRoots[i] = committeeRoot
	}
	return full
}
