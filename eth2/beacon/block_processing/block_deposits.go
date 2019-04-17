package block_processing

import (
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/hash"
	"github.com/protolambda/zrnt/eth2/util/merkle"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessBlockDeposits(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	depositCount := beacon.DepositIndex(len(block.Body.Deposits))
	expectedCount := state.LatestEth1Data.DepositCount - state.DepositIndex
	if expectedCount > beacon.MAX_DEPOSITS {
		expectedCount = beacon.MAX_DEPOSITS
	}
	if depositCount != expectedCount {
		return errors.New("block does not contain expected deposits amount")
	}
	for i := 0; i < len(block.Body.Deposits); i++ {
		dep := &block.Body.Deposits[i]
		if err := ProcessDeposit(state, dep); err != nil {
			return err
		}
		state.DepositIndex += 1
	}
	return nil
}

// Process a deposit from Ethereum 1.0.
// Used to add a validator or top up an existing validator's balance by some deposit amount.
func ProcessDeposit(state *beacon.BeaconState, dep *beacon.Deposit) error {
	// Deposits must be processed in order
	if dep.Index != state.DepositIndex {
		return errors.New(fmt.Sprintf("deposit has index %d that does not match with state index %d", dep.Index, state.DepositIndex))
	}

	serializedDepositData := dep.Data.Serialized()

	// Verify the Merkle branch
	if !merkle.VerifyMerkleBranch(
		hash.Hash(serializedDepositData),
		dep.Proof[:],
		beacon.DEPOSIT_CONTRACT_TREE_DEPTH,
		uint64(dep.Index),
		state.LatestEth1Data.DepositRoot) {
		return errors.New(fmt.Sprintf("deposit %d has merkle proof that failed to be verified", dep.Index))
	}

	// Increment the next deposit index we are expecting. Note that this
	// needs to be done here because while the deposit contract will never
	// create an invalid Merkle branch, it may admit an invalid deposit
	// object, and we need to be able to skip over it
	state.DepositIndex += 1

	valIndex := beacon.ValidatorIndexMarker
	for i, v := range state.ValidatorRegistry {
		if v.Pubkey == dep.Data.Pubkey {
			valIndex = beacon.ValidatorIndex(i)
			break
		}
	}

	// Check if it is a known validator that is depositing ("if pubkey not in validator_pubkeys")
	if valIndex == beacon.ValidatorIndexMarker {
		// only unknown pubkeys need to be verified, others are already trusted
		if !bls.BlsVerify(
			dep.Data.Pubkey,
			ssz.SigningRoot(dep.Data),
			dep.Data.ProofOfPossession,
			beacon.GetDomain(state.Fork, state.Epoch(), beacon.DOMAIN_DEPOSIT)) {
			return errors.New("could not verify BLS signature")
		}

		// Not a known pubkey, add new validator
		validator := &beacon.Validator{
			Pubkey:                dep.Data.Pubkey,
			WithdrawalCredentials: dep.Data.WithdrawalCredentials,
			ActivationEligibilityEpoch: beacon.FAR_FUTURE_EPOCH,
			ActivationEpoch:       beacon.FAR_FUTURE_EPOCH,
			ExitEpoch:             beacon.FAR_FUTURE_EPOCH,
			WithdrawableEpoch:     beacon.FAR_FUTURE_EPOCH,
			Slashed:               false,
			HighBalance:           0,
		}
		// Note: In phase 2 registry indices that have been withdrawn for a long time will be recycled.
		state.ValidatorRegistry = append(state.ValidatorRegistry, validator)
		state.Balances = append(state.Balances, 0)
		x := beacon.ValidatorIndex(len(state.ValidatorRegistry) - 1)
		xStr := x.String()
		fmt.Println(xStr)
		state.SetBalance(x, dep.Data.Amount)
	} else {
		// Increase balance by deposit amount
		state.IncreaseBalance(valIndex, dep.Data.Amount)
	}
	return nil
}
