package beacon

type AttesterFlag uint8

func (flags AttesterFlag) HasMarkers(markers AttesterFlag) bool {
	return flags&markers == markers
}

const (
	PrevSourceAttester AttesterFlag = 1 << iota
	PrevTargetAttester
	PrevHeadAttester

	CurrSourceAttester
	CurrTargetAttester
	CurrHeadAttester

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
	// Effective balance
	EffectiveBalance Gwei
	// If the validator is active
	Active bool
	// Exit epoch
	ExitEpoch Epoch
	ActivationEligibilityEpoch Epoch
	ActivationEpoch Epoch
}
