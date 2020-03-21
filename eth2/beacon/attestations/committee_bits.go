package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zssz/bitfields"
	. "github.com/protolambda/ztyp/view"
)

type CommitteeBits []byte

var CommitteeBitsType = BitlistType(MAX_VALIDATORS_PER_COMMITTEE)

func (cb CommitteeBits) BitLen() uint64 {
	return bitfields.BitlistLen(cb)
}

func (cb CommitteeBits) GetBit(i uint64) bool {
	return bitfields.GetBit(cb, i)
}

func (cb CommitteeBits) SetBit(i uint64, v bool) {
	bitfields.SetBit(cb, i, v)
}

func (cb *CommitteeBits) Limit() uint64 {
	return MAX_VALIDATORS_PER_COMMITTEE
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
