package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/ztyp/props"
	. "github.com/protolambda/ztyp/view"
)


var AttestationDataType = &ContainerType{
	{"slot", SlotType},
	{"index", CommitteeIndexType},
	// LMD GHOST vote
	{"beacon_block_root", RootType},
	// FFG vote
	{"source", CheckpointType},
	{"target", CheckpointType},
}

type AttestationDataNode struct { *ContainerView }

func (ad *AttestationDataNode) RawAttestationData() (res *AttestationData, err error) {
	res = &AttestationData{}
	if res.Slot, err = SlotReadProp(PropReader(ad, 0)).Slot(); err != nil {
		return
	}
	if res.Index, err = CommitteeIndexReadProp(PropReader(ad, 1)).CommitteeIndex(); err != nil {
		return
	}
	if res.BeaconBlockRoot, err = RootReadProp(PropReader(ad, 2)).Root(); err != nil {
		return
	}
	if res.Source, err = CheckpointProp(PropReader(ad, 3)).CheckPoint(); err != nil {
		return
	}
	if res.Target, err = CheckpointProp(PropReader(ad, 4)).CheckPoint(); err != nil {
		return
	}
	return
}


type PendingAttestationNode struct { *ContainerView }

func (pa *PendingAttestationNode) RawPendingAttestation() (res *PendingAttestation, err error) {
	res = &PendingAttestation{}
	// load aggregation bits
	{
		bits, err := BitListReadProp(PropReader(pa, 0)).BitList()
		if err != nil {
			return nil, err
		}
		bitLength, err := bits.Length()
		if err != nil {
			return nil, err
		}
		// rounded up, and then an extra bit for delimiting. ((bitLength + 7 + 1)/ 8)
		byteLength := (bitLength / 8) + 1
		res.AggregationBits = make(CommitteeBits, byteLength, byteLength)
		if err := bits.IntoBytes(res.AggregationBits); err != nil {
			return nil, err
		}
		// add delimiting bit
		res.AggregationBits.SetBit(bitLength, true)
	}
	{
		v, err := ContainerReadProp(PropReader(pa, 1)).Container()
		if err != nil {
			return nil, err
		}
		attDataNode := &AttestationDataNode{ContainerView: v}
		rawAttData, err := attDataNode.RawAttestationData()
		if err != nil {
			return nil, err
		}
		res.Data = *rawAttData
	}
	{
		delay, err := SlotReadProp(PropReader(pa, 2)).Slot()
		if err != nil {
			return nil, err
		}
		res.InclusionDelay = delay
	}
	{
		proposer, err := ValidatorIndexProp(PropReader(pa, 3)).ValidatorIndex()
		if err != nil {
			return nil, err
		}
		res.ProposerIndex = proposer
	}
	return res, nil
}

var PendingAttestationType = &ContainerType{
	{"aggregation_bits", CommitteeBitsType},
	{"data", AttestationDataType},
	{"inclusion_delay", SlotType},
	{"proposer_index", ValidatorIndexType},
}


type EpochPendingAttestations struct { *ListView }

func (ep *EpochPendingAttestations) CollectRawPendingAttestations() ([]*PendingAttestation, error) {
	ll, err := ep.Length()
	if err != nil {
		return nil, err
	}
	out := make([]*PendingAttestation, ll, ll)
	for i := uint64(0); i < ll; i++ {
		v, err := ContainerReadProp(PropReader(ep, i)).Container()
		if err != nil {
			return nil, err
		}
		pan := PendingAttestationNode{ContainerView: v}
		pa, err := pan.RawPendingAttestation()
		if err != nil {
			return nil, err
		}
		out[i] = pa
	}
	return out, nil
}

type EpochPendingAttestationsProp ListReadProp

func (p EpochPendingAttestationsProp) PendingAttestations() (*EpochPendingAttestations, error) {
	v, err := ListReadProp(p).List()
	if err != nil {
		return nil, err
	}
	return &EpochPendingAttestations{ListView: v}, nil
}

var PendingAttestationsType = ListType(PendingAttestationType, uint64(MAX_ATTESTATIONS*SLOTS_PER_EPOCH))

type AttestationsProps struct {
	PreviousEpochAttestations EpochPendingAttestationsProp
	CurrentEpochAttestations  EpochPendingAttestationsProp
}

// Rotate current/previous epoch attestations
func (state *AttestationsProps) RotateEpochAttestations() error {
	prev, err := state.PreviousEpochAttestations.PendingAttestations()
	if err != nil {
		return err
	}
	curr, err := state.CurrentEpochAttestations.PendingAttestations()
	if err != nil {
		return err
	}
	nextPrevV, err := PendingAttestationsType.ViewFromBacking(curr.Backing(), curr.ViewHook)
	if err != nil {
		return err
	}
	nextPrev := &EpochPendingAttestations{ListView: nextPrevV.(*ListView)}
	if err := prev.PropagateChange(nextPrev); err != nil {
		return err
	}
	nextCurr := &EpochPendingAttestations{ListView: PendingAttestationsType.New(curr.ViewHook)}
	if err := curr.PropagateChange(nextCurr); err != nil {
		return err
	}
	return nil
}

func (state *AttestationsProps) CurrentEpochPendingAttestations() ([]*PendingAttestation, error) {
	atts, err := state.CurrentEpochAttestations.PendingAttestations()
	if err != nil {
		return nil, err
	}
	return atts.CollectRawPendingAttestations()
}

func (state *AttestationsProps) PreviousEpochPendingAttestations() ([]*PendingAttestation, error) {
	atts, err := state.PreviousEpochAttestations.PendingAttestations()
	if err != nil {
		return nil, err
	}
	return atts.CollectRawPendingAttestations()
}
