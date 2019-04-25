package block_processing

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/merkle"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessBlockDeposits(state *BeaconState, block *BeaconBlock) error {
	depositCount := DepositIndex(len(block.Body.Deposits))
	expectedCount := state.LatestEth1Data.DepositCount - state.DepositIndex
	if expectedCount > MAX_DEPOSITS {
		expectedCount = MAX_DEPOSITS
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

// Process an Eth1 deposit, registering a validator or increasing its balance.
func ProcessDeposit(state *BeaconState, dep *Deposit) error {
	// Deposits must be processed in order
	if dep.Index != state.DepositIndex {
		return errors.New(fmt.Sprintf("deposit has index %d that does not match with state index %d", dep.Index, state.DepositIndex))
	}

	// Verify the Merkle branch
	if !merkle.VerifyMerkleBranch(
		ssz.HashTreeRoot(&dep.Data),
		dep.Proof[:],
		DEPOSIT_CONTRACT_TREE_DEPTH,
		uint64(dep.Index),
		state.LatestEth1Data.DepositRoot) {
		return errors.New(fmt.Sprintf("deposit %d has merkle proof that failed to be verified", dep.Index))
	}

	// Increment the next deposit index we are expecting. Note that this
	// needs to be done here because while the deposit contract will never
	// create an invalid Merkle branch, it may admit an invalid deposit
	// object, and we need to be able to skip over it
	state.DepositIndex += 1

	valIndex := ValidatorIndexMarker
	for i, v := range state.ValidatorRegistry {
		if v.Pubkey == dep.Data.Pubkey {
			valIndex = ValidatorIndex(i)
			break
		}
	}

	// Check if it is a known validator that is depositing ("if pubkey not in validator_pubkeys")
	if valIndex == ValidatorIndexMarker {
		// only unknown pubkeys need to be verified, others are already trusted
		if !bls.BlsVerify(
			dep.Data.Pubkey,
			ssz.SigningRoot(dep.Data),
			dep.Data.Signature,
			state.GetDomain(DOMAIN_DEPOSIT, state.Epoch())) {
			return errors.New("could not verify BLS signature")
		}

		// Add validator and balance entries
		validator := &Validator{
			Pubkey:                dep.Data.Pubkey,
			WithdrawalCredentials: dep.Data.WithdrawalCredentials,
			ActivationEligibilityEpoch: FAR_FUTURE_EPOCH,
			ActivationEpoch:       FAR_FUTURE_EPOCH,
			ExitEpoch:             FAR_FUTURE_EPOCH,
			WithdrawableEpoch:     FAR_FUTURE_EPOCH,
			EffectiveBalance: dep.Data.Amount - (dep.Data.Amount % EFFECTIVE_BALANCE_INCREMENT),
		}
		state.ValidatorRegistry = append(state.ValidatorRegistry, validator)
		state.Balances = append(state.Balances, 0)
	} else {
		// Increase balance by deposit amount
		state.IncreaseBalance(valIndex, dep.Data.Amount)
	}
	return nil
}
