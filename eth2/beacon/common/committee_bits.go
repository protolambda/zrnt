package common

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/protolambda/ztyp/bitfields"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
	"math/bits"
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

func (a *CommitteeBits) FixedLength(*Spec) uint64 {
	return 0
}

func (li CommitteeBits) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.BitListHTR(li, spec.MAX_VALIDATORS_PER_COMMITTEE)
}

func (cb CommitteeBits) MarshalText() ([]byte, error) {
	return conv.BytesMarshalText(cb[:])
}

func (cb *CommitteeBits) UnmarshalText(text []byte) error {
	return bytesUnmarshalText((*[]byte)(cb), text)
}

func bytesUnmarshalText(dst *[]byte, text []byte) error {
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	res := make([]byte, len(text)/2, len(text)/2)
	_, err := hex.Decode(res, text)
	if err != nil {
		return err
	}
	*dst = res
	return nil
}

func (cb CommitteeBits) String() string {
	return conv.BytesString(cb[:])
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
//
// WARNING: unsafe to use, panics if committee size does not match.
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
//
// WARNING: unsafe to use, panics if committee size does not match.
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

// Returns true if other only has bits set to 1 that this bitfield also has set to 1
func (cb CommitteeBits) Covers(other CommitteeBits) (bool, error) {
	if a, b := cb.BitLen(), other.BitLen(); a != b {
		return false, fmt.Errorf("bitfield length mismatch: %d <> %d", a, b)
	}
	// both bitfields have the same delimiter bit set (and same zero padding).
	// We can ignore it, it won't change the outcome.
	for i := 0; i < len(cb); i++ {
		// if there are any bits set in other, that are not also set in cb, then cb does not cover.
		if other[i]&^cb[i] != 0 {
			return false, nil
		}
	}
	return true, nil
}

func (cb CommitteeBits) OnesCount() uint64 {
	if len(cb) == 0 {
		return 0
	}
	count := uint64(0)
	for i := 0; i < len(cb)-1; i++ {
		count += uint64(bits.OnesCount8(cb[i]))
	}
	last := cb[len(cb)-1]
	if last == 0 {
		return count
	}
	// ignore the delimiter bit.
	last ^= uint8(1) << (cb.BitLen() % 8)
	count += uint64(bits.OnesCount8(last))
	return count
}

func (cb CommitteeBits) SingleParticipant(committee []ValidatorIndex) (ValidatorIndex, error) {
	bitLen := cb.BitLen()
	if bitLen != uint64(len(committee)) {
		return 0, fmt.Errorf("committee mismatch, bitfield length %d does not match committee size %d", bitLen, len(committee))
	}
	var found *ValidatorIndex
	for i := uint64(0); i < bitLen; i++ {
		if cb.GetBit(i) {
			if found == nil {
				found = &committee[i]
			} else {
				return 0, fmt.Errorf("found at least two participants: %d and %d", *found, committee[i])
			}
		}
	}
	if found == nil {
		return 0, fmt.Errorf("found no participants")
	}
	return *found, nil
}

func (cb CommitteeBits) Copy() CommitteeBits {
	// append won't find capacity, and thus put contents into new array, and then returns typed slice of it.
	return append(CommitteeBits(nil), cb...)
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

func (c *Phase0Config) CommitteeCount(activeValidators uint64) uint64 {
	validatorsPerSlot := activeValidators / uint64(c.SLOTS_PER_EPOCH)
	committeesPerSlot := validatorsPerSlot / c.TARGET_COMMITTEE_SIZE
	if c.MAX_COMMITTEES_PER_SLOT < committeesPerSlot {
		committeesPerSlot = c.MAX_COMMITTEES_PER_SLOT
	}
	if committeesPerSlot == 0 {
		committeesPerSlot = 1
	}
	return committeesPerSlot
}
