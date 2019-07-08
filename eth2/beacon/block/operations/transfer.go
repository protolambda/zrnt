package operations

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

var TransferSSZ = zssz.GetSSZ((*Transfer)(nil))

type Transfer struct {
	// Sender index
	Sender ValidatorIndex
	// Recipient index
	Recipient ValidatorIndex
	// Amount in Gwei
	Amount Gwei
	// Fee in Gwei for block proposer
	Fee Gwei
	// Inclusion slot
	Slot Slot
	// Sender withdrawal pubkey
	Pubkey BLSPubkey
	// Sender signature
	Signature BLSSignature
}

func (transfer *Transfer) Process(state *BeaconState) error {
	if !state.Validators.IsValidatorIndex(transfer.Sender) {
		return errors.New("cannot send funds from non-existent validator")
	}
	senderBalance := state.Balances[transfer.Sender]
	// Verify the amount and fee aren't individually too big (for anti-overflow purposes)
	if senderBalance < transfer.Amount {
		return errors.New("transfer amount is too big")
	}
	if senderBalance < transfer.Fee {
		return errors.New("transfer fee is too big")
	}
	if transfer.Sender == transfer.Recipient {
		return errors.New("no self-transfers (to enforce >= MIN_DEPOSIT_AMOUNT or zero balance invariant)")
	}
	// A transfer is valid in only one slot
	// (note: combined with unique transfers in a block, this functions as replay protection)
	if state.Slot != transfer.Slot {
		return errors.New("transfer is not valid in current slot")
	}
	sender := state.Validators[transfer.Sender]
	// Sender must be not yet eligible for activation, withdrawn, or transfer balance over MAX_EFFECTIVE_BALANCE
	if !(sender.ActivationEligibilityEpoch == FAR_FUTURE_EPOCH ||
		state.Epoch() >= sender.WithdrawableEpoch ||
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
		state.GetDomain(DOMAIN_TRANSFER, transfer.Slot.ToEpoch())) {
		return errors.New("transfer signature is invalid")
	}
	state.Balances.DecreaseBalance(transfer.Sender, transfer.Amount+transfer.Fee)
	state.Balances.IncreaseBalance(transfer.Recipient, transfer.Amount)
	propIndex := state.GetBeaconProposerIndex()
	state.Balances.IncreaseBalance(propIndex, transfer.Fee)
	// Verify balances are not dust
	if b := state.Balances[transfer.Sender]; 0 < b && b < MIN_DEPOSIT_AMOUNT {
		return errors.New("transfer is invalid: results in dust on sender address")
	}
	if b := state.Balances[transfer.Recipient]; 0 < b && b < MIN_DEPOSIT_AMOUNT {
		return errors.New("transfer is invalid: results in dust on recipient address")
	}
	return nil
}
