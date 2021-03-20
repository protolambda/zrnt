package phase0

import (
	"context"
	"errors"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/merkle"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

func BlockDepositsType(spec *common.Spec) ListTypeDef {
	return ListType(common.DepositType, spec.MAX_DEPOSITS)
}

type Deposits []common.Deposit

func (a *Deposits) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, common.Deposit{})
		return &((*a)[i])
	}, common.DepositType.TypeByteLength(), spec.MAX_DEPOSITS)
}

func (a Deposits) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, common.DepositType.TypeByteLength(), uint64(len(a)))
}

func (a Deposits) ByteLength(*common.Spec) (out uint64) {
	return common.DepositType.TypeByteLength() * uint64(len(a))
}

func (a *Deposits) FixedLength(*common.Spec) uint64 {
	return 0
}

func (li Deposits) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, spec.MAX_DEPOSITS)
}

// Verify that outstanding deposits are processed up to the maximum number of deposits, then process all in order.
func ProcessDeposits(ctx context.Context, spec *common.Spec, epc *EpochsContext, state common.BeaconState, ops []common.Deposit) error {
	inputCount := uint64(len(ops))
	eth1Data, err := state.Eth1Data()
	if err != nil {
		return err
	}
	depIndex, err := state.DepositIndex()
	if err != nil {
		return err
	}
	// state deposit count and deposit index are trusted not to underflow
	expectedInputCount := uint64(eth1Data.DepositCount - depIndex)
	if expectedInputCount > spec.MAX_DEPOSITS {
		expectedInputCount = spec.MAX_DEPOSITS
	}
	if inputCount != expectedInputCount {
		return errors.New("block does not contain expected deposits amount")
	}

	for i := range ops {
		select {
		case <-ctx.Done():
			return common.TransitionCancelErr
		default: // Don't block.
			break
		}
		if err := ProcessDeposit(spec, epc, state, &ops[i], false); err != nil {
			return err
		}
	}
	return nil
}

// Process an Eth1 deposit, registering a validator or increasing its balance.
func ProcessDeposit(spec *common.Spec, epc *EpochsContext, state common.BeaconState, dep *common.Deposit, ignoreSignatureAndProof bool) error {
	depositIndex, err := state.DepositIndex()
	if err != nil {
		return err
	}
	eth1Data, err := state.Eth1Data()
	if err != nil {
		return err
	}

	// Verify the Merkle branch
	if !ignoreSignatureAndProof && !merkle.VerifyMerkleBranch(
		dep.Data.HashTreeRoot(tree.GetHashFn()),
		dep.Proof[:],
		common.DEPOSIT_CONTRACT_TREE_DEPTH+1, // Add 1 for the `List` length mix-in
		uint64(depositIndex),
		eth1Data.DepositRoot) {
		return fmt.Errorf("deposit %d merkle proof failed to be verified", depositIndex)
	}

	// Increment the next deposit index we are expecting. Note that this
	// needs to be done here because while the deposit contract will never
	// create an invalid Merkle branch, it may admit an invalid deposit
	// object, and we need to be able to skip over it
	if err := state.IncrementDepositIndex(); err != nil {
		return err
	}

	validators, err := state.Validators()
	if err != nil {
		return err
	}

	valCount, err := validators.ValidatorCount()
	if err != nil {
		return err
	}
	valIndex, ok := epc.PubkeyCache.ValidatorIndex(dep.Data.Pubkey)
	// it exists if: it exists in the pubkey cache AND the validator index is lower than the current validator count.
	exists := ok && uint64(valIndex) < valCount
	if !exists {
		valIndex = common.ValidatorIndex(valCount)
	}

	// Check if it is a known validator that is depositing ("if pubkey not in validator_pubkeys")
	if !exists {
		// Verify the deposit signature (proof of possession) which is not checked by the deposit contract
		if !ignoreSignatureAndProof && !bls.Verify(
			&common.CachedPubkey{Compressed: dep.Data.Pubkey},
			common.ComputeSigningRoot(
				dep.Data.MessageRoot(),
				// Fork-agnostic domain since deposits are valid across forks
				common.ComputeDomain(spec.DOMAIN_DEPOSIT, spec.GENESIS_FORK_VERSION, common.Root{})),
			dep.Data.Signature) {
			// invalid signatures are OK,
			// the depositor will not receive anything because of their mistake,
			// and the chain continues.
			return nil
		}

		// Add validator and balance entries
		balance := dep.Data.Amount
		withdrawalCreds := dep.Data.WithdrawalCredentials
		pubkey := dep.Data.Pubkey
		if err := state.AddValidator(spec, pubkey, withdrawalCreds, balance); err != nil {
			return err
		}
		if pc, err := epc.PubkeyCache.AddValidator(valIndex, pubkey); err != nil {
			return err
		} else {
			epc.PubkeyCache = pc
		}
	} else {
		// Increase balance by deposit amount
		bals, err := state.Balances()
		if err != nil {
			return err
		}
		if err := common.IncreaseBalance(bals, valIndex, dep.Data.Amount); err != nil {
			return err
		}
	}
	return nil
}
