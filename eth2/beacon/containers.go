package beacon

import (
	"crypto/sha256"
	"github.com/protolambda/eth2-shuffle"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/stake"
	"github.com/protolambda/go-beacon-transition/eth2/util/bitfield"
)

// NOTE: these containers are going to be moved to sub-packages, per-topic.

type ProposerSlashing struct {
	// Proposer index
	Proposer_index ValidatorIndex
	// First proposal
	Header_1 BeaconBlockHeader
	// Second proposal
	Header_2 BeaconBlockHeader
}

type AttesterSlashing struct {
	// First slashable attestation
	Slashable_attestation_1 SlashableAttestation
	// Second slashable attestation
	Slashable_attestation_2 SlashableAttestation
}

type SlashableAttestation struct {
	// Validator indices
	Validator_indices []ValidatorIndex
	// Attestation data
	Data AttestationData
	// Custody bitfield
	Custody_bitfield bitfield.Bitfield
	// Aggregate signature
	Aggregate_signature BLSSignature `ssz:"signature"`
}

type Attestation struct {
	// Attester aggregation bitfield
	Aggregation_bitfield bitfield.Bitfield
	// Attestation data
	Data AttestationData
	// Custody bitfield
	Custody_bitfield bitfield.Bitfield
	// BLS aggregate signature
	Aggregate_signature BLSSignature `ssz:"signature"`
}

type AttestationData struct {
	// Slot number
	Slot Slot
	// Shard number
	Shard Shard
	// Root of the signed beacon block
	Beacon_block_root Root
	// Root of the ancestor at the epoch boundary
	Epoch_boundary_root Root
	// Data from the shard since the last attestation
	Crosslink_data_root Root
	// Last crosslink
	Latest_crosslink Crosslink
	// Last justified epoch in the beacon state
	Justified_epoch Epoch
	// Hash of the last justified beacon block
	Justified_block_root Root
}

type AttestationDataAndCustodyBit struct {
	// Attestation data
	Data AttestationData
	// Custody bit
	Custody_bit bool
}

type Crosslink struct {
	// Epoch number
	Epoch Epoch
	// Shard data since the previous crosslink
	Crosslink_data_root Root
}

type Deposit struct {
	// Branch in the deposit tree
	Proof []Root
	// Index in the deposit tree
	Index DepositIndex
	// Data
	Deposit_data DepositData
}

type DepositData struct {
	// Amount in Gwei
	Amount Gwei
	// Timestamp from deposit contract
	Timestamp Timestamp
	// Deposit input
	Deposit_input DepositInput
}

type DepositInput struct {
	// BLS pubkey
	Pubkey BLSPubkey
	// Withdrawal credentials
	Withdrawal_credentials Root
	// A BLS signature of this `DepositInput`
	Proof_of_possession BLSSignature `ssz:"signature"`
}

type VoluntaryExit struct {
	// Minimum epoch for processing exit
	Epoch Epoch
	// Index of the exiting validator
	Validator_index ValidatorIndex
	// Validator signature
	Signature BLSSignature `ssz:"signature"`
}

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
	Signature BLSSignature `ssz:"signature"`
}

type Validator struct {
	// BLS public key
	Pubkey BLSPubkey
	// Withdrawal credentials
	Withdrawal_credentials Root
	// Epoch when validator activated
	Activation_epoch Epoch
	// Epoch when validator exited
	Exit_epoch Epoch
	// Epoch when validator is eligible to withdraw
	Withdrawable_epoch Epoch
	// Did the validator initiate an exit
	Initiated_exit bool
	// Was the validator slashed
	Slashed bool
}

func (v *Validator) IsActive(epoch Epoch) bool {
	return v.Activation_epoch <= epoch && epoch < v.Exit_epoch
}

type PendingAttestation struct {
	// Attester aggregation bitfield
	Aggregation_bitfield bitfield.Bitfield
	// Attestation data
	Data AttestationData
	// Custody bitfield
	Custody_bitfield bitfield.Bitfield
	// Inclusion slot
	Inclusion_slot Slot
}

type Fork struct {
	// TODO: fork versions are 64 bits, but usage is 32 bits in BLS domain. Spec unclear about it.
	// Previous fork version
	Previous_version uint64
	// Current fork version
	Current_version uint64
	// Fork epoch number
	Epoch Epoch
}

// Return the fork version of the given epoch
func (f Fork) GetVersion(epoch Epoch) uint64 {
	if epoch < f.Epoch {
		return f.Previous_version
	}
	return f.Current_version
}

type Eth1Data struct {
	// Root of the deposit tree
	Deposit_root Root
	// Block hash
	Block_hash Root
}

type Eth1DataVote struct {
	// Data being voted for
	Eth1_data Eth1Data
	// Vote count
	Vote_count uint64
}

type ValidatorBalances []Gwei

func (balances ValidatorBalances) ApplyStakeDeltas(deltas *stake.Deltas) {
	if len(deltas.Penalties) != len(balances) || len(deltas.Rewards) != len(balances) {
		panic("cannot apply deltas to balances list with different length")
	}
	for i := 0; i < len(balances); i++ {
		balances[i] = Max(
			0,
			balances[i]+deltas.Rewards[i]-deltas.Penalties[i],
		)
	}
}

// Return the effective balance (also known as "balance at stake") for a validator with the given index.
func (balances ValidatorBalances) Get_effective_balance(index ValidatorIndex) Gwei {
	return Max(balances[index], MAX_DEPOSIT_AMOUNT)
}

// Return the combined effective balance of an array of validators.
func (balances ValidatorBalances) Get_total_balance(indices []ValidatorIndex) (sum Gwei) {
	for _, vIndex := range indices {
		sum += balances.Get_effective_balance(vIndex)
	}
	return sum
}

type ValidatorRegistry []Validator

func (vr ValidatorRegistry) Is_validator_index(index ValidatorIndex) bool {
	return index < ValidatorIndex(len(vr))
}

func (vr ValidatorRegistry) Get_active_validator_indices(epoch Epoch) ValidatorIndexList {
	res := make([]ValidatorIndex, 0, len(vr))
	for i, v := range vr {
		if v.IsActive(epoch) {
			res = append(res, ValidatorIndex(i))
		}
	}
	return res
}

func (vr ValidatorRegistry) Get_active_validator_count(epoch Epoch) (count uint64) {
	for _, v := range vr {
		if v.IsActive(epoch) {
			count++
		}
	}
	return
}

// Shuffle active validators and split into crosslink committees.
// Return a list of committees (each a list of validator indices).
func (vr ValidatorRegistry) Get_shuffling(seed Bytes32, epoch Epoch) [][]ValidatorIndex {
	active_validator_indices := vr.Get_active_validator_indices(epoch)
	committee_count := Get_epoch_committee_count(uint64(len(active_validator_indices)))
	commitees := make([][]ValidatorIndex, committee_count, committee_count)
	// Active validators, shuffled in-place.
	hash := sha256.New()
	hashFn := func(in []byte) []byte {
		hash.Reset()
		hash.Write(in)
		return hash.Sum(nil)
	}
	eth2_shuffle.ShuffleList(hashFn, ValidatorIndexList(active_validator_indices).RawIndexSlice(), SHUFFLE_ROUND_COUNT, seed)
	committee_size := uint64(len(active_validator_indices)) / committee_count
	for i := uint64(0); i < committee_count; i += committee_size {
		commitees[i] = active_validator_indices[i : i+committee_size]
	}
	return commitees
}
