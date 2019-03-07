package transfers

import (
	"errors"
	"fmt"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/hash"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessBlockTransfers(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.Transfers) > beacon.MAX_TRANSFERS {
		return errors.New("too many transfers")
	}
	// check if all TXs are distinct
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
	withdrawCred := beacon.Root(hash.Hash(transfer.Pubkey[:]))
	// overwrite first byte, remainder (the [1:] part, is still the hash)
	withdrawCred[0] = beacon.BLS_WITHDRAWAL_PREFIX_BYTE
	// verify transfer data + signature
	// TODO: fix formatting/quality
	if !(state.Validator_balances[transfer.Sender] >= transfer.Amount && state.Validator_balances[transfer.Sender] >= transfer.Fee &&
		((state.Validator_balances[transfer.Sender] == transfer.Amount+transfer.Fee) ||
			(state.Validator_balances[transfer.Sender] >= transfer.Amount+transfer.Fee+beacon.MIN_DEPOSIT_AMOUNT)) &&
		state.Slot == transfer.Slot &&
		(state.Epoch() >= state.Validator_registry[transfer.Sender].Withdrawable_epoch || state.Validator_registry[transfer.Sender].Activation_epoch == beacon.FAR_FUTURE_EPOCH) &&
		state.Validator_registry[transfer.Sender].Withdrawal_credentials == withdrawCred &&
		bls.Bls_verify(transfer.Pubkey, ssz.Signed_root(transfer), transfer.Signature, beacon.Get_domain(state.Fork, transfer.Slot.ToEpoch(), beacon.DOMAIN_TRANSFER))) {
		return errors.New("transfer is invalid")
	}
	state.Validator_balances[transfer.Sender] -= transfer.Amount + transfer.Fee
	state.Validator_balances[transfer.Recipient] += transfer.Amount
	propIndex := state.Get_beacon_proposer_index(state.Slot, false)
	state.Validator_balances[propIndex] += transfer.Fee
	return nil
}
