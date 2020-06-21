// +build !preset_mainnet,!preset_minimal

package generated

const PRESET_NAME string = "mainnet"

const MAX_COMMITTEES_PER_SLOT = 64

const TARGET_COMMITTEE_SIZE = 128

const MAX_VALIDATORS_PER_COMMITTEE = 2048

const MIN_PER_EPOCH_CHURN_LIMIT = 4

const CHURN_LIMIT_QUOTIENT = 65536

const SHUFFLE_ROUND_COUNT = 90

const MIN_GENESIS_ACTIVE_VALIDATOR_COUNT = 16384

const MIN_GENESIS_TIME = 1578009600

const HYSTERESIS_QUOTIENT = 4

const HYSTERESIS_DOWNWARD_MULTIPLIER = 1

const HYSTERESIS_UPWARD_MULTIPLIER = 5

const SAFE_SLOTS_TO_UPDATE_JUSTIFIED = 8

const ETH1_FOLLOW_DISTANCE = 1024

const TARGET_AGGREGATORS_PER_COMMITTEE = 16

const RANDOM_SUBNETS_PER_VALIDATOR = 1

const EPOCHS_PER_RANDOM_SUBNET_SUBSCRIPTION = 256

const SECONDS_PER_ETH1_BLOCK = 14

var DEPOSIT_CONTRACT_ADDRESS = [40]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90}

const MIN_DEPOSIT_AMOUNT = 1000000000

const MAX_EFFECTIVE_BALANCE = 32000000000

const EJECTION_BALANCE = 16000000000

const EFFECTIVE_BALANCE_INCREMENT = 1000000000

const GENESIS_FORK_VERSION = 0

const BLS_WITHDRAWAL_PREFIX = 0

const GENESIS_DELAY = 172800

const SECONDS_PER_SLOT = 12

const MIN_ATTESTATION_INCLUSION_DELAY = 1

const SLOTS_PER_EPOCH = 32

const MIN_SEED_LOOKAHEAD = 1

const MAX_SEED_LOOKAHEAD = 4

const EPOCHS_PER_ETH1_VOTING_PERIOD = 32

const SLOTS_PER_HISTORICAL_ROOT = 8192

const MIN_VALIDATOR_WITHDRAWABILITY_DELAY = 256

const SHARD_COMMITTEE_PERIOD = 256

const MAX_EPOCHS_PER_CROSSLINK = 64

const MIN_EPOCHS_TO_INACTIVITY_PENALTY = 4

const EPOCHS_PER_HISTORICAL_VECTOR = 65536

const EPOCHS_PER_SLASHINGS_VECTOR = 8192

const HISTORICAL_ROOTS_LIMIT = 16777216

const VALIDATOR_REGISTRY_LIMIT = 1099511627776

const BASE_REWARD_FACTOR = 64

const WHISTLEBLOWER_REWARD_QUOTIENT = 512

const PROPOSER_REWARD_QUOTIENT = 8

const INACTIVITY_PENALTY_QUOTIENT = 16777216

const MIN_SLASHING_PENALTY_QUOTIENT = 32

const MAX_PROPOSER_SLASHINGS = 16

const MAX_ATTESTER_SLASHINGS = 2

const MAX_ATTESTATIONS = 128

const MAX_DEPOSITS = 16

const MAX_VOLUNTARY_EXITS = 16

const DOMAIN_BEACON_PROPOSER = 0x00000000

const DOMAIN_BEACON_ATTESTER = 0x01000000

const DOMAIN_RANDAO = 0x02000000

const DOMAIN_DEPOSIT = 0x03000000

const DOMAIN_VOLUNTARY_EXIT = 0x04000000

const DOMAIN_SELECTION_PROOF = 0x05000000

const DOMAIN_AGGREGATE_AND_PROOF = 0x06000000

const DOMAIN_SHARD_PROPOSAL = 0x80000000

const DOMAIN_SHARD_COMMITTEE = 0x81000000

const DOMAIN_LIGHT_CLIENT = 0x82000000

const DOMAIN_CUSTODY_BIT_SLASHING = 0x83000000

const PHASE_1_FORK_VERSION = 16777216

const PHASE_1_GENESIS_SLOT = 32

const INITIAL_ACTIVE_SHARDS = 64

const MAX_SHARDS = 1024

const ONLINE_PERIOD = 8

const LIGHT_CLIENT_COMMITTEE_SIZE = 128

const LIGHT_CLIENT_COMMITTEE_PERIOD = 256

const SHARD_BLOCK_CHUNK_SIZE = 262144

const MAX_SHARD_BLOCK_CHUNKS = 4

const TARGET_SHARD_BLOCK_SIZE = 196608

// const SHARD_BLOCK_OFFSETS = (unrecognized type) [1 2 3 5 8 13 21 34 55 89 144 233]

const MAX_SHARD_BLOCKS_PER_ATTESTATION = 12

const MAX_GASPRICE = 16384

const MIN_GASPRICE = 8

const GASPRICE_ADJUSTMENT_COEFFICIENT = 8

const RANDAO_PENALTY_EPOCHS = 2

const EARLY_DERIVED_SECRET_PENALTY_MAX_FUTURE_EPOCHS = 16384

const EPOCHS_PER_CUSTODY_PERIOD = 2048

const CUSTODY_PERIOD_TO_RANDAO_PADDING = 2048

const MAX_REVEAL_LATENESS_DECREMENT = 128

const MAX_CUSTODY_KEY_REVEALS = 256

const MAX_EARLY_DERIVED_SECRET_REVEALS = 1

const MAX_CUSTODY_SLASHINGS = 1

const EARLY_DERIVED_SECRET_REVEAL_SLOT_REWARD_MULTIPLE = 2

const MINOR_REWARD_QUOTIENT = 256