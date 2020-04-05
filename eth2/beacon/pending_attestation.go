package beacon

import (
	"bytes"
	"fmt"

	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

type PendingAttestation struct {
	AggregationBits CommitteeBits
	Data            AttestationData
	InclusionDelay  Slot
	ProposerIndex   ValidatorIndex
}

var AttestationDataSSZ = zssz.GetSSZ((*AttestationData)(nil))

type AttestationData struct {
	Slot  Slot
	Index CommitteeIndex

	// LMD GHOST vote
	BeaconBlockRoot Root

	// FFG vote
	Source Checkpoint
	Target Checkpoint
}


var AttestationDataType = ContainerType("AttestationData", []FieldDef{
	{"slot", SlotType},
	{"index", CommitteeIndexType},
	// LMD GHOST vote
	{"beacon_block_root", RootType},
	// FFG vote
	{"source", CheckpointType},
	{"target", CheckpointType},
})

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
		bits, err := BitListProp(PropReader(pa, 0)).BitList()
		if err != nil {
			return nil, err
		}
		bitLength, err := bits.Length()
		if err != nil {
			return nil, err
		}
		// rounded up, and then an extra bit for delimiting. ((bitLength + 7 + 1)/ 8)
		byteLength := (bitLength / 8) + 1
		var buf bytes.Buffer
		if err := bits.Serialize(&buf); err != nil {
			return nil, err
		}
		res.AggregationBits = buf.Bytes()
		if uint64(len(res.AggregationBits)) != byteLength {
			return nil, fmt.Errorf("failed to convert raw attestation bitfield to pending att node bits")
		}
	}
	{
		v, err := ContainerProp(PropReader(pa, 1)).Container()
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

var PendingAttestationType = ContainerType("PendingAttestation", []FieldDef{
	{"aggregation_bits", CommitteeBitsType},
	{"data", AttestationDataType},
	{"inclusion_delay", SlotType},
	{"proposer_index", ValidatorIndexType},
})


type EpochPendingAttestations struct { *ComplexListView }

func (ep *EpochPendingAttestations) CollectRawPendingAttestations() ([]*PendingAttestation, error) {
	ll, err := ep.Length()
	if err != nil {
		return nil, err
	}
	out := make([]*PendingAttestation, ll, ll)
	for i := uint64(0); i < ll; i++ {
		v, err := ContainerProp(PropReader(ep, i)).Container()
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

type EpochPendingAttestationsProp ComplexListProp

func (p EpochPendingAttestationsProp) PendingAttestations() (*EpochPendingAttestations, error) {
	v, err := ComplexListProp(p).List()
	if err != nil {
		return nil, err
	}
	return &EpochPendingAttestations{ComplexListView: v}, nil
}

var PendingAttestationsType = ComplexListType(PendingAttestationType, uint64(MAX_ATTESTATIONS*SLOTS_PER_EPOCH))

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
	nextPrevV, err := PendingAttestationsType.ViewFromBacking(curr.Backing(), prev.Hook)
	if err != nil {
		return err
	}
	nextCurrV := PendingAttestationsType.Default(curr.Hook)

	nextPrev := &EpochPendingAttestations{ComplexListView: nextPrevV.(*ComplexListView)}
	if err := prev.SetBacking(nextPrev.Backing()); err != nil {
		return err
	}
	nextCurr := &EpochPendingAttestations{ComplexListView: nextCurrV.(*ComplexListView)}
	if err := curr.SetBacking(nextCurr.Backing()); err != nil {
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
