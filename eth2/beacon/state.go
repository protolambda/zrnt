package beacon

import (
	"encoding/binary"
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bitfield"
	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/zrnt/eth2/util/shuffling"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	"sort"
)

var BeaconStateSSZ = zssz.GetSSZ((*BeaconState)(nil))

type BeaconState struct {
	// Misc
	Slot        Slot
	GenesisTime Timestamp
	Fork        Fork

	// Validator registry
	ValidatorRegistry ValidatorRegistry
	Balances          []Gwei

	// Randomness and committees
	LatestRandaoMixes [LATEST_RANDAO_MIXES_LENGTH]Root
	LatestStartShard  Shard

	// Finality
	PreviousEpochAttestations []*PendingAttestation
	CurrentEpochAttestations  []*PendingAttestation
	PreviousJustifiedEpoch    Epoch
	CurrentJustifiedEpoch     Epoch
	PreviousJustifiedRoot     Root
	CurrentJustifiedRoot      Root
	JustificationBitfield     uint64
	FinalizedEpoch            Epoch
	FinalizedRoot             Root

	// Recent state
	CurrentCrosslinks      [SHARD_COUNT]Crosslink
	PreviousCrosslinks     [SHARD_COUNT]Crosslink
	LatestBlockRoots       [SLOTS_PER_HISTORICAL_ROOT]Root
	LatestStateRoots       [SLOTS_PER_HISTORICAL_ROOT]Root
	LatestActiveIndexRoots [LATEST_ACTIVE_INDEX_ROOTS_LENGTH]Root
	// Balances slashed at every withdrawal period
	LatestSlashedBalances [LATEST_SLASHED_EXIT_LENGTH]Gwei
	LatestBlockHeader     BeaconBlockHeader
	HistoricalRoots       []Root

	// Ethereum 1.0 chain data
	LatestEth1Data Eth1Data
	Eth1DataVotes  []Eth1Data
	DepositIndex   DepositIndex
}

// Make a deep copy of the state object
func (state *BeaconState) Copy() *BeaconState {
	// copy over state
	stUn := *state
	res := &stUn
	// manually copy over slices, and efficiently (i.e. explicitly make, but don't initially zero out, just overwrite)
	// validators
	res.ValidatorRegistry = make([]*Validator, 0, len(state.ValidatorRegistry))
	for _, v := range state.ValidatorRegistry {
		res.ValidatorRegistry = append(res.ValidatorRegistry, v.Copy())
	}
	res.Balances = append(make([]Gwei, 0, len(state.Balances)), state.Balances...)
	// finality
	res.PreviousEpochAttestations = append(make([]*PendingAttestation, 0, len(state.PreviousEpochAttestations)), state.PreviousEpochAttestations...)
	res.CurrentEpochAttestations = append(make([]*PendingAttestation, 0, len(state.CurrentEpochAttestations)), state.CurrentEpochAttestations...)
	// recent state
	res.HistoricalRoots = append(make([]Root, 0, len(state.HistoricalRoots)), state.HistoricalRoots...)
	// eth1
	res.Eth1DataVotes = append(make([]Eth1Data, 0, len(state.Eth1DataVotes)), state.Eth1DataVotes...)
	return res
}

// Get current epoch
func (state *BeaconState) Epoch() Epoch {
	return state.Slot.ToEpoch()
}

// Return previous epoch.
func (state *BeaconState) PreviousEpoch() Epoch {
	currentEpoch := state.Epoch()
	if currentEpoch == GENESIS_EPOCH {
		return GENESIS_EPOCH
	} else {
		return currentEpoch - 1
	}
}

func (state *BeaconState) GetChurnLimit() uint64 {
	return math.MaxU64(MIN_PER_EPOCH_CHURN_LIMIT,
		state.ValidatorRegistry.GetActiveValidatorCount(state.Epoch())/CHURN_LIMIT_QUOTIENT)
}

// Initiate the exit of the validator of the given index
func (state *BeaconState) InitiateValidatorExit(index ValidatorIndex) {
	validator := state.ValidatorRegistry[index]
	// Return if validator already initiated exit
	if validator.ExitEpoch != FAR_FUTURE_EPOCH {
		return
	}
	// Compute exit queue epoch
	exitQueueEnd := state.Epoch().GetDelayedActivationExitEpoch()
	for _, v := range state.ValidatorRegistry {
		if v.ExitEpoch != FAR_FUTURE_EPOCH && v.ExitEpoch > exitQueueEnd {
			exitQueueEnd = v.ExitEpoch
		}
	}
	exitQueueChurn := uint64(0)
	for _, v := range state.ValidatorRegistry {
		if v.ExitEpoch == exitQueueEnd {
			exitQueueChurn++
		}
	}
	if exitQueueChurn >= state.GetChurnLimit() {
		exitQueueEnd++
	}

	// Set validator exit epoch and withdrawable epoch
	validator.ExitEpoch = exitQueueEnd
	validator.WithdrawableEpoch = validator.ExitEpoch + MIN_VALIDATOR_WITHDRAWABILITY_DELAY
}

func (state *BeaconState) GetEpochStartShard(epoch Epoch) Shard {
	currentEpoch := state.Epoch()
	checkEpoch := currentEpoch + 1
	if epoch > checkEpoch {
		panic("cannot find start shard for epoch, epoch is too new")
	}
	shard := (state.LatestStartShard + state.GetShardDelta(currentEpoch)) % SHARD_COUNT
	for checkEpoch > epoch {
		checkEpoch--
		shard = (shard + SHARD_COUNT - state.GetShardDelta(checkEpoch)) % SHARD_COUNT
	}
	return shard
}

// TODO: spec should refer to AttestationData here instead of Attestation
func (state *BeaconState) GetAttestationSlot(attData *AttestationData) Slot {
	epoch := attData.TargetEpoch
	committeeCount := Slot(state.GetEpochCommitteeCount(epoch))
	offset := Slot((attData.Crosslink.Shard + SHARD_COUNT - state.GetEpochStartShard(epoch)) % SHARD_COUNT)
	return epoch.GetStartSlot() + (offset / (committeeCount / SLOTS_PER_EPOCH))
}

// Return the number of committees in one epoch.
func (state *BeaconState) GetEpochCommitteeCount(epoch Epoch) uint64 {
	activeValidatorCount := state.ValidatorRegistry.GetActiveValidatorCount(epoch)
	return math.MaxU64(1,
		math.MinU64(
			uint64(SHARD_COUNT)/uint64(SLOTS_PER_EPOCH),
			activeValidatorCount/uint64(SLOTS_PER_EPOCH)/TARGET_COMMITTEE_SIZE,
		)) * uint64(SLOTS_PER_EPOCH)
}

// Return the number of shards to increment state.latest_start_shard during epoch
func (state *BeaconState) GetShardDelta(epoch Epoch) Shard {
	return Shard(math.MinU64(
		state.GetEpochCommitteeCount(epoch),
		uint64(SHARD_COUNT-(SHARD_COUNT/Shard(SLOTS_PER_EPOCH)))))
}

// Return the beacon proposer index for the current slot
func (state *BeaconState) GetBeaconProposerIndex() ValidatorIndex {
	epoch := state.Epoch()
	committeesPerSlot := state.GetEpochCommitteeCount(epoch) / uint64(SLOTS_PER_EPOCH)
	offset := Shard(committeesPerSlot) * Shard(state.Slot%SLOTS_PER_EPOCH)
	shard := (state.GetEpochStartShard(epoch) + offset) % SHARD_COUNT
	firstCommittee := state.GetCrosslinkCommittee(epoch, shard)
	seed := state.GenerateSeed(epoch)
	buf := make([]byte, 32+8, 32+8)
	copy(buf[0:32], seed[:])
	for i := uint64(0); true; i++ {
		binary.LittleEndian.PutUint64(buf[32:], i)
		h := Hash(buf)
		for j := uint64(0); j < 32; j++ {
			randomByte := h[j]
			candidateIndex := firstCommittee[(uint64(epoch)+((i<<5)|j))%uint64(len(firstCommittee))]
			effectiveBalance := state.ValidatorRegistry[candidateIndex].EffectiveBalance
			if effectiveBalance*0xff >= MAX_EFFECTIVE_BALANCE*Gwei(randomByte) {
				return candidateIndex
			}
		}
	}
	return 0
}

func (state *BeaconState) GetRandaoMix(epoch Epoch) Root {
	// Epoch is expected to be between (current_epoch - LATEST_RANDAO_MIXES_LENGTH, current_epoch].
	// TODO: spec has expectations on input, but doesn't enforce them, and purposefully ignores them in some calls.
	return state.LatestRandaoMixes[epoch%LATEST_RANDAO_MIXES_LENGTH]
}

func (state *BeaconState) GetActiveIndexRoot(epoch Epoch) Root {
	// Epoch is expected to be between (current_epoch - LATEST_ACTIVE_INDEX_ROOTS_LENGTH + ACTIVATION_EXIT_DELAY, current_epoch + ACTIVATION_EXIT_DELAY].
	// TODO: spec has expectations on input, but doesn't enforce them, and purposefully ignores them in some calls.
	return state.LatestActiveIndexRoots[epoch%LATEST_ACTIVE_INDEX_ROOTS_LENGTH]
}

// Generate a seed for the given epoch
func (state *BeaconState) GenerateSeed(epoch Epoch) Root {
	buf := make([]byte, 32*3)
	mix := state.GetRandaoMix(epoch + LATEST_RANDAO_MIXES_LENGTH - MIN_SEED_LOOKAHEAD)
	copy(buf[0:32], mix[:])
	// get_active_index_root in spec, but only used once, and the assertion is unnecessary, since epoch input is always trusted
	activeIndexRoot := state.GetActiveIndexRoot(epoch)
	copy(buf[32:64], activeIndexRoot[:])
	binary.LittleEndian.PutUint64(buf[64:], uint64(epoch))
	return Hash(buf)
}

// Return the block root at the given slot (a recent one)
func (state *BeaconState) GetBlockRootAtSlot(slot Slot) (Root, error) {
	if !(slot < state.Slot && slot+SLOTS_PER_HISTORICAL_ROOT <= state.Slot) {
		return Root{}, errors.New("cannot get block root for given slot")
	}
	return state.LatestBlockRoots[slot%SLOTS_PER_HISTORICAL_ROOT], nil
}

// Return the block root at a recent epoch
func (state *BeaconState) GetBlockRoot(epoch Epoch) (Root, error) {
	return state.GetBlockRootAtSlot(epoch.GetStartSlot())
}

// Optimized compared to spec: takes pre-shuffled active indices as input, to not shuffle per-committee.
func computeCommittee(shuffled []ValidatorIndex, index uint64, count uint64) []ValidatorIndex {
	// Return the index'th shuffled committee out of the total committees data (shuffled active indices)
	startOffset := (uint64(len(shuffled)) * index) / count
	endOffset := (uint64(len(shuffled)) * (index + 1)) / count
	return shuffled[startOffset:endOffset]
}

func (state *BeaconState) GetCrosslinkCommittee(epoch Epoch, shard Shard) []ValidatorIndex {
	currentEpoch := state.Epoch()
	previousEpoch := state.PreviousEpoch()
	nextEpoch := currentEpoch + 1

	if !(previousEpoch <= epoch && epoch <= nextEpoch) {
		panic("could not retrieve crosslink committee for out of range slot")
	}

	seed := state.GenerateSeed(epoch)
	activeIndices := state.ValidatorRegistry.GetActiveValidatorIndices(epoch)
	// Active validators, shuffled in-place.
	// TODO: cache shuffling
	shuffling.UnshuffleList(activeIndices, seed)
	index := uint64((shard + SHARD_COUNT - state.GetEpochStartShard(epoch)) % SHARD_COUNT)
	count := state.GetEpochCommitteeCount(epoch)
	return computeCommittee(activeIndices, index, count)
}

func (state *BeaconState) GetAttesters(attestations []*PendingAttestation, filter func(att *AttestationData) bool) ValidatorSet {
	out := make(ValidatorSet, 0)
	for _, att := range attestations {
		// If the attestation is for the boundary:
		if filter(&att.Data) {
			participants, _ := state.GetAttestingIndicesUnsorted(&att.Data, &att.AggregationBitfield)
			out = append(out, participants...)
		}
	}
	out.Dedup()
	return out
}

func (state *BeaconState) GetWinningCrosslinkAndAttestingIndices(shard Shard, epoch Epoch) (*Crosslink, ValidatorSet) {
	pendingAttestations := state.PreviousEpochAttestations
	if epoch == state.Epoch() {
		pendingAttestations = state.CurrentEpochAttestations
	}

	latestCrosslinkRoot := ssz.HashTreeRoot(&state.CurrentCrosslinks[shard], CrosslinkSSZ)

	// keyed by raw crosslink object. Not too big, and simplifies reduction to unique crosslinks
	crosslinkAttesters := make(map[*Crosslink]ValidatorSet)
	for _, att := range pendingAttestations {
		if att.Data.Crosslink.Shard == shard {
			if att.Data.Crosslink.ParentRoot == latestCrosslinkRoot ||
				latestCrosslinkRoot == ssz.HashTreeRoot(&att.Data.Crosslink, CrosslinkSSZ) {
				participants, _ := state.GetAttestingIndices(&att.Data, &att.AggregationBitfield)
				crosslinkAttesters[&att.Data.Crosslink] = append(crosslinkAttesters[&att.Data.Crosslink], participants...)
			}
		}
	}
	// handle when no attestations for shard available
	if len(crosslinkAttesters) == 0 {
		return &Crosslink{}, nil
	}
	for k, v := range crosslinkAttesters {
		v.Dedup()
		crosslinkAttesters[k] = state.FilterUnslashed(v)
	}

	// Now determine the best crosslink, by total weight (votes, weighted by balance)
	var winningLink *Crosslink = nil
	winningWeight := Gwei(0)
	for crosslink, attesters := range crosslinkAttesters {
		// effectively "get_attesting_balance": attesters consists of only de-duplicated unslashed validators.
		weight := state.GetTotalBalanceOf(attesters)
		if winningLink == nil || weight > winningWeight {
			winningLink = crosslink
		}
		if winningLink != nil && weight == winningWeight {
			// break tie lexicographically
			for i := 0; i < 32; i++ {
				if crosslink.DataRoot[i] > winningLink.DataRoot[i] {
					winningLink = crosslink
					break
				}
			}
		}
	}

	// now retrieve all the attesters of this winning root
	winners := crosslinkAttesters[winningLink]

	return winningLink, winners
}

// Exit the validator with the given index
func (state *BeaconState) ExitValidator(index ValidatorIndex) {
	validator := state.ValidatorRegistry[index]
	// Update validator exit epoch if not previously exited
	if validator.ExitEpoch == FAR_FUTURE_EPOCH {
		validator.ExitEpoch = state.Epoch().GetDelayedActivationExitEpoch()
	}
}

// Return the sorted attesting indices at for the attestation_data and bitfield
func (state *BeaconState) GetAttestingIndicesUnsorted(attestationData *AttestationData, bitfield *bitfield.Bitfield) ([]ValidatorIndex, error) {
	// Find the committee in the list with the desired shard
	crosslinkCommittee := state.GetCrosslinkCommittee(attestationData.TargetEpoch, attestationData.Crosslink.Shard)

	if len(crosslinkCommittee) == 0 {
		return nil, fmt.Errorf("cannot find crosslink committee at target epoch %d for shard %d", attestationData.TargetEpoch, attestationData.Crosslink.Shard)
	}
	if !bitfield.VerifySize(uint64(len(crosslinkCommittee))) {
		return nil, errors.New("bitfield has wrong size for corresponding crosslink committee")
	}

	// Find the participating attesters in the committee
	participants := make([]ValidatorIndex, 0, len(crosslinkCommittee))
	for i, vIndex := range crosslinkCommittee {
		if bitfield.GetBit(uint64(i)) == 1 {
			participants = append(participants, vIndex)
		}
	}
	return participants, nil
}

// Return the sorted attesting indices at for the attestation_data and bitfield
func (state *BeaconState) GetAttestingIndices(attestationData *AttestationData, bitfield *bitfield.Bitfield) (ValidatorSet, error) {
	participants, err := state.GetAttestingIndicesUnsorted(attestationData, bitfield)
	if err != nil {
		return nil, err
	}
	out := ValidatorSet(participants)
	sort.Sort(out)
	return out, nil
}

// Slash the validator with the given index.
func (state *BeaconState) SlashValidator(slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex) error {
	currentEpoch := state.Epoch()
	validator := state.ValidatorRegistry[slashedIndex]
	state.InitiateValidatorExit(slashedIndex)
	validator.Slashed = true
	validator.WithdrawableEpoch = currentEpoch + LATEST_SLASHED_EXIT_LENGTH
	slashedBalance := validator.EffectiveBalance
	state.LatestSlashedBalances[currentEpoch%LATEST_SLASHED_EXIT_LENGTH] += slashedBalance

	propIndex := state.GetBeaconProposerIndex()
	if whistleblowerIndex == nil {
		whistleblowerIndex = &propIndex
	}
	whistleblowerReward := slashedBalance / WHISTLEBLOWING_REWARD_QUOTIENT
	proposerReward := whistleblowerReward / PROPOSER_REWARD_QUOTIENT
	state.IncreaseBalance(propIndex, proposerReward)
	state.IncreaseBalance(*whistleblowerIndex, whistleblowerReward-proposerReward)
	state.DecreaseBalance(slashedIndex, whistleblowerReward)
	return nil
}

// Filters a slice in-place. Only keeps the unslashed validators.
// If input is sorted, then the result will be sorted.
func (state *BeaconState) FilterUnslashed(indices []ValidatorIndex) []ValidatorIndex {
	unslashed := indices[:0]
	for _, x := range indices {
		if !state.ValidatorRegistry[x].Slashed {
			unslashed = append(unslashed, x)
		}
	}
	return unslashed
}

func (state *BeaconState) ApplyDeltas(deltas *Deltas) {
	if len(deltas.Penalties) != len(state.Balances) || len(deltas.Rewards) != len(state.Balances) {
		panic("cannot apply deltas to balances list with different length")
	}
	for i := ValidatorIndex(0); i < ValidatorIndex(len(state.Balances)); i++ {
		state.IncreaseBalance(i, deltas.Rewards[i])
		state.DecreaseBalance(i, deltas.Penalties[i])
	}
}

// Return the total balance sum (1 Gwei minimum to avoid divisions by zero.)
func (state *BeaconState) GetTotalActiveBalance() (sum Gwei) {
	epoch := state.Epoch()
	for _, v := range state.ValidatorRegistry {
		if v.IsActive(epoch) {
			sum += v.EffectiveBalance
		}
	}
	if sum == 0 {
		return 1
	}
	return sum
}

// Return the combined effective balance of an array of validators. (1 Gwei minimum to avoid divisions by zero.)
func (state *BeaconState) GetTotalBalanceOf(indices []ValidatorIndex) (sum Gwei) {
	for _, vIndex := range indices {
		sum += state.ValidatorRegistry[vIndex].EffectiveBalance
	}
	if sum == 0 {
		return 1
	}
	return sum
}

func (state *BeaconState) IncreaseBalance(index ValidatorIndex, delta Gwei) {
	state.Balances[index] += delta
}

func (state *BeaconState) DecreaseBalance(index ValidatorIndex, delta Gwei) {
	currentBalance := state.Balances[index]
	// prevent underflow, clip to 0
	if currentBalance >= delta {
		state.Balances[index] -= delta
	} else {
		state.Balances[index] = 0
	}
}

// Convert attestation to (almost) indexed-verifiable form
func (state *BeaconState) ConvertToIndexed(attestation *Attestation) (*IndexedAttestation, error) {
	if a, b := len(attestation.AggregationBitfield), len(attestation.CustodyBitfield); a != b {
		return nil, fmt.Errorf("aggregation bitfield does not match custody bitfield size: %d <> %d", a, b)
	} else if a > (MAX_INDICES_PER_ATTESTATION / 8) {
		return nil, fmt.Errorf("aggregation bitfield is too large: %d", a)
	}
	participants, err := state.GetAttestingIndices(&attestation.Data, &attestation.AggregationBitfield)
	if err != nil {
		return nil, errors.New("participants could not be derived from aggregation_bitfield")
	}
	custodyBit1Indices, err := state.GetAttestingIndices(&attestation.Data, &attestation.CustodyBitfield)
	if err != nil {
		return nil, errors.New("participants could not be derived from custody_bitfield")
	}
	if len(custodyBit1Indices) > len(participants) {
		return nil, fmt.Errorf("attestation has more custody bits set (%d) than participants allowed (%d)",
			len(custodyBit1Indices), len(participants))
	}
	// everyone who is a participant, and has not a custody bit set to 1, is part of the 0 custody bit indices.
	custodyBit0Indices := make([]ValidatorIndex, 0, len(participants)-len(custodyBit1Indices))
	participants.ZigZagJoin(custodyBit1Indices, nil, func(i ValidatorIndex) {
		custodyBit0Indices = append(custodyBit0Indices, i)
	})
	return &IndexedAttestation{
		CustodyBit0Indices: custodyBit0Indices,
		CustodyBit1Indices: custodyBit1Indices,
		Data:               attestation.Data,
		Signature:          attestation.Signature,
	}, nil
}

// Verify validity of slashable_attestation fields.
func (state *BeaconState) ValidateIndexedAttestation(indexedAttestation *IndexedAttestation) error {
	// wrap it in validator-sets. Does not sort it, but does make checking if it is a lot easier.
	bit0Indices := ValidatorSet(indexedAttestation.CustodyBit0Indices)
	bit1Indices := ValidatorSet(indexedAttestation.CustodyBit1Indices)

	// To be removed in Phase 1.
	if len(bit1Indices) != 0 {
		return errors.New("validators cannot have a custody bit set to 1 during phase 0")
	}

	// Verify max number of indices
	totalAttestingIndices := len(bit1Indices) + len(bit0Indices)
	if !(1 <= totalAttestingIndices && totalAttestingIndices <= MAX_INDICES_PER_ATTESTATION) {
		return fmt.Errorf("invalid indices count in indexed attestation: %d", totalAttestingIndices)
	}

	// The indices must be sorted
	if !sort.IsSorted(bit0Indices) {
		return errors.New("custody bit 0 indices are not sorted")
	}

	if !sort.IsSorted(bit1Indices) {
		return errors.New("custody bit 1 indices are not sorted")
	}

	// Verify index sets are disjoint
	if bit0Indices.Intersects(bit1Indices) {
		return errors.New("validator set for custody bit 1 intersects with validator set for custody bit 0")
	}

	// Check the last item of the sorted list to be a valid index,
	// if this one is valid, the others are as well.
	if len(bit0Indices) > 0 && !state.ValidatorRegistry.IsValidatorIndex(bit0Indices[len(bit0Indices)-1]) {
		return errors.New("index in custody bit 1 indices is invalid")
	}

	if len(bit1Indices) > 0 && !state.ValidatorRegistry.IsValidatorIndex(bit1Indices[len(bit1Indices)-1]) {
		return errors.New("index in custody bit 1 indices is invalid")
	}

	custodyBit0Pubkeys := make([]BLSPubkey, 0)
	for _, i := range bit0Indices {
		custodyBit0Pubkeys = append(custodyBit0Pubkeys, state.ValidatorRegistry[i].Pubkey)
	}
	custodyBit1Pubkeys := make([]BLSPubkey, 0)
	for _, i := range bit1Indices {
		custodyBit1Pubkeys = append(custodyBit1Pubkeys, state.ValidatorRegistry[i].Pubkey)
	}

	// don't trust, verify
	if bls.BlsVerifyMultiple(
		[]BLSPubkey{
			bls.BlsAggregatePubkeys(custodyBit0Pubkeys),
			bls.BlsAggregatePubkeys(custodyBit1Pubkeys)},
		[]Root{
			ssz.HashTreeRoot(&AttestationDataAndCustodyBit{Data: indexedAttestation.Data, CustodyBit: false}, AttestationDataAndCustodyBitSSZ),
			ssz.HashTreeRoot(&AttestationDataAndCustodyBit{Data: indexedAttestation.Data, CustodyBit: true}, AttestationDataAndCustodyBitSSZ)},
		indexedAttestation.Signature,
		state.GetDomain(DOMAIN_ATTESTATION, indexedAttestation.Data.TargetEpoch),
	) {
		return nil
	}

	return errors.New("could not verify BLS signature for indexed attestation")
}

// Return the signature domain (fork version concatenated with domain type) of a message.
func (state *BeaconState) GetDomain(dom BLSDomainType, messageEpoch Epoch) BLSDomain {
	v := state.Fork.CurrentVersion
	if messageEpoch < state.Fork.Epoch {
		v = state.Fork.PreviousVersion
	}
	// combine fork version with domain type.
	return BLSDomain((uint64(v.ToUint32()) << 32) | uint64(dom))
}
