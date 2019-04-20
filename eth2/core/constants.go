//go:generate go run ../../building/yaml_constants/main.go --presets-dir=../../../eth2.0-specs/configs/constant_presets --output-dir=../../constant_presets/

package core

import (
	"github.com/protolambda/zrnt/constant_presets"
)

// Misc.
const SHARD_COUNT Shard = constant_presets.SHARD_COUNT
const TARGET_COMMITTEE_SIZE = constant_presets.TARGET_COMMITTEE_SIZE
const MAX_ATTESTATION_PARTICIPANTS = constant_presets.MAX_ATTESTATION_PARTICIPANTS
const MIN_PER_EPOCH_CHURN_LIMIT = constant_presets.MIN_PER_EPOCH_CHURN_LIMIT
const CHURN_LIMIT_QUOTIENT = constant_presets.CHURN_LIMIT_QUOTIENT
const SHUFFLE_ROUND_COUNT uint8 = constant_presets.SHUFFLE_ROUND_COUNT

// Deposit contract
const DEPOSIT_CONTRACT_TREE_DEPTH = constant_presets.DEPOSIT_CONTRACT_TREE_DEPTH

// Gwei values
const MIN_DEPOSIT_AMOUNT Gwei = constant_presets.MIN_DEPOSIT_AMOUNT
const MAX_DEPOSIT_AMOUNT Gwei = constant_presets.MAX_DEPOSIT_AMOUNT
// unused const FORK_CHOICE_BALANCE_INCREMENT Gwei = constant_presets.FORK_CHOICE_BALANCE_INCREMENT
const EJECTION_BALANCE Gwei = constant_presets.EJECTION_BALANCE

const HIGH_BALANCE_INCREMENT Gwei = constant_presets.HIGH_BALANCE_INCREMENT
const HALF_INCREMENT = constant_presets.HIGH_BALANCE_INCREMENT / 2

// Initial values
const GENESIS_SLOT Slot = constant_presets.GENESIS_SLOT
const GENESIS_START_SHARD Shard = constant_presets.GENESIS_START_SHARD
const GENESIS_FORK_VERSION uint32 = constant_presets.GENESIS_FORK_VERSION
const GENESIS_EPOCH = Epoch(GENESIS_SLOT / SLOTS_PER_EPOCH)

const FAR_FUTURE_EPOCH Epoch = constant_presets.FAR_FUTURE_EPOCH

const BLS_WITHDRAWAL_PREFIX_BYTE byte = constant_presets.BLS_WITHDRAWAL_PREFIX_BYTE

// Time parameters
const SECONDS_PER_SLOT Timestamp = constant_presets.SECONDS_PER_SLOT
const MIN_ATTESTATION_INCLUSION_DELAY Slot = constant_presets.MIN_ATTESTATION_INCLUSION_DELAY
const SLOTS_PER_EPOCH Slot = constant_presets.SLOTS_PER_EPOCH
const MIN_SEED_LOOKAHEAD Epoch = constant_presets.MIN_SEED_LOOKAHEAD
const ACTIVATION_EXIT_DELAY Epoch = constant_presets.ACTIVATION_EXIT_DELAY
const SLOTS_PER_ETH1_VOTING_PERIOD Slot = constant_presets.SLOTS_PER_ETH1_VOTING_PERIOD
const MIN_VALIDATOR_WITHDRAWABILITY_DELAY Epoch = constant_presets.MIN_VALIDATOR_WITHDRAWABILITY_DELAY
const PERSISTENT_COMMITTEE_PERIOD Epoch = constant_presets.PERSISTENT_COMMITTEE_PERIOD
const MAX_CROSSLINK_EPOCHS Epoch = constant_presets.MAX_CROSSLINK_EPOCHS

// State list lengths
const SLOTS_PER_HISTORICAL_ROOT Slot = constant_presets.SLOTS_PER_HISTORICAL_ROOT
const LATEST_RANDAO_MIXES_LENGTH Epoch = constant_presets.LATEST_RANDAO_MIXES_LENGTH
const LATEST_ACTIVE_INDEX_ROOTS_LENGTH Epoch = constant_presets.LATEST_ACTIVE_INDEX_ROOTS_LENGTH
const LATEST_SLASHED_EXIT_LENGTH Epoch = constant_presets.LATEST_SLASHED_EXIT_LENGTH

// Reward and penalty quotients
const BASE_REWARD_QUOTIENT = constant_presets.BASE_REWARD_QUOTIENT
const WHISTLEBLOWING_REWARD_QUOTIENT = constant_presets.WHISTLEBLOWING_REWARD_QUOTIENT
const PROPOSER_REWARD_QUOTIENT = constant_presets.PROPOSER_REWARD_QUOTIENT
const INACTIVITY_PENALTY_QUOTIENT = constant_presets.INACTIVITY_PENALTY_QUOTIENT
const MIN_PENALTY_QUOTIENT = constant_presets.MIN_PENALTY_QUOTIENT

// Max transactions per block
const MAX_PROPOSER_SLASHINGS = constant_presets.MAX_PROPOSER_SLASHINGS
const MAX_ATTESTER_SLASHINGS = constant_presets.MAX_ATTESTER_SLASHINGS
const MAX_ATTESTATIONS = constant_presets.MAX_ATTESTATIONS
const MAX_DEPOSITS = constant_presets.MAX_DEPOSITS
const MAX_VOLUNTARY_EXITS = constant_presets.MAX_VOLUNTARY_EXITS
const MAX_TRANSFERS = constant_presets.MAX_TRANSFERS

// Signature domains
const (
	DOMAIN_BEACON_BLOCK BLSDomainType = constant_presets.DOMAIN_BEACON_BLOCK
	DOMAIN_RANDAO BLSDomainType = constant_presets.DOMAIN_RANDAO
	DOMAIN_ATTESTATION BLSDomainType = constant_presets.DOMAIN_ATTESTATION
	DOMAIN_DEPOSIT BLSDomainType = constant_presets.DOMAIN_DEPOSIT
	DOMAIN_VOLUNTARY_EXIT BLSDomainType = constant_presets.DOMAIN_VOLUNTARY_EXIT
	DOMAIN_TRANSFER BLSDomainType = constant_presets.DOMAIN_TRANSFER
)
