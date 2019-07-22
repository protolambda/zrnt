package deposits

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/merkle"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

type DepositFeature struct {
	Meta interface {
		meta.Pubkeys
		meta.Deposits
		meta.Balance
		meta.Onboarding
		meta.Depositing
	}
}

// Verify that outstanding deposits are processed up to the maximum number of deposits, then process all in order.
func (f *DepositFeature) ProcessDeposits(ops []Deposit) error {
	depositCount := DepositIndex(len(ops))
	expectedCount := f.Meta.DepCount() - f.Meta.DepIndex()
	if expectedCount > MAX_DEPOSITS {
		expectedCount = MAX_DEPOSITS
	}
	if depositCount != expectedCount {
		return errors.New("block does not contain expected deposits amount")
	}

	for i := range ops {
		if err := f.ProcessDeposit(&ops[i]); err != nil {
			return err
		}
	}
	return nil
}

type Deposit struct {
	Proof [DEPOSIT_CONTRACT_TREE_DEPTH + 1]Root // Merkle-path to deposit data list root
	Data  DepositData
}

// Process an Eth1 deposit, registering a validator or increasing its balance.
func (f *DepositFeature) ProcessDeposit(dep *Deposit) error {
	depositIndex := f.Meta.DepIndex()

	// Verify the Merkle branch
	if !merkle.VerifyMerkleBranch(
		ssz.HashTreeRoot(&dep.Data, DepositDataSSZ),
		dep.Proof[:],
		DEPOSIT_CONTRACT_TREE_DEPTH+1, // Add 1 for the `List` length mix-in
		uint64(depositIndex),
		f.Meta.DepRoot()) {
		return fmt.Errorf("deposit %d merkle proof failed to be verified", depositIndex)
	}

	// Increment the next deposit index we are expecting. Note that this
	// needs to be done here because while the deposit contract will never
	// create an invalid Merkle branch, it may admit an invalid deposit
	// object, and we need to be able to skip over it
	f.Meta.IncrementDepositIndex()

	valIndex, exists := f.Meta.ValidatorIndex(dep.Data.Pubkey)

	// Check if it is a known validator that is depositing ("if pubkey not in validator_pubkeys")
	if !exists {
		// Verify the deposit signature (proof of possession) for new validators.
		// Only unknown pubkeys need to be verified, others are already trusted
		// Note: The deposit contract does not check signatures.
		// Note: Deposits are valid across forks, thus the deposit domain is retrieved directly from ComputeDomain().
		if !bls.BlsVerify(
			dep.Data.Pubkey,
			ssz.SigningRoot(dep.Data, DepositDataSSZ),
			dep.Data.Signature,
			ComputeDomain(DOMAIN_DEPOSIT, Version{})) {
			return errors.New("could not verify BLS signature")
		}

		// Add validator and balance entries
		f.Meta.AddNewValidator(dep.Data.Pubkey, dep.Data.WithdrawalCredentials, dep.Data.Amount)

	} else {
		// Increase balance by deposit amount
		f.Meta.IncreaseBalance(valIndex, dep.Data.Amount)
	}
	return nil
}
