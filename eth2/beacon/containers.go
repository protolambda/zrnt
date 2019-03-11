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
	// First slashable attestation
	SlashableAttestation1 SlashableAttestation
	// Second slashable attestation
	SlashableAttestation2 SlashableAttestation
}

type SlashableAttestation struct {
	// Validator indices
	ValidatorIndices []ValidatorIndex
	// Attestation data
	Data AttestationData
	// Custody bitfield
	CustodyBitfield bitfield.Bitfield
	// Aggregate signature
	AggregateSignature BLSSignature `ssz:"signature"`
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
	// Slot number
	Slot Slot
	// Shard number
	Shard Shard
	// Root of the signed beacon block
	BeaconBlockRoot Root
	// Root of the ancestor at the epoch boundary
	EpochBoundaryRoot Root
	// Data from the shard since the last attestation
	CrosslinkDataRoot Root
	// Last crosslink
	LatestCrosslink Crosslink
	// Last justified epoch in the beacon state
	JustifiedEpoch Epoch
	// Hash of the last justified beacon block
	JustifiedBlockRoot Root
}

type AttestationDataAndCustodyBit struct {
	// Attestation data
	Data AttestationData
	// Custody bit
	CustodyBit bool
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
	DepositData DepositData
}

type DepositData struct {
	// Amount in Gwei
	Amount Gwei
	// Timestamp from deposit contract
	Timestamp Timestamp
	// Deposit input
	DepositInput DepositInput
}

// Let serialized_deposit_data be the serialized form of deposit.deposit_data.
// It should equal 8 bytes for deposit_data.amount +
//              8 bytes for deposit_data.timestamp +
//              176 bytes for deposit_data.deposit_input
// That is, it should match deposit_data in the Ethereum 1.0 deposit contract
//  of which the hash was placed into the Merkle tree.
func (d *DepositData) Serialized() []byte {
	depInputBytes := ssz.SSZEncode(d.DepositInput)
	serializedDepositData := make([]byte, 8+8+len(depInputBytes), 8+8+len(depInputBytes))
	binary.LittleEndian.PutUint64(serializedDepositData[0:8], uint64(d.Amount))
	binary.LittleEndian.PutUint64(serializedDepositData[8:16], uint64(d.Timestamp))
	copy(serializedDepositData[16:], depInputBytes)
	return serializedDepositData
}

type DepositInput struct {
	// BLS pubkey
	Pubkey BLSPubkey
	// Withdrawal credentials
	WithdrawalCredentials Root
	// A BLS signature of this `DepositInput`
	ProofOfPossession BLSSignature `ssz:"signature"`
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
}

func (v *Validator) IsActive(epoch Epoch) bool {
	return v.ActivationEpoch <= epoch && epoch < v.ExitEpoch
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

type Fork struct {
	// TODO: fork versions are 64 bits, but usage is 32 bits in BLS domain. Spec unclear about it.
	// Previous fork version
	PreviousVersion uint64
	// Current fork version
	CurrentVersion uint64
	// Fork epoch number
	Epoch Epoch
}

// Return the fork version of the given epoch
func (f Fork) GetVersion(epoch Epoch) uint64 {
	if epoch < f.Epoch {
		return f.PreviousVersion
	}
	return f.CurrentVersion
}

type Eth1Data struct {
	// Root of the deposit tree
	DepositRoot Root
	// Block hash
	BlockHash Root
}

type Eth1DataVote struct {
	// Data being voted for
	Eth1Data Eth1Data
	// Vote count
	VoteCount uint64
}

type ValidatorBalances []Gwei

func (balances ValidatorBalances) ApplyStakeDeltas(deltas *Deltas) {
	if len(deltas.Penalties) != len(balances) || len(deltas.Rewards) != len(balances) {
		panic("cannot apply deltas to balances list with different length")
	}
	for i := 0; i < len(balances); i++ {
		balances[i] = Max(
			0,
			balances[i]+deltas.Rewards[i]-deltas.Penalties[i],
		)
	}
}

// Return the effective balance (also known as "balance at stake") for a validator with the given index.
func (balances ValidatorBalances) GetEffectiveBalance(index ValidatorIndex) Gwei {
	return Min(balances[index], MAX_DEPOSIT_AMOUNT)
}

// Return the total balance sum
func (balances ValidatorBalances) GetBalanceSum() (sum Gwei) {
	for i := 0; i < len(balances); i++ {
		sum += balances.GetEffectiveBalance(ValidatorIndex(i))
	}
	return sum
}

// Return the combined effective balance of an array of validators.
func (balances ValidatorBalances) GetTotalBalance(indices []ValidatorIndex) (sum Gwei) {
	for _, vIndex := range indices {
		sum += balances.GetEffectiveBalance(vIndex)
	}
	return sum
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

// Shuffle active validators and split into crosslink committees.
// Return a list of committees (each a list of validator indices).
func (vr ValidatorRegistry) GetShuffling(seed Bytes32, epoch Epoch) [][]ValidatorIndex {
	activeValidatorIndices := vr.GetActiveValidatorIndices(epoch)
	committeeCount := GetEpochCommitteeCount(uint64(len(activeValidatorIndices)))
	if committeeCount > uint64(len(activeValidatorIndices)) {
		panic("not enough validators to form committees!")
	}
	committees := make([][]ValidatorIndex, committeeCount, committeeCount)
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
	committeeSize := uint64(len(rawIndexList)) / committeeCount
	for i := uint64(0); i < committeeCount; i += committeeSize {
		committee := make([]ValidatorIndex, committeeSize)
		for j := uint64(0); j < committeeSize; j++ {
			committee[j] = ValidatorIndex(rawIndexList[i+j])
		}
		committees[i] = committee
	}
	return committees
}

type HistoricalBatch struct {
	BlockRoots [SLOTS_PER_HISTORICAL_ROOT]Root
	StateRoots [SLOTS_PER_HISTORICAL_ROOT]Root
}
