package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

func PendingShardHeaderType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("PendingShardHeader", []FieldDef{
		{"commitment", DataCommitmentType},
		{"root", RootType},
		{"votes", phase0.AttestationBitsType(spec)},
		{"weight", common.GweiType},
		{"update_slot", common.SlotType},
	})
}

type PendingShardHeaderView struct {
	*ContainerView
}

func AsPendingShardHeader(v View, err error) (*PendingShardHeaderView, error) {
	c, err := AsContainer(v, err)
	return &PendingShardHeaderView{c}, err
}

func (v *PendingShardHeaderView) Commitment() (*DataCommitmentView, error) {
	return AsDataCommitment(v.Get(0))
}

func (v *PendingShardHeaderView) Root() (common.Root, error) {
	return AsRoot(v.Get(1))
}

func (v *PendingShardHeaderView) Votes() (*phase0.AttestationBitsView, error) {
	return phase0.AsAttestationBits(v.Get(2))
}

func (v *PendingShardHeaderView) SetVotes(bits *phase0.AttestationBitsView) error {
	return v.Set(2, bits)
}

func (v *PendingShardHeaderView) Weight() (common.Gwei, error) {
	return common.AsGwei(v.Get(3))
}

func (v *PendingShardHeaderView) SetWeight(w common.Gwei) error {
	return v.Set(3, Uint64View(w))
}

func (v *PendingShardHeaderView) UpdateSlot() (common.Slot, error) {
	return common.AsSlot(v.Get(4))
}

func (v *PendingShardHeaderView) SetUpdateSlot(slot common.Slot) error {
	return v.Set(4, Uint64View(slot))
}

type PendingShardHeader struct {
	// KZG10 commitment to the data
	Commitment DataCommitment `json:"commitment" yaml:"commitment"`
	// hash_tree_root of the ShardHeader (stored so that attestations can be checked against it)
	Root common.Root `json:"root" yaml:"root"`
	// Who voted for the header
	Votes phase0.AttestationBits `json:"votes" yaml:"votes"`
	// Sum of effective balances of votes
	Weight common.Gwei `json:"weight" yaml:"weight"`
	// When the header was last updated, as reference for weight accuracy
	UpdateSlot common.Slot `json:"update_slot" yaml:"update_slot"`
}

func (h *PendingShardHeader) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&h.Commitment, &h.Root, spec.Wrap(&h.Votes), &h.Weight, &h.UpdateSlot)
}

func (h *PendingShardHeader) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&h.Commitment, &h.Root, spec.Wrap(&h.Votes), &h.Weight, &h.UpdateSlot)
}

func (h *PendingShardHeader) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&h.Commitment, &h.Root, spec.Wrap(&h.Votes), &h.Weight, &h.UpdateSlot)
}

func (h *PendingShardHeader) FixedLength(spec *common.Spec) uint64 {
	return 0
}

func (h *PendingShardHeader) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&h.Commitment, &h.Root, spec.Wrap(&h.Votes), &h.Weight, &h.UpdateSlot)
}

func (h *PendingShardHeader) View(spec *common.Spec) *PendingShardHeaderView {
	r := RootView(h.Root)
	psh, _ := AsPendingShardHeader(PendingShardHeaderType(spec).FromFields(
		h.Commitment.View(),
		&r,
		h.Votes.View(spec),
		Uint64View(h.Weight),
		Uint64View(h.UpdateSlot)))
	return psh
}

func PendingShardHeadersType(spec *common.Spec) *ComplexListTypeDef {
	return ComplexListType(PendingShardHeaderType(spec), spec.MAX_SHARD_HEADERS_PER_SHARD)
}

type PendingShardHeadersView struct {
	*ComplexListView
}

func AsPendingShardHeaders(v View, err error) (*PendingShardHeadersView, error) {
	c, err := AsComplexList(v, err)
	return &PendingShardHeadersView{c}, err
}

func (v *PendingShardHeadersView) Header(i uint64) (*PendingShardHeaderView, error) {
	return AsPendingShardHeader(v.Get(i))
}

type PendingShardHeaders []PendingShardHeader

func (hl *PendingShardHeaders) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*hl)
		*hl = append(*hl, PendingShardHeader{})
		return spec.Wrap(&((*hl)[i]))
	}, 0, spec.MAX_SHARD_HEADERS_PER_SHARD)
}

func (hl PendingShardHeaders) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&hl[i])
	}, 0, uint64(len(hl)))
}

func (hl PendingShardHeaders) ByteLength(spec *common.Spec) (out uint64) {
	for _, v := range hl {
		out += v.ByteLength(spec) + codec.OFFSET_SIZE
	}
	return
}

func (hl *PendingShardHeaders) FixedLength(*common.Spec) uint64 {
	return 0
}

func (hl PendingShardHeaders) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(hl))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&hl[i])
		}
		return nil
	}, length, spec.MAX_SHARD_HEADERS_PER_SHARD)
}

func (hl PendingShardHeaders) View(spec *common.Spec) (*PendingShardHeadersView, error) {
	elements := make([]View, len(hl), len(hl))
	for i := 0; i < len(hl); i++ {
		elements[i] = hl[i].View(spec)
	}
	return AsPendingShardHeaders(PendingShardHeadersType(spec).FromElements(elements...))
}
