package altair

import (
	"bytes"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type BeaconState struct {
	// Versioning
	GenesisTime           common.Timestamp `json:"genesis_time" yaml:"genesis_time"`
	GenesisValidatorsRoot common.Root      `json:"genesis_validators_root" yaml:"genesis_validators_root"`
	Slot                  common.Slot      `json:"slot" yaml:"slot"`
	Fork                  common.Fork      `json:"fork" yaml:"fork"`
	// History
	LatestBlockHeader phase0.BeaconBlockHeader    `json:"latest_block_header" yaml:"latest_block_header"`
	BlockRoots        phase0.HistoricalBatchRoots `json:"block_roots" yaml:"block_roots"`
	StateRoots        phase0.HistoricalBatchRoots `json:"state_roots" yaml:"state_roots"`
	HistoricalRoots   phase0.HistoricalRoots      `json:"historical_roots" yaml:"historical_roots"`
	// Eth1
	Eth1Data      common.Eth1Data      `json:"eth1_data" yaml:"eth1_data"`
	Eth1DataVotes phase0.Eth1DataVotes `json:"eth1_data_votes" yaml:"eth1_data_votes"`
	DepositIndex  common.DepositIndex  `json:"eth1_deposit_index" yaml:"eth1_deposit_index"`
	// Registry
	Validators  phase0.ValidatorRegistry `json:"validators" yaml:"validators"`
	Balances    phase0.Balances          `json:"balances" yaml:"balances"`
	RandaoMixes phase0.RandaoMixes       `json:"randao_mixes" yaml:"randao_mixes"`
	Slashings   phase0.SlashingsHistory  `json:"slashings" yaml:"slashings"`
	// Participation
	PreviousEpochParticipation ParticipationRegistry `json:"previous_epoch_participation" yaml:"previous_epoch_participation"`
	CurrentEpochParticipation  ParticipationRegistry `json:"current_epoch_participation" yaml:"current_epoch_participation"`
	// Finality
	JustificationBits           phase0.JustificationBits `json:"justification_bits" yaml:"justification_bits"`
	PreviousJustifiedCheckpoint common.Checkpoint        `json:"previous_justified_checkpoint" yaml:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint  common.Checkpoint        `json:"current_justified_checkpoint" yaml:"current_justified_checkpoint"`
	FinalizedCheckpoint         common.Checkpoint        `json:"finalized_checkpoint" yaml:"finalized_checkpoint"`
	// Light client sync committees
	CurrentSyncCommittee SyncCommittee `json:"current_sync_committee" yaml:"current_sync_committee"`
	NextSyncCommittee    SyncCommittee `json:"next_sync_committee" yaml:"next_sync_committee"`
	// Leak
	LeakScores LeakScores `json:"leak_scores" yaml:"leak_scores"`
}

func (v *BeaconState) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), spec.Wrap(&v.HistoricalRoots),
		&v.Eth1Data, spec.Wrap(&v.Eth1DataVotes), &v.DepositIndex,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances),
		spec.Wrap(&v.RandaoMixes), spec.Wrap(&v.Slashings),
		spec.Wrap(&v.PreviousEpochParticipation), spec.Wrap(&v.CurrentEpochParticipation),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint,
		spec.Wrap(&v.CurrentSyncCommittee), spec.Wrap(&v.NextSyncCommittee),
		spec.Wrap(&v.LeakScores))
}

func (v *BeaconState) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), spec.Wrap(&v.HistoricalRoots),
		&v.Eth1Data, spec.Wrap(&v.Eth1DataVotes), &v.DepositIndex,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances),
		spec.Wrap(&v.RandaoMixes), spec.Wrap(&v.Slashings),
		spec.Wrap(&v.PreviousEpochParticipation), spec.Wrap(&v.CurrentEpochParticipation),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint,
		spec.Wrap(&v.CurrentSyncCommittee), spec.Wrap(&v.NextSyncCommittee),
		spec.Wrap(&v.LeakScores))
}

func (v *BeaconState) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), spec.Wrap(&v.HistoricalRoots),
		&v.Eth1Data, spec.Wrap(&v.Eth1DataVotes), &v.DepositIndex,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances),
		spec.Wrap(&v.RandaoMixes), spec.Wrap(&v.Slashings),
		spec.Wrap(&v.PreviousEpochParticipation), spec.Wrap(&v.CurrentEpochParticipation),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint,
		spec.Wrap(&v.CurrentSyncCommittee), spec.Wrap(&v.NextSyncCommittee),
		spec.Wrap(&v.LeakScores))
}

func (*BeaconState) FixedLength(*common.Spec) uint64 {
	return 0 // dynamic size
}

func (v *BeaconState) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), spec.Wrap(&v.HistoricalRoots),
		&v.Eth1Data, spec.Wrap(&v.Eth1DataVotes), &v.DepositIndex,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances),
		spec.Wrap(&v.RandaoMixes), spec.Wrap(&v.Slashings),
		spec.Wrap(&v.PreviousEpochParticipation), spec.Wrap(&v.CurrentEpochParticipation),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint,
		spec.Wrap(&v.CurrentSyncCommittee), spec.Wrap(&v.NextSyncCommittee),
		spec.Wrap(&v.LeakScores))
}

// Hack to make state fields consistent and verifiable without using many hardcoded indices
// A trade-off to interpret the state as tree, without generics, and access fields by index very fast.
const (
	_stateGenesisTime = iota
	_stateGenesisValidatorsRoot
	_stateSlot
	_stateFork
	_stateLatestBlockHeader
	_stateBlockRoots
	_stateStateRoots
	_stateHistoricalRoots
	_stateEth1Data
	_stateEth1DataVotes
	_stateDepositIndex
	_stateValidators
	_stateBalances
	_stateRandaoMixes
	_stateSlashings
	_statePreviousEpochParticipation
	_stateCurrentEpochParticipation
	_stateJustificationBits
	_statePreviousJustifiedCheckpoint
	_stateCurrentJustifiedCheckpoint
	_stateFinalizedCheckpoint
	_currentSyncCommittee
	_nextSyncCommittee
	_leakScores
)

func BeaconStateType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("BeaconState", []FieldDef{
		// Versioning
		{"genesis_time", Uint64Type},
		{"genesis_validators_root", RootType},
		{"slot", common.SlotType},
		{"fork", common.ForkType},
		// History
		{"latest_block_header", phase0.BeaconBlockHeaderType},
		{"block_roots", phase0.BatchRootsType(spec)},
		{"state_roots", phase0.BatchRootsType(spec)},
		{"historical_roots", phase0.HistoricalRootsType(spec)},
		// Eth1
		{"eth1_data", common.Eth1DataType},
		{"eth1_data_votes", phase0.Eth1DataVotesType(spec)},
		{"eth1_deposit_index", Uint64Type},
		// Registry
		{"validators", phase0.ValidatorsRegistryType(spec)},
		{"balances", phase0.RegistryBalancesType(spec)},
		// Randomness
		{"randao_mixes", phase0.RandaoMixesType(spec)},
		// Slashings
		{"slashings", phase0.SlashingsType(spec)},
		// Participation
		{"previous_epoch_participation", ParticipationRegistryType(spec)},
		{"current_epoch_participation", ParticipationRegistryType(spec)},
		// Finality
		{"justification_bits", phase0.JustificationBitsType},
		{"previous_justified_checkpoint", common.CheckpointType},
		{"current_justified_checkpoint", common.CheckpointType},
		{"finalized_checkpoint", common.CheckpointType},
		// Light client sync committees
		{"current_sync_committee", SyncCommitteeType(spec)},
		{"next_sync_committee", SyncCommitteeType(spec)},
		{"leak_scores", LeakScoresType(spec)},
	})
}

// To load a state:
//
//     state, err := beacon.AsBeaconStateView(beacon.BeaconStateType.Deserialize(codec.NewDecodingReader(reader, size)))
func AsBeaconStateView(v View, err error) (*BeaconStateView, error) {
	c, err := AsContainer(v, err)
	return &BeaconStateView{c}, err
}

type BeaconStateView struct {
	*ContainerView
}

func NewBeaconStateView(spec *common.Spec) *BeaconStateView {
	return &BeaconStateView{ContainerView: BeaconStateType(spec).New()}
}

func (state *BeaconStateView) GenesisTime() (common.Timestamp, error) {
	return common.AsTimestamp(state.Get(_stateGenesisTime))
}

func (state *BeaconStateView) SetGenesisTime(t common.Timestamp) error {
	return state.Set(_stateGenesisTime, Uint64View(t))
}

func (state *BeaconStateView) GenesisValidatorsRoot() (common.Root, error) {
	return AsRoot(state.Get(_stateGenesisValidatorsRoot))
}

func (state *BeaconStateView) SetGenesisValidatorsRoot(r common.Root) error {
	rv := RootView(r)
	return state.Set(_stateGenesisValidatorsRoot, &rv)
}

func (state *BeaconStateView) Slot() (common.Slot, error) {
	return common.AsSlot(state.Get(_stateSlot))
}

func (state *BeaconStateView) SetSlot(slot common.Slot) error {
	return state.Set(_stateSlot, Uint64View(slot))
}

func (state *BeaconStateView) Fork() (*common.ForkView, error) {
	return common.AsFork(state.Get(_stateFork))
}

func (state *BeaconStateView) SetFork(f common.Fork) error {
	return state.Set(_stateFork, f.View())
}

func (state *BeaconStateView) LatestBlockHeader() (*phase0.BeaconBlockHeaderView, error) {
	return phase0.AsBeaconBlockHeader(state.Get(_stateLatestBlockHeader))
}

func (state *BeaconStateView) SetLatestBlockHeader(v *phase0.BeaconBlockHeaderView) error {
	return state.Set(_stateLatestBlockHeader, v)
}

func (state *BeaconStateView) BlockRoots() (*phase0.BatchRootsView, error) {
	return phase0.AsBatchRoots(state.Get(_stateBlockRoots))
}

func (state *BeaconStateView) StateRoots() (*phase0.BatchRootsView, error) {
	return phase0.AsBatchRoots(state.Get(_stateStateRoots))
}

func (state *BeaconStateView) HistoricalRoots() (*phase0.HistoricalRootsView, error) {
	return phase0.AsHistoricalRoots(state.Get(_stateHistoricalRoots))
}

func (state *BeaconStateView) Eth1Data() (*common.Eth1DataView, error) {
	return common.AsEth1Data(state.Get(_stateEth1Data))
}

func (state *BeaconStateView) SetEth1Data(v *common.Eth1DataView) error {
	return state.Set(_stateEth1Data, v)
}

func (state *BeaconStateView) Eth1DataVotes() (*phase0.Eth1DataVotesView, error) {
	return phase0.AsEth1DataVotes(state.Get(_stateEth1DataVotes))
}

func (state *BeaconStateView) DepositIndex() (common.DepositIndex, error) {
	return common.AsDepositIndex(state.Get(_stateDepositIndex))
}

func (state *BeaconStateView) IncrementDepositIndex() error {
	depIndex, err := state.DepositIndex()
	if err != nil {
		return err
	}
	return state.Set(_stateDepositIndex, Uint64View(depIndex+1))
}

func (state *BeaconStateView) Validators() (*phase0.ValidatorsRegistryView, error) {
	return phase0.AsValidatorsRegistry(state.Get(_stateValidators))
}

func (state *BeaconStateView) Balances() (*phase0.RegistryBalancesView, error) {
	return phase0.AsRegistryBalances(state.Get(_stateBalances))
}

func (state *BeaconStateView) RandaoMixes() (*phase0.RandaoMixesView, error) {
	return phase0.AsRandaoMixes(state.Get(_stateRandaoMixes))
}

func (state *BeaconStateView) SetRandaoMixes(v *phase0.RandaoMixesView) error {
	return state.Set(_stateRandaoMixes, v)
}

func (state *BeaconStateView) Slashings() (*phase0.SlashingsView, error) {
	return phase0.AsSlashings(state.Get(_stateSlashings))
}

func (state *BeaconStateView) PreviousEpochParticipation() (*ParticipationRegistryView, error) {
	return AsParticipationRegistry(state.Get(_statePreviousEpochParticipation))
}

func (state *BeaconStateView) CurrentEpochParticipation() (*ParticipationRegistryView, error) {
	return AsParticipationRegistry(state.Get(_stateCurrentEpochParticipation))
}

func (state *BeaconStateView) JustificationBits() (*phase0.JustificationBitsView, error) {
	return phase0.AsJustificationBits(state.Get(_stateJustificationBits))
}

func (state *BeaconStateView) PreviousJustifiedCheckpoint() (*common.CheckpointView, error) {
	return common.AsCheckPoint(state.Get(_statePreviousJustifiedCheckpoint))
}

func (state *BeaconStateView) CurrentJustifiedCheckpoint() (*common.CheckpointView, error) {
	return common.AsCheckPoint(state.Get(_stateCurrentJustifiedCheckpoint))
}

func (state *BeaconStateView) FinalizedCheckpoint() (*common.CheckpointView, error) {
	return common.AsCheckPoint(state.Get(_stateFinalizedCheckpoint))
}

func (state *BeaconStateView) CurrentSyncCommittee() (*SyncCommitteeView, error) {
	return AsSyncCommittee(state.Get(_currentSyncCommittee))
}

func (state *BeaconStateView) NextSyncCommittee() (*SyncCommitteeView, error) {
	return AsSyncCommittee(state.Get(_nextSyncCommittee))
}

func (state *BeaconStateView) LeakScores() (*LeakScoresView, error) {
	return AsLeakScores(state.Get(_leakScores))
}

func (state *BeaconStateView) IsValidIndex(index common.ValidatorIndex) (bool, error) {
	vals, err := state.Validators()
	if err != nil {
		return false, err
	}
	count, err := vals.Length()
	if err != nil {
		return false, err
	}
	return uint64(index) < count, nil
}

// Raw converts the tree-structured state into a flattened native Go structure.
func (state *BeaconStateView) Raw(spec *common.Spec) (*BeaconState, error) {
	var buf bytes.Buffer
	if err := state.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	var raw BeaconState
	err := raw.Deserialize(spec, codec.NewDecodingReader(bytes.NewReader(buf.Bytes()), uint64(len(buf.Bytes()))))
	if err != nil {
		return nil, err
	}
	return &raw, nil
}

func (state *BeaconStateView) SetRecentRoots(slot common.Slot, blockRoot common.Root, stateRoot common.Root) error {
	blockRootsBatch, err := state.BlockRoots()
	if err != nil {
		return err
	}
	stateRootsBatch, err := state.StateRoots()
	if err != nil {
		return err
	}
	if err := blockRootsBatch.SetRoot(slot%common.Slot(blockRootsBatch.VectorLength), blockRoot); err != nil {
		return err
	}
	if err := stateRootsBatch.SetRoot(slot%common.Slot(stateRootsBatch.VectorLength), stateRoot); err != nil {
		return err
	}
	return nil
}

func (state *BeaconStateView) UpdateHistoricalRoots() error {
	histRoots, err := state.HistoricalRoots()
	if err != nil {
		return err
	}
	blockRoots, err := state.BlockRoots()
	if err != nil {
		return err
	}
	stateRoots, err := state.StateRoots()
	if err != nil {
		return err
	}
	// emulating HistoricalBatch here
	hFn := tree.GetHashFn()
	newHistoricalRoot := RootView(tree.Hash(blockRoots.HashTreeRoot(hFn), stateRoots.HashTreeRoot(hFn)))
	return histRoots.Append(&newHistoricalRoot)
}

// Return the signature domain (fork version concatenated with domain type) of a message.
func (state *BeaconStateView) GetDomain(dom common.BLSDomainType, messageEpoch common.Epoch) (common.BLSDomain, error) {
	forkView, err := state.Fork()
	if err != nil {
		return common.BLSDomain{}, err
	}
	fork, err := forkView.Raw()
	if err != nil {
		return common.BLSDomain{}, err
	}
	genesisValRoot, err := state.GenesisValidatorsRoot()
	if err != nil {
		return common.BLSDomain{}, err
	}
	return fork.GetDomain(dom, genesisValRoot, messageEpoch)
}
