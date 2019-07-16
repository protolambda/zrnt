package operations

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/merkle"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

type Deposits []Deposit

func (_ *Deposits) Limit() uint32 {
	return MAX_DEPOSITS
}

func (ops Deposits) Process(state *BeaconState) error {
	depositCount := DepositIndex(len(ops))
	expectedCount := state.LatestEth1Data.DepositCount - state.DepositIndex
	if expectedCount > MAX_DEPOSITS {
		expectedCount = MAX_DEPOSITS
	}
	if depositCount != expectedCount {
		return errors.New("block does not contain expected deposits amount")
	}

	for _, op := range ops {
		if err := op.Process(state); err != nil {
			return err
		}
	}
	return nil
}

type Deposit struct {
	Proof [DEPOSIT_CONTRACT_TREE_DEPTH + 1]Root // Merkle-path to deposit data list root
	Data DepositData
}

// Process an Eth1 deposit, registering a validator or increasing its balance.
func (dep *Deposit) Process(state *BeaconState) error {

	// Verify the Merkle branch
	if !merkle.VerifyMerkleBranch(
		ssz.HashTreeRoot(&dep.Data, DepositDataSSZ),
		dep.Proof[:],
		DEPOSIT_CONTRACT_TREE_DEPTH + 1,
		uint64(state.DepositIndex),
		state.LatestEth1Data.DepositRoot) {
		return fmt.Errorf("deposit %d merkle proof failed to be verified", state.DepositIndex)
	}

	// Increment the next deposit index we are expecting. Note that this
	// needs to be done here because while the deposit contract will never
	// create an invalid Merkle branch, it may admit an invalid deposit
	// object, and we need to be able to skip over it
	state.DepositIndex += 1

	valIndex := ValidatorIndexMarker
	for i, v := range state.Validators {
		if v.Pubkey == dep.Data.Pubkey {
			valIndex = ValidatorIndex(i)
			break
		}
	}

	// Check if it is a known validator that is depositing ("if pubkey not in validator_pubkeys")
	if valIndex == ValidatorIndexMarker {
		// only unknown pubkeys need to be verified, others are already trusted
		// Note: Deposits are valid across forks, thus the deposit domain is retrieved directly from ComputeDomain().
		if !bls.BlsVerify(
			dep.Data.Pubkey,
			ssz.SigningRoot(dep.Data, DepositDataSSZ),
			dep.Data.Signature,
			ComputeDomain(DOMAIN_DEPOSIT, Version{})) {
			return errors.New("could not verify BLS signature")
		}

		effBalance := dep.Data.Amount - (dep.Data.Amount % EFFECTIVE_BALANCE_INCREMENT)
		if effBalance > MAX_EFFECTIVE_BALANCE {
			effBalance = MAX_EFFECTIVE_BALANCE
		}
		// Add validator and balance entries
		validator := &Validator{
			Pubkey:                     dep.Data.Pubkey,
			WithdrawalCredentials:      dep.Data.WithdrawalCredentials,
			ActivationEligibilityEpoch: FAR_FUTURE_EPOCH,
			ActivationEpoch:            FAR_FUTURE_EPOCH,
			ExitEpoch:                  FAR_FUTURE_EPOCH,
			WithdrawableEpoch:          FAR_FUTURE_EPOCH,
			EffectiveBalance:           effBalance,
		}
		state.Validators = append(state.Validators, validator)
		state.Balances = append(state.Balances, dep.Data.Amount)
	} else {
		// Increase balance by deposit amount
		state.Balances.IncreaseBalance(valIndex, dep.Data.Amount)
	}
	return nil
}
