package fulu

import (
	"bytes"

	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
	"github.com/protolambda/zrnt/eth2/beacon/electra"
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
	}, 0, (spec.MIN_SEED_LOOKAHEAD + 1) * uint64(spec.SLOTS_PER_EPOCH))
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
	}, length, (spec.MIN_SEED_LOOKAHEAD + 1) * uint64(spec.SLOTS_PER_EPOCH))
}

type BeaconStateView struct {
	*ContainerView
}

func NewBeaconStateView(spec *common.Spec) *BeaconStateView {
	return &BeaconStateView{ContainerView: NewContainerView(BeaconStateType(spec))}
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
		{"proposer_lookahead", ProposerLookaheadType(spec)},
	})
}

func ProposerLookaheadType(spec *common.Spec) *ListTypeDef {
	return ListType(common.ValidatorIndexType, (spec.MIN_SEED_LOOKAHEAD + 1) * uint64(spec.SLOTS_PER_EPOCH))
}

func (state *BeaconStateView) GenesisTime() (common.Timestamp, error) {
	return common.AsTimestamp(state.Get(0))
}

func (state *BeaconStateView) GenesisValidatorsRoot() (common.Root, error) {
	return common.AsRoot(state.Get(1))
}

func (state *BeaconStateView) Slot() (common.Slot, error) {
	return common.AsSlot(state.Get(2))
}

func (state *BeaconStateView) Fork() (common.ForkView, error) {
	return common.AsFork(state.Get(3))
}

func (state *BeaconStateView) LatestBlockHeader() (common.BeaconBlockHeaderView, error) {
	return common.AsBeaconBlockHeader(state.Get(4))
}

func (state *BeaconStateView) BlockRoots() (*ListView, error) {
	return AsList(state.Get(5))
}

func (state *BeaconStateView) StateRoots() (*ListView, error) {
	return AsList(state.Get(6))
}

func (state *BeaconStateView) HistoricalRoots() (phase0.HistoricalRootsView, error) {
	return phase0.AsHistoricalRoots(state.Get(7))
}

func (state *BeaconStateView) Eth1Data() (common.Eth1DataView, error) {
	return common.AsEth1Data(state.Get(8))
}

func (state *BeaconStateView) Eth1DataVotes() (phase0.Eth1DataVotesView, error) {
	return phase0.AsEth1DataVotes(state.Get(9))
}

func (state *BeaconStateView) Eth1DepositIndex() (common.DepositIndex, error) {
	return common.AsDepositIndex(state.Get(10))
}

func (state *BeaconStateView) Validators() (*ListView, error) {
	return AsList(state.Get(11))
}

func (state *BeaconStateView) Balances() (*ListView, error) {
	return AsList(state.Get(12))
}

func (state *BeaconStateView) RandaoMixes() (phase0.RandaoMixesView, error) {
	return phase0.AsRandaoMixes(state.Get(13))
}

func (state *BeaconStateView) Slashings() (*ListView, error) {
	return AsList(state.Get(14))
}

func (state *BeaconStateView) PreviousEpochParticipation() (altair.ParticipationRegistryView, error) {
	return altair.AsParticipationRegistry(state.Get(15))
}

func (state *BeaconStateView) CurrentEpochParticipation() (altair.ParticipationRegistryView, error) {
	return altair.AsParticipationRegistry(state.Get(16))
}

func (state *BeaconStateView) JustificationBits() (common.JustificationBits, error) {
	return common.AsJustificationBits(state.Get(17))
}

func (state *BeaconStateView) PreviousJustifiedCheckpoint() (common.CheckpointView, error) {
	return common.AsCheckpoint(state.Get(18))
}

func (state *BeaconStateView) CurrentJustifiedCheckpoint() (common.CheckpointView, error) {
	return common.AsCheckpoint(state.Get(19))
}

func (state *BeaconStateView) FinalizedCheckpoint() (common.CheckpointView, error) {
	return common.AsCheckpoint(state.Get(20))
}

func (state *BeaconStateView) InactivityScores() (altair.InactivityScoresView, error) {
	return altair.AsInactivityScores(state.Get(21))
}

func (state *BeaconStateView) CurrentSyncCommittee() (common.SyncCommitteeView, error) {
	return common.AsSyncCommittee(state.Get(22))
}

func (state *BeaconStateView) NextSyncCommittee() (common.SyncCommitteeView, error) {
	return common.AsSyncCommittee(state.Get(23))
}

func (state *BeaconStateView) LatestExecutionPayloadHeader() (ExecutionPayloadHeaderView, error) {
	return AsExecutionPayloadHeader(state.Get(24))
}

func (state *BeaconStateView) NextWithdrawalIndex() (common.WithdrawalIndex, error) {
	return common.AsWithdrawalIndex(state.Get(25))
}

func (state *BeaconStateView) NextWithdrawalValidatorIndex() (common.ValidatorIndex, error) {
	return common.AsValidatorIndex(state.Get(26))
}

func (state *BeaconStateView) HistoricalSummaries() (capella.HistoricalSummariesView, error) {
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

func (state *BeaconStateView) PendingBalanceDeposits() (*ListView, error) {
	return AsList(state.Get(32))
}

func (state *BeaconStateView) PendingPartialWithdrawals() (*ListView, error) {
	return AsList(state.Get(33))
}

func (state *BeaconStateView) PendingConsolidations() (*ListView, error) {
	return AsList(state.Get(34))
}

func (state *BeaconStateView) ProposerLookahead() (ProposerLookaheadView, error) {
	return AsProposerLookahead(state.Get(35))
}


type ProposerLookaheadView struct{ *ListView }

func AsProposerLookahead(v View, err error) (*ProposerLookaheadView, error) {
	c, err := AsList(v, err)
	return &ProposerLookaheadView{c}, err
}

func (v *ProposerLookaheadView) Get(i uint64) (common.ValidatorIndex, error) {
	return common.AsValidatorIndex(v.ListView.Get(i))
}

func (v *ProposerLookaheadView) Set(i uint64, val common.ValidatorIndex) error {
	return v.ListView.Set(i, Uint64View(val))
}

