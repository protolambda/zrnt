package block_processing

import (
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/hash"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessBlockTransfers(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.Transfers) > beacon.MAX_TRANSFERS {
		return errors.New("too many transfers")
	}
	// check if all transfers are distinct
	distinctionCheckSet := make(map[beacon.BLSSignature]struct{})
	for i, v := range block.Body.Transfers {
		if existing, ok := distinctionCheckSet[v.Signature]; ok {
			return errors.New(fmt.Sprintf("transfer %d is the same as transfer %d, aborting", i, existing))
		}
		distinctionCheckSet[v.Signature] = struct{}{}
	}

	for _, transfer := range block.Body.Transfers {
		if err := ProcessTransfer(state, &transfer); err != nil {
			return err
		}
	}
	return nil
}

func ProcessTransfer(state *beacon.BeaconState, transfer *beacon.Transfer) error {
	// verify transfer data + signature
	senderBalance := state.GetBalance(transfer.Sender)
	// Verify the amount and fee aren't individually too big (for anti-overflow purposes)
	if !(senderBalance >= transfer.Amount && senderBalance >= transfer.Fee) {
		return errors.New("transfer value parameter (amount and/or fee) is too big")
	}
	// Verify that we have enough ETH to send, and that after the transfer the balance will be either
	// exactly zero or at least MIN_DEPOSIT_AMOUNT
	if !(senderBalance == transfer.Amount + transfer.Fee ||
		senderBalance >= transfer.Amount + transfer.Fee + beacon.MIN_DEPOSIT_AMOUNT) {
		return errors.New("transfer value is invalid, results in non-zero balance, but insufficient to stake with")
	}
	if transfer.Sender == transfer.Recipient {
		return errors.New("no self-transfers (to enforce >= MIN_DEPOSIT_AMOUNT or zero balance invariant)")
	}
	// A transfer is valid in only one slot
	// (note: combined with unique transfers in a block, this functions as replay protection)
	if state.Slot != transfer.Slot {
		return errors.New("transfer is not valid in current slot")
	}
	// Only withdrawn or not-yet-deposited accounts can transfer
	if !(state.Epoch() >= state.ValidatorRegistry[transfer.Sender].WithdrawableEpoch ||
		state.ValidatorRegistry[transfer.Sender].ActivationEpoch == beacon.FAR_FUTURE_EPOCH) {
		return errors.New("transfer sender is not eligible to make a transfer, it has to be withdrawn, or yet to be activated")
	}
	// Verify that the pubkey is valid
	withdrawCred := beacon.Root(hash.Hash(transfer.Pubkey[:]))
	// overwrite first byte, remainder (the [1:] part, is still the hash)
	withdrawCred[0] = beacon.BLS_WITHDRAWAL_PREFIX_BYTE
	if state.ValidatorRegistry[transfer.Sender].WithdrawalCredentials != withdrawCred {
		return errors.New("transfer pubkey is invalid")
	}
	// Verify that the signature is valid
	if !bls.BlsVerify(transfer.Pubkey, ssz.SignedRoot(transfer), transfer.Signature,
		beacon.GetDomain(state.Fork, transfer.Slot.ToEpoch(), beacon.DOMAIN_TRANSFER)) {
		return errors.New("transfer signature is invalid")
	}
	state.DecreaseBalance(transfer.Sender, transfer.Amount + transfer.Fee)
	state.IncreaseBalance(transfer.Recipient, transfer.Amount)
	propIndex := state.GetBeaconProposerIndex(state.Slot)
	state.IncreaseBalance(propIndex, transfer.Fee)
	return nil
}
