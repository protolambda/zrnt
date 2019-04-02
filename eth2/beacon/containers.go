package beacon

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/protolambda/eth2-shuffle"
	"github.com/protolambda/zrnt/eth2/util/bitfield"
	"github.com/protolambda/zrnt/eth2/util/ssz"
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
	AggregateSignature BLSSignature `ssz:"signature"`
}

type AttestationData struct {
	//  LMD GHOST vote
	Slot Slot
	// Root of the signed beacon block
	BeaconBlockRoot Root

	// FFG vote
	SourceEpoch Epoch
	SourceRoot  Root
	TargetRoot  Root

	// Crosslink vote
	Shard             Shard
	PreviousCrosslink Crosslink
	CrosslinkDataRoot Root
}

type AttestationDataAndCustodyBit struct {
	// Attestation data
	Data AttestationData
	// Custody bit
	CustodyBit bool
}

type IndexedAttestation struct {
	// Validator Indices
	CustodyBit0Indexes []ValidatorIndex
	CustodyBit1Indexes []ValidatorIndex
	// Attestation data
	Data AttestationData
	// BLS aggregate signature
	AggregateSignature BLSSignature `ssz:"signature"`
}

type Crosslink struct {
	// Epoch number
	Epoch Epoch
	// Shard data since the previous crosslink
	CrosslinkDataRoot Root
}

type Deposit struct {
	// Branch in the deposit tree
	Proof [DEPOSIT_CONTRACT_TREE_DEPTH][32]byte
	// Index in the deposit tree
	Index DepositIndex
	// Data
	Data DepositData
}

type DepositData struct {
	// BLS pubkey
	Pubkey BLSPubkey
	// Withdrawal credentials
	WithdrawalCredentials Root
	// Amount in Gwei
	Amount Gwei
	// Container self-signature
	ProofOfPossession BLSSignature `ssz:"signature"`
}

// Let serialized_deposit_data be the serialized form of deposit.deposit_data.
//
// It should equal to:
//  48 bytes for pubkey
//  32 bytes for withdrawal credentials
//  8 bytes for amount
//  96 bytes for proof of possession
//
// This should match deposit_data in the Ethereum 1.0 deposit contract
//  of which the hash was placed into the Merkle tree.
func (d *DepositData) Serialized() []byte {
	depInputBytes := ssz.SSZEncode(d)
	serializedDepositData := make([]byte, 8+8+len(depInputBytes), 8+8+len(depInputBytes))
	binary.LittleEndian.PutUint64(serializedDepositData[0:8], uint64(d.Amount))
	copy(serializedDepositData[8:], depInputBytes)
	return serializedDepositData
}

type VoluntaryExit struct {
	// Minimum epoch for processing exit
	Epoch Epoch
	// Index of the exiting validator
	ValidatorIndex ValidatorIndex
	// Validator signature
	Signature BLSSignature `ssz:"signature"`
}

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
	Signature BLSSignature `ssz:"signature"`
}

type Validator struct {
	// BLS public key
	Pubkey BLSPubkey
	// Withdrawal credentials
	WithdrawalCredentials Root
	// Epoch when validator activated
	ActivationEpoch Epoch
	// Epoch when validator exited
	ExitEpoch Epoch
	// Epoch when validator is eligible to withdraw
	WithdrawableEpoch Epoch
	// Did the validator initiate an exit
	InitiatedExit bool
	// Was the validator slashed
	Slashed bool
	// Rounded balance
	HighBalance Gwei
}

func (v *Validator) IsActive(epoch Epoch) bool {
	return v.ActivationEpoch <= epoch && epoch < v.ExitEpoch
}

func (v *Validator) IsSlashable(epoch Epoch) bool {
	return v.ActivationEpoch <= epoch && epoch < v.WithdrawableEpoch && !v.Slashed
}

type PendingAttestation struct {
	// Attester aggregation bitfield
	AggregationBitfield bitfield.Bitfield
	// Attestation data
	Data AttestationData
	// Custody bitfield
	CustodyBitfield bitfield.Bitfield
	// Inclusion slot
	InclusionSlot Slot
}

// 32 bits, not strictly an integer, hence represented as 4 bytes
// (bytes not necessarily corresponding to versions)
type ForkVersion [4]byte

func Int32ToForkVersion(v uint32) ForkVersion {
	return [4]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
}

type Fork struct {
	// Previous fork version
	PreviousVersion ForkVersion
	// Current fork version
	CurrentVersion ForkVersion
	// Fork epoch number
	Epoch Epoch
}

// Return the fork version of the given epoch
func (f Fork) GetVersion(epoch Epoch) ForkVersion {
	if epoch < f.Epoch {
		return f.PreviousVersion
	}
	return f.CurrentVersion
}

type Eth1Data struct {
	// Root of the deposit tree
	DepositRoot Root
	// Total number of deposits
	DepositCount DepositIndex
	// Block hash
	BlockHash Root
}

type Eth1DataVote struct {
	// Data being voted for
	Eth1Data Eth1Data
	// Vote count
	VoteCount uint64
}

type ValidatorRegistry []Validator

func (vr ValidatorRegistry) IsValidatorIndex(index ValidatorIndex) bool {
	return index < ValidatorIndex(len(vr))
}

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

// Shuffle active validators
func (vr ValidatorRegistry) GetShuffled(seed Bytes32, epoch Epoch) []ValidatorIndex {
	activeValidatorIndices := vr.GetActiveValidatorIndices(epoch)
	committeeCount := GetEpochCommitteeCount(uint64(len(activeValidatorIndices)))
	if committeeCount > uint64(len(activeValidatorIndices)) {
		panic("not enough validators to form committees!")
	}
	// Active validators, shuffled in-place.
	hash := sha256.New()
	hashFn := func(in []byte) []byte {
		hash.Reset()
		hash.Write(in)
		return hash.Sum(nil)
	}
	rawIndexList := make([]uint64, len(vr))
	for i := 0; i < len(activeValidatorIndices); i++ {
		rawIndexList[i] = uint64(activeValidatorIndices[i])
	}
	eth2_shuffle.ShuffleList(hashFn, rawIndexList, SHUFFLE_ROUND_COUNT, seed)
	shuffled := make([]ValidatorIndex, len(vr))
	for i := 0; i < len(activeValidatorIndices); i++ {
		shuffled[i] = ValidatorIndex(rawIndexList[i])
	}
	return shuffled
}

type HistoricalBatch struct {
	BlockRoots [SLOTS_PER_HISTORICAL_ROOT]Root
	StateRoots [SLOTS_PER_HISTORICAL_ROOT]Root
}
