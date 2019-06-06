package beacon

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bitfield"
	"github.com/protolambda/zssz"
)

// NOTE: these containers are going to be moved to sub-packages, per-topic.

type ProposerSlashing struct {
	// Proposer index
	ProposerIndex ValidatorIndex
	// First proposal
	Header1 BeaconBlockHeader
	// Second proposal
	Header2 BeaconBlockHeader
}

type AttesterSlashing struct {
	// First attestation
	Attestation1 IndexedAttestation
	// Second attestation
	Attestation2 IndexedAttestation
}

type Attestation struct {
	// Attester aggregation bitfield
	AggregationBitfield bitfield.Bitfield
	// Attestation data
	Data AttestationData
	// Custody bitfield
	CustodyBitfield bitfield.Bitfield
	// BLS aggregate signature
	Signature BLSSignature
}

type AttestationData struct {
	// Root of the signed beacon block
	BeaconBlockRoot Root

	// FFG vote
	SourceEpoch Epoch
	SourceRoot  Root
	TargetEpoch Epoch
	TargetRoot  Root

	// Crosslink vote
	Shard             Shard
	PreviousCrosslinkRoot Root
	CrosslinkDataRoot Root
}

var AttestationDataAndCustodyBitSSZ = zssz.GetSSZ((*AttestationDataAndCustodyBit)(nil))

type AttestationDataAndCustodyBit struct {
	// Attestation data
	Data AttestationData
	// Custody bit
	CustodyBit bool
}

type IndexedAttestation struct {
	// Validator Indices
	CustodyBit0Indices []ValidatorIndex
	CustodyBit1Indices []ValidatorIndex
	// Attestation data
	Data AttestationData
	// BLS aggregate signature
	Signature BLSSignature
}

var CrosslinkSSZ = zssz.GetSSZ((*Crosslink)(nil))

type Crosslink struct {
	// Epoch number
	Epoch Epoch
	// Root of the previous crosslink
	PreviousCrosslinkRoot Root
	// Root of the crosslinked shard data since the previous crosslink
	CrosslinkDataRoot Root
}

type Deposit struct {
	// Branch in the deposit tree
	Proof [DEPOSIT_CONTRACT_TREE_DEPTH]Root
	// Index in the deposit tree
	Index DepositIndex
	// Data
	Data DepositData
}

var DepositDataSSZ = zssz.GetSSZ((*DepositData)(nil))

type DepositData struct {
	// BLS pubkey
	Pubkey BLSPubkey
	// Withdrawal credentials
	WithdrawalCredentials Root
	// Amount in Gwei
	Amount Gwei
	// Container self-signature
	Signature BLSSignature
}

var VoluntaryExitSSZ = zssz.GetSSZ((*VoluntaryExit)(nil))

type VoluntaryExit struct {
	// Minimum epoch for processing exit
	Epoch Epoch
	// Index of the exiting validator
	ValidatorIndex ValidatorIndex
	// Validator signature
	Signature BLSSignature
}

var TransferSSZ = zssz.GetSSZ((*Transfer)(nil))

type Transfer struct {
	// Sender index
	Sender ValidatorIndex
	// Recipient index
	Recipient ValidatorIndex
	// Amount in Gwei
	Amount Gwei
	// Fee in Gwei for block proposer
	Fee Gwei
	// Inclusion slot
	Slot Slot
	// Sender withdrawal pubkey
	Pubkey BLSPubkey
	// Sender signature
	Signature BLSSignature
}

type Validator struct {
	// BLS public key
	Pubkey BLSPubkey
	// Withdrawal credentials
	WithdrawalCredentials Root
	// Epoch when became eligible for activation
	ActivationEligibilityEpoch Epoch
	// Epoch when validator activated
	ActivationEpoch Epoch
	// Epoch when validator exited
	ExitEpoch Epoch
	// Epoch when validator is eligible to withdraw
	WithdrawableEpoch Epoch
	// Was the validator slashed
	Slashed bool
	// Effective balance
	EffectiveBalance Gwei
}

func (v *Validator) Copy() *Validator {
	copyV := *v
	return &copyV
}

func (v *Validator) IsActive(epoch Epoch) bool {
	return v.ActivationEpoch <= epoch && epoch < v.ExitEpoch
}

func (v *Validator) IsSlashable(epoch Epoch) bool {
	return !v.Slashed && v.ActivationEpoch <= epoch && epoch < v.WithdrawableEpoch
}

type PendingAttestation struct {
	// Attester aggregation bitfield
	AggregationBitfield bitfield.Bitfield
	// Attestation data
	Data AttestationData
	// Inclusion delay
	InclusionDelay Slot
	// Proposer index
	ProposerIndex ValidatorIndex
}

type Fork struct {
	// Previous fork version
	PreviousVersion ForkVersion
	// Current fork version
	CurrentVersion ForkVersion
	// Fork epoch number
	Epoch Epoch
}

type Eth1Data struct {
	// Root of the deposit tree
	DepositRoot Root
	// Total number of deposits
	DepositCount DepositIndex
	// Block hash
	BlockHash Root
}

type ValidatorRegistry []*Validator

func (vr ValidatorRegistry) IsValidatorIndex(index ValidatorIndex) bool {
	return index < ValidatorIndex(len(vr))
}

var ValidatorIndexListSSZ = zssz.GetSSZ((*[]ValidatorIndex)(nil))

func (vr ValidatorRegistry) GetActiveValidatorIndices(epoch Epoch) []ValidatorIndex {
	res := make([]ValidatorIndex, 0, len(vr))
	for i, v := range vr {
		if v.IsActive(epoch) {
			res = append(res, ValidatorIndex(i))
		}
	}
	return res
}

func (vr ValidatorRegistry) GetActiveValidatorCount(epoch Epoch) (count uint64) {
	for _, v := range vr {
		if v.IsActive(epoch) {
			count++
		}
	}
	return
}

var HistoricalBatchSSZ = zssz.GetSSZ((*HistoricalBatch)(nil))

type HistoricalBatch struct {
	BlockRoots [SLOTS_PER_HISTORICAL_ROOT]Root
	StateRoots [SLOTS_PER_HISTORICAL_ROOT]Root
}
