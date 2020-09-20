package beacon

import (
	"bytes"
	"fmt"
	"github.com/protolambda/zssz/bitfields"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

// CommitteeBits is formatted as a serialized SSZ bitlist, including the delimit bit
type CommitteeBits []byte

func (c *Phase0Config) View(cb CommitteeBits) *CommitteeBitsView {
	v, _ := c.CommitteeBits().Deserialize(codec.NewDecodingReader(bytes.NewReader(cb), uint64(len(cb))))
	return &CommitteeBitsView{v.(*BitListView)}
}

func (li CommitteeBits) View(spec *Spec) *CommitteeBitsView {
	v, _ := BitListType(spec.MAX_VALIDATORS_PER_COMMITTEE).Deserialize(
		codec.NewDecodingReader(bytes.NewReader(li), uint64(len(li))))
	return &CommitteeBitsView{v.(*BitListView)}
}

func (li *CommitteeBits) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.BitList((*[]byte)(li), spec.MAX_VALIDATORS_PER_COMMITTEE)
}

func (a CommitteeBits) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.BitList(a[:])
}

func (a CommitteeBits) ByteLength(spec *Spec) uint64 {
	return uint64(len(a))
}

func (a *CommitteeBits) FixedLength() uint64 {
	return 0
}

func (li CommitteeBits) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.BitListHTR(li, spec.MAX_VALIDATORS_PER_COMMITTEE)
}

func (cb CommitteeBits) BitLen() uint64 {
	return bitfields.BitlistLen(cb)
}

func (cb CommitteeBits) GetBit(i uint64) bool {
	return bitfields.GetBit(cb, i)
}

func (cb CommitteeBits) SetBit(i uint64, v bool) {
	bitfields.SetBit(cb, i, v)
}

// Sets the bits to true that are true in other. (in place)
func (cb CommitteeBits) Or(other CommitteeBits) {
	for i := 0; i < len(cb); i++ {
		cb[i] |= other[i]
	}
}

// In-place filters a list of committees indices to only keep the bitfield participants.
// The result is not sorted. Returns the re-sliced filtered participants list.
func (cb CommitteeBits) FilterParticipants(committee []ValidatorIndex) []ValidatorIndex {
	bitLen := cb.BitLen()
	out := committee[:0]
	if bitLen != uint64(len(committee)) {
		panic("committee mismatch, bitfield length does not match")
	}
	for i := uint64(0); i < bitLen; i++ {
		if cb.GetBit(i) {
			out = append(out, committee[i])
		}
	}
	return out
}

// In-place filters a list of committees indices to only keep the bitfield NON-participants.
// The result is not sorted. Returns the re-sliced filtered non-participants list.
func (cb CommitteeBits) FilterNonParticipants(committee []ValidatorIndex) []ValidatorIndex {
	bitLen := cb.BitLen()
	out := committee[:0]
	if bitLen != uint64(len(committee)) {
		panic("committee mismatch, bitfield length does not match")
	}
	for i := uint64(0); i < bitLen; i++ {
		if !cb.GetBit(i) {
			out = append(out, committee[i])
		}
	}
	return out
}

func (c *Phase0Config) CommitteeBits() *BitListTypeDef {
	return BitListType(c.MAX_VALIDATORS_PER_COMMITTEE)
}

type CommitteeBitsView struct {
	*BitListView
}

func AsCommitteeBits(v View, err error) (*CommitteeBitsView, error) {
	c, err := AsBitList(v, err)
	return &CommitteeBitsView{c}, err
}

func (v *CommitteeBitsView) Raw() (CommitteeBits, error) {
	bitLength, err := v.Length()
	if err != nil {
		return nil, err
	}
	// rounded up, and then an extra bit for delimiting. ((bitLength + 7 + 1)/ 8)
	byteLength := (bitLength / 8) + 1
	var buf bytes.Buffer
	if err := v.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	out := CommitteeBits(buf.Bytes())
	if uint64(len(out)) != byteLength {
		return nil, fmt.Errorf("failed to convert attestation tree bits view to raw bits")
	}
	return out, nil
}
