package common

import (
	"encoding/json"

	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
	"gopkg.in/yaml.v3"
)

const TARGET_AGGREGATORS_PER_COMMITTEE = 16
const RANDOM_SUBNETS_PER_VALIDATOR = 1
const EPOCHS_PER_RANDOM_SUBNET_SUBSCRIPTION = 256
const BLS_WITHDRAWAL_PREFIX = 0
const ETH1_ADDRESS_WITHDRAWAL_PREFIX = 1
const SYNC_COMMITTEE_SUBNET_COUNT = 4
const TARGET_AGGREGATORS_PER_SYNC_SUBCOMMITTEE = 16

// Phase0
var DOMAIN_BEACON_PROPOSER = BLSDomainType{0x00, 0x00, 0x00, 0x00}
var DOMAIN_BEACON_ATTESTER = BLSDomainType{0x01, 0x00, 0x00, 0x00}
var DOMAIN_RANDAO = BLSDomainType{0x02, 0x00, 0x00, 0x00}
var DOMAIN_DEPOSIT = BLSDomainType{0x03, 0x00, 0x00, 0x00}
var DOMAIN_VOLUNTARY_EXIT = BLSDomainType{0x04, 0x00, 0x00, 0x00}
var DOMAIN_SELECTION_PROOF = BLSDomainType{0x05, 0x00, 0x00, 0x00}
var DOMAIN_AGGREGATE_AND_PROOF = BLSDomainType{0x06, 0x00, 0x00, 0x00}

// Altair
var DOMAIN_SYNC_COMMITTEE = BLSDomainType{0x07, 0x00, 0x00, 0x00}
var DOMAIN_SYNC_COMMITTEE_SELECTION_PROOF = BLSDomainType{0x08, 0x00, 0x00, 0x00}
var DOMAIN_CONTRIBUTION_AND_PROOF = BLSDomainType{0x09, 0x00, 0x00, 0x00}

// Capella
var DOMAIN_BLS_TO_EXECUTION_CHANGE = BLSDomainType{0x0A, 0x00, 0x00, 0x00}

// Deneb
const BLOB_TX_TYPE = 0x03
const VERSIONED_HASH_VERSION_KZG = 0x01

type Phase0Preset struct {
	// Misc.
	MAX_COMMITTEES_PER_SLOT      Uint64View `yaml:"MAX_COMMITTEES_PER_SLOT" json:"MAX_COMMITTEES_PER_SLOT"`
	TARGET_COMMITTEE_SIZE        Uint64View `yaml:"TARGET_COMMITTEE_SIZE" json:"TARGET_COMMITTEE_SIZE"`
	MAX_VALIDATORS_PER_COMMITTEE Uint64View `yaml:"MAX_VALIDATORS_PER_COMMITTEE" json:"MAX_VALIDATORS_PER_COMMITTEE"`
	SHUFFLE_ROUND_COUNT          Uint8View  `yaml:"SHUFFLE_ROUND_COUNT" json:"SHUFFLE_ROUND_COUNT"`

	// Balance math
	HYSTERESIS_QUOTIENT            Uint64View `yaml:"HYSTERESIS_QUOTIENT" json:"HYSTERESIS_QUOTIENT"`
	HYSTERESIS_DOWNWARD_MULTIPLIER Uint64View `yaml:"HYSTERESIS_DOWNWARD_MULTIPLIER" json:"HYSTERESIS_DOWNWARD_MULTIPLIER"`
	HYSTERESIS_UPWARD_MULTIPLIER   Uint64View `yaml:"HYSTERESIS_UPWARD_MULTIPLIER" json:"HYSTERESIS_UPWARD_MULTIPLIER"`

	// Gwei values
	MIN_DEPOSIT_AMOUNT          Gwei `yaml:"MIN_DEPOSIT_AMOUNT" json:"MIN_DEPOSIT_AMOUNT"`
	MAX_EFFECTIVE_BALANCE       Gwei `yaml:"MAX_EFFECTIVE_BALANCE" json:"MAX_EFFECTIVE_BALANCE"`
	EFFECTIVE_BALANCE_INCREMENT Gwei `yaml:"EFFECTIVE_BALANCE_INCREMENT" json:"EFFECTIVE_BALANCE_INCREMENT"`

	// Time parameters
	MIN_ATTESTATION_INCLUSION_DELAY  Slot  `yaml:"MIN_ATTESTATION_INCLUSION_DELAY" json:"MIN_ATTESTATION_INCLUSION_DELAY"`
	SLOTS_PER_EPOCH                  Slot  `yaml:"SLOTS_PER_EPOCH" json:"SLOTS_PER_EPOCH"`
	MIN_SEED_LOOKAHEAD               Epoch `yaml:"MIN_SEED_LOOKAHEAD" json:"MIN_SEED_LOOKAHEAD"`
	MAX_SEED_LOOKAHEAD               Epoch `yaml:"MAX_SEED_LOOKAHEAD" json:"MAX_SEED_LOOKAHEAD"`
	EPOCHS_PER_ETH1_VOTING_PERIOD    Epoch `yaml:"EPOCHS_PER_ETH1_VOTING_PERIOD" json:"EPOCHS_PER_ETH1_VOTING_PERIOD"`
	SLOTS_PER_HISTORICAL_ROOT        Slot  `yaml:"SLOTS_PER_HISTORICAL_ROOT" json:"SLOTS_PER_HISTORICAL_ROOT"`
	MIN_EPOCHS_TO_INACTIVITY_PENALTY Epoch `yaml:"MIN_EPOCHS_TO_INACTIVITY_PENALTY" json:"MIN_EPOCHS_TO_INACTIVITY_PENALTY"`

	// State vector lengths
	EPOCHS_PER_HISTORICAL_VECTOR Epoch      `yaml:"EPOCHS_PER_HISTORICAL_VECTOR" json:"EPOCHS_PER_HISTORICAL_VECTOR"`
	EPOCHS_PER_SLASHINGS_VECTOR  Epoch      `yaml:"EPOCHS_PER_SLASHINGS_VECTOR" json:"EPOCHS_PER_SLASHINGS_VECTOR"`
	HISTORICAL_ROOTS_LIMIT       Uint64View `yaml:"HISTORICAL_ROOTS_LIMIT" json:"HISTORICAL_ROOTS_LIMIT"`
	VALIDATOR_REGISTRY_LIMIT     Uint64View `yaml:"VALIDATOR_REGISTRY_LIMIT" json:"VALIDATOR_REGISTRY_LIMIT"`

	// Reward and penalty quotients
	BASE_REWARD_FACTOR               Uint64View `yaml:"BASE_REWARD_FACTOR" json:"BASE_REWARD_FACTOR"`
	WHISTLEBLOWER_REWARD_QUOTIENT    Uint64View `yaml:"WHISTLEBLOWER_REWARD_QUOTIENT" json:"WHISTLEBLOWER_REWARD_QUOTIENT"`
	PROPOSER_REWARD_QUOTIENT         Uint64View `yaml:"PROPOSER_REWARD_QUOTIENT" json:"PROPOSER_REWARD_QUOTIENT"`
	INACTIVITY_PENALTY_QUOTIENT      Uint64View `yaml:"INACTIVITY_PENALTY_QUOTIENT" json:"INACTIVITY_PENALTY_QUOTIENT"`
	MIN_SLASHING_PENALTY_QUOTIENT    Uint64View `yaml:"MIN_SLASHING_PENALTY_QUOTIENT" json:"MIN_SLASHING_PENALTY_QUOTIENT"`
	PROPORTIONAL_SLASHING_MULTIPLIER Uint64View `yaml:"PROPORTIONAL_SLASHING_MULTIPLIER" json:"PROPORTIONAL_SLASHING_MULTIPLIER"`

	// Max operations per block
	MAX_PROPOSER_SLASHINGS Uint64View `yaml:"MAX_PROPOSER_SLASHINGS" json:"MAX_PROPOSER_SLASHINGS"`
	MAX_ATTESTER_SLASHINGS Uint64View `yaml:"MAX_ATTESTER_SLASHINGS" json:"MAX_ATTESTER_SLASHINGS"`
	MAX_ATTESTATIONS       Uint64View `yaml:"MAX_ATTESTATIONS" json:"MAX_ATTESTATIONS"`
	MAX_DEPOSITS           Uint64View `yaml:"MAX_DEPOSITS" json:"MAX_DEPOSITS"`
	MAX_VOLUNTARY_EXITS    Uint64View `yaml:"MAX_VOLUNTARY_EXITS" json:"MAX_VOLUNTARY_EXITS"`
}

type AltairPreset struct {
	// Updated penalty values
	INACTIVITY_PENALTY_QUOTIENT_ALTAIR      Uint64View `yaml:"INACTIVITY_PENALTY_QUOTIENT_ALTAIR" json:"INACTIVITY_PENALTY_QUOTIENT_ALTAIR"`
	MIN_SLASHING_PENALTY_QUOTIENT_ALTAIR    Uint64View `yaml:"MIN_SLASHING_PENALTY_QUOTIENT_ALTAIR" json:"MIN_SLASHING_PENALTY_QUOTIENT_ALTAIR"`
	PROPORTIONAL_SLASHING_MULTIPLIER_ALTAIR Uint64View `yaml:"PROPORTIONAL_SLASHING_MULTIPLIER_ALTAIR" json:"PROPORTIONAL_SLASHING_MULTIPLIER_ALTAIR"`

	// Sync committee
	SYNC_COMMITTEE_SIZE              Uint64View `yaml:"SYNC_COMMITTEE_SIZE" json:"SYNC_COMMITTEE_SIZE"`
	EPOCHS_PER_SYNC_COMMITTEE_PERIOD Epoch      `yaml:"EPOCHS_PER_SYNC_COMMITTEE_PERIOD" json:"EPOCHS_PER_SYNC_COMMITTEE_PERIOD"`

	// Sync committees and light clients
	MIN_SYNC_COMMITTEE_PARTICIPANTS Uint64View `yaml:"MIN_SYNC_COMMITTEE_PARTICIPANTS" json:"MIN_SYNC_COMMITTEE_PARTICIPANTS"`
}

type BellatrixPreset struct {
	INACTIVITY_PENALTY_QUOTIENT_BELLATRIX      Uint64View `yaml:"INACTIVITY_PENALTY_QUOTIENT_BELLATRIX" json:"INACTIVITY_PENALTY_QUOTIENT_BELLATRIX"`
	MIN_SLASHING_PENALTY_QUOTIENT_BELLATRIX    Uint64View `yaml:"MIN_SLASHING_PENALTY_QUOTIENT_BELLATRIX" json:"MIN_SLASHING_PENALTY_QUOTIENT_BELLATRIX"`
	PROPORTIONAL_SLASHING_MULTIPLIER_BELLATRIX Uint64View `yaml:"PROPORTIONAL_SLASHING_MULTIPLIER_BELLATRIX" json:"PROPORTIONAL_SLASHING_MULTIPLIER_BELLATRIX"`
	MAX_BYTES_PER_TRANSACTION                  Uint64View `yaml:"MAX_BYTES_PER_TRANSACTION" json:"MAX_BYTES_PER_TRANSACTION"`
	MAX_TRANSACTIONS_PER_PAYLOAD               Uint64View `yaml:"MAX_TRANSACTIONS_PER_PAYLOAD" json:"MAX_TRANSACTIONS_PER_PAYLOAD"`
	BYTES_PER_LOGS_BLOOM                       Uint64View `yaml:"BYTES_PER_LOGS_BLOOM" json:"BYTES_PER_LOGS_BLOOM"`
	MAX_EXTRA_DATA_BYTES                       Uint64View `yaml:"MAX_EXTRA_DATA_BYTES" json:"MAX_EXTRA_DATA_BYTES"`
}

type CapellaPreset struct {
	MAX_BLS_TO_EXECUTION_CHANGES         Uint64View `yaml:"MAX_BLS_TO_EXECUTION_CHANGES" json:"MAX_BLS_TO_EXECUTION_CHANGES"`
	MAX_WITHDRAWALS_PER_PAYLOAD          Uint64View `yaml:"MAX_WITHDRAWALS_PER_PAYLOAD" json:"MAX_WITHDRAWALS_PER_PAYLOAD"`
	MAX_VALIDATORS_PER_WITHDRAWALS_SWEEP Uint64View `yaml:"MAX_VALIDATORS_PER_WITHDRAWALS_SWEEP" json:"MAX_VALIDATORS_PER_WITHDRAWALS_SWEEP"`
}

type DenebPreset struct {
	FIELD_ELEMENTS_PER_BLOB              Uint64View `yaml:"FIELD_ELEMENTS_PER_BLOB" json:"FIELD_ELEMENTS_PER_BLOB"`
	MAX_BLOB_COMMITMENTS_PER_BLOCK       Uint64View `yaml:"MAX_BLOB_COMMITMENTS_PER_BLOCK" json:"MAX_BLOB_COMMITMENTS_PER_BLOCK"`
	KZG_COMMITMENT_INCLUSION_PROOF_DEPTH Uint64View `yaml:"KZG_COMMITMENT_INCLUSION_PROOF_DEPTH" json:"KZG_COMMITMENT_INCLUSION_PROOF_DEPTH"`
}

type ElectraPreset struct {
	// Gwei values
	MIN_ACTIVATION_BALANCE        Gwei `yaml:"MIN_ACTIVATION_BALANCE" json:"MIN_ACTIVATION_BALANCE"`
	MAX_EFFECTIVE_BALANCE_ELECTRA Gwei `yaml:"MAX_EFFECTIVE_BALANCE_ELECTRA" json:"MAX_EFFECTIVE_BALANCE_ELECTRA"`

	// State list lengths
	PENDING_DEPOSITS_LIMIT            Uint64View `yaml:"PENDING_DEPOSITS_LIMIT" json:"PENDING_DEPOSITS_LIMIT"`
	PENDING_PARTIAL_WITHDRAWALS_LIMIT Uint64View `yaml:"PENDING_PARTIAL_WITHDRAWALS_LIMIT" json:"PENDING_PARTIAL_WITHDRAWALS_LIMIT"`
	PENDING_CONSOLIDATIONS_LIMIT      Uint64View `yaml:"PENDING_CONSOLIDATIONS_LIMIT" json:"PENDING_CONSOLIDATIONS_LIMIT"`

	// Reward and penalty quotients
	MIN_SLASHING_PENALTY_QUOTIENT_ELECTRA Uint64View `yaml:"MIN_SLASHING_PENALTY_QUOTIENT_ELECTRA" json:"MIN_SLASHING_PENALTY_QUOTIENT_ELECTRA"`
	WHISTLEBLOWER_REWARD_QUOTIENT_ELECTRA Uint64View `yaml:"WHISTLEBLOWER_REWARD_QUOTIENT_ELECTRA" json:"WHISTLEBLOWER_REWARD_QUOTIENT_ELECTRA"`

	// Max operations per block
	MAX_ATTESTER_SLASHINGS_ELECTRA         Uint64View `yaml:"MAX_ATTESTER_SLASHINGS_ELECTRA" json:"MAX_ATTESTER_SLASHINGS_ELECTRA"`
	MAX_ATTESTATIONS_ELECTRA               Uint64View `yaml:"MAX_ATTESTATIONS_ELECTRA" json:"MAX_ATTESTATIONS_ELECTRA"`
	MAX_CONSOLIDATION_REQUESTS_PER_PAYLOAD Uint64View `yaml:"MAX_CONSOLIDATION_REQUESTS_PER_PAYLOAD" json:"MAX_CONSOLIDATION_REQUESTS_PER_PAYLOAD"`

	// Execution
	MAX_DEPOSIT_REQUESTS_PER_PAYLOAD    Uint64View `yaml:"MAX_DEPOSIT_REQUESTS_PER_PAYLOAD" json:"MAX_DEPOSIT_REQUESTS_PER_PAYLOAD"`
	MAX_WITHDRAWAL_REQUESTS_PER_PAYLOAD Uint64View `yaml:"MAX_WITHDRAWAL_REQUESTS_PER_PAYLOAD" json:"MAX_WITHDRAWAL_REQUESTS_PER_PAYLOAD"`

	// Withdrawals processing
	MAX_PENDING_PARTIALS_PER_WITHDRAWALS_SWEEP Uint64View `yaml:"MAX_PENDING_PARTIALS_PER_WITHDRAWALS_SWEEP" json:"MAX_PENDING_PARTIALS_PER_WITHDRAWALS_SWEEP"`

	// Pending deposits processing
	MAX_PENDING_DEPOSITS_PER_EPOCH Uint64View `yaml:"MAX_PENDING_DEPOSITS_PER_EPOCH" json:"MAX_PENDING_DEPOSITS_PER_EPOCH"`
}

type Config struct {
	PRESET_BASE string `yaml:"PRESET_BASE" json:"PRESET_BASE"`

	CONFIG_NAME string `yaml:"CONFIG_NAME" json:"CONFIG_NAME"`

	// Transition
	TERMINAL_TOTAL_DIFFICULTY            Uint256View `yaml:"TERMINAL_TOTAL_DIFFICULTY" json:"TERMINAL_TOTAL_DIFFICULTY"`
	TERMINAL_BLOCK_HASH                  Hash32      `yaml:"TERMINAL_BLOCK_HASH" json:"TERMINAL_BLOCK_HASH"`
	TERMINAL_BLOCK_HASH_ACTIVATION_EPOCH Epoch       `yaml:"TERMINAL_BLOCK_HASH_ACTIVATION_EPOCH" json:"TERMINAL_BLOCK_HASH_ACTIVATION_EPOCH"`

	// Genesis.
	MIN_GENESIS_ACTIVE_VALIDATOR_COUNT Uint64View `yaml:"MIN_GENESIS_ACTIVE_VALIDATOR_COUNT" json:"MIN_GENESIS_ACTIVE_VALIDATOR_COUNT"`
	MIN_GENESIS_TIME                   Timestamp  `yaml:"MIN_GENESIS_TIME" json:"MIN_GENESIS_TIME"`
	GENESIS_FORK_VERSION               Version    `yaml:"GENESIS_FORK_VERSION" json:"GENESIS_FORK_VERSION"`
	GENESIS_DELAY                      Timestamp  `yaml:"GENESIS_DELAY" json:"GENESIS_DELAY"`

	// Altair
	ALTAIR_FORK_VERSION Version `yaml:"ALTAIR_FORK_VERSION" json:"ALTAIR_FORK_VERSION"`
	ALTAIR_FORK_EPOCH   Epoch   `yaml:"ALTAIR_FORK_EPOCH" json:"ALTAIR_FORK_EPOCH"`

	// Bellatrix
	BELLATRIX_FORK_VERSION Version `yaml:"BELLATRIX_FORK_VERSION" json:"BELLATRIX_FORK_VERSION"`
	BELLATRIX_FORK_EPOCH   Epoch   `yaml:"BELLATRIX_FORK_EPOCH" json:"BELLATRIX_FORK_EPOCH"`

	// Capella
	CAPELLA_FORK_VERSION Version `yaml:"CAPELLA_FORK_VERSION" json:"CAPELLA_FORK_VERSION"`
	CAPELLA_FORK_EPOCH   Epoch   `yaml:"CAPELLA_FORK_EPOCH" json:"CAPELLA_FORK_EPOCH"`

	// Deneb
	DENEB_FORK_VERSION Version `yaml:"DENEB_FORK_VERSION" json:"DENEB_FORK_VERSION"`
	DENEB_FORK_EPOCH   Epoch   `yaml:"DENEB_FORK_EPOCH" json:"DENEB_FORK_EPOCH"`

	// Electra
	ELECTRA_FORK_VERSION Version `yaml:"ELECTRA_FORK_VERSION" json:"ELECTRA_FORK_VERSION"`
	ELECTRA_FORK_EPOCH   Epoch   `yaml:"ELECTRA_FORK_EPOCH" json:"ELECTRA_FORK_EPOCH"`

	// Fulu
	FULU_FORK_VERSION Version `yaml:"FULU_FORK_VERSION" json:"FULU_FORK_VERSION"`
	FULU_FORK_EPOCH   Epoch   `yaml:"FULU_FORK_EPOCH" json:"FULU_FORK_EPOCH"`

	// EIP7441
	EIP7441_FORK_VERSION Version `yaml:"EIP7441_FORK_VERSION" json:"EIP7441_FORK_VERSION"`
	EIP7441_FORK_EPOCH   Epoch   `yaml:"EIP7441_FORK_EPOCH" json:"EIP7441_FORK_EPOCH"`
	// EIP7732
	EIP7732_FORK_VERSION Version `yaml:"EIP7732_FORK_VERSION" json:"EIP7732_FORK_VERSION"`
	EIP7732_FORK_EPOCH   Epoch   `yaml:"EIP7732_FORK_EPOCH" json:"EIP7732_FORK_EPOCH"`

	// Time parameters
	SECONDS_PER_SLOT                    Timestamp  `yaml:"SECONDS_PER_SLOT" json:"SECONDS_PER_SLOT"`
	SECONDS_PER_ETH1_BLOCK              Uint64View `yaml:"SECONDS_PER_ETH1_BLOCK" json:"SECONDS_PER_ETH1_BLOCK"`
	MIN_VALIDATOR_WITHDRAWABILITY_DELAY Epoch      `yaml:"MIN_VALIDATOR_WITHDRAWABILITY_DELAY" json:"MIN_VALIDATOR_WITHDRAWABILITY_DELAY"`
	SHARD_COMMITTEE_PERIOD              Epoch      `yaml:"SHARD_COMMITTEE_PERIOD" json:"SHARD_COMMITTEE_PERIOD"`
	ETH1_FOLLOW_DISTANCE                Uint64View `yaml:"ETH1_FOLLOW_DISTANCE" json:"ETH1_FOLLOW_DISTANCE"`

	// Validator cycle
	INACTIVITY_SCORE_BIAS                Uint64View `yaml:"INACTIVITY_SCORE_BIAS" json:"INACTIVITY_SCORE_BIAS"`
	INACTIVITY_SCORE_RECOVERY_RATE       Uint64View `yaml:"INACTIVITY_SCORE_RECOVERY_RATE" json:"INACTIVITY_SCORE_RECOVERY_RATE"`
	EJECTION_BALANCE                     Gwei       `yaml:"EJECTION_BALANCE" json:"EJECTION_BALANCE"`
	MIN_PER_EPOCH_CHURN_LIMIT            Uint64View `yaml:"MIN_PER_EPOCH_CHURN_LIMIT" json:"MIN_PER_EPOCH_CHURN_LIMIT"`
	CHURN_LIMIT_QUOTIENT                 Uint64View `yaml:"CHURN_LIMIT_QUOTIENT" json:"CHURN_LIMIT_QUOTIENT"`
	MAX_PER_EPOCH_ACTIVATION_CHURN_LIMIT Uint64View `yaml:"MAX_PER_EPOCH_ACTIVATION_CHURN_LIMIT" json:"MAX_PER_EPOCH_ACTIVATION_CHURN_LIMIT"`

	// Fork choice
	PROPOSER_SCORE_BOOST                Uint64View `yaml:"PROPOSER_SCORE_BOOST" json:"PROPOSER_SCORE_BOOST"`
	REORG_HEAD_WEIGHT_THRESHOLD         Uint64View `yaml:"REORG_HEAD_WEIGHT_THRESHOLD" json:"REORG_HEAD_WEIGHT_THRESHOLD"`
	REORG_PARENT_WEIGHT_THRESHOLD       Uint64View `yaml:"REORG_PARENT_WEIGHT_THRESHOLD" json:"REORG_PARENT_WEIGHT_THRESHOLD"`
	REORG_MAX_EPOCHS_SINCE_FINALIZATION Epoch      `yaml:"REORG_MAX_EPOCHS_SINCE_FINALIZATION" json:"REORG_MAX_EPOCHS_SINCE_FINALIZATION"`

	// Deposit contract
	DEPOSIT_CHAIN_ID         Uint64View  `yaml:"DEPOSIT_CHAIN_ID" json:"DEPOSIT_CHAIN_ID"`
	DEPOSIT_NETWORK_ID       Uint64View  `yaml:"DEPOSIT_NETWORK_ID" json:"DEPOSIT_NETWORK_ID"`
	DEPOSIT_CONTRACT_ADDRESS Eth1Address `yaml:"DEPOSIT_CONTRACT_ADDRESS" json:"DEPOSIT_CONTRACT_ADDRESS"`

	// Networking
	MAX_PAYLOAD_SIZE                   Uint64View           `yaml:"MAX_PAYLOAD_SIZE" json:"MAX_PAYLOAD_SIZE"`
	MAX_REQUEST_BLOCKS                 Uint64View           `yaml:"MAX_REQUEST_BLOCKS" json:"MAX_REQUEST_BLOCKS"`
	EPOCHS_PER_SUBNET_SUBSCRIPTION     Uint64View           `yaml:"EPOCHS_PER_SUBNET_SUBSCRIPTION" json:"EPOCHS_PER_SUBNET_SUBSCRIPTION"`
	MIN_EPOCHS_FOR_BLOCK_REQUESTS      Uint64View           `yaml:"MIN_EPOCHS_FOR_BLOCK_REQUESTS" json:"MIN_EPOCHS_FOR_BLOCK_REQUESTS"`
	TTFB_TIMEOUT                       Uint64View           `yaml:"TTFB_TIMEOUT" json:"TTFB_TIMEOUT"`
	RESP_TIMEOUT                       Uint64View           `yaml:"RESP_TIMEOUT" json:"RESP_TIMEOUT"`
	ATTESTATION_PROPAGATION_SLOT_RANGE Uint64View           `yaml:"ATTESTATION_PROPAGATION_SLOT_RANGE" json:"ATTESTATION_PROPAGATION_SLOT_RANGE"`
	MAXIMUM_GOSSIP_CLOCK_DISPARITY     Uint64View           `yaml:"MAXIMUM_GOSSIP_CLOCK_DISPARITY" json:"MAXIMUM_GOSSIP_CLOCK_DISPARITY"`
	MESSAGE_DOMAIN_INVALID_SNAPPY      NetworkMessageDomain `yaml:"MESSAGE_DOMAIN_INVALID_SNAPPY" json:"MESSAGE_DOMAIN_INVALID_SNAPPY"`
	MESSAGE_DOMAIN_VALID_SNAPPY        NetworkMessageDomain `yaml:"MESSAGE_DOMAIN_VALID_SNAPPY" json:"MESSAGE_DOMAIN_VALID_SNAPPY"`
	SUBNETS_PER_NODE                   Uint64View           `yaml:"SUBNETS_PER_NODE" json:"SUBNETS_PER_NODE"`
	ATTESTATION_SUBNET_COUNT           Uint64View           `yaml:"ATTESTATION_SUBNET_COUNT" json:"ATTESTATION_SUBNET_COUNT"`
	ATTESTATION_SUBNET_EXTRA_BITS      Uint64View           `yaml:"ATTESTATION_SUBNET_EXTRA_BITS" json:"ATTESTATION_SUBNET_EXTRA_BITS"`
	ATTESTATION_SUBNET_PREFIX_BITS     Uint64View           `yaml:"ATTESTATION_SUBNET_PREFIX_BITS" json:"ATTESTATION_SUBNET_PREFIX_BITS"`

	// Deneb
	MAX_REQUEST_BLOCKS_DENEB              Uint64View `yaml:"MAX_REQUEST_BLOCKS_DENEB" json:"MAX_REQUEST_BLOCKS_DENEB"`
	MIN_EPOCHS_FOR_BLOB_SIDECARS_REQUESTS Uint64View `yaml:"MIN_EPOCHS_FOR_BLOB_SIDECARS_REQUESTS" json:"MIN_EPOCHS_FOR_BLOB_SIDECARS_REQUESTS"`
	BLOB_SIDECAR_SUBNET_COUNT             Uint64View `yaml:"BLOB_SIDECAR_SUBNET_COUNT" json:"BLOB_SIDECAR_SUBNET_COUNT"`
	MAX_BLOBS_PER_BLOCK                   Uint64View `yaml:"MAX_BLOBS_PER_BLOCK" json:"MAX_BLOBS_PER_BLOCK"`
	MAX_REQUEST_BLOB_SIDECARS             Uint64View `yaml:"MAX_REQUEST_BLOB_SIDECARS" json:"MAX_REQUEST_BLOB_SIDECARS"`

	// Electra
	MIN_PER_EPOCH_CHURN_LIMIT_ELECTRA         Uint64View `yaml:"MIN_PER_EPOCH_CHURN_LIMIT_ELECTRA" json:"MIN_PER_EPOCH_CHURN_LIMIT_ELECTRA"`
	MAX_PER_EPOCH_ACTIVATION_EXIT_CHURN_LIMIT Uint64View `yaml:"MAX_PER_EPOCH_ACTIVATION_EXIT_CHURN_LIMIT" json:"MAX_PER_EPOCH_ACTIVATION_EXIT_CHURN_LIMIT"`
	BLOB_SIDECAR_SUBNET_COUNT_ELECTRA         Uint64View `yaml:"BLOB_SIDECAR_SUBNET_COUNT_ELECTRA" json:"BLOB_SIDECAR_SUBNET_COUNT_ELECTRA"`
	MAX_BLOBS_PER_BLOCK_ELECTRA               Uint64View `yaml:"MAX_BLOBS_PER_BLOCK_ELECTRA" json:"MAX_BLOBS_PER_BLOCK_ELECTRA"`
	MAX_REQUEST_BLOB_SIDECARS_ELECTRA         Uint64View `yaml:"MAX_REQUEST_BLOB_SIDECARS_ELECTRA" json:"MAX_REQUEST_BLOB_SIDECARS_ELECTRA"`

	// Fulu
	NUMBER_OF_COLUMNS                            Uint64View `yaml:"NUMBER_OF_COLUMNS" json:"NUMBER_OF_COLUMNS"`
	NUMBER_OF_CUSTODY_GROUPS                     Uint64View `yaml:"NUMBER_OF_CUSTODY_GROUPS" json:"NUMBER_OF_CUSTODY_GROUPS"`
	DATA_COLUMN_SIDECAR_SUBNET_COUNT             Uint64View `yaml:"DATA_COLUMN_SIDECAR_SUBNET_COUNT" json:"DATA_COLUMN_SIDECAR_SUBNET_COUNT"`
	MAX_REQUEST_DATA_COLUMN_SIDECARS             Uint64View `yaml:"MAX_REQUEST_DATA_COLUMN_SIDECARS" json:"MAX_REQUEST_DATA_COLUMN_SIDECARS"`
	SAMPLES_PER_SLOT                             Uint64View `yaml:"SAMPLES_PER_SLOT" json:"SAMPLES_PER_SLOT"`
	CUSTODY_REQUIREMENT                          Uint64View `yaml:"CUSTODY_REQUIREMENT" json:"CUSTODY_REQUIREMENT"`
	VALIDATOR_CUSTODY_REQUIREMENT                Uint64View `yaml:"VALIDATOR_CUSTODY_REQUIREMENT" json:"VALIDATOR_CUSTODY_REQUIREMENT"`
	BALANCE_PER_ADDITIONAL_CUSTODY_GROUP         Uint64View `yaml:"BALANCE_PER_ADDITIONAL_CUSTODY_GROUP" json:"BALANCE_PER_ADDITIONAL_CUSTODY_GROUP"`
	MAX_BLOBS_PER_BLOCK_FULU                     Uint64View `yaml:"MAX_BLOBS_PER_BLOCK_FULU" json:"MAX_BLOBS_PER_BLOCK_FULU"`
	MIN_EPOCHS_FOR_DATA_COLUMN_SIDECARS_REQUESTS Uint64View `yaml:"MIN_EPOCHS_FOR_DATA_COLUMN_SIDECARS_REQUESTS" json:"MIN_EPOCHS_FOR_DATA_COLUMN_SIDECARS_REQUESTS"`

	// EIP7441
	EPOCHS_PER_SHUFFLING_PHASE Uint64View `yaml:"EPOCHS_PER_SHUFFLING_PHASE" json:"EPOCHS_PER_SHUFFLING_PHASE"`
	PROPOSER_SELECTION_GAP     Uint64View `yaml:"PROPOSER_SELECTION_GAP" json:"PROPOSER_SELECTION_GAP"`

	// EIP7732
	MAX_REQUEST_PAYLOADS Uint64View `yaml:"MAX_REQUEST_PAYLOADS" json:"MAX_REQUEST_PAYLOADS"`
}

type SpecObj interface {
	Deserialize(spec *Spec, dr *codec.DecodingReader) error
	Serialize(spec *Spec, w *codec.EncodingWriter) error
	ByteLength(spec *Spec) uint64
	HashTreeRoot(spec *Spec, h tree.HashFn) Root
	FixedLength(spec *Spec) uint64
}

type SSZObj interface {
	codec.Serializable
	codec.Deserializable
	codec.FixedLength
	tree.HTR
}

type WrappedSpecObj interface {
	SSZObj
	Unwrap() (*Spec, SpecObj)
}

type specObj struct {
	spec *Spec
	des  SpecObj
}

func (s *specObj) Deserialize(dr *codec.DecodingReader) error {
	return s.des.Deserialize(s.spec, dr)
}

func (s *specObj) Serialize(w *codec.EncodingWriter) error {
	return s.des.Serialize(s.spec, w)
}

func (s *specObj) ByteLength() uint64 {
	return s.des.ByteLength(s.spec)
}

func (s *specObj) HashTreeRoot(h tree.HashFn) Root {
	return s.des.HashTreeRoot(s.spec, h)
}

func (s *specObj) FixedLength() uint64 {
	return s.des.FixedLength(s.spec)
}

func (s *specObj) Unwrap() (*Spec, SpecObj) {
	return s.spec, s.des
}

func (s *specObj) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, s.des)
}

func (s *specObj) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.des)
}

func (s *specObj) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(s.des)
}

func (s *specObj) MarshalYAML() (interface{}, error) {
	return s.des, nil
}

type Spec struct {
	Phase0Preset    `json:",inline" yaml:",inline"`
	AltairPreset    `json:",inline" yaml:",inline"`
	BellatrixPreset `json:",inline" yaml:",inline"`
	CapellaPreset   `json:",inline" yaml:",inline"`
	DenebPreset     `json:",inline" yaml:",inline"`
	ElectraPreset   `json:",inline" yaml:",inline"`
	Config          `json:",inline" yaml:",inline"`

	ExecutionEngine `json:"-" yaml:"-"`
}

// Wraps the object to parametrize with given spec. JSON and YAML functionality is proxied to the inner value.
func (spec *Spec) Wrap(des SpecObj) SSZObj {
	return &specObj{spec, des}
}

func (spec *Spec) ForkVersion(slot Slot) Version {
	epoch := spec.SlotToEpoch(slot)
	if epoch < spec.ALTAIR_FORK_EPOCH {
		return spec.GENESIS_FORK_VERSION
	} else if epoch < spec.BELLATRIX_FORK_EPOCH {
		return spec.ALTAIR_FORK_VERSION
	} else if epoch < spec.CAPELLA_FORK_EPOCH {
		return spec.BELLATRIX_FORK_VERSION
	} else if epoch < spec.DENEB_FORK_EPOCH {
		return spec.DENEB_FORK_VERSION
	} else if epoch < spec.ELECTRA_FORK_EPOCH {
		return spec.ELECTRA_FORK_VERSION
	} else {
		return spec.FULU_FORK_VERSION
	}
}
