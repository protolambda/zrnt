//go:generate go run ../../../presets/cmd/main.go --presets-dir=../../../presets/configs --output-dir=../../presets/generated

package common

import (
	"encoding/json"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"gopkg.in/yaml.v3"
)

const TARGET_AGGREGATORS_PER_COMMITTEE = 16
const RANDOM_SUBNETS_PER_VALIDATOR = 1
const EPOCHS_PER_RANDOM_SUBNET_SUBSCRIPTION = 256
const BLS_WITHDRAWAL_PREFIX = 0

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

// Sharding
var DOMAIN_SHARD_PROPOSER = BLSDomainType{0x80, 0x00, 0x00, 0x00}
var DOMAIN_SHARD_COMMITTEE = BLSDomainType{0x81, 0x00, 0x00, 0x00}

type Phase0Preset struct {
	// Misc.
	MAX_COMMITTEES_PER_SLOT      uint64 `yaml:"MAX_COMMITTEES_PER_SLOT" json:"MAX_COMMITTEES_PER_SLOT"`
	TARGET_COMMITTEE_SIZE        uint64 `yaml:"TARGET_COMMITTEE_SIZE" json:"TARGET_COMMITTEE_SIZE"`
	MAX_VALIDATORS_PER_COMMITTEE uint64 `yaml:"MAX_VALIDATORS_PER_COMMITTEE" json:"MAX_VALIDATORS_PER_COMMITTEE"`
	SHUFFLE_ROUND_COUNT          uint8  `yaml:"SHUFFLE_ROUND_COUNT" json:"SHUFFLE_ROUND_COUNT"`

	// Balance math
	HYSTERESIS_QUOTIENT            uint64 `yaml:"HYSTERESIS_QUOTIENT" json:"HYSTERESIS_QUOTIENT"`
	HYSTERESIS_DOWNWARD_MULTIPLIER uint64 `yaml:"HYSTERESIS_DOWNWARD_MULTIPLIER" json:"HYSTERESIS_DOWNWARD_MULTIPLIER"`
	HYSTERESIS_UPWARD_MULTIPLIER   uint64 `yaml:"HYSTERESIS_UPWARD_MULTIPLIER" json:"HYSTERESIS_UPWARD_MULTIPLIER"`

	// Fork Choice
	SAFE_SLOTS_TO_UPDATE_JUSTIFIED uint64 `yaml:"SAFE_SLOTS_TO_UPDATE_JUSTIFIED" json:"SAFE_SLOTS_TO_UPDATE_JUSTIFIED"`

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
	EPOCHS_PER_HISTORICAL_VECTOR Epoch  `yaml:"EPOCHS_PER_HISTORICAL_VECTOR" json:"EPOCHS_PER_HISTORICAL_VECTOR"`
	EPOCHS_PER_SLASHINGS_VECTOR  Epoch  `yaml:"EPOCHS_PER_SLASHINGS_VECTOR" json:"EPOCHS_PER_SLASHINGS_VECTOR"`
	HISTORICAL_ROOTS_LIMIT       uint64 `yaml:"HISTORICAL_ROOTS_LIMIT" json:"HISTORICAL_ROOTS_LIMIT"`
	VALIDATOR_REGISTRY_LIMIT     uint64 `yaml:"VALIDATOR_REGISTRY_LIMIT" json:"VALIDATOR_REGISTRY_LIMIT"`

	// Reward and penalty quotients
	BASE_REWARD_FACTOR               uint64 `yaml:"BASE_REWARD_FACTOR" json:"BASE_REWARD_FACTOR"`
	WHISTLEBLOWER_REWARD_QUOTIENT    uint64 `yaml:"WHISTLEBLOWER_REWARD_QUOTIENT" json:"WHISTLEBLOWER_REWARD_QUOTIENT"`
	PROPOSER_REWARD_QUOTIENT         uint64 `yaml:"PROPOSER_REWARD_QUOTIENT" json:"PROPOSER_REWARD_QUOTIENT"`
	INACTIVITY_PENALTY_QUOTIENT      uint64 `yaml:"INACTIVITY_PENALTY_QUOTIENT" json:"INACTIVITY_PENALTY_QUOTIENT"`
	MIN_SLASHING_PENALTY_QUOTIENT    uint64 `yaml:"MIN_SLASHING_PENALTY_QUOTIENT" json:"MIN_SLASHING_PENALTY_QUOTIENT"`
	PROPORTIONAL_SLASHING_MULTIPLIER uint64 `yaml:"PROPORTIONAL_SLASHING_MULTIPLIER" json:"PROPORTIONAL_SLASHING_MULTIPLIER"`

	// Max operations per block
	MAX_PROPOSER_SLASHINGS uint64 `yaml:"MAX_PROPOSER_SLASHINGS" json:"MAX_PROPOSER_SLASHINGS"`
	MAX_ATTESTER_SLASHINGS uint64 `yaml:"MAX_ATTESTER_SLASHINGS" json:"MAX_ATTESTER_SLASHINGS"`
	MAX_ATTESTATIONS       uint64 `yaml:"MAX_ATTESTATIONS" json:"MAX_ATTESTATIONS"`
	MAX_DEPOSITS           uint64 `yaml:"MAX_DEPOSITS" json:"MAX_DEPOSITS"`
	MAX_VOLUNTARY_EXITS    uint64 `yaml:"MAX_VOLUNTARY_EXITS" json:"MAX_VOLUNTARY_EXITS"`
}

type AltairPreset struct {
	// Updated penalty values
	INACTIVITY_PENALTY_QUOTIENT_ALTAIR      uint64 `yaml:"INACTIVITY_PENALTY_QUOTIENT_ALTAIR" json:"INACTIVITY_PENALTY_QUOTIENT_ALTAIR"`
	MIN_SLASHING_PENALTY_QUOTIENT_ALTAIR    uint64 `yaml:"MIN_SLASHING_PENALTY_QUOTIENT_ALTAIR" json:"MIN_SLASHING_PENALTY_QUOTIENT_ALTAIR"`
	PROPORTIONAL_SLASHING_MULTIPLIER_ALTAIR uint64 `yaml:"PROPORTIONAL_SLASHING_MULTIPLIER_ALTAIR" json:"PROPORTIONAL_SLASHING_MULTIPLIER_ALTAIR"`

	// Sync committee
	SYNC_COMMITTEE_SIZE              uint64 `yaml:"SYNC_COMMITTEE_SIZE" json:"SYNC_COMMITTEE_SIZE"`
	EPOCHS_PER_SYNC_COMMITTEE_PERIOD Epoch  `yaml:"EPOCHS_PER_SYNC_COMMITTEE_PERIOD" json:"EPOCHS_PER_SYNC_COMMITTEE_PERIOD"`

	// Sync committees and light clients
	MIN_SYNC_COMMITTEE_PARTICIPANTS uint64 `yaml:"MIN_SYNC_COMMITTEE_PARTICIPANTS" json:"MIN_SYNC_COMMITTEE_PARTICIPANTS"`
}

type MergePreset struct {
}

type ShardingPreset struct {
	// Misc.
	MAX_SHARDS                      uint64 `yaml:"MAX_SHARDS" json:"MAX_SHARDS"`
	INITIAL_ACTIVE_SHARDS           uint64 `yaml:"INITIAL_ACTIVE_SHARDS" json:"INITIAL_ACTIVE_SHARDS"`
	GASPRICE_ADJUSTMENT_COEFFICIENT uint64 `yaml:"GASPRICE_ADJUSTMENT_COEFFICIENT" json:"GASPRICE_ADJUSTMENT_COEFFICIENT"`
	MAX_SHARD_PROPOSER_SLASHINGS    uint64 `yaml:"MAX_SHARD_PROPOSER_SLASHINGS" json:"MAX_SHARD_PROPOSER_SLASHINGS"`

	// Shard block configs
	MAX_SHARD_HEADERS_PER_SHARD uint64 `yaml:"MAX_SHARD_HEADERS_PER_SHARD" json:"MAX_SHARD_HEADERS_PER_SHARD"`
	SHARD_STATE_MEMORY_SLOTS    Slot   `yaml:"SHARD_STATE_MEMORY_SLOTS" json:"SHARD_STATE_MEMORY_SLOTS"`
	MAX_SAMPLES_PER_BLOCK       uint64 `yaml:"MAX_SAMPLES_PER_BLOCK" json:"MAX_SAMPLES_PER_BLOCK"`
	TARGET_SAMPLES_PER_BLOCK    uint64 `yaml:"TARGET_SAMPLES_PER_BLOCK" json:"TARGET_SAMPLES_PER_BLOCK"`

	// Gwei values
	MAX_GASPRICE Gwei `yaml:"MAX_GASPRICE" json:"MAX_GASPRICE"`
	MIN_GASPRICE Gwei `yaml:"MIN_GASPRICE" json:"MIN_GASPRICE"`
}

type Config struct {
	PRESET_BASE string `yaml:"PRESET_BASE" json:"PRESET_BASE"`

	// Genesis.
	MIN_GENESIS_ACTIVE_VALIDATOR_COUNT uint64    `yaml:"MIN_GENESIS_ACTIVE_VALIDATOR_COUNT" json:"MIN_GENESIS_ACTIVE_VALIDATOR_COUNT"`
	MIN_GENESIS_TIME                   Timestamp `yaml:"MIN_GENESIS_TIME" json:"MIN_GENESIS_TIME"`
	GENESIS_FORK_VERSION               Version   `yaml:"GENESIS_FORK_VERSION" json:"GENESIS_FORK_VERSION"`
	GENESIS_DELAY                      Timestamp `yaml:"GENESIS_DELAY" json:"GENESIS_DELAY"`

	// Altair
	ALTAIR_FORK_VERSION Version `yaml:"ALTAIR_FORK_VERSION" json:"ALTAIR_FORK_VERSION"`
	ALTAIR_FORK_EPOCH   Epoch   `yaml:"ALTAIR_FORK_EPOCH" json:"ALTAIR_FORK_EPOCH"`

	// Merge
	MERGE_FORK_VERSION Version `yaml:"MERGE_FORK_VERSION" json:"MERGE_FORK_VERSION"`
	MERGE_FORK_EPOCH   Epoch   `yaml:"MERGE_FORK_EPOCH" json:"MERGE_FORK_EPOCH"`

	// Sharding
	SHARDING_FORK_VERSION Version `yaml:"SHARDING_FORK_VERSION" json:"SHARDING_FORK_VERSION"`
	SHARDING_FORK_EPOCH   Epoch   `yaml:"SHARDING_FORK_EPOCH" json:"SHARDING_FORK_EPOCH"`

	// Merge transition
	MIN_ANCHOR_POW_BLOCK_DIFFICULTY uint64 `yaml:"MIN_ANCHOR_POW_BLOCK_DIFFICULTY" json:"MIN_ANCHOR_POW_BLOCK_DIFFICULTY"`

	// Time parameters
	SECONDS_PER_SLOT                    Timestamp `yaml:"SECONDS_PER_SLOT" json:"SECONDS_PER_SLOT"`
	SECONDS_PER_ETH1_BLOCK              uint64    `yaml:"SECONDS_PER_ETH1_BLOCK" json:"SECONDS_PER_ETH1_BLOCK"`
	MIN_VALIDATOR_WITHDRAWABILITY_DELAY Epoch     `yaml:"MIN_VALIDATOR_WITHDRAWABILITY_DELAY" json:"MIN_VALIDATOR_WITHDRAWABILITY_DELAY"`
	SHARD_COMMITTEE_PERIOD              Epoch     `yaml:"SHARD_COMMITTEE_PERIOD" json:"SHARD_COMMITTEE_PERIOD"`
	ETH1_FOLLOW_DISTANCE                uint64    `yaml:"ETH1_FOLLOW_DISTANCE" json:"ETH1_FOLLOW_DISTANCE"`

	// Validator cycle
	INACTIVITY_SCORE_BIAS          uint64 `yaml:"INACTIVITY_SCORE_BIAS" json:"INACTIVITY_SCORE_BIAS"`
	INACTIVITY_SCORE_RECOVERY_RATE uint64 `yaml:"INACTIVITY_SCORE_RECOVERY_RATE" json:"INACTIVITY_SCORE_RECOVERY_RATE"`
	EJECTION_BALANCE               Gwei   `yaml:"EJECTION_BALANCE" json:"EJECTION_BALANCE"`
	MIN_PER_EPOCH_CHURN_LIMIT      uint64 `yaml:"MIN_PER_EPOCH_CHURN_LIMIT" json:"MIN_PER_EPOCH_CHURN_LIMIT"`
	CHURN_LIMIT_QUOTIENT           uint64 `yaml:"CHURN_LIMIT_QUOTIENT" json:"CHURN_LIMIT_QUOTIENT"`

	// Deposit contract
	DEPOSIT_CHAIN_ID         uint64      `yaml:"DEPOSIT_CHAIN_ID" json:"DEPOSIT_CHAIN_ID"`
	DEPOSIT_NETWORK_ID       uint64      `yaml:"DEPOSIT_NETWORK_ID" json:"DEPOSIT_NETWORK_ID"`
	DEPOSIT_CONTRACT_ADDRESS Eth1Address `yaml:"DEPOSIT_CONTRACT_ADDRESS" json:"DEPOSIT_CONTRACT_ADDRESS"`
}

// TODO
//type TrustedSetup struct {
//	G1_SETUP []BLSPubkey `yaml:"G1_SETUP" json:"G1_SETUP"`
//	G2_SETUP []BLSSignature `yaml:"G2_SETUP" json:"G2_SETUP"`
//}

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
	Phase0Preset   `yaml:",inline"`
	AltairPreset   `yaml:",inline"`
	MergePreset    `yaml:",inline"`
	ShardingPreset `yaml:",inline"`
	Config         `yaml:",inline"`
	//TrustedSetup   `yaml:",inline"`

	// Experimental, for merge purposes
	ExecutionEngine `yaml:"-"`
}

// Wraps the object to parametrize with given spec. JSON and YAML functionality is proxied to the inner value.
func (spec *Spec) Wrap(des SpecObj) SSZObj {
	return &specObj{spec, des}
}

func (spec *Spec) ForkVersion(slot Slot) Version {
	epoch := spec.SlotToEpoch(slot)
	if epoch < spec.ALTAIR_FORK_EPOCH {
		return spec.GENESIS_FORK_VERSION
	} else if epoch < spec.MERGE_FORK_EPOCH {
		return spec.ALTAIR_FORK_VERSION
	} else if epoch < spec.SHARDING_FORK_EPOCH {
		return spec.MERGE_FORK_VERSION
	} else {
		return spec.SHARDING_FORK_VERSION
	}
}

func (spec *Spec) ActiveShardCount(epoch Epoch) uint64 {
	// TODO: this may become more dynamic, based on state, fork, etc.
	return spec.INITIAL_ACTIVE_SHARDS
}

func (spec *Spec) ComputeUpdatedGasPrice(prevGasPrice Gwei, shardBlockLength uint64, adjustmentQuotient uint64) Gwei {
	if shardBlockLength > spec.TARGET_SAMPLES_PER_BLOCK {
		delta := Gwei(uint64(prevGasPrice) * (shardBlockLength - spec.TARGET_SAMPLES_PER_BLOCK) /
			spec.TARGET_SAMPLES_PER_BLOCK / adjustmentQuotient)
		if delta < 1 {
			delta = 1
		}
		out := prevGasPrice + delta
		if out > spec.MAX_GASPRICE {
			out = spec.MAX_GASPRICE
		}
		return out
	} else {
		delta := Gwei(uint64(prevGasPrice) * (spec.TARGET_SAMPLES_PER_BLOCK - shardBlockLength) /
			spec.TARGET_SAMPLES_PER_BLOCK / adjustmentQuotient)
		if delta < 1 {
			delta = 1
		}
		out := spec.MIN_GASPRICE + delta
		if out < prevGasPrice {
			out = prevGasPrice
		}
		out -= delta
		return out
	}
}
