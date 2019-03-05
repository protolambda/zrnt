package transition

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/hash"
	"github.com/protolambda/go-beacon-transition/eth2/util/merkle"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

// NOTE: temporary: block transition is going to be split up,
// this will just be a function calling all block-handle "entry-points".


func ApplyBlock(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	// Verify slot
	if block.Slot != state.Slot {
		return errors.New("cannot apply block to pre-block-state at different slot")
	}

	proposer := state.Validator_registry[get_beacon_proposer_index(state, state.Slot)]
	// Block signature
	{
		proposal := beacon.Proposal{Slot: block.Slot, Shard: eth2.BEACON_CHAIN_SHARD_NUMBER, Block_root: ssz.Signed_root(block, "Signature"), Signature: block.Signature}
		if !bls.Bls_verify(proposer.Pubkey, ssz.Signed_root(proposal, "Signature"), proposal.Signature, get_domain(state.Fork, state.Epoch(), eth2.DOMAIN_PROPOSAL)) {
			return errors.New("block signature invalid")
		}
	}

	// RANDAO
	{
		if !bls.Bls_verify(proposer.Pubkey, ssz.Hash_tree_root(state.Epoch()), block.Randao_reveal, get_domain(state.Fork, state.Epoch(), eth2.DOMAIN_RANDAO)) {
			return errors.New("randao invalid")
		}
		state.Latest_randao_mixes[state.Epoch()%eth2.LATEST_RANDAO_MIXES_LENGTH] = hash.XorBytes32(get_randao_mix(state, state.Epoch()), hash.Hash(block.Randao_reveal[:]))
	}

	// Eth1 data
	{
		// If there exists an eth1_data_vote in state.Eth1_data_votes for which eth1_data_vote.eth1_data == block.Eth1_data (there will be at most one), set eth1_data_vote.vote_count += 1.
		// Otherwise, append to state.Eth1_data_votes a new Eth1DataVote(eth1_data=block.Eth1_data, vote_count=1).
		found := false
		for i, vote := range state.Eth1_data_votes {
			if vote.Eth1_data == block.Eth1_data {
				state.Eth1_data_votes[i].Vote_count += 1
				found = true
				break
			}
		}
		if !found {
			state.Eth1_data_votes = append(state.Eth1_data_votes, beacon.Eth1DataVote{Eth1_data: block.Eth1_data, Vote_count: 1})
		}
	}

	// Transactions
	// START ------------------------------

	// Proposer slashings
	{
		if len(block.Body.Proposer_slashings) > eth2.MAX_PROPOSER_SLASHINGS {
			return errors.New("too many proposer slashings")
		}
		for i, ps := range block.Body.Proposer_slashings {
			if !is_validator_index(state, ps.Proposer_index) {
				return errors.New("invalid proposer index")
			}
			proposer := state.Validator_registry[ps.Proposer_index]
			if !(ps.Proposal_1.Slot == ps.Proposal_2.Slot && ps.Proposal_1.Shard == ps.Proposal_2.Shard &&
				ps.Proposal_1.Block_root != ps.Proposal_2.Block_root && proposer.Slashed == false &&
				bls.Bls_verify(proposer.Pubkey, ssz.Signed_root(ps.Proposal_1, "Signature"), ps.Proposal_1.Signature, get_domain(state.Fork, ps.Proposal_1.Slot.ToEpoch(), eth2.DOMAIN_PROPOSAL)) &&
				bls.Bls_verify(proposer.Pubkey, ssz.Signed_root(ps.Proposal_2, "Signature"), ps.Proposal_2.Signature, get_domain(state.Fork, ps.Proposal_2.Slot.ToEpoch(), eth2.DOMAIN_PROPOSAL))) {
				return errors.New(fmt.Sprintf("proposer slashing %d is invalid", i))
			}
			if err := slash_validator(state, ps.Proposer_index); err != nil {
				return err
			}
		}
	}

	// Attester slashings
	{
		if len(block.Body.Attester_slashings) > eth2.MAX_ATTESTER_SLASHINGS {
			return errors.New("too many attester slashings")
		}
		for i, attester_slashing := range block.Body.Attester_slashings {
			sa1, sa2 := &attester_slashing.Slashable_attestation_1, &attester_slashing.Slashable_attestation_2
			// verify the attester_slashing
			if !(sa1.Data != sa2.Data && (is_double_vote(&sa1.Data, &sa2.Data) || is_surround_vote(&sa1.Data, &sa2.Data)) &&
				verify_slashable_attestation(state, sa1) && verify_slashable_attestation(state, sa2)) {
				return errors.New(fmt.Sprintf("attester slashing %d is invalid", i))
			}
			// keep track of effectiveness
			slashedAny := false
			// run slashings where applicable
		ValLoop:
			// indices are trusted, they have been verified by verify_slashable_attestation(...)
			for _, v1 := range sa1.Validator_indices {
				for _, v2 := range sa2.Validator_indices {
					if v1 == v2 && !state.Validator_registry[v1].Slashed {
						if err := slash_validator(state, v1); err != nil {
							return err
						}
						slashedAny = true
						// continue to look for next validator in outer loop (because there are no duplicates in attestation)
						continue ValLoop
					}
				}
			}
			// "Verify that len(slashable_indices) >= 1."
			if !slashedAny {
				return errors.New(fmt.Sprintf("attester slashing %d is not effective, hence invalid", i))
			}
		}
	}

	// Attestations
	{
		if len(block.Body.Attestations) > eth2.MAX_ATTESTATIONS {
			return errors.New("too many attestations")
		}
		for i, attestation := range block.Body.Attestations {

			justified_epoch := state.Previous_justified_epoch
			if (attestation.Data.Slot + 1).ToEpoch() >= state.Epoch() {
				justified_epoch = state.Justified_epoch
			}
			blockRoot, blockRootErr := get_block_root(state, attestation.Data.Justified_epoch.GetStartSlot())
			if !(attestation.Data.Slot >= eth2.GENESIS_SLOT && attestation.Data.Slot+eth2.MIN_ATTESTATION_INCLUSION_DELAY <= state.Slot &&
				state.Slot < attestation.Data.Slot+eth2.SLOTS_PER_EPOCH && attestation.Data.Justified_epoch == justified_epoch &&
				(blockRootErr == nil && attestation.Data.Justified_block_root == blockRoot) &&
				(state.Latest_crosslinks[attestation.Data.Shard] == attestation.Data.Latest_crosslink ||
					state.Latest_crosslinks[attestation.Data.Shard] == beacon.Crosslink{Crosslink_data_root: attestation.Data.Crosslink_data_root, Epoch: attestation.Data.Slot.ToEpoch()})) {
				return errors.New(fmt.Sprintf("attestation %d is not valid", i))
			}
			// Verify bitfields and aggregate signature
			// custody bitfield is phase 0 only:
			if attestation.Aggregation_bitfield.IsZero() || !attestation.Custody_bitfield.IsZero() {
				return errors.New(fmt.Sprintf("attestation %d has incorrect bitfield(s)", i))
			}

			crosslink_committees, err := get_crosslink_committees_at_slot(state, attestation.Data.Slot, false)
			if err != nil {
				return err
			}
			crosslink_committee := CrosslinkCommittee{}
			for _, committee := range crosslink_committees {
				if committee.Shard == attestation.Data.Shard {
					crosslink_committee = committee
					break
				}
			}
			// TODO spec is weak here: it's not very explicit about length of bitfields.
			//  Let's just make sure they are the size of the committee
			if !attestation.Aggregation_bitfield.VerifySize(uint64(len(crosslink_committee.Committee))) ||
				!attestation.Custody_bitfield.VerifySize(uint64(len(crosslink_committee.Committee))) {
				return errors.New(fmt.Sprintf("attestation %d has bitfield(s) with incorrect size", i))
			}
			// phase 0 only
			if !attestation.Aggregation_bitfield.IsZero() || !attestation.Custody_bitfield.IsZero() {
				return errors.New(fmt.Sprintf("attestation %d has non-zero bitfield(s)", i))
			}

			participants, err := get_attestation_participants(state, &attestation.Data, &attestation.Aggregation_bitfield)
			if err != nil {
				return errors.New("participants could not be derived from aggregation_bitfield")
			}
			custody_bit_1_participants, err := get_attestation_participants(state, &attestation.Data, &attestation.Custody_bitfield)
			if err != nil {
				return errors.New("participants could not be derived from custody_bitfield")
			}
			custody_bit_0_participants := participants.Minus(custody_bit_1_participants)

			// get lists of pubkeys for both 0 and 1 custody-bits
			custody_bit_0_pubkeys := make([]eth2.BLSPubkey, len(custody_bit_0_participants))
			for i, v := range custody_bit_0_participants {
				custody_bit_0_pubkeys[i] = state.Validator_registry[v].Pubkey
			}
			custody_bit_1_pubkeys := make([]eth2.BLSPubkey, len(custody_bit_1_participants))
			for i, v := range custody_bit_1_participants {
				custody_bit_1_pubkeys[i] = state.Validator_registry[v].Pubkey
			}
			// aggregate each of the two lists
			pubKeys := []eth2.BLSPubkey{bls.Bls_aggregate_pubkeys(custody_bit_0_pubkeys), bls.Bls_aggregate_pubkeys(custody_bit_1_pubkeys)}
			// hash the attestation data with 0 and 1 as bit
			hashes := []eth2.Root{
				ssz.Hash_tree_root(beacon.AttestationDataAndCustodyBit{attestation.Data, false}),
				ssz.Hash_tree_root(beacon.AttestationDataAndCustodyBit{attestation.Data, true}),
			}
			// now verify the two
			if !bls.Bls_verify_multiple(pubKeys, hashes, attestation.Aggregate_signature,
				get_domain(state.Fork, attestation.Data.Slot.ToEpoch(), eth2.DOMAIN_ATTESTATION)) {
				return errors.New(fmt.Sprintf("attestation %d has invalid aggregated BLS signature", i))
			}

			// phase 0 only:
			if attestation.Data.Crosslink_data_root != (eth2.Root{}) {
				return errors.New(fmt.Sprintf("attestation %d has invalid crosslink: root must be 0 in phase 0", i))
			}
		}
	}

	// Deposits
	{
		if len(block.Body.Deposits) > eth2.MAX_DEPOSITS {
			return errors.New("too many deposits")
		}
		for i, dep := range block.Body.Deposits {
			if dep.Index != state.Deposit_index {
				return errors.New(fmt.Sprintf("deposit %d has index %d that does not match with state index %d", i, dep.Index, state.Deposit_index))
			}
			// Let serialized_deposit_data be the serialized form of deposit.deposit_data.
			// It should be 8 bytes for deposit_data.amount
			//  followed by 8 bytes for deposit_data.timestamp
			//  and then the DepositInput bytes.
			// That is, it should match deposit_data in the Ethereum 1.0 deposit contract
			//  of which the hash was placed into the Merkle tree.
			dep_input_bytes := ssz.SSZEncode(dep.Deposit_data.Deposit_input)
			serialized_deposit_data := make([]byte, 8+8+len(dep_input_bytes), 8+8+len(dep_input_bytes))
			binary.LittleEndian.PutUint64(serialized_deposit_data[0:8], uint64(dep.Deposit_data.Amount))
			binary.LittleEndian.PutUint64(serialized_deposit_data[8:16], uint64(dep.Deposit_data.Timestamp))
			copy(serialized_deposit_data[16:], dep_input_bytes)

			// verify the deposit
			if !merkle.Verify_merkle_branch(hash.Hash(serialized_deposit_data), dep.Branch, eth2.DEPOSIT_CONTRACT_TREE_DEPTH,
				uint64(dep.Index), state.Latest_eth1_data.Deposit_root) {
				return errors.New(fmt.Sprintf("deposit %d has merkle proof that failed to be verified", i))
			}
			if err := process_deposit(state, &dep); err != nil {
				return err
			}
			state.Deposit_index += 1
		}
	}

	// Voluntary exits
	{
		if len(block.Body.Voluntary_exits) > eth2.MAX_VOLUNTARY_EXITS {
			return errors.New("too many voluntary exits")
		}
		for i, exit := range block.Body.Voluntary_exits {
			validator := state.Validator_registry[exit.Validator_index]
			if !(validator.Exit_epoch > get_delayed_activation_exit_epoch(state.Epoch()) &&
				state.Epoch() > exit.Epoch &&
				bls.Bls_verify(validator.Pubkey, ssz.Signed_root(exit, "Signature"),
					exit.Signature, get_domain(state.Fork, exit.Epoch, eth2.DOMAIN_EXIT))) {
				return errors.New(fmt.Sprintf("voluntary exit %d could not be verified", i))
			}
			initiate_validator_exit(state, exit.Validator_index)
		}
	}

	// Transfers
	{
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

		for i, transfer := range block.Body.Transfers {
			withdrawCred := eth2.Root(hash.Hash(transfer.Pubkey[:]))
			// overwrite first byte, remainder (the [1:] part, is still the hash)
			withdrawCred[0] = eth2.BLS_WITHDRAWAL_PREFIX_BYTE
			// verify transfer data + signature. No separate error messages for line limit challenge...
			if !(state.Validator_balances[transfer.From] >= transfer.Amount && state.Validator_balances[transfer.From] >= transfer.Fee &&
				((state.Validator_balances[transfer.From] == transfer.Amount+transfer.Fee) ||
					(state.Validator_balances[transfer.From] >= transfer.Amount+transfer.Fee+eth2.MIN_DEPOSIT_AMOUNT)) &&
				state.Slot == transfer.Slot &&
				(state.Epoch() >= state.Validator_registry[transfer.From].Withdrawable_epoch || state.Validator_registry[transfer.From].Activation_epoch == eth2.FAR_FUTURE_EPOCH) &&
				state.Validator_registry[transfer.From].Withdrawal_credentials == withdrawCred &&
				bls.Bls_verify(transfer.Pubkey, ssz.Signed_root(transfer, "Signature"), transfer.Signature, get_domain(state.Fork, transfer.Slot.ToEpoch(), eth2.DOMAIN_TRANSFER))) {
				return errors.New(fmt.Sprintf("transfer %d is invalid", i))
			}
			state.Validator_balances[transfer.From] -= transfer.Amount + transfer.Fee
			state.Validator_balances[transfer.To] += transfer.Amount
			state.Validator_balances[get_beacon_proposer_index(state, state.Slot)] += transfer.Fee
		}
	}

	// END ------------------------------

	return nil
}