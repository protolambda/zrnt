package fulu

import (
	"context"
	"fmt"

	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/protolambda/zrnt/eth2/beacon/common"
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
	// Execution-layer
	LatestExecutionPayloadHeader ExecutionPayloadHeader `json:"latest_execution_payload_header" yaml:"latest_execution_payload_header"`
	// Withdrawal indices
	NextWithdrawalIndex     common.WithdrawalIndex `json:"next_withdrawal_index" yaml:"next_withdrawal_index"`
	NextWithdrawalValidatorIndex common.ValidatorIndex  `json:"next_withdrawal_validator_index" yaml:"next_withdrawal_validator_index"`
	// Deep history (frozen in Capella)
	HistoricalSummaries capella.HistoricalSummaries `json:"historical_summaries" yaml:"historical_summaries"`
	// Deposit requests
	DepositReceiptsStartIndex common.DepositIndex `json:"deposit_receipts_start_index" yaml:"deposit_receipts_start_index"`
	// Exits
	EarliestExitEpoch                 common.Epoch `json:"earliest_exit_epoch" yaml:"earliest_exit_epoch"`
	ConsolidationBalanceToConsume     common.Gwei  `json:"consolidation_balance_to_consume" yaml:"consolidation_balance_to_consume"`
	EarliestConsolidationEpoch        common.Epoch `json:"earliest_consolidation_epoch" yaml:"earliest_consolidation_epoch"`
	PendingBalanceDeposits            PendingBalanceDeposits `json:"pending_balance_deposits" yaml:"pending_balance_deposits"`
	PendingPartialWithdrawals         PendingPartialWithdrawals `json:"pending_partial_withdrawals" yaml:"pending_partial_withdrawals"`
	PendingConsolidations             PendingConsolidations `json:"pending_consolidations" yaml:"pending_consolidations"`
	// Fulu
	ProposerLookahead                 ProposerLookahead `json:"proposer_lookahead" yaml:"proposer_lookahead"`
}

type PendingBalanceDeposits = common.PendingDeposits
type PendingPartialWithdrawals = common.PendingPartialWithdrawals  
type PendingConsolidations = common.PendingConsolidations

type ProposerLookahead []common.ValidatorIndex

func (v ProposerLookahead) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(v)
		v = append(v, common.ValidatorIndex(0))
		return &v[i]
	}, 0, (uint64(spec.MIN_SEED_LOOKAHEAD) + 1) * uint64(spec.SLOTS_PER_EPOCH))
}

func (v ProposerLookahead) Serialize(w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return v[i]
	}, 0, uint64(len(v)))
}

func (v ProposerLookahead) ByteLength() (out uint64) {
	return uint64(len(v)) * 8
}

func (v ProposerLookahead) FixedLength() uint64 {
	return 0
}

func (a ProposerLookahead) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(a))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return a[i]
		}
		return nil
	}, length, (uint64(spec.MIN_SEED_LOOKAHEAD) + 1) * uint64(spec.SLOTS_PER_EPOCH))
}

type BeaconStateView struct {
	*ContainerView
}

func NewBeaconStateView(spec *common.Spec) *BeaconStateView {
	return &BeaconStateView{ContainerView: BeaconStateType(spec).New()}
}

func AsBeaconState(v View, err error) (*BeaconStateView, error) {
	c, err := AsContainer(v, err)
	return &BeaconStateView{c}, err
}

func BeaconStateType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("BeaconState", []FieldDef{
		{"genesis_time", Uint64Type},
		{"genesis_validators_root", RootType},
		{"slot", common.SlotType},
		{"fork", common.ForkType},
		{"latest_block_header", common.BeaconBlockHeaderType},
		{"block_roots", phase0.BatchRootsType(spec)},
		{"state_roots", phase0.BatchRootsType(spec)},
			{"historical_roots", phase0.HistoricalRootsType(spec)},
		{"eth1_data", common.Eth1DataType},
		{"eth1_data_votes", phase0.Eth1DataVotesType(spec)},
		{"eth1_deposit_index", Uint64Type},
		{"validators", phase0.ValidatorsRegistryType(spec)},
		{"balances", phase0.RegistryBalancesType(spec)},
		{"randao_mixes", phase0.RandaoMixesType(spec)},
		{"slashings", phase0.SlashingsType(spec)},
		{"previous_epoch_participation", altair.ParticipationRegistryType(spec)},
		{"current_epoch_participation", altair.ParticipationRegistryType(spec)},
		{"justification_bits", common.JustificationBitsType},
		{"previous_justified_checkpoint", common.CheckpointType},
		{"current_justified_checkpoint", common.CheckpointType},
		{"finalized_checkpoint", common.CheckpointType},
		{"inactivity_scores", altair.InactivityScoresType(spec)},
		{"current_sync_committee", common.SyncCommitteeType(spec)},
		{"next_sync_committee", common.SyncCommitteeType(spec)},
		{"latest_execution_payload_header", ExecutionPayloadHeaderType},
		{"next_withdrawal_index", common.WithdrawalIndexType},
		{"next_withdrawal_validator_index", common.ValidatorIndexType},
		{"historical_summaries", capella.HistoricalSummariesType(spec)},
		{"deposit_receipts_start_index", Uint64Type},
		{"earliest_exit_epoch", common.EpochType},
		{"consolidation_balance_to_consume", common.GweiType},
		{"earliest_consolidation_epoch", common.EpochType},
		{"pending_balance_deposits", common.PendingDepositsType(spec)},
		{"pending_partial_withdrawals", common.PendingPartialWithdrawalsType(spec)},
		{"pending_consolidations", common.PendingConsolidationsType(spec)},
		{"proposer_lookahead", ListType(common.ValidatorIndexType, uint64(uint64(spec.MIN_SEED_LOOKAHEAD + 1) * uint64(spec.SLOTS_PER_EPOCH)))},
	})
}

func ProposerLookaheadType(spec *common.Spec) ListTypeDef {
	return ListType(common.ValidatorIndexType, uint64(uint64(spec.MIN_SEED_LOOKAHEAD + 1) * uint64(spec.SLOTS_PER_EPOCH)))
}

func (state *BeaconStateView) GenesisTime() (common.Timestamp, error) {
	return common.AsTimestamp(state.Get(0))
}

func (state *BeaconStateView) SetGenesisTime(t common.Timestamp) error {
	return state.Set(0, Uint64View(t))
}

func (state *BeaconStateView) GenesisValidatorsRoot() (common.Root, error) {
	return AsRoot(state.Get(1))
}

func (state *BeaconStateView) SetGenesisValidatorsRoot(r common.Root) error {
	rv := RootView(r)
	return state.Set(1, &rv)
}

func (state *BeaconStateView) Slot() (common.Slot, error) {
	return common.AsSlot(state.Get(2))
}

func (state *BeaconStateView) SetSlot(slot common.Slot) error {
	return state.Set(2, Uint64View(slot))
}

func (state *BeaconStateView) Fork() (common.Fork, error) {
	fv, err := common.AsFork(state.Get(3))
	if err != nil {
		return common.Fork{}, err
	}
	return fv.Raw()
}

func (state *BeaconStateView) SetFork(f common.Fork) error {
	return state.Set(3, f.View())
}

func (state *BeaconStateView) LatestBlockHeader() (*common.BeaconBlockHeader, error) {
	h, err := common.AsBeaconBlockHeader(state.Get(4))
	if err != nil {
		return nil, err
	}
	return h.Raw()
}

func (state *BeaconStateView) SetLatestBlockHeader(v *common.BeaconBlockHeader) error {
	return state.Set(4, v.View())
}

func (state *BeaconStateView) BlockRoots() (common.BatchRoots, error) {
	return phase0.AsBatchRoots(state.Get(5))
}

func (state *BeaconStateView) StateRoots() (common.BatchRoots, error) {
	return phase0.AsBatchRoots(state.Get(6))
}

func (state *BeaconStateView) HistoricalRoots() (common.HistoricalRoots, error) {
	return phase0.AsHistoricalRoots(state.Get(7))
}

func (state *BeaconStateView) Eth1Data() (common.Eth1Data, error) {
	dat, err := common.AsEth1Data(state.Get(8))
	if err != nil {
		return common.Eth1Data{}, err
	}
	return dat.Raw()
}

func (state *BeaconStateView) SetEth1Data(v common.Eth1Data) error {
	return state.Set(8, v.View())
}

func (state *BeaconStateView) Eth1DataVotes() (common.Eth1DataVotes, error) {
	return phase0.AsEth1DataVotes(state.Get(9))
}

func (state *BeaconStateView) Eth1DepositIndex() (common.DepositIndex, error) {
	return common.AsDepositIndex(state.Get(10))
}

func (state *BeaconStateView) IncrementDepositIndex() error {
	depIndex, err := state.Eth1DepositIndex()
	if err != nil {
		return err
	}
	return state.Set(10, Uint64View(depIndex+1))
}

func (state *BeaconStateView) Validators() (common.ValidatorRegistry, error) {
	return phase0.AsValidatorsRegistry(state.Get(11))
}

func (state *BeaconStateView) Balances() (common.BalancesRegistry, error) {
	return phase0.AsRegistryBalances(state.Get(12))
}

func (state *BeaconStateView) SetBalances(balances []common.Gwei) error {
	typ := state.Fields[12].Type.(*BasicListTypeDef)
	balancesView, err := phase0.Balances(balances).View(typ.ListLimit)
	if err != nil {
		return err
	}
	return state.Set(12, balancesView)
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
	validators, err := phase0.AsValidatorsRegistry(state.Get(11))
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
	return phase0.AsRandaoMixes(state.Get(13))
}

func (state *BeaconStateView) SeedRandao(spec *common.Spec, seed common.Root) error {
	v, err := phase0.SeedRandao(spec, seed)
	if err != nil {
		return err
	}
	return state.Set(13, v)
}

func (state *BeaconStateView) Slashings() (common.Slashings, error) {
	return phase0.AsSlashings(state.Get(14))
}

func (state *BeaconStateView) PreviousEpochParticipation() (*altair.ParticipationRegistryView, error) {
	return altair.AsParticipationRegistry(state.Get(15))
}

func (state *BeaconStateView) CurrentEpochParticipation() (*altair.ParticipationRegistryView, error) {
	return altair.AsParticipationRegistry(state.Get(16))
}

func (state *BeaconStateView) JustificationBits() (common.JustificationBits, error) {
	b, err := common.AsJustificationBits(state.Get(17))
	if err != nil {
		return common.JustificationBits{}, err
	}
	return b.Raw()
}

func (state *BeaconStateView) SetJustificationBits(bits common.JustificationBits) error {
	b, err := common.AsJustificationBits(state.Get(17))
	if err != nil {
		return err
	}
	return b.Set(bits)
}

func (state *BeaconStateView) PreviousJustifiedCheckpoint() (common.Checkpoint, error) {
	c, err := common.AsCheckPoint(state.Get(18))
	if err != nil {
		return common.Checkpoint{}, err
	}
	return c.Raw()
}

func (state *BeaconStateView) SetPreviousJustifiedCheckpoint(c common.Checkpoint) error {
	v, err := common.AsCheckPoint(state.Get(18))
	if err != nil {
		return err
	}
	return v.Set(&c)
}

func (state *BeaconStateView) CurrentJustifiedCheckpoint() (common.Checkpoint, error) {
	c, err := common.AsCheckPoint(state.Get(19))
	if err != nil {
		return common.Checkpoint{}, err
	}
	return c.Raw()
}

func (state *BeaconStateView) SetCurrentJustifiedCheckpoint(c common.Checkpoint) error {
	v, err := common.AsCheckPoint(state.Get(19))
	if err != nil {
		return err
	}
	return v.Set(&c)
}

func (state *BeaconStateView) FinalizedCheckpoint() (common.Checkpoint, error) {
	c, err := common.AsCheckPoint(state.Get(20))
	if err != nil {
		return common.Checkpoint{}, err
	}
	return c.Raw()
}

func (state *BeaconStateView) SetFinalizedCheckpoint(c common.Checkpoint) error {
	v, err := common.AsCheckPoint(state.Get(20))
	if err != nil {
		return err
	}
	return v.Set(&c)
}

func (state *BeaconStateView) InactivityScores() (*altair.InactivityScoresView, error) {
	return altair.AsInactivityScores(state.Get(21))
}

func (state *BeaconStateView) CurrentSyncCommittee() (*common.SyncCommitteeView, error) {
	return common.AsSyncCommittee(state.Get(22))
}

func (state *BeaconStateView) NextSyncCommittee() (*common.SyncCommitteeView, error) {
	return common.AsSyncCommittee(state.Get(23))
}

func (state *BeaconStateView) LatestExecutionPayloadHeader() (*ExecutionPayloadHeaderView, error) {
	return AsExecutionPayloadHeader(state.Get(24))
}

func (state *BeaconStateView) NextWithdrawalIndex() (common.WithdrawalIndex, error) {
	return common.AsWithdrawalIndex(state.Get(25))
}

func (state *BeaconStateView) NextWithdrawalValidatorIndex() (common.ValidatorIndex, error) {
	return common.AsValidatorIndex(state.Get(26))
}

func (state *BeaconStateView) HistoricalSummaries() (*capella.HistoricalSummariesView, error) {
	return capella.AsHistoricalSummaries(state.Get(27))
}

func (state *BeaconStateView) DepositReceiptsStartIndex() (common.DepositIndex, error) {
	return common.AsDepositIndex(state.Get(28))
}

func (state *BeaconStateView) EarliestExitEpoch() (common.Epoch, error) {
	return common.AsEpoch(state.Get(29))
}

func (state *BeaconStateView) ConsolidationBalanceToConsume() (common.Gwei, error) {
	return common.AsGwei(state.Get(30))
}

func (state *BeaconStateView) EarliestConsolidationEpoch() (common.Epoch, error) {
	return common.AsEpoch(state.Get(31))
}

func (state *BeaconStateView) PendingBalanceDeposits() (View, error) {
	return state.Get(32)
}

func (state *BeaconStateView) PendingPartialWithdrawals() (View, error) {
	return state.Get(33)
}

func (state *BeaconStateView) PendingConsolidations() (View, error) {
	return state.Get(34)
}

func (state *BeaconStateView) ProposerLookahead() (View, error) {
	return state.Get(35)
}

func (state *BeaconStateView) HashTreeRoot(fn tree.HashFn) common.Root {
	return fn.HashTreeRoot(state.ContainerView)
}

func (state *BeaconStateView) CopyState() (common.BeaconState, error) {
	return AsBeaconState(state.ContainerView.Copy())
}

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

func (state *BeaconStateView) ProcessEpoch(ctx context.Context, spec *common.Spec, epc *common.EpochsContext) error {
	// TODO: Implement Fulu epoch processing
	return fmt.Errorf("Fulu epoch processing not yet implemented")
}

func (state *BeaconStateView) ProcessBlock(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, benv *common.BeaconBlockEnvelope) error {
	// TODO: Implement Fulu block processing  
	return fmt.Errorf("Fulu block processing not yet implemented")
}

