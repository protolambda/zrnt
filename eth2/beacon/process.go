package beacon

import (
	. "github.com/protolambda/zrnt/eth2/beacon/attestations"
	. "github.com/protolambda/zrnt/eth2/beacon/deposits"
	. "github.com/protolambda/zrnt/eth2/beacon/eth1"
	. "github.com/protolambda/zrnt/eth2/beacon/exits"
	. "github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings/attslash"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings/propslash"
	. "github.com/protolambda/zrnt/eth2/beacon/transfers"
	. "github.com/protolambda/zrnt/eth2/core"
)

// TODO: split up
type BlockMeta interface {
	ProcessAttestations(ops []Attestation) error
	ProcessAttestation(attestation *Attestation) error
	ProcessDeposits(ops []Deposit) error
	ProcessDeposit(dep *Deposit) error
	ProcessEth1Vote(data Eth1Data) error
	ProcessVoluntaryExits(ops []VoluntaryExit) error
	ProcessVoluntaryExit(exit *VoluntaryExit) error
	ProcessHeader(header *BeaconBlockHeader) error
	ProcessRandaoReveal(reveal BLSSignature) error
	ProcessAttesterSlashings(ops []AttesterSlashing) error
	ProcessAttesterSlashing(attesterSlashing *AttesterSlashing) error
	ProcessProposerSlashings(ops []ProposerSlashing) error
	ProcessProposerSlashing(ps *ProposerSlashing) error
	ProcessTransfers(ops []Transfer) error
	ProcessTransfer(transfer *Transfer) error
}

type EpochMeta interface {
	ProcessEpochJustification()
	ProcessEpochCrosslinks()
	ProcessEpochRewardsAndPenalties()
	ProcessEpochRegistryUpdates()
	ProcessEpochSlashings()
	ProcessEpochFinalUpdates()
}
