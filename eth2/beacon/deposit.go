package beacon

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"

	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/merkle"
	. "github.com/protolambda/ztyp/view"
)

var DepositDataType = ContainerType("DepositData", []FieldDef{
	{"pubkey", BLSPubkeyType},
	{"withdrawal_credentials", Bytes32Type},
	{"amount", GweiType},
	{"signature", BLSSignatureType},
})

type DepositData struct {
	Pubkey                BLSPubkey `json:"pubkey" yaml:"pubkey"`
	WithdrawalCredentials Root      `json:"withdrawal_credentials" yaml:"withdrawal_credentials"`
	Amount                Gwei      `json:"amount" yaml:"amount"`
	// Signing over DepositMessage
	Signature BLSSignature `json:"signature" yaml:"signature"`
}

func (d *DepositData) ToMessage() *DepositMessage {
	return &DepositMessage{
		Pubkey:                d.Pubkey,
		WithdrawalCredentials: d.WithdrawalCredentials,
		Amount:                d.Amount,
	}
}

func (d *DepositData) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&d.Pubkey, &d.WithdrawalCredentials, &d.Amount, &d.Signature)
}

func (d *DepositData) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&d.Pubkey, &d.WithdrawalCredentials, &d.Amount, &d.Signature)
}

func (a *DepositData) ByteLength() uint64 {
	return DepositDataType.TypeByteLength()
}

func (a *DepositData) FixedLength() uint64 {
	return DepositDataType.TypeByteLength()
}

// hash-tree-root including the signature
func (d *DepositData) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(d.Pubkey, d.WithdrawalCredentials, d.Amount, d.Signature)
}

// hash-tree-root excluding the signature
func (d *DepositData) MessageRoot() Root {
	return d.ToMessage().HashTreeRoot(tree.GetHashFn())
}

type DepositMessage struct {
	Pubkey                BLSPubkey `json:"pubkey" yaml:"pubkey"`
	WithdrawalCredentials Root      `json:"withdrawal_credentials" yaml:"withdrawal_credentials"`
	Amount                Gwei      `json:"amount" yaml:"amount"`
}

func (b *DepositMessage) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(b.Pubkey, b.WithdrawalCredentials, b.Amount)
}

var DepositProofType = VectorType(Bytes32Type, DEPOSIT_CONTRACT_TREE_DEPTH+1)

// DepositProof contains the proof for the merkle-path to deposit root, including list mix-in.
type DepositProof [DEPOSIT_CONTRACT_TREE_DEPTH + 1]Root

func (d *DepositProof) Deserialize(dr *codec.DecodingReader) error {
	return dr.Vector(func(i uint64) codec.Deserializable {
		return &d[i]
	}, RootType.TypeByteLength(), DepositProofType.Length())
}

func (d *DepositProof) Serialize(w *codec.EncodingWriter) error {
	return w.Vector(func(i uint64) codec.Serializable {
		return &d[i]
	}, RootType.TypeByteLength(), DepositProofType.Length())
}

func (a *DepositProof) ByteLength() uint64 {
	return DepositProofType.TypeByteLength()
}

func (a *DepositProof) FixedLength() uint64 {
	return DepositProofType.TypeByteLength()
}

func (b *DepositProof) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.ChunksHTR(func(i uint64) tree.Root {
		return b[i]
	}, uint64(len(b)), uint64(len(b)))
}

type Deposit struct {
	Proof DepositProof `json:"proof" yaml:"proof"`
	Data  DepositData  `json:"data" yaml:"data"`
}

func (d *Deposit) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&d.Proof, &d.Data)
}

func (d *Deposit) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&d.Proof, &d.Data)
}

func (a *Deposit) ByteLength() uint64 {
	return Eth1DataType.TypeByteLength()
}

func (a *Deposit) FixedLength() uint64 {
	return DepositType.TypeByteLength()
}

func (b *Deposit) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(&b.Proof, &b.Data)
}

var DepositType = ContainerType("Deposit", []FieldDef{
	{"proof", DepositProofType}, // Merkle path to deposit data list root
	{"data", DepositDataType},
})

func (c *Phase0Config) BlockDeposits() ListTypeDef {
	return ListType(DepositType, c.MAX_DEPOSITS)
}

type Deposits []Deposit

func (a *Deposits) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, Deposit{})
		return &((*a)[i])
	}, DepositType.TypeByteLength(), spec.MAX_DEPOSITS)
}

func (a Deposits) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, DepositType.TypeByteLength(), uint64(len(a)))
}

func (a Deposits) ByteLength(*Spec) (out uint64) {
	return DepositType.TypeByteLength() * uint64(len(a))
}

func (a *Deposits) FixedLength(*Spec) uint64 {
	return 0
}

func (li Deposits) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, spec.MAX_DEPOSITS)
}

// Verify that outstanding deposits are processed up to the maximum number of deposits, then process all in order.
func (spec *Spec) ProcessDeposits(ctx context.Context, epc *EpochsContext, state *BeaconStateView, ops []Deposit) error {
	inputCount := uint64(len(ops))
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
	// state deposit count and deposit index are trusted not to underflow
	expectedInputCount := uint64(depCount - depIndex)
	if expectedInputCount > spec.MAX_DEPOSITS {
		expectedInputCount = spec.MAX_DEPOSITS
	}
	if inputCount != expectedInputCount {
		return errors.New("block does not contain expected deposits amount")
	}

	for i := range ops {
		select {
		case <-ctx.Done():
			return TransitionCancelErr
		default: // Don't block.
			break
		}
		if err := spec.ProcessDeposit(epc, state, &ops[i], false); err != nil {
			return err
		}
	}
	return nil
}

// Process an Eth1 deposit, registering a validator or increasing its balance.
func (spec *Spec) ProcessDeposit(epc *EpochsContext, state *BeaconStateView, dep *Deposit, ignoreSignatureAndProof bool) error {
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
	if !ignoreSignatureAndProof && !merkle.VerifyMerkleBranch(
		dep.Data.HashTreeRoot(tree.GetHashFn()),
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

	validators, err := state.Validators()
	if err != nil {
		return err
	}

	valCount, err := validators.Length()
	if err != nil {
		return err
	}
	valIndex, ok := epc.PubkeyCache.ValidatorIndex(dep.Data.Pubkey)
	// it exists if: it exists in the pubkey cache AND the validator index is lower than the current validator count.
	exists := ok && uint64(valIndex) < valCount
	if !exists {
		valIndex = ValidatorIndex(valCount)
	}

	// Check if it is a known validator that is depositing ("if pubkey not in validator_pubkeys")
	if !exists {
		// Verify the deposit signature (proof of possession) which is not checked by the deposit contract
		if !ignoreSignatureAndProof && !bls.Verify(
			&CachedPubkey{Compressed: dep.Data.Pubkey},
			ComputeSigningRoot(
				dep.Data.MessageRoot(),
				// Fork-agnostic domain since deposits are valid across forks
				ComputeDomain(spec.DOMAIN_DEPOSIT, spec.GENESIS_FORK_VERSION, Root{})),
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
		effBalance := balance - (balance % spec.EFFECTIVE_BALANCE_INCREMENT)
		if effBalance > spec.MAX_EFFECTIVE_BALANCE {
			effBalance = spec.MAX_EFFECTIVE_BALANCE
		}
		validatorRaw := Validator{
			Pubkey:                     pubkey,
			WithdrawalCredentials:      withdrawalCreds,
			ActivationEligibilityEpoch: FAR_FUTURE_EPOCH,
			ActivationEpoch:            FAR_FUTURE_EPOCH,
			ExitEpoch:                  FAR_FUTURE_EPOCH,
			WithdrawableEpoch:          FAR_FUTURE_EPOCH,
			EffectiveBalance:           effBalance,
		}
		validator := validatorRaw.View()
		if err := validators.Append(validator); err != nil {
			return err
		}
		bals, err := state.Balances()
		if err != nil {
			return err
		}
		if err := bals.Append(Uint64View(balance)); err != nil {
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
		if err := bals.IncreaseBalance(valIndex, dep.Data.Amount); err != nil {
			return err
		}
	}
	return nil
}
