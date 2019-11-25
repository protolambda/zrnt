package beacon

import (
	. "github.com/protolambda/zrnt/experimental/types"
)

var Version = Vector(Byte, 4)

const Epoch = Uint64

var BLSSignature = Vector(Byte, 96)
var BLSPubkey = Vector(Byte, 48)

const Bytes32 = Root
const Gwei = Uint64
const Slot = Uint64
const CommitteeIndex = Uint64
const ValidatorIndex = Uint64

var Fork = &Container{
	{"previous_version", Version},
	{"current_version", Version},
	{"epoch", Epoch}, // Epoch of latest fork
}
var Checkpoint = &Container{
	{"epoch", Epoch},
	{"root", Root},
}
var Validator = &Container{
	{"pubkey", BLSPubkey},
	{"withdrawal_credentials", Bytes32}, // Commitment to pubkey for withdrawals
	{"effective_balance", Gwei},         // Balance at stake
	{"slashed", Bool},
	// Status epochs
	{"activation_eligibility_epoch", Epoch}, // When criteria for activation were met
	{"activation_epoch", Epoch},
	{"exit_epoch", Epoch},
	{"withdrawable_epoch", Epoch}, // When validator can withdraw funds
}

var AttestationData = &Container{
	{"slot", Slot},
	{"index", CommitteeIndex},
	// LMD GHOST vote
	{"beacon_block_root", Root},
	// FFG vote
	{"source", Checkpoint},
	{"target", Checkpoint},
}
var CommitteeIndices = List(ValidatorIndex, MAX_VALIDATORS_PER_COMMITTEE)
var IndexedAttestation = &Container{
	{"attesting_indices", CommitteeIndices},
	{"data", AttestationData},
	{"signature", BLSSignature},
}
var CommitteeBits = Bitlist(MAX_VALIDATORS_PER_COMMITTEE)
var PendingAttestation = &Container{
	{"aggregation_bits", CommitteeBits},
	{"data", AttestationData},
	{"inclusion_delay", Slot},
	{"proposer_index", ValidatorIndex},
}
var Eth1Data = &Container{
	{"deposit_root", Root},
	{"deposit_count", Uint64},
	{"block_hash", Bytes32},
}
var BatchRoots = Vector(Root, SLOTS_PER_HISTORICAL_ROOT)
var HistoricalBatch = &Container{
	{"block_roots", BatchRoots},
	{"state_roots", BatchRoots},
}
var DepositData = &Container{
	{"pubkey", BLSPubkey},
	{"withdrawal_credentials", Bytes32},
	{"amount", Gwei},
	{"signature", BLSSignature},
}
var BeaconBlockHeader = &Container{
	{"slot", Slot},
	{"parent_root", Root},
	{"state_root", Root},
	{"body_root", Root},
	{"signature", BLSSignature},
}

// Beacon operations
var ProposerSlashing = &Container{
	{"proposer_index", ValidatorIndex},
	{"header_1", BeaconBlockHeader},
	{"header_2", BeaconBlockHeader},
}
var AttesterSlashing = &Container{
	{"attestation_1", IndexedAttestation},
	{"attestation_2", IndexedAttestation},
}
var Attestation = &Container{
	{"aggregation_bits", CommitteeBits},
	{"data", AttestationData},
	{"signature", BLSSignature},
}
var DepositProof = Vector(Bytes32, DEPOSIT_CONTRACT_TREE_DEPTH+1)
var Deposit = &Container{
	{"proof", DepositProof}, // Merkle path to deposit data list root
	{"data", DepositData},
}
var VoluntaryExit = &Container{
	{"epoch", Epoch}, // Earliest epoch when voluntary exit can be processed
	{"validator_index", ValidatorIndex},
	{"signature", BLSSignature},
}

var ProposerSlashings = List(ProposerSlashing, MAX_PROPOSER_SLASHINGS)
var AttesterSlashings = List(AttesterSlashing, MAX_ATTESTER_SLASHINGS)
var Attestations = List(Attestation, MAX_ATTESTATIONS)
var Deposits = List(Deposit, MAX_DEPOSITS)
var VoluntaryExits = List(VoluntaryExit, MAX_VOLUNTARY_EXITS)

// Beacon blocks
var BeaconBlockBody = &Container{
	{"randao_reveal", BLSSignature},
	{"eth1_data", Eth1Data}, // Eth1 data vote
	{"graffiti", Bytes32},   // Arbitrary data
	// Operations
	{"proposer_slashings", ProposerSlashings},
	{"attester_slashings", AttesterSlashings},
	{"attestations", Attestations},
	{"deposits", Deposits},
	{"voluntary_exits", VoluntaryExits},
}
var BeaconBlock = &Container{
	{"slot", Slot},
	{"parent_root", Root},
	{"state_root", Root},
	{"body", BeaconBlockBody},
	{"signature", BLSSignature},
}
var HistoricalRoots = List(Root, HISTORICAL_ROOTS_LIMIT)
var Eth1DataVotes = List(Eth1Data, SLOTS_PER_ETH1_VOTING_PERIOD)
var RegistryValidators = List(Validator, VALIDATOR_REGISTRY_LIMIT)
var RegistryBalances = List(Gwei, VALIDATOR_REGISTRY_LIMIT)
var RandaoMixes = Vector(Bytes32, EPOCHS_PER_HISTORICAL_VECTOR)
var Slashings = Vector(Gwei, EPOCHS_PER_SLASHINGS_VECTOR)
var PendingAttestations = List(PendingAttestation, MAX_ATTESTATIONS*SLOTS_PER_EPOCH)
var JustificationBits = Bitvector(JUSTIFICATION_BITS_LENGTH)

// Beacon state
var BeaconState = &Container{
	// Versioning
	{"genesis_time", Uint64},
	{"slot", Slot},
	{"fork", Fork},
	// History
	{"latest_block_header", BeaconBlockHeader},
	{"block_roots", BatchRoots},
	{"state_roots", BatchRoots},
	{"historical_roots", HistoricalRoots},
	// Eth1
	{"eth1_data", Eth1Data},
	{"eth1_data_votes", Eth1DataVotes},
	{"eth1_deposit_index", Uint64},
	// Registry
	{"validators", RegistryValidators},
	{"balances", RegistryBalances},
	// Randomness
	{"randao_mixes", RandaoMixes},
	// Slashings
	{"slashings", Slashings}, // Per-epoch sums of slashed effective balances
	// Attestations
	{"previous_epoch_attestations", PendingAttestations},
	{"current_epoch_attestations", PendingAttestation},
	// Finality
	{"justification_bits", JustificationBits},     // Bit set for every recent justified epoch
	{"previous_justified_checkpoint", Checkpoint}, // Previous epoch snapshot
	{"current_justified_checkpoint", Checkpoint},
	{"finalized_checkpoint", Checkpoint},
}
