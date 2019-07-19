package transfers

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type TransferReq interface {
	VersioningMeta
	ProposingMeta
	RegistrySizeMeta
	ValidatorMeta
	BalanceMeta
}

// Verifies that there are no duplicate transfers, then processes in-order.
func ProcessTransfers(meta TransferReq, ops []Transfer) error {
	// check if all transfers are distinct
	distinctionCheckSet := make(map[BLSSignature]struct{})
	for i, v := range ops {
		if existing, ok := distinctionCheckSet[v.Signature]; ok {
			return fmt.Errorf("transfer %d is the same as transfer %d, aborting", i, existing)
		}
		distinctionCheckSet[v.Signature] = struct{}{}
	}

	for i := range ops {
		if err := ops[i].Process(meta); err != nil {
			return err
		}
	}
	return nil
}

var TransferSSZ = zssz.GetSSZ((*Transfer)(nil))

type Transfer struct {
	Sender    ValidatorIndex
	Recipient ValidatorIndex
	Amount    Gwei
	Fee       Gwei
	Slot      Slot         // Slot at which transfer must be processed
	Pubkey    BLSPubkey    // Sender withdrawal pubkey
	Signature BLSSignature // Signature checked against withdrawal pubkey
}

func (transfer *Transfer) Process(meta TransferReq) error {
	if !meta.IsValidIndex(transfer.Sender) {
		return errors.New("cannot send funds from non-existent validator")
	}
	senderBalance := meta.GetBalance(transfer.Sender)
	// Verify the amount and fee aren't individually too big (for anti-overflow purposes)
	if senderBalance < transfer.Amount {
		return errors.New("transfer amount is too big")
	}
	if senderBalance < transfer.Fee {
		return errors.New("transfer fee is too big")
	}
	if senderBalance < transfer.Fee+transfer.Amount {
		return errors.New("transfer total is too big")
	}
	if transfer.Sender == transfer.Recipient {
		return errors.New("no self-transfers (to enforce >= MIN_DEPOSIT_AMOUNT or zero balance invariant)")
	}
	// A transfer is valid in only one slot
	// (note: combined with unique transfers in a block, this functions as replay protection)
	if meta.Slot() != transfer.Slot {
		return errors.New("transfer is not valid in current slot")
	}
	sender := meta.Validator(transfer.Sender)
	// Sender must be not yet eligible for activation, withdrawn, or transfer balance over MAX_EFFECTIVE_BALANCE
	if !(sender.ActivationEligibilityEpoch == FAR_FUTURE_EPOCH ||
		meta.Epoch() >= sender.WithdrawableEpoch ||
		(transfer.Amount+transfer.Fee+MAX_EFFECTIVE_BALANCE) <= senderBalance) {
		return errors.New("transfer sender is not eligible to make a transfer, it has to be withdrawn, or yet to be activated")
	}
	// Verify that the pubkey is valid
	withdrawCred := Hash(transfer.Pubkey[:])
	// overwrite first byte, remainder (the [1:] part, is still the hash)
	withdrawCred[0] = BLS_WITHDRAWAL_PREFIX
	if sender.WithdrawalCredentials != withdrawCred {
		return errors.New("transfer pubkey is invalid")
	}
	// Verify that the signature is valid
	if !bls.BlsVerify(transfer.Pubkey, ssz.SigningRoot(transfer, TransferSSZ), transfer.Signature,
		meta.GetDomain(DOMAIN_TRANSFER, transfer.Slot.ToEpoch())) {
		return errors.New("transfer signature is invalid")
	}
	meta.DecreaseBalance(transfer.Sender, transfer.Amount+transfer.Fee)
	meta.IncreaseBalance(transfer.Recipient, transfer.Amount)
	propIndex := meta.GetBeaconProposerIndex()
	meta.IncreaseBalance(propIndex, transfer.Fee)
	// Verify balances are not dust
	if b := meta.GetBalance(transfer.Sender); 0 < b && b < MIN_DEPOSIT_AMOUNT {
		return errors.New("transfer is invalid: results in dust on sender address")
	}
	if b := meta.GetBalance(transfer.Recipient); 0 < b && b < MIN_DEPOSIT_AMOUNT {
		return errors.New("transfer is invalid: results in dust on recipient address")
	}
	return nil
}
