package beacon

import (
	"errors"
	"fmt"


	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/merkle"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

var DepositDataType = ContainerType("DepositData", []FieldDef{
	{"pubkey", BLSPubkeyType},
	{"withdrawal_credentials", Bytes32Type},
	{"amount", GweiType},
	{"signature", BLSSignatureType},
})

var DepositDataSSZ = zssz.GetSSZ((*DepositData)(nil))

type DepositData struct {
	Pubkey                BLSPubkey
	WithdrawalCredentials Root
	Amount                Gwei
	// Signing over DepositMessage
	Signature BLSSignature
}

func (d *DepositData) ToMessage() *DepositMessage {
	return &DepositMessage{
		Pubkey:                d.Pubkey,
		WithdrawalCredentials: d.WithdrawalCredentials,
		Amount:                d.Amount,
	}
}

func (d *DepositData) MessageRoot() Root {
	return ssz.HashTreeRoot(d.ToMessage(), DepositMessageSSZ)
}

var DepositMessageSSZ = zssz.GetSSZ((*DepositMessage)(nil))

type DepositMessage struct {
	Pubkey                BLSPubkey
	WithdrawalCredentials Root
	Amount                Gwei
}

var DepositProofType = VectorType(Bytes32Type, DEPOSIT_CONTRACT_TREE_DEPTH+1)

var DepositSSZ = zssz.GetSSZ((*Deposit)(nil))

type Deposit struct {
	Proof [DEPOSIT_CONTRACT_TREE_DEPTH + 1]Root // Merkle-path to deposit root
	Data  DepositData
}

var DepositType = ContainerType("Deposit", []FieldDef{
	{"proof", DepositProofType}, // Merkle path to deposit data list root
	{"data", DepositDataType},
})

// Verify that outstanding deposits are processed up to the maximum number of deposits, then process all in order.
func (state *BeaconStateView) ProcessDeposits(ops []Deposit) error {
	inputCount := DepositIndex(len(ops))
	eth1Data, err := state.Eth1Data()
	if err != nil {
		return err
	}
	depCount, err := eth1Data.DepositCount()
	if err != nil {
		return err
	}
	depIndex, err := state.DepositIndex()
	if err != nil {
		return err
	}
	expectedInputCount := depCount - depIndex
	if expectedInputCount > MAX_DEPOSITS {
		expectedInputCount = MAX_DEPOSITS
	}
	if inputCount != expectedInputCount {
		return errors.New("block does not contain expected deposits amount")
	}

	for i := range ops {
		if err := state.ProcessDeposit(&ops[i]); err != nil {
			return err
		}
	}
	return nil
}

// Process an Eth1 deposit, registering a validator or increasing its balance.
func (state *BeaconStateView) ProcessDeposit(dep *Deposit) error {
	depositIndex, err := state.DepositIndex()
	if err != nil {
		return err
	}
	eth1Data, err := state.Eth1Data()
	if err != nil {
		return err
	}
	depositsRoot, err := eth1Data.DepositRoot()
	if err != nil {
		return err
	}

	// Verify the Merkle branch
	if !merkle.VerifyMerkleBranch(
		ssz.HashTreeRoot(&dep.Data, DepositDataSSZ),
		dep.Proof[:],
		DEPOSIT_CONTRACT_TREE_DEPTH+1, // Add 1 for the `List` length mix-in
		uint64(depositIndex),
		depositsRoot) {
		return fmt.Errorf("deposit %d merkle proof failed to be verified", depositIndex)
	}

	// Increment the next deposit index we are expecting. Note that this
	// needs to be done here because while the deposit contract will never
	// create an invalid Merkle branch, it may admit an invalid deposit
	// object, and we need to be able to skip over it
	if err := state.IncrementDepositIndex(); err != nil {
		return err
	}

	valIndex, exists, err := input.ValidatorIndex(dep.Data.Pubkey)

	// Check if it is a known validator that is depositing ("if pubkey not in validator_pubkeys")
	if !exists {
		// Verify the deposit signature (proof of possession) which is not checked by the deposit contract
		if !bls.Verify(
			dep.Data.Pubkey,
			ComputeSigningRoot(
				dep.Data.MessageRoot(),
				// Fork-agnostic domain since deposits are valid across forks
				ComputeDomain(DOMAIN_DEPOSIT, GENESIS_FORK_VERSION, Root{})),
			dep.Data.Signature) {
			// invalid signatures are OK,
			// the depositor will not receive anything because of their mistake,
			// and the chain continues.
			return nil
		}

		// Add validator and balance entries
		return input.AddNewValidator(dep.Data.Pubkey, dep.Data.WithdrawalCredentials, dep.Data.Amount)

	} else {
		// Increase balance by deposit amount
		return input.IncreaseBalance(valIndex, dep.Data.Amount)
	}
}
