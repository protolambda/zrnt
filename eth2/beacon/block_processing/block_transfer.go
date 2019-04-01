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
	withdrawCred := beacon.Root(hash.Hash(transfer.Pubkey[:]))
	// overwrite first byte, remainder (the [1:] part, is still the hash)
	withdrawCred[0] = beacon.BLS_WITHDRAWAL_PREFIX_BYTE
	// verify transfer data + signature
	// TODO: fix formatting/quality
	if !(state.ValidatorBalances[transfer.Sender] >= transfer.Amount && state.ValidatorBalances[transfer.Sender] >= transfer.Fee &&
		((state.ValidatorBalances[transfer.Sender] == transfer.Amount+transfer.Fee) ||
			(state.ValidatorBalances[transfer.Sender] >= transfer.Amount+transfer.Fee+beacon.MIN_DEPOSIT_AMOUNT)) &&
		state.Slot == transfer.Slot &&
		(state.Epoch() >= state.ValidatorRegistry[transfer.Sender].WithdrawableEpoch || state.ValidatorRegistry[transfer.Sender].ActivationEpoch == beacon.FAR_FUTURE_EPOCH) &&
		state.ValidatorRegistry[transfer.Sender].WithdrawalCredentials == withdrawCred &&
		bls.BlsVerify(transfer.Pubkey, ssz.SignedRoot(transfer), transfer.Signature, beacon.GetDomain(state.Fork, transfer.Slot.ToEpoch(), beacon.DOMAIN_TRANSFER))) {
		return errors.New("transfer is invalid")
	}
	state.ValidatorBalances[transfer.Sender] -= transfer.Amount + transfer.Fee
	state.ValidatorBalances[transfer.Recipient] += transfer.Amount
	propIndex := state.GetBeaconProposerIndex(state.Slot, false)
	state.ValidatorBalances[propIndex] += transfer.Fee
	return nil
}
