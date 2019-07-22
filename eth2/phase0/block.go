package phase0

import (
	. "github.com/protolambda/zrnt/eth2/beacon/attestations"
	. "github.com/protolambda/zrnt/eth2/beacon/deposits"
	. "github.com/protolambda/zrnt/eth2/beacon/eth1"
	. "github.com/protolambda/zrnt/eth2/beacon/exits"
	. "github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/beacon/randao"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings/attslash"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings/propslash"
	. "github.com/protolambda/zrnt/eth2/beacon/transfers"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type BeaconBlock struct {
	Slot       Slot
	ParentRoot Root
	StateRoot  Root
	Body       BeaconBlockBody
	Signature  BLSSignature
}

func (block *BeaconBlock) Header() *BeaconBlockHeader {
	return &BeaconBlockHeader{
		Slot:       block.Slot,
		ParentRoot: block.ParentRoot,
		StateRoot:  block.StateRoot,
		BodyRoot:   ssz.HashTreeRoot(block.Body, BeaconBlockBodySSZ),
		Signature:  block.Signature,
	}
}

var BeaconBlockBodySSZ = zssz.GetSSZ((*BeaconBlockBody)(nil))

type BeaconBlockBody struct {
	RandaoReveal BLSSignature
	Eth1Data     Eth1Data // Eth1 data vote
	Graffiti     Root     // Arbitrary data

	ProposerSlashings ProposerSlashings
	AttesterSlashings AttesterSlashings
	Attestations      Attestations
	Deposits          Deposits
	VoluntaryExits    VoluntaryExits
	Transfers         Transfers
}

type BlockProcessFeature struct {
	Block *BeaconBlock
	Meta interface {
		HeaderProcessor
		Eth1VoteProcessor
		AttestationProcessor
		DepositProcessor
		VoluntaryExitProcessor
		RandaoProcessor
		AttesterSlashingProcessor
		ProposerSlashingProcessor
		TransferProcessor
	}
}

func (f *BlockProcessFeature) Slot() Slot {
	return f.Block.Slot
}

func (f *BlockProcessFeature) StateRoot() Root {
	return f.Block.StateRoot
}

func (f *BlockProcessFeature) Process() error {
	header := f.Block.Header()
	if err := f.Meta.ProcessHeader(header); err != nil {
		return err
	}
	body := &f.Block.Body
	if err := f.Meta.ProcessRandaoReveal(body.RandaoReveal); err != nil {
		return err
	}
	if err := f.Meta.ProcessEth1Vote(body.Eth1Data); err != nil {
		return err
	}
	if err := f.Meta.ProcessProposerSlashings(body.ProposerSlashings); err != nil {
		return err
	}
	if err := f.Meta.ProcessAttesterSlashings(body.AttesterSlashings); err != nil {
		return err
	}
	if err := f.Meta.ProcessDeposits(body.Deposits); err != nil {
		return err
	}
	if err := f.Meta.ProcessVoluntaryExits(body.VoluntaryExits); err != nil {
		return err
	}
	if err := f.Meta.ProcessTransfers(body.Transfers); err != nil {
		return err
	}
	return nil
}
