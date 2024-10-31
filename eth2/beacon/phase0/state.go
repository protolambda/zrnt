package phase0

import (
	"bytes"

	"github.com/protolambda/zrnt/eth2/beacon/common"
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
	LatestBlockHeader      common.BeaconBlockHeader `json:"latest_block_header" yaml:"latest_block_header"`
	BlockRoots             HistoricalBatchRoots     `json:"block_roots" yaml:"block_roots"`
	StateRoots             HistoricalBatchRoots     `json:"state_roots" yaml:"state_roots"`
	RewardAdjustmentFactor common.Number            `json:"reward_adjustment_factor" yaml:"reward_adjustment_factor"`
	// Registry
	Validators  ValidatorRegistry `json:"validators" yaml:"validators"`
	Balances    Balances          `json:"balances" yaml:"balances"`
	Reserves    common.Number     `json:"reserves" yaml:"reserves"`
	RandaoMixes RandaoMixes       `json:"randao_mixes" yaml:"randao_mixes"`
	// Attestations
	PreviousEpochAttestations PendingAttestations `json:"previous_epoch_attestations" yaml:"previous_epoch_attestations"`
	CurrentEpochAttestations  PendingAttestations `json:"current_epoch_attestations" yaml:"current_epoch_attestations"`
	// Finality
	JustificationBits           common.JustificationBits `json:"justification_bits" yaml:"justification_bits"`
	PreviousJustifiedCheckpoint common.Checkpoint        `json:"previous_justified_checkpoint" yaml:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint  common.Checkpoint        `json:"current_justified_checkpoint" yaml:"current_justified_checkpoint"`
	FinalizedCheckpoint         common.Checkpoint        `json:"finalized_checkpoint" yaml:"finalized_checkpoint"`
}

func (v *BeaconState) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), &v.RewardAdjustmentFactor,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances), &v.Reserves,
		spec.Wrap(&v.RandaoMixes),
		spec.Wrap(&v.PreviousEpochAttestations), spec.Wrap(&v.CurrentEpochAttestations),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint)
}

func (v *BeaconState) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), &v.RewardAdjustmentFactor,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances), &v.Reserves,
		spec.Wrap(&v.RandaoMixes),
		spec.Wrap(&v.PreviousEpochAttestations), spec.Wrap(&v.CurrentEpochAttestations),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint)
}

func (v *BeaconState) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), &v.RewardAdjustmentFactor,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances), &v.Reserves,
		spec.Wrap(&v.RandaoMixes),
		spec.Wrap(&v.PreviousEpochAttestations), spec.Wrap(&v.CurrentEpochAttestations),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint)
}

func (*BeaconState) FixedLength(*common.Spec) uint64 {
	return 0 // dynamic size
}

func (v *BeaconState) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), &v.RewardAdjustmentFactor,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances), &v.Reserves,
		spec.Wrap(&v.RandaoMixes),
		spec.Wrap(&v.PreviousEpochAttestations), spec.Wrap(&v.CurrentEpochAttestations),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint)
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
	_stateRewardAdjustmentFactor
	_stateValidators
	_stateBalances
	_stateReserves
	_stateRandaoMixes
	_statePreviousEpochAttestations
	_stateCurrentEpochAttestations
	_stateJustificationBits
	_statePreviousJustifiedCheckpoint
	_stateCurrentJustifiedCheckpoint
	_stateFinalizedCheckpoint
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
		{"block_roots", BatchRootsType(spec)},
		{"state_roots", BatchRootsType(spec)},
		// Tokenomics
		{"reward_adjustment_factor", Uint64Type},
		// Registry
		{"validators", ValidatorsRegistryType(spec)},
		{"balances", RegistryBalancesType(spec)},
		{"reserves", Uint64Type},
		// Randomness
		{"randao_mixes", RandaoMixesType(spec)},
		// Attestations
		{"previous_epoch_attestations", PendingAttestationsType(spec)},
		{"current_epoch_attestations", PendingAttestationsType(spec)},
		// Finality
		{"justification_bits", common.JustificationBitsType},     // Bit set for every recent justified epoch
		{"previous_justified_checkpoint", common.CheckpointType}, // Previous epoch snapshot
		{"current_justified_checkpoint", common.CheckpointType},
		{"finalized_checkpoint", common.CheckpointType},
	})
}

// To load a state:
//
//	state, err := beacon.AsBeaconStateView(beacon.BeaconStateType.Deserialize(codec.NewDecodingReader(reader, size)))
func AsBeaconStateView(v View, err error) (*BeaconStateView, error) {
	c, err := AsContainer(v, err)
	return &BeaconStateView{c}, err
}

type BeaconStateView struct {
	*ContainerView
}

var _ common.BeaconState = (*BeaconStateView)(nil)

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
	return AsBatchRoots(state.Get(_stateBlockRoots))
}

func (state *BeaconStateView) StateRoots() (common.BatchRoots, error) {
	return AsBatchRoots(state.Get(_stateStateRoots))
}

func (state *BeaconStateView) RewardAdjustmentFactor() (common.Number, error) {
	return common.AsNumber(state.Get(_stateRewardAdjustmentFactor))
}

func (state *BeaconStateView) SetRewardAdjustmentFactor(v common.Number) error {
	return state.Set(_stateRewardAdjustmentFactor, Uint64View(v))
}

func (state *BeaconStateView) Validators() (common.ValidatorRegistry, error) {
	return AsValidatorsRegistry(state.Get(_stateValidators))
}

func (state *BeaconStateView) Balances() (common.BalancesRegistry, error) {
	return AsRegistryBalances(state.Get(_stateBalances))
}

func (state *BeaconStateView) SetBalances(balances []common.Gwei) error {
	typ := state.Fields[_stateBalances].Type.(*BasicListTypeDef)
	balancesView, err := Balances(balances).View(typ.ListLimit)
	if err != nil {
		return err
	}
	return state.Set(_stateBalances, balancesView)
}

func (state *BeaconStateView) Reserves() (common.Number, error) {
	return common.AsNumber(state.Get(_stateReserves))
}

func (state *BeaconStateView) SetReserves(v common.Number) error {
	return state.Set(_stateReserves, Uint64View(v))
}

func (state *BeaconStateView) AddValidator(spec *common.Spec, pub common.BLSPubkey, withdrawalCreds common.Root, balance common.Gwei) error {
	effBalance := balance - (balance % spec.EFFECTIVE_BALANCE_INCREMENT)
	if effBalance > spec.MAX_EFFECTIVE_BALANCE {
		effBalance = spec.MAX_EFFECTIVE_BALANCE
	}
	validatorRaw := Validator{
		Pubkey:                     pub,
		WithdrawalCredentials:      withdrawalCreds,
		ActivationEligibilityEpoch: common.FAR_FUTURE_EPOCH,
		ActivationEpoch:            common.FAR_FUTURE_EPOCH,
		ExitEpoch:                  common.FAR_FUTURE_EPOCH,
		EffectiveBalance:           effBalance,
		PrincipalBalance:           balance,
	}
	validators, err := AsValidatorsRegistry(state.Get(_stateValidators))
	if err != nil {
		return err
	}
	if err := validators.Append(validatorRaw.View()); err != nil {
		return err
	}
	bals, err := state.Balances()
	if err != nil {
		return err
	}
	if err := bals.AppendBalance(balance); err != nil {
		return err
	}
	return nil
}

func (state *BeaconStateView) RandaoMixes() (common.RandaoMixes, error) {
	return AsRandaoMixes(state.Get(_stateRandaoMixes))
}

func (state *BeaconStateView) SeedRandao(spec *common.Spec, seed common.Root) error {
	v, err := SeedRandao(spec, seed)
	if err != nil {
		return err
	}
	return state.Set(_stateRandaoMixes, v)
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

func (state *BeaconStateView) ForkSettings(spec *common.Spec) *common.ForkSettings {
	return &common.ForkSettings{
		MinSlashingPenaltyQuotient: uint64(spec.MIN_SLASHING_PENALTY_QUOTIENT),
		InactivityPenaltyQuotient:  uint64(spec.INACTIVITY_PENALTY_QUOTIENT),
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
