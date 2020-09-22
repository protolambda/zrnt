package beacon

import (
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type PendingAttestation struct {
	AggregationBits CommitteeBits
	Data            AttestationData
	InclusionDelay  Slot
	ProposerIndex   ValidatorIndex
}

func (a *PendingAttestation) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&a.AggregationBits), &a.Data, &a.InclusionDelay, &a.ProposerIndex)
}

func (a *PendingAttestation) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&a.AggregationBits), &a.Data, a.InclusionDelay, a.ProposerIndex)
}

func (a *PendingAttestation) ByteLength(spec *Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&a.AggregationBits), &a.Data, a.InclusionDelay, a.ProposerIndex)
}

func (a *PendingAttestation) FixedLength(*Spec) uint64 {
	return 0
}

func (a *PendingAttestation) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(spec.Wrap(&a.AggregationBits), &a.Data, a.InclusionDelay, a.ProposerIndex)
}

func (att *PendingAttestation) View(spec *Spec) *PendingAttestationView {
	bits := att.AggregationBits.View(spec)
	t := ContainerType("PendingAttestation", []FieldDef{
		{"aggregation_bits", bits.Type()},
		{"data", AttestationDataType},
		{"inclusion_delay", SlotType},
		{"proposer_index", ValidatorIndexType},
	})
	c, _ := t.FromFields(
		bits,
		att.Data.View(),
		Uint64View(att.InclusionDelay),
		Uint64View(att.ProposerIndex),
	)
	return &PendingAttestationView{c}
}

type AttestationData struct {
	Slot  Slot
	Index CommitteeIndex

	// LMD GHOST vote
	BeaconBlockRoot Root

	// FFG vote
	Source Checkpoint
	Target Checkpoint
}

func (a *AttestationData) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&a.Slot, &a.Index, &a.BeaconBlockRoot, &a.Source, &a.Target)
}

func (a *AttestationData) Serialize(w *codec.EncodingWriter) error {
	return w.Container(a.Slot, a.Index, &a.BeaconBlockRoot, &a.Source, &a.Target)
}

func (a *AttestationData) ByteLength() uint64 {
	return AttestationDataType.TypeByteLength()
}

func (*AttestationData) FixedLength() uint64 {
	return AttestationDataType.TypeByteLength()
}

func (p *AttestationData) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(p.Slot, p.Index, p.BeaconBlockRoot, &p.Source, &p.Target)
}

func (data *AttestationData) View() *AttestationDataView {
	rv := RootView(data.BeaconBlockRoot)
	c, _ := AttestationDataType.FromFields(
		Uint64View(data.Slot),
		Uint64View(data.Index),
		&rv,
		data.Source.View(),
		data.Target.View(),
	)
	return &AttestationDataView{c}
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

type AttestationDataView struct{ *ContainerView }

func (v *AttestationDataView) Raw() (*AttestationData, error) {
	fields, err := v.FieldValues()
	slot, err := AsSlot(fields[0], err)
	comm, err := AsCommitteeIndex(fields[1], err)
	root, err := AsRoot(fields[2], err)
	source, err := AsCheckPoint(fields[3], err)
	target, err := AsCheckPoint(fields[4], err)
	if err != nil {
		return nil, err
	}
	rawSource, err := source.Raw()
	if err != nil {
		return nil, err
	}
	rawTarget, err := target.Raw()
	if err != nil {
		return nil, err
	}
	return &AttestationData{
		Slot:            slot,
		Index:           comm,
		BeaconBlockRoot: root,
		Source:          rawSource,
		Target:          rawTarget,
	}, nil
}

func AsAttestationData(v View, err error) (*AttestationDataView, error) {
	c, err := AsContainer(v, err)
	return &AttestationDataView{c}, err
}

func (c *Phase0Config) PendingAttestation() *ContainerTypeDef {
	return ContainerType("PendingAttestation", []FieldDef{
		{"aggregation_bits", c.CommitteeBits()},
		{"data", AttestationDataType},
		{"inclusion_delay", SlotType},
		{"proposer_index", ValidatorIndexType},
	})
}

type PendingAttestationView struct{ *ContainerView }

func (v *PendingAttestationView) Raw() (*PendingAttestation, error) {
	// load aggregation bits
	fields, err := v.FieldValues()
	bits, err := AsCommitteeBits(fields[0], err)
	data, err := AsAttestationData(fields[1], err)
	delay, err := AsSlot(fields[2], err)
	proposerIndex, err := AsValidatorIndex(fields[3], err)
	if err != nil {
		return nil, err
	}
	rawBits, err := bits.Raw()
	if err != nil {
		return nil, err
	}
	rawData, err := data.Raw()
	if err != nil {
		return nil, err
	}
	return &PendingAttestation{
		AggregationBits: rawBits,
		Data:            *rawData,
		InclusionDelay:  delay,
		ProposerIndex:   proposerIndex,
	}, nil
}

func AsPendingAttestation(v View, err error) (*PendingAttestationView, error) {
	c, err := AsContainer(v, err)
	return &PendingAttestationView{c}, err
}

type PendingAttestations []*PendingAttestation

func (a *PendingAttestations) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, &PendingAttestation{})
		return spec.Wrap((*a)[i])
	}, 0, spec.MAX_ATTESTATIONS*uint64(spec.SLOTS_PER_EPOCH))
}

func (a PendingAttestations) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(a[i])
	}, 0, uint64(len(a)))
}

func (p PendingAttestations) ByteLength(spec *Spec) (out uint64) {
	for _, a := range p {
		out += a.ByteLength(spec) + codec.OFFSET_SIZE
	}
	return
}

func (a *PendingAttestations) FixedLength(*Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li PendingAttestations) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(li[i])
		}
		return nil
	}, length, spec.MAX_ATTESTATIONS*uint64(spec.SLOTS_PER_EPOCH))
}

func (c *Phase0Config) PendingAttestations() ListTypeDef {
	return ComplexListType(c.PendingAttestation(), c.MAX_ATTESTATIONS*uint64(c.SLOTS_PER_EPOCH))
}

type PendingAttestationsView struct{ *ComplexListView }

func AsPendingAttestations(v View, err error) (*PendingAttestationsView, error) {
	c, err := AsComplexList(v, err)
	return &PendingAttestationsView{c}, err
}
