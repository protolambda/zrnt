package beacon

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/util/bitfield"
	"github.com/protolambda/zrnt/eth2/util/hash"
	"github.com/protolambda/zrnt/eth2/util/math"
	"sort"
)

type BeaconState struct {
	// Misc
	Slot        Slot
	GenesisTime Timestamp
	Fork        Fork

	// Validator registry
	ValidatorRegistry            ValidatorRegistry
	ValidatorBalances            ValidatorBalances
	ValidatorRegistryUpdateEpoch Epoch

	// Randomness and committees
	LatestRandaoMixes           [LATEST_RANDAO_MIXES_LENGTH]Bytes32
	PreviousShufflingStartShard Shard
	CurrentShufflingStartShard  Shard
	PreviousShufflingEpoch      Epoch
	CurrentShufflingEpoch       Epoch
	PreviousShufflingSeed       Bytes32
	CurrentShufflingSeed        Bytes32

	// Finality
	PreviousEpochAttestations []PendingAttestation
	CurrentEpochAttestations  []PendingAttestation
	PreviousJustifiedEpoch    Epoch
	CurrentJustifiedEpoch     Epoch
	PreviousJustifiedRoot     Root
	CurrentJustifiedRoot      Root
	JustificationBitfield     uint64
	FinalizedEpoch            Epoch
	FinalizedRoot             Root

	// Recent state
	LatestCrosslinks       [SHARD_COUNT]Crosslink
	LatestBlockRoots       [SLOTS_PER_HISTORICAL_ROOT]Root
	LatestStateRoots       [SLOTS_PER_HISTORICAL_ROOT]Root
	LatestActiveIndexRoots [LATEST_ACTIVE_INDEX_ROOTS_LENGTH]Root
	// Balances slashed at every withdrawal period
	LatestSlashedBalances [LATEST_SLASHED_EXIT_LENGTH]Gwei
	LatestBlockHeader     BeaconBlockHeader
	HistoricalRoots       []Root

	// Ethereum 1.0 chain data
	LatestEth1Data Eth1Data
	Eth1DataVotes  []Eth1DataVote
	DepositIndex   DepositIndex
}

// Make a deep copy of the state object
func (state *BeaconState) Copy() *BeaconState {
	// copy over state
	stUn := *state
	res := &stUn
	// manually copy over slices, and efficiently (i.e. explicitly make, but don't initially zero out, just overwrite)
	// validators
	res.ValidatorRegistry = append(make([]Validator, 0, len(state.ValidatorRegistry)), state.ValidatorRegistry...)
	res.ValidatorBalances = append(make([]Gwei, 0, len(state.ValidatorBalances)), state.ValidatorBalances...)
	// finality
	res.PreviousEpochAttestations = append(make([]PendingAttestation, 0, len(state.PreviousEpochAttestations)), state.PreviousEpochAttestations...)
	res.CurrentEpochAttestations = append(make([]PendingAttestation, 0, len(state.CurrentEpochAttestations)), state.CurrentEpochAttestations...)
	// recent state
	res.HistoricalRoots = append(make([]Root, 0, len(state.HistoricalRoots)), state.HistoricalRoots...)
	// eth1
	res.Eth1DataVotes = append(make([]Eth1DataVote, 0, len(state.Eth1DataVotes)), state.Eth1DataVotes...)
	return res
}

// Get current epoch
func (state *BeaconState) Epoch() Epoch {
	return state.Slot.ToEpoch()
}

// Return previous epoch.
func (state *BeaconState) PreviousEpoch() Epoch {
	return state.Epoch() - 1
}

// Set the validator with the given index as withdrawable
// MIN_VALIDATOR_WITHDRAWABILITY_DELAY after the current epoch.
func (state *BeaconState) PrepareValidatorForWithdrawal(index ValidatorIndex) {
	state.ValidatorRegistry[index].WithdrawableEpoch = state.Epoch() + MIN_VALIDATOR_WITHDRAWABILITY_DELAY
}

// Initiate the validator of the given index
func (state *BeaconState) InitiateValidatorExit(index ValidatorIndex) {
	state.ValidatorRegistry[index].InitiatedExit = true
}

// Activate the validator of the given index
func (state *BeaconState) ActivateValidator(index ValidatorIndex, isGenesis bool) {
	validator := &state.ValidatorRegistry[index]

	if isGenesis {
		validator.ActivationEpoch = GENESIS_EPOCH
	} else {
		validator.ActivationEpoch = state.Epoch().GetDelayedActivationExitEpoch()
	}
}

func GetEpochCommitteeCount(activeValidatorCount uint64) uint64 {
	return math.MaxU64(1,
		math.MinU64(
			uint64(SHARD_COUNT)/uint64(SLOTS_PER_EPOCH),
			activeValidatorCount/uint64(SLOTS_PER_EPOCH)/TARGET_COMMITTEE_SIZE,
		)) * uint64(SLOTS_PER_EPOCH)
}

// Return the number of committees in the previous epoch
func (state *BeaconState) GetPreviousEpochCommitteeCount() uint64 {
	return GetEpochCommitteeCount(
		state.ValidatorRegistry.GetActiveValidatorCount(
			state.PreviousShufflingEpoch,
		))
}

// Return the number of committees in the current epoch
func (state *BeaconState) GetCurrentEpochCommitteeCount() uint64 {
	return GetEpochCommitteeCount(
		state.ValidatorRegistry.GetActiveValidatorCount(
			state.CurrentShufflingEpoch,
		))
}

// Return the number of committees in the next epoch
func (state *BeaconState) GetNextEpochCommitteeCount() uint64 {
	return GetEpochCommitteeCount(
		state.ValidatorRegistry.GetActiveValidatorCount(
			state.Epoch() + 1,
		))
}

// Return the beacon proposer index for the slot.
func (state *BeaconState) GetBeaconProposerIndex(slot Slot, registryChange bool) ValidatorIndex {
	epoch := slot.ToEpoch()
	currentEpoch := state.Epoch()
	if !(currentEpoch-1 <= epoch && epoch <= currentEpoch+1) {
		panic("epoch of given slot out of range")
	}
	committeeData := state.GetCrosslinkCommitteesAtSlot(slot, registryChange)
	firstCommitteeData := committeeData[0]
	return firstCommitteeData.Committee[epoch%Epoch(len(firstCommitteeData.Committee))]
}

//  Return the randao mix at a recent epoch
func (state *BeaconState) GetRandaoMix(epoch Epoch) Bytes32 {
	// Every usage is a trusted input (i.e. state is already up to date to handle the requested epoch number).
	// If something is wrong due to unforeseen usage, panic to catch it during development.
	if !(state.Epoch()-LATEST_RANDAO_MIXES_LENGTH < epoch && epoch <= state.Epoch()) {
		panic("cannot get randao mix for out-of-bounds epoch")
	}
	return state.LatestRandaoMixes[epoch%LATEST_RANDAO_MIXES_LENGTH]
}

func (state *BeaconState) GetActiveIndexRoot(epoch Epoch) Root {
	return state.LatestActiveIndexRoots[epoch%LATEST_ACTIVE_INDEX_ROOTS_LENGTH]
}

// Generate a seed for the given epoch
func (state *BeaconState) GenerateSeed(epoch Epoch) Bytes32 {
	buf := make([]byte, 32*3)
	mix := state.GetRandaoMix(epoch - MIN_SEED_LOOKAHEAD)
	copy(buf[0:32], mix[:])
	// get_active_index_root in spec, but only used once, and the assertion is unnecessary, since epoch input is always trusted
	activeIndexRoot := state.GetActiveIndexRoot(epoch)
	copy(buf[32:64], activeIndexRoot[:])
	binary.LittleEndian.PutUint64(buf[64:], uint64(epoch))
	return hash.Hash(buf)
}

// Return the block root at a recent slot
func (state *BeaconState) GetBlockRoot(slot Slot) (Root, error) {
	if slot+SLOTS_PER_HISTORICAL_ROOT < state.Slot || slot > state.Slot {
		return Root{}, errors.New("cannot get block root for given slot")
	}
	return state.LatestBlockRoots[slot%SLOTS_PER_HISTORICAL_ROOT], nil
}

// Return the state root at a recent
func (state *BeaconState) GetStateRoot(slot Slot) (Root, error) {
	if slot+SLOTS_PER_HISTORICAL_ROOT < state.Slot || slot > state.Slot {
		return Root{}, errors.New("cannot get state root for given slot")
	}
	return state.LatestStateRoots[slot%SLOTS_PER_HISTORICAL_ROOT], nil
}

type CrosslinkCommittee struct {
	Committee []ValidatorIndex
	Shard     Shard
}

// Return the list of (committee, shard) tuples for the slot.
//
// Note: There are two possible shufflings for crosslink committees for a
//  slot in the next epoch -- with and without a registryChange
func (state *BeaconState) GetCrosslinkCommitteesAtSlot(slot Slot, registryChange bool) []CrosslinkCommittee {
	epoch, currentEpoch, previousEpoch := slot.ToEpoch(), state.Epoch(), state.PreviousEpoch()
	nextEpoch := currentEpoch + 1

	if !(previousEpoch <= epoch && epoch <= nextEpoch) {
		panic("could not retrieve crosslink committee for out of range slot")
	}

	var committeesPerEpoch uint64
	var seed Bytes32
	var shufflingEpoch Epoch
	var shufflingStartShard Shard
	if epoch == currentEpoch {
		committeesPerEpoch = state.GetCurrentEpochCommitteeCount()
		seed = state.CurrentShufflingSeed
		shufflingEpoch = state.CurrentShufflingEpoch
		shufflingStartShard = state.CurrentShufflingStartShard
	} else if epoch == previousEpoch {
		committeesPerEpoch = state.GetPreviousEpochCommitteeCount()
		seed = state.PreviousShufflingSeed
		shufflingEpoch = state.PreviousShufflingEpoch
		shufflingStartShard = state.PreviousShufflingStartShard
	} else if epoch == nextEpoch {
		epochsSinceLastRegistryUpdate := currentEpoch - state.ValidatorRegistryUpdateEpoch
		if registryChange {
			committeesPerEpoch = state.GetNextEpochCommitteeCount()
			seed = state.GenerateSeed(nextEpoch)
			shufflingEpoch = nextEpoch
			currentCommitteesPerEpoch := state.GetCurrentEpochCommitteeCount()
			shufflingStartShard = (state.CurrentShufflingStartShard + Shard(currentCommitteesPerEpoch)) % SHARD_COUNT
		} else if epochsSinceLastRegistryUpdate > 1 && math.IsPowerOfTwo(uint64(epochsSinceLastRegistryUpdate)) {
			committeesPerEpoch = state.GetNextEpochCommitteeCount()
			seed = state.GenerateSeed(nextEpoch)
			shufflingEpoch = nextEpoch
			shufflingStartShard = state.CurrentShufflingStartShard
		} else {
			committeesPerEpoch = state.GetCurrentEpochCommitteeCount()
			seed = state.CurrentShufflingSeed
			shufflingEpoch = state.CurrentShufflingEpoch
			shufflingStartShard = state.CurrentShufflingStartShard
		}
	}
	// TODO: this shuffling could be cached
	shuffling := state.ValidatorRegistry.GetShuffling(seed, shufflingEpoch)
	offset := slot % SLOTS_PER_EPOCH
	committeesPerSlot := committeesPerEpoch / uint64(SLOTS_PER_EPOCH)
	slotStartShard := (shufflingStartShard + Shard(committeesPerSlot)*Shard(offset)) % SHARD_COUNT

	crosslinkCommittees := make([]CrosslinkCommittee, committeesPerSlot)
	for i := uint64(0); i < committeesPerSlot; i++ {
		crosslinkCommittees[i] = CrosslinkCommittee{
			Committee: shuffling[committeesPerSlot*uint64(offset)+i],
			Shard:     (slotStartShard + Shard(i)) % SHARD_COUNT}
	}
	return crosslinkCommittees
}

func (state *BeaconState) GetWinningRootAndParticipants(shard Shard) (Root, []ValidatorIndex) {
	weightedCrosslinks := make(map[Root]Gwei)

	updateCrosslinkWeights := func(att *PendingAttestation) {
		if att.Data.PreviousCrosslink == state.LatestCrosslinks[shard] {
			participants, _ := state.GetAttestationParticipants(&att.Data, &att.AggregationBitfield)
			for _, participant := range participants {
				weightedCrosslinks[att.Data.CrosslinkDataRoot] += state.ValidatorBalances.GetEffectiveBalance(participant)
			}
		}
	}
	for i := 0; i < len(state.PreviousEpochAttestations); i++ {
		updateCrosslinkWeights(&state.PreviousEpochAttestations[i])
	}
	for i := 0; i < len(state.CurrentEpochAttestations); i++ {
		updateCrosslinkWeights(&state.CurrentEpochAttestations[i])
	}

	// handle when no attestations for shard available
	if len(weightedCrosslinks) == 0 {
		return Root{}, nil
	}
	// Now determine the best root, by total weight (votes, weighted by balance)
	var winningRoot Root
	winningWeight := Gwei(0)
	for root, weight := range weightedCrosslinks {
		if weight > winningWeight {
			winningRoot = root
		}
		if weight == winningWeight {
			// break tie lexicographically
			for i := 0; i < 32; i++ {
				if root[i] > winningRoot[i] {
					winningRoot = root
					break
				}
			}
		}
	}

	// now retrieve all the attesters of this winning root
	winningAttestersSet := make(map[ValidatorIndex]struct{})
	findWinners := func(att *PendingAttestation) {
		if att.Data.CrosslinkDataRoot == winningRoot {
			participants, _ := state.GetAttestationParticipants(&att.Data, &att.AggregationBitfield)
			for _, participant := range participants {
				winningAttestersSet[participant] = struct{}{}
			}
		}
	}
	for i := 0; i < len(state.PreviousEpochAttestations); i++ {
		findWinners(&state.PreviousEpochAttestations[i])
	}
	for i := 0; i < len(state.CurrentEpochAttestations); i++ {
		findWinners(&state.CurrentEpochAttestations[i])
	}
	winningAttesters := make([]ValidatorIndex, len(winningAttestersSet))
	i := 0
	for attester := range winningAttestersSet {
		winningAttesters[i] = attester
		i++
	}
	// Spec returns it in sorted order, although not strictly necessary (TODO)
	sort.Slice(winningAttesters, func(i int, j int) bool {
		return winningAttesters[i] < winningAttesters[j]
	})

	return winningRoot, winningAttesters
}

// Exit the validator of the given index
func (state *BeaconState) ExitValidator(index ValidatorIndex) {
	validator := &state.ValidatorRegistry[index]
	delayedActivationExitEpoch := state.Epoch().GetDelayedActivationExitEpoch()
	// The following updates only occur if not previous exited
	if validator.ExitEpoch > delayedActivationExitEpoch {
		return
	}
	validator.ExitEpoch = delayedActivationExitEpoch
}

// Update validator registry.
func (state *BeaconState) UpdateValidatorRegistry() {
	// The total effective balance of active validators
	totalBalance := state.ValidatorBalances.GetTotalBalance(state.ValidatorRegistry.GetActiveValidatorIndices(state.Epoch()))

	// The maximum balance churn in Gwei (for deposits and exits separately)
	maxBalanceChurn := Max(MAX_DEPOSIT_AMOUNT, totalBalance/(2*MAX_BALANCE_CHURN_QUOTIENT))

	// Activate validators within the allowable balance churn
	{
		balanceChurn := Gwei(0)
		for index, validator := range state.ValidatorRegistry {
			if validator.ActivationEpoch == FAR_FUTURE_EPOCH && state.ValidatorBalances[index] >= MAX_DEPOSIT_AMOUNT {
				// Check the balance churn would be within the allowance
				balanceChurn += state.ValidatorBalances.GetEffectiveBalance(ValidatorIndex(index))
				if balanceChurn > maxBalanceChurn {
					break
				}
				//  Activate validator
				validator.ActivationEpoch = state.Epoch().GetDelayedActivationExitEpoch()
			}
		}
	}

	// Exit validators within the allowable balance churn
	{
		balanceChurn := Gwei(0)
		for index, validator := range state.ValidatorRegistry {
			if validator.ExitEpoch == FAR_FUTURE_EPOCH && validator.InitiatedExit {
				// Check the balance churn would be within the allowance
				balanceChurn += state.ValidatorBalances.GetEffectiveBalance(ValidatorIndex(index))
				if balanceChurn > maxBalanceChurn {
					break
				}
				// Exit validator
				state.ExitValidator(ValidatorIndex(index))
			}
		}
	}
}

// Return the participant indices at for the attestation_data and bitfield
func (state *BeaconState) GetAttestationParticipants(attestationData *AttestationData, bitfield *bitfield.Bitfield) ([]ValidatorIndex, error) {
	// Find the committee in the list with the desired shard
	crosslinkCommittees := state.GetCrosslinkCommitteesAtSlot(attestationData.Slot, false)

	var crosslinkCommittee []ValidatorIndex
	for _, crossComm := range crosslinkCommittees {
		if crossComm.Shard == attestationData.Shard {
			crosslinkCommittee = crossComm.Committee
			break
		}
	}
	if len(crosslinkCommittee) == 0 {
		return nil, errors.New(fmt.Sprintf("cannot find crosslink committee at slot %d for shard %d", attestationData.Slot, attestationData.Shard))
	}
	if !bitfield.VerifySize(uint64(len(crosslinkCommittee))) {
		return nil, errors.New("bitfield has wrong size for corresponding crosslink committee")
	}

	// Find the participating attesters in the committee
	participants := make([]ValidatorIndex, 0)
	for i, vIndex := range crosslinkCommittee {
		if bitfield.GetBit(uint64(i)) == 1 {
			participants = append(participants, vIndex)
		}
	}
	return participants, nil
}

// Slash the validator with index index.
func (state *BeaconState) SlashValidator(index ValidatorIndex) error {
	validator := &state.ValidatorRegistry[index]
	// [TO BE REMOVED IN PHASE 2] (this is to make phase 0 and phase 1 consistent with behavior in phase 2)
	if state.Slot >= validator.WithdrawableEpoch.GetStartSlot() {
		return errors.New("cannot slash validator after withdrawal epoch")
	}
	state.ExitValidator(index)
	state.LatestSlashedBalances[state.Epoch()%LATEST_SLASHED_EXIT_LENGTH] += state.ValidatorBalances.GetEffectiveBalance(index)

	whistleblowerReward := state.ValidatorBalances.GetEffectiveBalance(index) / WHISTLEBLOWER_REWARD_QUOTIENT
	propIndex := state.GetBeaconProposerIndex(state.Slot, false)
	state.ValidatorBalances[propIndex] += whistleblowerReward
	state.ValidatorBalances[index] -= whistleblowerReward
	validator.Slashed = true
	validator.WithdrawableEpoch = state.Epoch() + LATEST_SLASHED_EXIT_LENGTH
	return nil
}
