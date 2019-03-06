package block_processing

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/transition"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/hash"
	"github.com/protolambda/go-beacon-transition/eth2/util/merkle"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessDeposits(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.Deposits) > eth2.MAX_DEPOSITS {
		return errors.New("too many deposits")
	}
	for _, dep := range block.Body.Deposits {
		if err := ProcessDeposit(state, &dep); err != nil {
			return err
		}
		state.Deposit_index += 1
	}
}


// Process a deposit from Ethereum 1.0.
func ProcessDeposit(state *beacon.BeaconState, dep *beacon.Deposit) error {
	deposit_input := &dep.Deposit_data.Deposit_input

	// Deposits must be processed in order
	if dep.Index != state.Deposit_index {
		return errors.New(fmt.Sprintf("deposit has index %d that does not match with state index %d", dep.Index, state.Deposit_index))
	}

	// Let serialized_deposit_data be the serialized form of deposit.deposit_data.
	// It should equal 8 bytes for deposit_data.amount +
	//              8 bytes for deposit_data.timestamp +
	//              176 bytes for deposit_data.deposit_input
	// That is, it should match deposit_data in the Ethereum 1.0 deposit contract
	//  of which the hash was placed into the Merkle tree.
	dep_input_bytes := ssz.SSZEncode(dep.Deposit_data.Deposit_input)
	serialized_deposit_data := make([]byte, 8+8+len(dep_input_bytes), 8+8+len(dep_input_bytes))
	binary.LittleEndian.PutUint64(serialized_deposit_data[0:8], uint64(dep.Deposit_data.Amount))
	binary.LittleEndian.PutUint64(serialized_deposit_data[8:16], uint64(dep.Deposit_data.Timestamp))
	copy(serialized_deposit_data[16:], dep_input_bytes)

	// Verify the Merkle branch
	if !merkle.Verify_merkle_branch(
		hash.Hash(serialized_deposit_data),
		dep.Proof,
		eth2.DEPOSIT_CONTRACT_TREE_DEPTH,
		uint64(dep.Index),
		state.Latest_eth1_data.Deposit_root) {
		return errors.New(fmt.Sprintf("deposit %d has merkle proof that failed to be verified", dep.Index))
	}

	// Increment the next deposit index we are expecting. Note that this
	// needs to be done here because while the deposit contract will never
	// create an invalid Merkle branch, it may admit an invalid deposit
	// object, and we need to be able to skip over it
	state.Deposit_index += 1

	if !bls.Bls_verify(
		deposit_input.Pubkey,
		ssz.Signed_root(deposit_input),
		deposit_input.Proof_of_possession,
		transition.Get_domain(state.Fork, state.Epoch(), eth2.DOMAIN_DEPOSIT)) {
		// simply don't handle the deposit. (TODO: should this be an error (making block invalid)?)
		return nil
	}

	val_index := eth2.ValidatorIndexMarker
	for i, v := range state.Validator_registry {
		if v.Pubkey == deposit_input.Pubkey {
			val_index = eth2.ValidatorIndex(i)
			break
		}
	}

	pubkey := state.Validator_registry[val_index].Pubkey
	amount := dep.Deposit_data.Amount
	withdrawalCredentials := deposit_input.Withdrawal_credentials
	// Check if it is a known validator that is depositing ("if pubkey not in validator_pubkeys")
	if val_index == eth2.ValidatorIndexMarker {
		// Not a known pubkey, add new validator
		validator := beacon.Validator{
			Pubkey:                 pubkey,
			Withdrawal_credentials: withdrawalCredentials,
			Activation_epoch:       eth2.FAR_FUTURE_EPOCH, Exit_epoch: eth2.FAR_FUTURE_EPOCH, Withdrawable_epoch: eth2.FAR_FUTURE_EPOCH,
			Initiated_exit: false, Slashed: false,
		}
		// Note: In phase 2 registry indices that have been withdrawn for a long time will be recycled.
		state.Validator_registry = append(state.Validator_registry, validator)
		state.Validator_balances = append(state.Validator_balances, amount)
	} else {
		// known pubkey, check withdrawal credentials first, then increase balance.
		if state.Validator_registry[val_index].Withdrawal_credentials != deposit_input.Withdrawal_credentials {
			return errors.New("deposit has wrong withdrawal credentials")
		}
		// Increase balance by deposit amount
		state.Validator_balances[val_index] += dep.Deposit_data.Amount
	}
	return nil
}