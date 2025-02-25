package electra

import (
	"bytes"

	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
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
	HistoricalRoots   phase0.HistoricalRoots      `json:"historical_roots" yaml:"historical_roots"` // Frozen in Capella, replaced by historical_summaries
	// Eth1
	Eth1Data         common.Eth1Data      `json:"eth1_data" yaml:"eth1_data"`
	Eth1DataVotes    phase0.Eth1DataVotes `json:"eth1_data_votes" yaml:"eth1_data_votes"`
	Eth1DepositIndex common.DepositIndex  `json:"eth1_deposit_index" yaml:"eth1_deposit_index"`
	// Registry
	Validators  phase0.ValidatorRegistry `json:"validators" yaml:"validators"`
	Balances    phase0.Balances          `json:"balances" yaml:"balances"`
	RandaoMixes phase0.RandaoMixes       `json:"randao_mixes" yaml:"randao_mixes"`
	Slashings   phase0.SlashingsHistory  `json:"slashings" yaml:"slashings"`
	// Participation
	PreviousEpochParticipation altair.ParticipationRegistry `json:"previous_epoch_participation" yaml:"previous_epoch_participation"`
	CurrentEpochParticipation  altair.ParticipationRegistry `json:"current_epoch_participation" yaml:"current_epoch_participation"`
	// Finality
	JustificationBits           common.JustificationBits `json:"justification_bits" yaml:"justification_bits"`
	PreviousJustifiedCheckpoint common.Checkpoint        `json:"previous_justified_checkpoint" yaml:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint  common.Checkpoint        `json:"current_justified_checkpoint" yaml:"current_justified_checkpoint"`
	FinalizedCheckpoint         common.Checkpoint        `json:"finalized_checkpoint" yaml:"finalized_checkpoint"`
	// Inactivity
	InactivityScores altair.InactivityScores `json:"inactivity_scores" yaml:"inactivity_scores"`
	// Light client sync committees
	CurrentSyncCommittee common.SyncCommittee `json:"current_sync_committee" yaml:"current_sync_committee"`
	NextSyncCommittee    common.SyncCommittee `json:"next_sync_committee" yaml:"next_sync_committee"`
	// Execution-layer  (modified in EIP-4844)
	LatestExecutionPayloadHeader deneb.ExecutionPayloadHeader `json:"latest_execution_payload_header" yaml:"latest_execution_payload_header"`
	// Withdrawals
	NextWithdrawalIndex          common.WithdrawalIndex `json:"next_withdrawal_index" yaml:"next_withdrawal_index"`
	NextWithdrawalValidatorIndex common.ValidatorIndex  `json:"next_withdrawal_validator_index" yaml:"next_withdrawal_validator_index"`
	// Deep history valid from Capella onwards
	HistoricalSummaries capella.HistoricalSummaries `json:"historical_summaries"`
	// [New in Electra:EIP6110]
	DepositRequestsStartIndex Uint64View `json:"deposit_requests_start_index" yaml:"deposit_requests_start_index"`
	// [New in Electra:EIP7251]
	DepositBalanceToConsume common.Gwei `json:"deposit_balance_to_consume" yaml:"deposit_balance_to_consume"`
	// [New in Electra:EIP7251]
	ExitBalanceToConsume common.Gwei `json:"exit_balance_to_consume" yaml:"exit_balance_to_consume"`
	// [New in Electra:EIP7251]
	EarliestExitEpoch common.Epoch `json:"earliest_exit_epoch" yaml:"earliest_exit_epoch"`
	// [New in Electra:EIP7251]
	ConsolidationBalanceToConsume common.Gwei `json:"consolidation_balance_to_consume" yaml:"consolidation_balance_to_consume"`
	// [New in Electra:EIP7251]
	EarliestConsolidationEpoch common.Epoch `json:"earliest_consolidation_epoch" yaml:"earliest_consolidation_epoch"`
	// [New in Electra:EIP7251]
	PendingDeposits common.PendingDeposits `json:"pending_deposits" yaml:"pending_deposits"`
	// [New in Electra:EIP7251]
	PendingPartialWithdrawals common.PendingPartialWithdrawals `json:"pending_partial_withdrawals" yaml:"pending_partial_withdrawals"`
	// [New in Electra:EIP7251]
	PendingConsolidations common.PendingConsolidations `json:"pending_consolidations" yaml:"pending_consolidations"`
}

func (v *BeaconState) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), spec.Wrap(&v.HistoricalRoots),
		&v.Eth1Data, spec.Wrap(&v.Eth1DataVotes), &v.Eth1DepositIndex,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances),
		spec.Wrap(&v.RandaoMixes), spec.Wrap(&v.Slashings),
		spec.Wrap(&v.PreviousEpochParticipation), spec.Wrap(&v.CurrentEpochParticipation),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint,
		spec.Wrap(&v.InactivityScores),
		spec.Wrap(&v.CurrentSyncCommittee), spec.Wrap(&v.NextSyncCommittee),
		&v.LatestExecutionPayloadHeader,
		&v.NextWithdrawalIndex, &v.NextWithdrawalValidatorIndex,
		spec.Wrap(&v.HistoricalSummaries),
		&v.DepositRequestsStartIndex, &v.DepositBalanceToConsume,
		&v.ExitBalanceToConsume, &v.EarliestExitEpoch,
		&v.ConsolidationBalanceToConsume, &v.EarliestConsolidationEpoch,
		spec.Wrap(&v.PendingDeposits),
		spec.Wrap(&v.PendingPartialWithdrawals),
		spec.Wrap(&v.PendingConsolidations),
	)
}

func (v *BeaconState) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), spec.Wrap(&v.HistoricalRoots),
		&v.Eth1Data, spec.Wrap(&v.Eth1DataVotes), &v.Eth1DepositIndex,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances),
		spec.Wrap(&v.RandaoMixes), spec.Wrap(&v.Slashings),
		spec.Wrap(&v.PreviousEpochParticipation), spec.Wrap(&v.CurrentEpochParticipation),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint,
		spec.Wrap(&v.InactivityScores),
		spec.Wrap(&v.CurrentSyncCommittee), spec.Wrap(&v.NextSyncCommittee),
		&v.LatestExecutionPayloadHeader,
		&v.NextWithdrawalIndex, &v.NextWithdrawalValidatorIndex,
		spec.Wrap(&v.HistoricalSummaries),
		&v.DepositRequestsStartIndex, &v.DepositBalanceToConsume,
		&v.ExitBalanceToConsume, &v.EarliestExitEpoch,
		&v.ConsolidationBalanceToConsume, &v.EarliestConsolidationEpoch,
		spec.Wrap(&v.PendingDeposits),
		spec.Wrap(&v.PendingPartialWithdrawals),
		spec.Wrap(&v.PendingConsolidations),
	)
}

func (v *BeaconState) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), spec.Wrap(&v.HistoricalRoots),
		&v.Eth1Data, spec.Wrap(&v.Eth1DataVotes), &v.Eth1DepositIndex,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances),
		spec.Wrap(&v.RandaoMixes), spec.Wrap(&v.Slashings),
		spec.Wrap(&v.PreviousEpochParticipation), spec.Wrap(&v.CurrentEpochParticipation),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint,
		spec.Wrap(&v.InactivityScores),
		spec.Wrap(&v.CurrentSyncCommittee), spec.Wrap(&v.NextSyncCommittee),
		&v.LatestExecutionPayloadHeader,
		&v.NextWithdrawalIndex, &v.NextWithdrawalValidatorIndex,
		spec.Wrap(&v.HistoricalSummaries),
		&v.DepositRequestsStartIndex, &v.DepositBalanceToConsume,
		&v.ExitBalanceToConsume, &v.EarliestExitEpoch,
		&v.ConsolidationBalanceToConsume, &v.EarliestConsolidationEpoch,
		spec.Wrap(&v.PendingDeposits),
		spec.Wrap(&v.PendingPartialWithdrawals),
		spec.Wrap(&v.PendingConsolidations),
	)
}

func (*BeaconState) FixedLength(*common.Spec) uint64 {
	return 0 // dynamic size
}

func (v *BeaconState) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.GenesisTime, &v.GenesisValidatorsRoot,
		&v.Slot, &v.Fork, &v.LatestBlockHeader,
		spec.Wrap(&v.BlockRoots), spec.Wrap(&v.StateRoots), spec.Wrap(&v.HistoricalRoots),
		&v.Eth1Data, spec.Wrap(&v.Eth1DataVotes), &v.Eth1DepositIndex,
		spec.Wrap(&v.Validators), spec.Wrap(&v.Balances),
		spec.Wrap(&v.RandaoMixes), spec.Wrap(&v.Slashings),
		spec.Wrap(&v.PreviousEpochParticipation), spec.Wrap(&v.CurrentEpochParticipation),
		&v.JustificationBits,
		&v.PreviousJustifiedCheckpoint, &v.CurrentJustifiedCheckpoint,
		&v.FinalizedCheckpoint,
		spec.Wrap(&v.InactivityScores),
		spec.Wrap(&v.CurrentSyncCommittee), spec.Wrap(&v.NextSyncCommittee),
		&v.LatestExecutionPayloadHeader,
		&v.NextWithdrawalIndex, &v.NextWithdrawalValidatorIndex,
		spec.Wrap(&v.HistoricalSummaries),
		&v.DepositRequestsStartIndex, &v.DepositBalanceToConsume,
		&v.ExitBalanceToConsume, &v.EarliestExitEpoch,
		&v.ConsolidationBalanceToConsume, &v.EarliestConsolidationEpoch,
		spec.Wrap(&v.PendingDeposits),
		spec.Wrap(&v.PendingPartialWithdrawals),
		spec.Wrap(&v.PendingConsolidations),
	)
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
	_stateEth1DepositIndex
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
	_inactivityScores
	_currentSyncCommittee
	_nextSyncCommittee
	_latestExecutionPayloadHeader
	_nextWithdrawalIndex
	_nextWithdrawalValidatorIndex
	_historicalSummaries
	_depositRequestsStartIndex
	_depositBalanceToConsume
	_exitBalanceToConsume
	_earliestExitEpoch
	_consolidationBalanceToConsume
	_earliestConsolidationEpoch
	_pendingDeposits
	_pendingPartialWithdrawals
	_pendingConsolidations
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
		{"slashings", phase0.SlashingsType(spec)},
		// Participation
		{"previous_epoch_participation", altair.ParticipationRegistryType(spec)},
		{"current_epoch_participation", altair.ParticipationRegistryType(spec)},
		// Finality
		{"justification_bits", common.JustificationBitsType},
		{"previous_justified_checkpoint", common.CheckpointType},
		{"current_justified_checkpoint", common.CheckpointType},
		{"finalized_checkpoint", common.CheckpointType},
		// Inactivity
		{"inactivity_scores", altair.InactivityScoresType(spec)},
		// Sync
		{"current_sync_committee", common.SyncCommitteeType(spec)},
		{"next_sync_committee", common.SyncCommitteeType(spec)},
		// Execution-layer
		{"latest_execution_payload_header", deneb.ExecutionPayloadHeaderType},
		// Withdrawals
		{"next_withdrawal_index", common.WithdrawalIndexType},
		{"next_withdrawal_validator_index", common.ValidatorIndexType},
		// Deep history valid from Capella onwards
		{"historical_summaries", capella.HistoricalSummariesType(spec)},
		{"deposit_requests_start_index", Uint64Type},
		{"deposit_balance_to_consume", common.GweiType},
		{"exit_balance_to_consume", common.GweiType},
		{"earliest_exit_epoch", common.EpochType},
		{"consolidation_balance_to_consume", common.GweiType},
		{"earliest_consolidation_epoch", common.EpochType},
		{"pending_deposits", common.PendingDepositsType(spec)},
		{"pending_partial_withdrawals", common.PendingPartialWithdrawalsType(spec)},
		{"pending_consolidations", common.PendingConsolidationsType(spec)},
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

func (state *BeaconStateView) Eth1DepositIndex() (common.DepositIndex, error) {
	return common.AsDepositIndex(state.Get(_stateEth1DepositIndex))
}

func (state *BeaconStateView) IncrementDepositIndex() error {
	depIndex, err := state.Eth1DepositIndex()
	if err != nil {
		return err
	}
	return state.Set(_stateEth1DepositIndex, Uint64View(depIndex+1))
}

func (state *BeaconStateView) Validators() (common.ValidatorRegistry, error) {
	return phase0.AsValidatorsRegistry(state.Get(_stateValidators))
}

func (state *BeaconStateView) Balances() (common.BalancesRegistry, error) {
	return phase0.AsRegistryBalances(state.Get(_stateBalances))
}

func (state *BeaconStateView) SetBalances(balances []common.Gwei) error {
	typ := state.Fields[_stateBalances].Type.(*BasicListTypeDef)
	balancesView, err := phase0.Balances(balances).View(typ.ListLimit)
	if err != nil {
		return err
	}
	return state.Set(_stateBalances, balancesView)
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
	bals, err := state.Balances()
	if err != nil {
		return err
	}
	if err := bals.AppendBalance(balance); err != nil {
		return err
	}
	// New in Altair: init participation
	prevPart, err := state.PreviousEpochParticipation()
	if err != nil {
		return err
	}
	if err := prevPart.Append(Uint8View(altair.ParticipationFlags(0))); err != nil {
		return err
	}
	currPart, err := state.CurrentEpochParticipation()
	if err != nil {
		return err
	}
	if err := currPart.Append(Uint8View(altair.ParticipationFlags(0))); err != nil {
		return err
	}
	inActivityScores, err := state.InactivityScores()
	if err != nil {
		return err
	}
	if err := inActivityScores.Append(Uint8View(0)); err != nil {
		return err
	}
	// New in Altair: init inactivity score
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

func (state *BeaconStateView) PreviousEpochParticipation() (*altair.ParticipationRegistryView, error) {
	return altair.AsParticipationRegistry(state.Get(_statePreviousEpochParticipation))
}

func (state *BeaconStateView) CurrentEpochParticipation() (*altair.ParticipationRegistryView, error) {
	return altair.AsParticipationRegistry(state.Get(_stateCurrentEpochParticipation))
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

func (state *BeaconStateView) InactivityScores() (*altair.InactivityScoresView, error) {
	return altair.AsInactivityScores(state.Get(_inactivityScores))
}

func (state *BeaconStateView) CurrentSyncCommittee() (*common.SyncCommitteeView, error) {
	return common.AsSyncCommittee(state.Get(_currentSyncCommittee))
}

func (state *BeaconStateView) SetCurrentSyncCommittee(v *common.SyncCommitteeView) error {
	return state.Set(_currentSyncCommittee, v)
}

func (state *BeaconStateView) NextSyncCommittee() (*common.SyncCommitteeView, error) {
	return common.AsSyncCommittee(state.Get(_nextSyncCommittee))
}

func (state *BeaconStateView) SetNextSyncCommittee(v *common.SyncCommitteeView) error {
	return state.Set(_nextSyncCommittee, v)
}

func (state *BeaconStateView) RotateSyncCommittee(next *common.SyncCommitteeView) error {
	v, err := state.Get(_nextSyncCommittee)
	if err != nil {
		return err
	}
	if err := state.Set(_currentSyncCommittee, v); err != nil {
		return err
	}
	return state.Set(_nextSyncCommittee, next)
}

func (state *BeaconStateView) LatestExecutionPayloadHeader() (*deneb.ExecutionPayloadHeaderView, error) {
	return deneb.AsExecutionPayloadHeader(state.Get(_latestExecutionPayloadHeader))
}

func (state *BeaconStateView) SetLatestExecutionPayloadHeader(h *deneb.ExecutionPayloadHeader) error {
	return state.Set(_latestExecutionPayloadHeader, h.View())
}

func (state *BeaconStateView) NextWithdrawalIndex() (common.WithdrawalIndex, error) {
	v, err := state.Get(_nextWithdrawalIndex)
	return common.AsWithdrawalIndex(v, err)
}

func (state *BeaconStateView) IncrementNextWithdrawalIndex() error {
	nextIndex, err := state.NextWithdrawalIndex()
	if err != nil {
		return err
	}
	return state.Set(_nextWithdrawalIndex, Uint64View(nextIndex+1))
}

func (state *BeaconStateView) SetNextWithdrawalIndex(nextIndex common.WithdrawalIndex) error {
	return state.Set(_nextWithdrawalIndex, Uint64View(nextIndex))
}

func (state *BeaconStateView) NextWithdrawalValidatorIndex() (common.ValidatorIndex, error) {
	v, err := state.Get(_nextWithdrawalValidatorIndex)
	return common.AsValidatorIndex(v, err)
}

func (state *BeaconStateView) SetNextWithdrawalValidatorIndex(nextValidator common.ValidatorIndex) error {
	return state.Set(_nextWithdrawalValidatorIndex, Uint64View(nextValidator))
}

func (state *BeaconStateView) HistoricalSummaries() (capella.HistoricalSummariesList, error) {
	v, err := state.Get(_historicalSummaries)
	return capella.AsHistoricalSummaries(v, err)
}

func (state *BeaconStateView) DepositRequestsStartIndex() (Uint64View, error) {
	v, err := state.Get(_depositRequestsStartIndex)
	return AsUint64(v, err)
}

func (state *BeaconStateView) SetDepositRequestsStartIndex(v Uint64View) error {
	return state.Set(_depositRequestsStartIndex, &v)
}

func (state *BeaconStateView) DepositBalanceToConsume() (common.Gwei, error) {
	v, err := state.Get(_depositBalanceToConsume)
	return common.AsGwei(v, err)
}

func (state *BeaconStateView) SetDepositBalanceToConsume(v common.Gwei) error {
	return state.Set(_depositBalanceToConsume, (*Uint64View)(&v))
}

func (state *BeaconStateView) ExitBalanceToConsume() (common.Gwei, error) {
	v, err := state.Get(_exitBalanceToConsume)
	return common.AsGwei(v, err)
}

func (state *BeaconStateView) SetExitBalanceToConsume(v common.Gwei) error {
	return state.Set(_exitBalanceToConsume, Uint64View(v))
}

func (state *BeaconStateView) EarliestExitEpoch() (common.Epoch, error) {
	v, err := state.Get(_earliestExitEpoch)
	return common.AsEpoch(v, err)
}

func (state *BeaconStateView) SetEarliestExitEpoch(v common.Epoch) error {
	return state.Set(_earliestExitEpoch, Uint64View(v))
}

func (state *BeaconStateView) ConsolidationBalanceToConsume() (common.Gwei, error) {
	v, err := state.Get(_consolidationBalanceToConsume)
	return common.AsGwei(v, err)
}

func (state *BeaconStateView) SetConsolidationBalanceToConsume(v common.Gwei) error {
	return state.Set(_consolidationBalanceToConsume, Uint64View(v))
}

func (state *BeaconStateView) EarliestConsolidationEpoch() (common.Epoch, error) {
	v, err := state.Get(_earliestConsolidationEpoch)
	return common.AsEpoch(v, err)
}

func (state *BeaconStateView) SetEarliestConsolidationEpoch(v common.Epoch) error {
	return state.Set(_earliestConsolidationEpoch, Uint64View(v))
}

// not exposed via getters yet (yet):
//_pendingDeposits
//_pendingPartialWithdrawals
//_pendingConsolidations

func (state *BeaconStateView) ForkSettings(spec *common.Spec) *common.ForkSettings {
	return &common.ForkSettings{
		MinSlashingPenaltyQuotient:     uint64(spec.MIN_SLASHING_PENALTY_QUOTIENT_BELLATRIX),
		ProportionalSlashingMultiplier: uint64(spec.PROPORTIONAL_SLASHING_MULTIPLIER_BELLATRIX),
		InactivityPenaltyQuotient:      uint64(spec.INACTIVITY_PENALTY_QUOTIENT_BELLATRIX),
		CalcProposerShare: func(whistleblowerReward common.Gwei) common.Gwei {
			return whistleblowerReward * altair.PROPOSER_WEIGHT / altair.WEIGHT_DENOMINATOR
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

type ExecutionTrackingBeaconState interface {
	common.BeaconState

	LatestExecutionPayloadHeader() (*deneb.ExecutionPayloadHeaderView, error)
	SetLatestExecutionPayloadHeader(h *deneb.ExecutionPayloadHeader) error
}
