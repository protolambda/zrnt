package deposits

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/merkle"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

// Verify that outstanding deposits are processed up to the maximum number of deposits, then process all in order.
func ProcessDeposits(meta DepositReq, ops []Deposit) error {
	depositCount := DepositIndex(len(ops))
	expectedCount := meta.DepositCount() - meta.DepositIndex()
	if expectedCount > MAX_DEPOSITS {
		expectedCount = MAX_DEPOSITS
	}
	if depositCount != expectedCount {
		return errors.New("block does not contain expected deposits amount")
	}

	for i := range ops {
		if err := ProcessDeposit(meta, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

type Deposit struct {
	Proof [DEPOSIT_CONTRACT_TREE_DEPTH + 1]Root // Merkle-path to deposit data list root
	Data  DepositData
}

type DepositReq interface {
	PubkeyMeta
	Eth1Meta
	DepositMeta
}

// Process an Eth1 deposit, registering a validator or increasing its balance.
func ProcessDeposit(meta DepositReq, dep *Deposit) error {
	depositIndex := meta.DepositIndex()

	// Verify the Merkle branch
	if !merkle.VerifyMerkleBranch(
		ssz.HashTreeRoot(&dep.Data, DepositDataSSZ),
		dep.Proof[:],
		DEPOSIT_CONTRACT_TREE_DEPTH+1, // Add 1 for the `List` length mix-in
		uint64(depositIndex),
		meta.DepositRoot()) {
		return fmt.Errorf("deposit %d merkle proof failed to be verified", depositIndex)
	}

	// Increment the next deposit index we are expecting. Note that this
	// needs to be done here because while the deposit contract will never
	// create an invalid Merkle branch, it may admit an invalid deposit
	// object, and we need to be able to skip over it
	meta.IncrementDepositIndex()

	valIndex, exists := meta.ValidatorIndex(dep.Data.Pubkey)

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
		meta.AddNewValidator(dep.Data.Pubkey, dep.Data.WithdrawalCredentials, dep.Data.Amount)

	} else {
		// Increase balance by deposit amount
		meta.IncreaseBalance(valIndex, dep.Data.Amount)
	}
	return nil
}
