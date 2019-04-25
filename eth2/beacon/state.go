package beacon

import (
	"encoding/binary"
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bitfield"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/hash"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/zrnt/eth2/util/shuffling"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"sort"
)

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
	for i, v := range state.ValidatorRegistry {
		res.ValidatorRegistry[i] = v.Copy()
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
	if currentEpoch > GENESIS_EPOCH {
		return currentEpoch - 1
	} else {
		return currentEpoch
	}
}

func (state *BeaconState) GetChurnLimit() uint64 {
	return math.MaxU64(MIN_PER_EPOCH_CHURN_LIMIT,
		state.ValidatorRegistry.GetActiveValidatorCount(state.Epoch())/CHURN_LIMIT_QUOTIENT)
}

// Initiate the validator of the given index
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
	currentEpoch := state.Epoch()
	firstCommittee := state.GetCrosslinkCommitteesAtSlot(state.Slot)[0].Committee
	seed := state.GenerateSeed(currentEpoch)
	buf := make([]byte, 32+8, 32+8)
	copy(buf[0:32], seed[:])
	for i := uint64(0); true; i++ {
		binary.LittleEndian.PutUint64(buf[32:], i)
		h := hash.Hash(buf)
		for j := uint64(0); j < 32; j++ {
			randomByte := h[j]
			candidate := firstCommittee[(uint64(currentEpoch)+(i<<5|j))%uint64(len(firstCommittee))]
			if state.GetEffectiveBalance(candidate)<<8 > MAX_EFFECTIVE_BALANCE*Gwei(randomByte) {
				return candidate
			}
		}
	}
	return 0
}

func (state *BeaconState) GetCrosslinkFromAttestationData(data *AttestationData) *Crosslink {
	epoch := state.CurrentCrosslinks[data.Shard].Epoch + MAX_CROSSLINK_EPOCHS
	if currentEpoch := data.Slot.ToEpoch(); currentEpoch < epoch {
		epoch = currentEpoch
	}
	return &Crosslink{
		epoch,
		data.PreviousCrosslinkRoot,
		data.CrosslinkDataRoot,
	}
}

//  Return the randao mix at a recent epoch
func (state *BeaconState) GetRandaoMix(epoch Epoch) Root {
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
func (state *BeaconState) GenerateSeed(epoch Epoch) Root {
	buf := make([]byte, 32*3)
	mix := state.GetRandaoMix(epoch - MIN_SEED_LOOKAHEAD)
	copy(buf[0:32], mix[:])
	// get_active_index_root in spec, but only used once, and the assertion is unnecessary, since epoch input is always trusted
	activeIndexRoot := state.GetActiveIndexRoot(epoch)
	copy(buf[32:64], activeIndexRoot[:])
	binary.LittleEndian.PutUint64(buf[64:], uint64(epoch))
	return hash.HashRoot(buf)
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

// Returns a value such that for a list L, chunk count k and index i,
//  split(L, k)[i] == L[get_split_offset(len(L), k, i): get_split_offset(len(L), k, i+1)]
func getSplitOffset(listSize uint64, chunks uint64, index uint64) uint64 {
	return (listSize * index) / chunks
}

// Return the list of (committee, shard) tuples for the slot.
func (state *BeaconState) GetCrosslinkCommitteesAtSlot(slot Slot) []CrosslinkCommittee {
	epoch := slot.ToEpoch()
	currentEpoch := state.Epoch()
	previousEpoch := state.PreviousEpoch()
	nextEpoch := currentEpoch + 1

	if !(previousEpoch <= epoch && epoch <= nextEpoch) {
		panic("could not retrieve crosslink committee for out of range slot")
	}

	var startShard Shard
	if epoch == currentEpoch {
		startShard = state.LatestStartShard
	} else if epoch == previousEpoch {
		previousShardDelta := state.GetShardDelta(previousEpoch)
		startShard = (state.LatestStartShard - previousShardDelta) % SHARD_COUNT
	} else if epoch == nextEpoch {
		currentShardDelta := state.GetShardDelta(currentEpoch)
		startShard = (state.LatestStartShard + currentShardDelta) % SHARD_COUNT
	}

	committeesPerEpoch := state.GetEpochCommitteeCount(epoch)
	committeesPerSlot := committeesPerEpoch / uint64(SLOTS_PER_EPOCH)
	offset := uint64(slot % SLOTS_PER_EPOCH)
	slotStartShard := (startShard + Shard(committeesPerSlot)*Shard(offset)) % SHARD_COUNT
	seed := state.GenerateSeed(epoch)

	crosslinkCommittees := make([]CrosslinkCommittee, committeesPerSlot)
	{
		shuffled := state.GetShuffled(seed, epoch)
		activeValidatorCount := state.ValidatorRegistry.GetActiveValidatorCount(epoch)

		// Return the index'th shuffled committee out of a total total_committees
		computeCommittee := func(index uint64) []ValidatorIndex {
			startOffset := getSplitOffset(activeValidatorCount, committeesPerEpoch, index)
			endOffset := getSplitOffset(activeValidatorCount, committeesPerEpoch, index)
			return shuffled[startOffset:endOffset]
		}

		for i := uint64(0); i < committeesPerSlot; i++ {
			crosslinkCommittees[i] = CrosslinkCommittee{
				Committee: computeCommittee(committeesPerSlot*offset + i),
				Shard:     (slotStartShard + Shard(i)) % SHARD_COUNT,
			}
		}
	}
	return crosslinkCommittees
}

func (state *BeaconState) GetAttesters(attestations []*PendingAttestation, filter func (att *AttestationData) bool) ValidatorSet {
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

func (state *BeaconState) GetWinningCrosslinkAndAttestingIndices(shard Shard, epoch Epoch) (Crosslink, ValidatorSet) {
	pendingAttestations := state.PreviousEpochAttestations
	if epoch == state.Epoch() {
		pendingAttestations = state.CurrentEpochAttestations
	}

	latestCrosslinkRoot := ssz.HashTreeRoot(state.CurrentCrosslinks[shard])

	// keyed by raw crosslink object. Not too big, and simplifies reduction to unique crosslinks
	crosslinkAttesters := make(map[Crosslink]ValidatorSet)
	for _, att := range pendingAttestations {
		if att.Data.Shard == shard {
			c := state.GetCrosslinkFromAttestationData(&att.Data)
			if c.PreviousCrosslinkRoot == latestCrosslinkRoot ||
				latestCrosslinkRoot == ssz.HashTreeRoot(c) {
				participants, _ := state.GetAttestingIndices(&att.Data, &att.AggregationBitfield)
				crosslinkAttesters[*c] = append(crosslinkAttesters[*c], participants...)
			}
		}
	}
	// handle when no attestations for shard available
	if len(crosslinkAttesters) == 0 {
		return Crosslink{Epoch: GENESIS_EPOCH}, nil
	}
	for k, v := range crosslinkAttesters {
		v.Dedup()
		crosslinkAttesters[k] = state.FilterUnslashed(v)
	}

	// Now determine the best crosslink, by total weight (votes, weighted by balance)
	winningLink := Crosslink{}
	winningWeight := Gwei(0)
	for crosslink, attesters := range crosslinkAttesters {
		// effectively "get_attesting_balance": attesters consists of only de-duplicated unslashed validators.
		weight := state.GetTotalBalanceOf(attesters)
		if weight > winningWeight {
			winningLink = crosslink
		}
		if weight == winningWeight {
			// break tie lexicographically
			for i := 0; i < 32; i++ {
				if crosslink.CrosslinkDataRoot[i] > winningLink.CrosslinkDataRoot[i] {
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
	crosslinkCommittees := state.GetCrosslinkCommitteesAtSlot(attestationData.Slot)
	crosslinkCommittee := crosslinkCommittees[(SHARD_COUNT-crosslinkCommittees[0].Shard+Shard(attestationData.Slot))%SHARD_COUNT].Committee

	if len(crosslinkCommittee) == 0 {
		return nil, errors.New(fmt.Sprintf("cannot find crosslink committee at slot %d for shard %d", attestationData.Slot, attestationData.Shard))
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
func (state *BeaconState) SlashValidator(slashedIndex ValidatorIndex) error {
	validator := state.ValidatorRegistry[slashedIndex]
	state.InitiateValidatorExit(slashedIndex)
	state.LatestSlashedBalances[state.Epoch()%LATEST_SLASHED_EXIT_LENGTH] += state.GetEffectiveBalance(index)

	propIndex := state.GetBeaconProposerIndex()
	whistleblowerReward := state.GetEffectiveBalance(slashedIndex) / WHISTLEBLOWING_REWARD_QUOTIENT
	state.IncreaseBalance(propIndex, whistleblowerReward)
	state.DecreaseBalance(slashedIndex, whistleblowerReward)
	validator.Slashed = true
	validator.WithdrawableEpoch = state.Epoch() + LATEST_SLASHED_EXIT_LENGTH
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

// Return the effective balance (also known as "balance at stake") for a validator with the given index.
func (state *BeaconState) GetEffectiveBalance(index ValidatorIndex) Gwei {
	return Min(state.GetBalance(index), MAX_EFFECTIVE_BALANCE)
}

// Return the total balance sum
func (state *BeaconState) GetTotalBalance() (sum Gwei) {
	for i := 0; i < len(state.Balances); i++ {
		sum += state.GetEffectiveBalance(ValidatorIndex(i))
	}
	return sum
}

// Return the combined effective balance of an array of validators.
func (state *BeaconState) GetTotalBalanceOf(indices []ValidatorIndex) (sum Gwei) {
	for _, vIndex := range indices {
		sum += state.GetEffectiveBalance(vIndex)
	}
	return sum
}

func (state *BeaconState) GetBalance(index ValidatorIndex) Gwei {
	return state.Balances[index]
}

// Set the balance for a validator with the given ``index`` in both ``BeaconState``
//  and validator's rounded balance ``high_balance``.
func (state *BeaconState) SetBalance(index ValidatorIndex, balance Gwei) {
	validator := state.ValidatorRegistry[index]
	if validator.HighBalance > balance || validator.HighBalance+3*HALF_INCREMENT < balance {
		validator.HighBalance = balance - (balance % HIGH_BALANCE_INCREMENT)
	}
	state.Balances[index] = balance
}

func (state *BeaconState) IncreaseBalance(index ValidatorIndex, delta Gwei) {
	state.SetBalance(index, state.GetBalance(index)+delta)
}

func (state *BeaconState) DecreaseBalance(index ValidatorIndex, delta Gwei) {
	currentBalance := state.GetBalance(index)
	// prevent underflow, clip to 0
	if currentBalance >= delta {
		state.SetBalance(index, currentBalance-delta)
	} else {
		state.SetBalance(index, 0)
	}
}

// Convert attestation to (almost) indexed-verifiable form
func (state *BeaconState) ConvertToIndexed(attestation *Attestation) (*IndexedAttestation, error) {
	participants, err := state.GetAttestingIndices(&attestation.Data, &attestation.AggregationBitfield)
	if err != nil {
		return nil, errors.New("participants could not be derived from aggregation_bitfield")
	}
	custodyBit1Indices, err := state.GetAttestingIndices(&attestation.Data, &attestation.CustodyBitfield)
	if err != nil {
		return nil, errors.New("participants could not be derived from custody_bitfield")
	}
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
func (state *BeaconState) VerifyIndexedAttestation(indexedAttestation *IndexedAttestation) error {
	// wrap it in validator-sets. Does not sort it, but does make checking if it is a lot easier.
	custodyBit0Indices := ValidatorSet(indexedAttestation.CustodyBit0Indices)
	custodyBit1Indices := ValidatorSet(indexedAttestation.CustodyBit1Indices)

	// Ensure no duplicate indices across custody bits
	if custodyBit0Indices.Intersects(custodyBit1Indices) {
		return errors.New("validator set for custody bit 1 intersects with validator set for custody bit 0")
	}

	// [TO BE REMOVED IN PHASE 1]
	if len(custodyBit1Indices) != 0 {
		return errors.New("validators cannot have a custody bit set to 1 during phase 0")
	}

	totalAttestingIndices := len(custodyBit1Indices) + len(custodyBit0Indices)
	if !(1 <= totalAttestingIndices && totalAttestingIndices <= MAX_INDICES_PER_ATTESTATION) {
		return errors.New(fmt.Sprintf("invalid indices count in indexed attestation: %d", totalAttestingIndices))
	}

	// The indices must be sorted
	if sort.IsSorted(custodyBit0Indices) {
		return errors.New("custody bit 0 indices are not sorted")
	}

	if sort.IsSorted(custodyBit1Indices) {
		return errors.New("custody bit 1 indices are not sorted")
	}

	// Check the last item of the sorted list to be a valid index,
	// if this one is valid, the others are as well.
	if !state.ValidatorRegistry.IsValidatorIndex(custodyBit0Indices[len(custodyBit0Indices)-1]) {
		return errors.New("index in custody bit 1 indices is invalid")
	}

	if !state.ValidatorRegistry.IsValidatorIndex(custodyBit0Indices[len(custodyBit0Indices)-1]) {
		return errors.New("index in custody bit 1 indices is invalid")
	}

	custodyBit0Pubkeys := make([]BLSPubkey, 0)
	for _, i := range custodyBit0Indices {
		custodyBit0Pubkeys = append(custodyBit0Pubkeys, state.ValidatorRegistry[i].Pubkey)
	}
	custodyBit1Pubkeys := make([]BLSPubkey, 0)
	for _, i := range custodyBit1Indices {
		custodyBit1Pubkeys = append(custodyBit1Pubkeys, state.ValidatorRegistry[i].Pubkey)
	}

	// don't trust, verify
	if bls.BlsVerifyMultiple(
		[]BLSPubkey{
			bls.BlsAggregatePubkeys(custodyBit0Pubkeys),
			bls.BlsAggregatePubkeys(custodyBit1Pubkeys)},
		[]Root{
			ssz.HashTreeRoot(AttestationDataAndCustodyBit{Data: indexedAttestation.Data, CustodyBit: false}),
			ssz.HashTreeRoot(AttestationDataAndCustodyBit{Data: indexedAttestation.Data, CustodyBit: true})},
		indexedAttestation.Signature,
		state.GetDomain(DOMAIN_ATTESTATION, indexedAttestation.Data.Slot.ToEpoch()),
	) {
		return nil
	} else {
		return errors.New("could not verify BLS signature for indexed attestation")
	}
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

// Shuffle active validators
func (state *BeaconState) GetShuffled(seed Root, epoch Epoch) []ValidatorIndex {
	activeValidatorIndices := state.ValidatorRegistry.GetActiveValidatorIndices(epoch)
	committeeCount := state.GetEpochCommitteeCount(epoch)
	if committeeCount > uint64(len(activeValidatorIndices)) {
		panic("not enough validators to form committees!")
	}
	indexList := make([]ValidatorIndex, len(state.ValidatorRegistry))
	for i := 0; i < len(activeValidatorIndices); i++ {
		indexList[i] = activeValidatorIndices[i]
	}
	// Active validators, shuffled in-place.
	shuffling.ShuffleList(indexList, seed)
	return indexList
}
