package block_processing

import (
	"errors"
	"fmt"
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/transition"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/hash"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessTransfers(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.Transfers) > eth2.MAX_TRANSFERS {
		return errors.New("too many transfers")
	}
	// check if all TXs are distinct
	distinctionCheckSet := make(map[eth2.BLSSignature]uint64)
	for i, v := range block.Body.Transfers {
		if existing, ok := distinctionCheckSet[v.Signature]; ok {
			return errors.New(fmt.Sprintf("transfer %d is the same as transfer %d, aborting", i, existing))
		}
		distinctionCheckSet[v.Signature] = uint64(i)
	}

	for _, transfer := range block.Body.Transfers {
		if err := ProcessTransfer(state, &transfer); err != nil {
			return err
		}
	}
	return nil
}

func ProcessTransfer(state *beacon.BeaconState, transfer *beacon.Transfer) error {
	withdrawCred := eth2.Root(hash.Hash(transfer.Pubkey[:]))
	// overwrite first byte, remainder (the [1:] part, is still the hash)
	withdrawCred[0] = eth2.BLS_WITHDRAWAL_PREFIX_BYTE
	// verify transfer data + signature. No separate error messages for line limit challenge...
	if !(state.Validator_balances[transfer.Sender] >= transfer.Amount && state.Validator_balances[transfer.Sender] >= transfer.Fee &&
		((state.Validator_balances[transfer.Sender] == transfer.Amount+transfer.Fee) ||
			(state.Validator_balances[transfer.Sender] >= transfer.Amount+transfer.Fee+eth2.MIN_DEPOSIT_AMOUNT)) &&
		state.Slot == transfer.Slot &&
		(state.Epoch() >= state.Validator_registry[transfer.Sender].Withdrawable_epoch || state.Validator_registry[transfer.Sender].Activation_epoch == eth2.FAR_FUTURE_EPOCH) &&
		state.Validator_registry[transfer.Sender].Withdrawal_credentials == withdrawCred &&
		bls.Bls_verify(transfer.Pubkey, ssz.Signed_root(transfer), transfer.Signature, transition.Get_domain(state.Fork, transfer.Slot.ToEpoch(), eth2.DOMAIN_TRANSFER))) {
		return errors.New("transfer is invalid")
	}
	state.Validator_balances[transfer.Sender] -= transfer.Amount + transfer.Fee
	state.Validator_balances[transfer.Recipient] += transfer.Amount
	propIndex, err := transition.Get_beacon_proposer_index(state, state.Slot, false)
	if err != nil {
		return err
	}
	state.Validator_balances[propIndex] += transfer.Fee
	return nil
}
