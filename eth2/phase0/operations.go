package phase0

import (
	. "github.com/protolambda/zrnt/eth2/beacon/attestations"
	. "github.com/protolambda/zrnt/eth2/beacon/deposits"
	. "github.com/protolambda/zrnt/eth2/beacon/exits"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings/attslash"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings/propslash"
	. "github.com/protolambda/zrnt/eth2/beacon/transfers"
	. "github.com/protolambda/zrnt/eth2/core"
)

type ProposerSlashings []ProposerSlashing

func (*ProposerSlashings) Limit() uint32 {
	return MAX_PROPOSER_SLASHINGS
}

type AttesterSlashings []AttesterSlashing

func (*AttesterSlashings) Limit() uint32 {
	return MAX_ATTESTER_SLASHINGS
}

type Attestations []Attestation

func (*Attestations) Limit() uint32 {
	return MAX_ATTESTATIONS
}

type Deposits []Deposit

func (*Deposits) Limit() uint32 {
	return MAX_DEPOSITS
}

type Transfers []Transfer

func (*Transfers) Limit() uint32 {
	return MAX_TRANSFERS
}

type VoluntaryExits []VoluntaryExit

func (*VoluntaryExits) Limit() uint32 {
	return MAX_VOLUNTARY_EXITS
}
