package beacon

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/protolambda/eth2-shuffle"
	"github.com/protolambda/zrnt/eth2/util/bitfield"
	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/zrnt/eth2/util/data_types"
	"github.com/protolambda/zrnt/eth2/util/hash"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"sort"
)

type BeaconState struct {
	// Misc
	Slot        Slot
	GenesisTime Timestamp
	Fork        Fork

	// Validator registry
	ValidatorRegistry            ValidatorRegistry
	Balances                     []Gwei

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
	return state.Epoch() - 1
}

func (state *BeaconState) GetChurnLimit() uint64 {
	return math.MaxU64(MIN_PER_EPOCH_CHURN_LIMIT,
		state.ValidatorRegistry.GetActiveValidatorCount(state.Epoch()) / CHURN_LIMIT_QUOTIENT)
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

// Activate the validator of the given index
func (state *BeaconState) ActivateValidator(index ValidatorIndex, isGenesis bool) {
	validator := state.ValidatorRegistry[index]

	if isGenesis {
		validator.ActivationEligibilityEpoch = GENESIS_EPOCH
		validator.ActivationEpoch = GENESIS_EPOCH
	} else {
		validator.ActivationEpoch = state.Epoch().GetDelayedActivationExitEpoch()
	}
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

func (state *BeaconState) GetShardDelta(epoch Epoch) Shard {
	return Shard(math.MinU64(
		state.GetEpochCommitteeCount(epoch),
		uint64(SHARD_COUNT - (SHARD_COUNT / Shard(SLOTS_PER_EPOCH)))))
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
			if state.GetEffectiveBalance(candidate)<<8 > MAX_DEPOSIT_AMOUNT*Gwei(randomByte) {
				return candidate
			}
		}
	}
	return 0
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

func (state *BeaconState) GetWinningRootAndParticipants(shard Shard) (Root, ValidatorSet) {
	weightedCrosslinks := make(map[Root]Gwei)

	updateCrosslinkWeights := func(att *PendingAttestation) {
		if att.Data.PreviousCrosslink == state.LatestCrosslinks[shard] {
			participants, _ := state.GetAttestationParticipants(&att.Data, &att.AggregationBitfield)
			for _, participant := range participants {
				weightedCrosslinks[att.Data.CrosslinkDataRoot] += state.GetEffectiveBalance(participant)
			}
		}
	}
	for _, att := range state.PreviousEpochAttestations {
		updateCrosslinkWeights(att)
	}
	for _, att := range state.CurrentEpochAttestations {
		updateCrosslinkWeights(att)
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
	for _, att := range state.PreviousEpochAttestations {
		findWinners(att)
	}
	for _, att := range state.CurrentEpochAttestations {
		findWinners(att)
	}
	winningAttesters := make(ValidatorSet, len(winningAttestersSet))
	i := 0
	for attester := range winningAttestersSet {
		winningAttesters[i] = attester
		i++
	}
	// Needs to be sorted, for efficiency and consistency
	sort.Sort(winningAttesters)

	return winningRoot, winningAttesters
}

// Exit the validator with the given index
func (state *BeaconState) ExitValidator(index ValidatorIndex) {
	validator := state.ValidatorRegistry[index]
	// Update validator exit epoch if not previously exited
	if validator.ExitEpoch == FAR_FUTURE_EPOCH {
		validator.ExitEpoch = state.Epoch().GetDelayedActivationExitEpoch()
	}
}

// Return the crosslink committee corresponding to the given attestationData
func (state *BeaconState) GetCrosslinkCommitteeForAttestation(attestationData *AttestationData) []ValidatorIndex {
	crosslinkCommittees := state.GetCrosslinkCommitteesAtSlot(attestationData.Slot)
	startShard := crosslinkCommittees[0].Shard
	// Find the committee in the list with the desired shard
	// TODO: spec committee lookup can be much more efficient here
	//  by exploiting the (modulo) consecutive shard ordering, see below
	crosslinkCommittee := crosslinkCommittees[(SHARD_COUNT+attestationData.Shard-startShard)%SHARD_COUNT]
	if crosslinkCommittee.Shard != attestationData.Shard {
		panic("either crosslink committees data is invalid, or supplied attestation has invalid shard")
	}
	return crosslinkCommittee.Committee
}

// Return the participant indices at for the attestation_data and bitfield
func (state *BeaconState) GetAttestationParticipants(attestationData *AttestationData, bitfield *bitfield.Bitfield) (ValidatorSet, error) {
	// Find the committee in the list with the desired shard
	crosslinkCommittee := state.GetCrosslinkCommitteeForAttestation(attestationData)

	if len(crosslinkCommittee) == 0 {
		return nil, errors.New(fmt.Sprintf("cannot find crosslink committee at slot %d for shard %d", attestationData.Slot, attestationData.Shard))
	}
	if !bitfield.VerifySize(uint64(len(crosslinkCommittee))) {
		return nil, errors.New("bitfield has wrong size for corresponding crosslink committee")
	}

	// Find the participating attesters in the committee
	participants := make(ValidatorSet, 0)
	for i, vIndex := range crosslinkCommittee {
		if bitfield.GetBit(uint64(i)) == 1 {
			participants = append(participants, vIndex)
		}
	}
	sort.Sort(participants)
	return participants, nil
}

// Slash the validator with index index.
func (state *BeaconState) SlashValidator(index ValidatorIndex) error {
	validator := state.ValidatorRegistry[index]
	state.InitiateValidatorExit(index)
	state.LatestSlashedBalances[state.Epoch()%LATEST_SLASHED_EXIT_LENGTH] += state.GetEffectiveBalance(index)

	whistleblowerReward := state.GetEffectiveBalance(index) / WHISTLEBLOWING_REWARD_QUOTIENT
	propIndex := state.GetBeaconProposerIndex()
	state.IncreaseBalance(propIndex, whistleblowerReward)
	state.DecreaseBalance(index, whistleblowerReward)
	validator.Slashed = true
	validator.WithdrawableEpoch = state.Epoch() + LATEST_SLASHED_EXIT_LENGTH
	return nil
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
	return Min(state.GetBalance(index), MAX_DEPOSIT_AMOUNT)
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
	participants, err := state.GetAttestationParticipants(&attestation.Data, &attestation.AggregationBitfield)
	if err != nil {
		return nil, errors.New("participants could not be derived from aggregation_bitfield")
	}
	custodyBit1Indices, err := state.GetAttestationParticipants(&attestation.Data, &attestation.CustodyBitfield)
	if err != nil {
		return nil, errors.New("participants could not be derived from custody_bitfield")
	}
	custodyBit0Indices := make([]ValidatorIndex, 0, len(participants) - len(custodyBit1Indices))
	participants.ZigZagJoin(custodyBit1Indices, nil, func(i ValidatorIndex) {
		custodyBit0Indices = append(custodyBit0Indices, i)
	})
	return &IndexedAttestation{
		CustodyBit0Indexes: custodyBit0Indices,
		CustodyBit1Indexes: custodyBit1Indices,
		Data:               attestation.Data,
		AggregateSignature: attestation.AggregateSignature,
	}, nil
}

// Verify validity of slashable_attestation fields.
func (state *BeaconState) VerifyIndexedAttestation(indexedAttestation *IndexedAttestation) bool {
	custodyBit0Indices := indexedAttestation.CustodyBit0Indexes
	custodyBit1Indices := indexedAttestation.CustodyBit1Indexes

	// [TO BE REMOVED IN PHASE 1]
	if len(custodyBit1Indices) != 0 {
		return false
	}

	totalAttestingIndices := len(custodyBit1Indices) + len(custodyBit0Indices)
	if !(1 <= totalAttestingIndices && totalAttestingIndices <= MAX_ATTESTATION_PARTICIPANTS) {
		return false
	}

	// simple check if the lists are sorted.
	verifyAttestIndexList := func(indices []ValidatorIndex) bool {
		end := len(indices) - 1
		for i := 0; i < end; i++ {
			if indices[i] >= indices[i+1] {
				return false
			}
		}

		// Check the last item of the sorted list to be a valid index
		if !state.ValidatorRegistry.IsValidatorIndex(indices[end]) {
			return false
		}
		return true
	}
	if !verifyAttestIndexList(custodyBit0Indices) || !verifyAttestIndexList(custodyBit1Indices) {
		return false
	}

	custodyBit0Pubkeys := make([]bls.BLSPubkey, 0)
	for _, i := range custodyBit0Indices {
		custodyBit0Pubkeys = append(custodyBit0Pubkeys, state.ValidatorRegistry[i].Pubkey)
	}
	custodyBit1Pubkeys := make([]bls.BLSPubkey, 0)
	for _, i := range custodyBit1Indices {
		custodyBit1Pubkeys = append(custodyBit1Pubkeys, state.ValidatorRegistry[i].Pubkey)
	}

	// don't trust, verify
	return bls.BlsVerifyMultiple(
		[]bls.BLSPubkey{
			bls.BlsAggregatePubkeys(custodyBit0Pubkeys),
			bls.BlsAggregatePubkeys(custodyBit1Pubkeys)},
		[][32]byte{
			ssz.HashTreeRoot(AttestationDataAndCustodyBit{Data: indexedAttestation.Data, CustodyBit: false}),
			ssz.HashTreeRoot(AttestationDataAndCustodyBit{Data: indexedAttestation.Data, CustodyBit: true})},
		indexedAttestation.AggregateSignature,
		GetDomain(state.Fork, indexedAttestation.Data.Slot.ToEpoch(), DOMAIN_ATTESTATION),
	)
}

// Shuffle active validators
func (state *BeaconState) GetShuffled(seed Root, epoch Epoch) []ValidatorIndex {
	activeValidatorIndices := state.ValidatorRegistry.GetActiveValidatorIndices(epoch)
	committeeCount := state.GetEpochCommitteeCount(epoch)
	if committeeCount > uint64(len(activeValidatorIndices)) {
		panic("not enough validators to form committees!")
	}
	// Active validators, shuffled in-place.
	hFnObj := sha256.New()
	hashFn := func(in []byte) []byte {
		hFnObj.Reset()
		hFnObj.Write(in)
		return hFnObj.Sum(nil)
	}
	rawIndexList := make([]uint64, len(state.ValidatorRegistry))
	for i := 0; i < len(activeValidatorIndices); i++ {
		rawIndexList[i] = uint64(activeValidatorIndices[i])
	}
	eth2_shuffle.ShuffleList(hashFn, rawIndexList, SHUFFLE_ROUND_COUNT, seed)
	shuffled := make([]ValidatorIndex, len(state.ValidatorRegistry))
	for i := 0; i < len(activeValidatorIndices); i++ {
		shuffled[i] = ValidatorIndex(rawIndexList[i])
	}
	return shuffled
}

