package beacon

import (
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/util/bitfield"
)

// NOTE: these containers are going to be moved to sub-packages, per-topic.

type ProposerSlashing struct {
	// Proposer index
	Proposer_index eth2.ValidatorIndex
	// First proposal
	Proposal_1 Proposal
	// Second proposal
	Proposal_2 Proposal
}

type Proposal struct {
	// Slot number
	Slot eth2.Slot
	// Shard number (`BEACON_CHAIN_SHARD_NUMBER` for beacon chain)
	Shard eth2.Shard
	// Block root
	Block_root eth2.Root
	// Signature
	Signature eth2.BLSSignature `ssz:"signature"`
}

type AttesterSlashing struct {
	// First slashable attestation
	Slashable_attestation_1 SlashableAttestation
	// Second slashable attestation
	Slashable_attestation_2 SlashableAttestation
}

type SlashableAttestation struct {
	// Validator indices
	Validator_indices []eth2.ValidatorIndex
	// Attestation data
	Data AttestationData
	// Custody bitfield
	Custody_bitfield bitfield.Bitfield
	// Aggregate signature
	Aggregate_signature eth2.BLSSignature `ssz:"signature"`
}

type Attestation struct {
	// Attester aggregation bitfield
	Aggregation_bitfield bitfield.Bitfield
	// Attestation data
	Data AttestationData
	// Custody bitfield
	Custody_bitfield bitfield.Bitfield
	// BLS aggregate signature
	Aggregate_signature eth2.BLSSignature `ssz:"signature"`
}

type AttestationData struct {
	// Slot number
	Slot eth2.Slot
	// Shard number
	Shard eth2.Shard
	// Root of the signed beacon block
	Beacon_block_root eth2.Root
	// Root of the ancestor at the epoch boundary
	Epoch_boundary_root eth2.Root
	// Data from the shard since the last attestation
	Crosslink_data_root eth2.Root
	// Last crosslink
	Latest_crosslink Crosslink
	// Last justified epoch in the beacon state
	Justified_epoch eth2.Epoch
	// Hash of the last justified beacon block
	Justified_block_root eth2.Root
}

type AttestationDataAndCustodyBit struct {
	// Attestation data
	Data AttestationData
	// Custody bit
	Custody_bit bool
}

type Crosslink struct {
	// Epoch number
	Epoch eth2.Epoch
	// Shard data since the previous crosslink
	Crosslink_data_root eth2.Root
}

type Deposit struct {
	// Branch in the deposit tree
	Branch []eth2.Root
	// Index in the deposit tree
	Index eth2.DepositIndex
	// Data
	Deposit_data DepositData
}

type DepositData struct {
	// Amount in Gwei
	Amount eth2.Gwei
	// Timestamp from deposit contract
	Timestamp eth2.Timestamp
	// Deposit input
	Deposit_input DepositInput
}

type DepositInput struct {
	// BLS pubkey
	Pubkey eth2.BLSPubkey
	// Withdrawal credentials
	Withdrawal_credentials eth2.Root
	// A BLS signature of this `DepositInput`
	Proof_of_possession eth2.BLSSignature `ssz:"signature"`
}

type VoluntaryExit struct {
	// Minimum epoch for processing exit
	Epoch eth2.Epoch
	// Index of the exiting validator
	Validator_index eth2.ValidatorIndex
	// Validator signature
	Signature eth2.BLSSignature `ssz:"signature"`
}

type Transfer struct {
	// Sender index
	From eth2.ValidatorIndex
	// Recipient index
	To eth2.ValidatorIndex
	// Amount in Gwei
	Amount eth2.Gwei
	// Fee in Gwei for block proposer
	Fee eth2.Gwei
	// Inclusion slot
	Slot eth2.Slot
	// Sender withdrawal pubkey
	Pubkey eth2.BLSPubkey
	// Sender signature
	Signature eth2.BLSSignature `ssz:"signature"`
}

type Validator struct {
	// BLS public key
	Pubkey eth2.BLSPubkey
	// Withdrawal credentials
	Withdrawal_credentials eth2.Root
	// Epoch when validator activated
	Activation_epoch eth2.Epoch
	// Epoch when validator exited
	Exit_epoch eth2.Epoch
	// Epoch when validator is eligible to withdraw
	Withdrawable_epoch eth2.Epoch
	// Did the validator initiate an exit
	Initiated_exit bool
	// Was the validator slashed
	Slashed bool
}

func (v *Validator) IsActive(epoch eth2.Epoch) bool {
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
	Inclusion_slot eth2.Slot
}

type Fork struct {
	// TODO: fork versions are 64 bits, but usage is 32 bits in BLS domain. Spec unclear about it.
	// Previous fork version
	Previous_version uint64
	// Current fork version
	Current_version uint64
	// Fork epoch number
	Epoch eth2.Epoch
}

// Return the fork version of the given epoch
func (f Fork) GetVersion(epoch eth2.Epoch) uint64 {
	if epoch < f.Epoch {
		return f.Previous_version
	}
	return f.Current_version
}

type Eth1Data struct {
	// Root of the deposit tree
	Deposit_root eth2.Root
	// Block hash
	Block_hash eth2.Root
}

type Eth1DataVote struct {
	// Data being voted for
	Eth1_data Eth1Data
	// Vote count
	Vote_count uint64
}
