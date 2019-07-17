package core

type AttesterFlag uint64

func (flags AttesterFlag) HasMarkers(markers AttesterFlag) bool {
	return flags&markers == markers
}

const (
	PrevEpochAttester AttesterFlag = 1 << iota
	MatchingHeadAttester
	PrevEpochBoundaryAttester
	CurrEpochBoundaryAttester
	UnslashedAttester
	EligibleAttester
)

type AttesterStatus struct {
	// The delay of inclusion of the latest attestation by the attester.
	// No delay (i.e. 0) by default
	InclusionDelay Slot
	// The validator index of the proposer of the attested beacon block.
	// Only valid if the validator has an attesting flag set.
	AttestedProposer ValidatorIndex
	// A bitfield of markers describing the recent actions of the validator
	Flags AttesterFlag
}
