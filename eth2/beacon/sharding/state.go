package sharding

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
	LatestBlockHeader common.BeaconBlockHeader    `json:"latest_block_header" yaml:"latest_block_header"`
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
	// Attestations
	PreviousEpochAttestations PendingAttestations `json:"previous_epoch_attestations" yaml:"previous_epoch_attestations"`
	CurrentEpochAttestations  PendingAttestations `json:"current_epoch_attestations" yaml:"current_epoch_attestations"`
	// Finality
	JustificationBits           common.JustificationBits `json:"justification_bits" yaml:"justification_bits"`
	PreviousJustifiedCheckpoint common.Checkpoint        `json:"previous_justified_checkpoint" yaml:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint  common.Checkpoint        `json:"current_justified_checkpoint" yaml:"current_justified_checkpoint"`
	FinalizedCheckpoint         common.Checkpoint        `json:"finalized_checkpoint" yaml:"finalized_checkpoint"`
	// Execution-layer
	LatestExecutionPayloadHeader common.ExecutionPayloadHeader `json:"latest_execution_payload_header" yaml:"latest_execution_payload_header"`
	// Sharding
	ShardBuffer            ShardBuffer  `json:"shard_buffer" yaml:"shard_buffer"`
	ShardGasPrice          Uint64View   `json:"shard_gasprice" yaml:"shard_gasprice"`
	CurrentEpochStartShard common.Shard `json:"current_epoch_start_shard" yaml:"current_epoch_start_shard"`
}

func (v *BeaconState) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), spec.Wrap(&v.HistoricalRoots),
		&v.Eth1Data, spec.Wrap(&v.Eth1DataVotes), &v.DepositIndex,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances),
		spec.Wrap(&v.RandaoMixes), spec.Wrap(&v.Slashings),
		spec.Wrap(&v.PreviousEpochAttestations), spec.Wrap(&v.CurrentEpochAttestations),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint,
		&v.LatestExecutionPayloadHeader,
		spec.Wrap(&v.ShardBuffer),
		&v.ShardGasPrice,
		&v.CurrentEpochStartShard)
}

func (v *BeaconState) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), spec.Wrap(&v.HistoricalRoots),
		&v.Eth1Data, spec.Wrap(&v.Eth1DataVotes), &v.DepositIndex,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances),
		spec.Wrap(&v.RandaoMixes), spec.Wrap(&v.Slashings),
		spec.Wrap(&v.PreviousEpochAttestations), spec.Wrap(&v.CurrentEpochAttestations),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint,
		&v.LatestExecutionPayloadHeader,
		spec.Wrap(&v.ShardBuffer),
		&v.ShardGasPrice,
		&v.CurrentEpochStartShard)
}

func (v *BeaconState) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), spec.Wrap(&v.HistoricalRoots),
		&v.Eth1Data, spec.Wrap(&v.Eth1DataVotes), &v.DepositIndex,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances),
		spec.Wrap(&v.RandaoMixes), spec.Wrap(&v.Slashings),
		spec.Wrap(&v.PreviousEpochAttestations), spec.Wrap(&v.CurrentEpochAttestations),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint,
		&v.LatestExecutionPayloadHeader,
		spec.Wrap(&v.ShardBuffer),
		&v.ShardGasPrice,
		&v.CurrentEpochStartShard)
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
		spec.Wrap(&v.PreviousEpochAttestations), spec.Wrap(&v.CurrentEpochAttestations),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint,
		&v.LatestExecutionPayloadHeader,
		spec.Wrap(&v.ShardBuffer),
		&v.ShardGasPrice,
		&v.CurrentEpochStartShard)
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
	_statePreviousEpochAttestations
	_stateCurrentEpochAttestations
	_stateJustificationBits
	_statePreviousJustifiedCheckpoint
	_stateCurrentJustifiedCheckpoint
	_stateFinalizedCheckpoint
	_latestExecutionPayloadHeader
	_shardBuffer
	_shardGasprice
	_currentEpochStartShard
)

func BeaconStateType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("BeaconState", []FieldDef{
		// Versioning
		{"genesis_time", Uint64Type},
		{"genesis_validators_root", RootType},
		{"slot", common.SlotType},
		{"fork", common.ForkType},
		// History
		{"latest_block_header", common.BeaconBlockHeaderType},
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
		{"slashings", phase0.SlashingsType(spec)}, // Per-epoch sums of slashed effective balances
		// Attestations
		{"previous_epoch_attestations", PendingAttestationsType(spec)},
		{"current_epoch_attestations", PendingAttestationsType(spec)},
		// Finality
		{"justification_bits", common.JustificationBitsType},     // Bit set for every recent justified epoch
		{"previous_justified_checkpoint", common.CheckpointType}, // Previous epoch snapshot
		{"current_justified_checkpoint", common.CheckpointType},
		{"finalized_checkpoint", common.CheckpointType},
		// Execution-layer
		{"latest_execution_payload_header", common.ExecutionPayloadHeaderType},
		// Sharding
		{"shard_buffer", ShardBufferType(spec)},
		{"shard_gasprice", Uint64Type},
		{"current_epoch_start_shard", common.ShardType},
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

var _ common.BeaconState = (*phase0.BeaconStateView)(nil)

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

func (state *BeaconStateView) Fork() (common.Fork, error) {
	fv, err := common.AsFork(state.Get(_stateFork))
	if err != nil {
		return common.Fork{}, err
	}
	return fv.Raw()
}

func (state *BeaconStateView) SetFork(f common.Fork) error {
	return state.Set(_stateFork, f.View())
}

func (state *BeaconStateView) LatestBlockHeader() (*common.BeaconBlockHeader, error) {
	h, err := common.AsBeaconBlockHeader(state.Get(_stateLatestBlockHeader))
	if err != nil {
		return nil, err
	}
	return h.Raw()
}

func (state *BeaconStateView) SetLatestBlockHeader(v *common.BeaconBlockHeader) error {
	return state.Set(_stateLatestBlockHeader, v.View())
}

func (state *BeaconStateView) BlockRoots() (common.BatchRoots, error) {
	return phase0.AsBatchRoots(state.Get(_stateBlockRoots))
}

func (state *BeaconStateView) StateRoots() (common.BatchRoots, error) {
	return phase0.AsBatchRoots(state.Get(_stateStateRoots))
}

func (state *BeaconStateView) HistoricalRoots() (common.HistoricalRoots, error) {
	return phase0.AsHistoricalRoots(state.Get(_stateHistoricalRoots))
}

func (state *BeaconStateView) Eth1Data() (common.Eth1Data, error) {
	dat, err := common.AsEth1Data(state.Get(_stateEth1Data))
	if err != nil {
		return common.Eth1Data{}, err
	}
	return dat.Raw()
}

func (state *BeaconStateView) SetEth1Data(v common.Eth1Data) error {
	return state.Set(_stateEth1Data, v.View())
}

func (state *BeaconStateView) Eth1DataVotes() (common.Eth1DataVotes, error) {
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

func (state *BeaconStateView) Validators() (common.ValidatorRegistry, error) {
	return phase0.AsValidatorsRegistry(state.Get(_stateValidators))
}

func (state *BeaconStateView) Balances() (common.Balances, error) {
	return phase0.AsRegistryBalances(state.Get(_stateBalances))
}

func (state *BeaconStateView) SetBalances(balances *phase0.RegistryBalancesView) error {
	return state.Set(_stateBalances, balances)
}

func (state *BeaconStateView) AddValidator(spec *common.Spec, pub common.BLSPubkey, withdrawalCreds common.Root, balance common.Gwei) error {
	effBalance := balance - (balance % spec.EFFECTIVE_BALANCE_INCREMENT)
	if effBalance > spec.MAX_EFFECTIVE_BALANCE {
		effBalance = spec.MAX_EFFECTIVE_BALANCE
	}
	validatorRaw := phase0.Validator{
		Pubkey:                     pub,
		WithdrawalCredentials:      withdrawalCreds,
		ActivationEligibilityEpoch: common.FAR_FUTURE_EPOCH,
		ActivationEpoch:            common.FAR_FUTURE_EPOCH,
		ExitEpoch:                  common.FAR_FUTURE_EPOCH,
		WithdrawableEpoch:          common.FAR_FUTURE_EPOCH,
		EffectiveBalance:           effBalance,
	}
	validators, err := phase0.AsValidatorsRegistry(state.Get(_stateValidators))
	if err != nil {
		return err
	}
	if err := validators.Append(validatorRaw.View()); err != nil {
		return err
	}
	bals, err := phase0.AsRegistryBalances(state.Get(_stateBalances))
	if err != nil {
		return err
	}
	if err := bals.Append(Uint64View(balance)); err != nil {
		return err
	}
	return nil
}

func (state *BeaconStateView) RandaoMixes() (common.RandaoMixes, error) {
	return phase0.AsRandaoMixes(state.Get(_stateRandaoMixes))
}

func (state *BeaconStateView) SeedRandao(spec *common.Spec, seed common.Root) error {
	v, err := phase0.SeedRandao(spec, seed)
	if err != nil {
		return err
	}
	return state.Set(_stateRandaoMixes, v)
}

func (state *BeaconStateView) Slashings() (common.Slashings, error) {
	return phase0.AsSlashings(state.Get(_stateSlashings))
}

func (state *BeaconStateView) PreviousEpochAttestations() (*PendingAttestationsView, error) {
	return AsPendingAttestations(state.Get(_statePreviousEpochAttestations))
}

func (state *BeaconStateView) CurrentEpochAttestations() (*PendingAttestationsView, error) {
	return AsPendingAttestations(state.Get(_stateCurrentEpochAttestations))
}

func (state *BeaconStateView) JustificationBits() (common.JustificationBits, error) {
	b, err := common.AsJustificationBits(state.Get(_stateJustificationBits))
	if err != nil {
		return common.JustificationBits{}, err
	}
	return b.Raw()
}

func (state *BeaconStateView) SetJustificationBits(bits common.JustificationBits) error {
	b, err := common.AsJustificationBits(state.Get(_stateJustificationBits))
	if err != nil {
		return err
	}
	return b.Set(bits)
}

func (state *BeaconStateView) PreviousJustifiedCheckpoint() (common.Checkpoint, error) {
	c, err := common.AsCheckPoint(state.Get(_statePreviousJustifiedCheckpoint))
	if err != nil {
		return common.Checkpoint{}, err
	}
	return c.Raw()
}

func (state *BeaconStateView) SetPreviousJustifiedCheckpoint(c common.Checkpoint) error {
	v, err := common.AsCheckPoint(state.Get(_statePreviousJustifiedCheckpoint))
	if err != nil {
		return err
	}
	return v.Set(&c)
}

func (state *BeaconStateView) CurrentJustifiedCheckpoint() (common.Checkpoint, error) {
	c, err := common.AsCheckPoint(state.Get(_stateCurrentJustifiedCheckpoint))
	if err != nil {
		return common.Checkpoint{}, err
	}
	return c.Raw()
}

func (state *BeaconStateView) SetCurrentJustifiedCheckpoint(c common.Checkpoint) error {
	v, err := common.AsCheckPoint(state.Get(_stateCurrentJustifiedCheckpoint))
	if err != nil {
		return err
	}
	return v.Set(&c)
}

func (state *BeaconStateView) FinalizedCheckpoint() (common.Checkpoint, error) {
	c, err := common.AsCheckPoint(state.Get(_stateFinalizedCheckpoint))
	if err != nil {
		return common.Checkpoint{}, err
	}
	return c.Raw()
}

func (state *BeaconStateView) SetFinalizedCheckpoint(c common.Checkpoint) error {
	v, err := common.AsCheckPoint(state.Get(_stateFinalizedCheckpoint))
	if err != nil {
		return err
	}
	return v.Set(&c)
}

func (state *BeaconStateView) LatestExecutionPayloadHeader() (*common.ExecutionPayloadHeaderView, error) {
	return common.AsExecutionPayloadHeader(state.Get(_latestExecutionPayloadHeader))
}

func (state *BeaconStateView) SetLatestExecutionPayloadHeader(h *common.ExecutionPayloadHeader) error {
	return state.Set(_latestExecutionPayloadHeader, h.View())
}

func (state *BeaconStateView) ShardBuffer() (*ShardBufferView, error) {
	return AsShardBuffer(state.Get(_shardBuffer))
}

func (state *BeaconStateView) ShardGasPrice() (common.Gwei, error) {
	return common.AsGwei(state.Get(_shardGasprice))
}

func (state *BeaconStateView) SetShardGasPrice(price common.Gwei) error {
	return state.Set(_shardGasprice, Uint64View(price))
}

func (state *BeaconStateView) CurrentEpochStartShard() (common.Shard, error) {
	return common.AsShard(state.Get(_currentEpochStartShard))
}

func (state *BeaconStateView) SetCurrentEpochStartShard(shard common.Shard) error {
	return state.Set(_currentEpochStartShard, Uint64View(shard))
}

func (state *BeaconStateView) ForkSettings(spec *common.Spec) *common.ForkSettings {
	return &common.ForkSettings{
		MinSlashingPenaltyQuotient:     spec.MIN_SLASHING_PENALTY_QUOTIENT,
		ProportionalSlashingMultiplier: spec.PROPORTIONAL_SLASHING_MULTIPLIER,
		CalcProposerShare: func(whistleblowerReward common.Gwei) common.Gwei {
			return whistleblowerReward / common.Gwei(spec.PROPOSER_REWARD_QUOTIENT)
		},
	}
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

func (state *BeaconStateView) CopyState() (common.BeaconState, error) {
	return AsBeaconStateView(state.ContainerView.Copy())
}
