package beacon

//import (
//	"errors"
//	. "github.com/protolambda/zrnt/eth2/beacon/deposits"
//	. "github.com/protolambda/zrnt/eth2/beacon/eth1"
//	"github.com/protolambda/zrnt/eth2/beacon/header"
//	"github.com/protolambda/zrnt/eth2/beacon/registry"
//	. "github.com/protolambda/zrnt/eth2/beacon/versioning"
//
//	"github.com/protolambda/zrnt/eth2/util/ssz"
//	"github.com/protolambda/zssz"
//)
//
//// TODO genesis tree state
//
//type DepositRoots []Root
//
//func (_ *DepositRoots) Limit() uint64 {
//	return 1 << DEPOSIT_CONTRACT_TREE_DEPTH
//}
//
//var DepositRootsSSZ = zssz.GetSSZ((*DepositRoots)(nil))
//
//func GenesisFromEth1(eth1BlockHash Root, time Timestamp, deps []Deposit, verifyDeposits bool) (*FullFeaturedState, error) {
//	state := &BeaconState{
//		VersioningState: VersioningState{
//			GenesisTime: time - (time % MIN_GENESIS_DELAY) + (2 * MIN_GENESIS_DELAY),
//			Fork: Fork{
//				PreviousVersion: GENESIS_FORK_VERSION,
//				CurrentVersion:  GENESIS_FORK_VERSION,
//				Epoch:           GENESIS_EPOCH,
//			},
//		},
//		// Ethereum 1.0 chain data
//		Eth1State: Eth1State{
//			Eth1Data: Eth1Data{
//				DepositRoot:  Root{}, // incrementally overwritten during deposit processing
//				DepositCount: DepositIndex(len(deps)),
//				BlockHash:    eth1BlockHash,
//			},
//		},
//		BlockHeaderState: header.BlockHeaderState{
//			LatestBlockHeader: header.BeaconBlockHeader{
//				BodyRoot: ssz.HashTreeRoot(BeaconBlockBody{}, BeaconBlockBodySSZ),
//			},
//		},
//	}
//	// Seed RANDAO with Eth1 entropy
//	state.SeedRandao(eth1BlockHash)
//
//	depProcessor := &DepositFeature{Meta: state}
//
//	depRoots := make(DepositRoots, 0, len(deps))
//	// Pre-process deposits: get roots
//	for i := range deps {
//		depRoots = append(depRoots, ssz.HashTreeRoot(&deps[i].Data, DepositDataSSZ))
//	}
//	if verifyDeposits {
//		// Process deposits
//		for i := range deps {
//			roots := DepositRoots(depRoots[:i+1])
//			state.Eth1Data.DepositRoot = ssz.HashTreeRoot(&roots, DepositRootsSSZ)
//			// in the rare case someone tries to create a genesis block using invalid data, panic.
//			if err := depProcessor.ProcessDeposit(&deps[i]); err != nil {
//				return nil, err
//			}
//		}
//	} else {
//		// Pre-process deposits: get roots
//		for i := range deps {
//			dat := &deps[i].Data
//			state.AddNewValidator(dat.Pubkey, dat.WithdrawalCredentials, dat.Amount)
//		}
//		state.DepositIndex = DepositIndex(len(deps))
//	}
//	state.Eth1Data.DepositRoot = ssz.HashTreeRoot(&depRoots, DepositRootsSSZ)
//	return InitState(state)
//}
//
//// After creating a state and onboarding validators,
//// process the new validators as genesis-validators, and initialize other state variables.
//func InitState(state *BeaconState) (*FullFeaturedState, error) {
//	if Slot(len(state.Validators)) < SLOTS_PER_EPOCH {
//		return nil, errors.New("not enough validators to init full featured BeaconState")
//	}
//	// Process activations
//	state.UpdateEffectiveBalances()
//	for _, v := range state.Validators {
//		if v.EffectiveBalance == MAX_EFFECTIVE_BALANCE {
//			v.ActivationEligibilityEpoch = GENESIS_EPOCH
//			v.ActivationEpoch = GENESIS_EPOCH
//		}
//	}
//	state.GenesisValidatorsRoot = ssz.HashTreeRoot(state.Validators, registry.ValidatorRegistrySSZ)
//	// Now that validators are activated, we can load the full feature set.
//	// Committees will now be pre-computed.
//	full := NewFullFeaturedState(state)
//	// pre-compute the committee data
//	full.LoadPrecomputedData()
//	return full, nil
//}
//
//func IsValidGenesisState(state *BeaconState) bool {
//	if state.GenesisTime < MIN_GENESIS_TIME {
//		return false
//	}
//	if state.GetActiveValidatorCount(GENESIS_EPOCH) < MIN_GENESIS_ACTIVE_VALIDATOR_COUNT {
//		return false
//	}
//	return true
//}
