package beacon

// Misc.
const SHARD_COUNT Shard = 1 << 10         // =  1,024
const TARGET_COMMITTEE_SIZE = 1 << 7      // =  128
const MAX_BALANCE_CHURN_QUOTIENT = 1 << 5 // =  32
const BEACON_CHAIN_SHARD_NUMBER Shard = 1<<64 - 1
const MAX_INDICES_PER_SLASHABLE_VOTE = 1 << 12 // =  4,096
const MAX_EXIT_DEQUEUES_PER_EPOCH = 1 << 2     // =  4
const SHUFFLE_ROUND_COUNT = 90

// Deposit contract
const DEPOSIT_CONTRACT_TREE_DEPTH = 1 << 5 // =  32

// Gwei values
const MIN_DEPOSIT_AMOUNT Gwei = (1 << 0) * 1000000000 // =  1,000,000,000  Gwei
const MAX_DEPOSIT_AMOUNT Gwei = (1 << 5) * 1000000000 // =  32,000,000,000  Gwei
// unused const FORK_CHOICE_BALANCE_INCREMENT Gwei = (1 << 0) * MILLION_GWEI // =  1,000,000,000  Gwei
const EJECTION_BALANCE Gwei = (1 << 4) * 1000000000 // =  16,000,000,000  Gwei

// Initial values
const GENESIS_FORK_VERSION uint64 = 0
const GENESIS_SLOT Slot = 1 << 32
const GENESIS_EPOCH = Epoch(GENESIS_SLOT / SLOTS_PER_EPOCH)

// unused const GENESIS_START_SHARD Shard = 0
const FAR_FUTURE_EPOCH Epoch = 1<<64 - 1

const BLS_WITHDRAWAL_PREFIX_BYTE byte = 0

// Time parameters
// unused const SECONDS_PER_SLOT Seconds = 6                       //  seconds  6 seconds
const MIN_ATTESTATION_INCLUSION_DELAY Slot = 1 << 2      // =  4  slots  24 seconds
const SLOTS_PER_EPOCH Slot = 1 << 6                      // =  64  slots  6.4 minutes
const MIN_SEED_LOOKAHEAD Epoch = 1 << 0                  // =  1  epochs  6.4 minutes
const ACTIVATION_EXIT_DELAY Epoch = 1 << 2               // =  4  epochs  25.6 minutes
const EPOCHS_PER_ETH1_VOTING_PERIOD Epoch = 1 << 4       // =  16  epochs  ~1.7 hours
const MIN_VALIDATOR_WITHDRAWABILITY_DELAY Epoch = 1 << 8 // =  256  epochs  ~27 hours

// State list lengths
const SLOTS_PER_HISTORICAL_ROOT Slot = 1 << 13         // =  8,192  slots  ~13 hours
const LATEST_RANDAO_MIXES_LENGTH Epoch = 1 << 13       // =  8,192  epochs  ~36 days
const LATEST_ACTIVE_INDEX_ROOTS_LENGTH Epoch = 1 << 13 // =  8,192  epochs  ~36 days
const LATEST_SLASHED_EXIT_LENGTH Epoch = 1 << 13       // =  8,192  epochs  ~36 days

// Reward and penalty quotients
const BASE_REWARD_QUOTIENT = 1 << 5                  // =  32
const WHISTLEBLOWER_REWARD_QUOTIENT = 1 << 9         // =  512
const ATTESTATION_INCLUSION_REWARD_QUOTIENT = 1 << 3 // =  8
const INACTIVITY_PENALTY_QUOTIENT = 1 << 24          // =  16,777,216
const MIN_PENALTY_QUOTIENT = 1 << 5                  // =  32

// Max transactions per block
const MAX_PROPOSER_SLASHINGS = 1 << 4 // =  16
const MAX_ATTESTER_SLASHINGS = 1 << 0 // =  1
const MAX_ATTESTATIONS = 1 << 7       // =  128
const MAX_DEPOSITS = 1 << 4           // =  16
const MAX_VOLUNTARY_EXITS = 1 << 4    // =  16
const MAX_TRANSFERS = 1 << 4          // =  16

// Signature domains
const (
	DOMAIN_BEACON_BLOCK BLSDomain = iota
	DOMAIN_RANDAO
	DOMAIN_ATTESTATION
	DOMAIN_DEPOSIT
	DOMAIN_VOLUNTARY_EXIT
	DOMAIN_TRANSFER
)

// Custom constant, not in spec: An impossible validator index used to mark special internal cases. (all 1s binary)
const ValidatorIndexMarker = ValidatorIndex(^uint64(0))
