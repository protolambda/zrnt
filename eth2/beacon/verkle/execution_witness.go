package verkle

import (
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
)

type Stem [31]byte
type Value []byte
type Suffix byte

func (s *Stem) Deserialize(dr *codec.DecodingReader) error {
	_, err := dr.Read(s[:])
	return err
}

func (*Stem) FixedLength() uint64 {
	return 31
}

func (s *Stem) HashTreeRoot(h tree.HashFn) tree.Root {
	return h.ByteVectorHTR(s[:])
}

func (s *Stem) Serialize(w *codec.EncodingWriter) error {
	return w.Write(s[:])
}

func (s *Stem) ByteLength() uint64 {
	return 31
}

func (s *Suffix) Deserialize(dr *codec.DecodingReader) error {
	b, err := dr.ReadByte()
	*s = Suffix(b)
	return err
}

func (s *Suffix) HashTreeRoot(h tree.HashFn) tree.Root {
	return tree.Root{0: byte(*s)}
}

func (*Suffix) FixedLength() uint64 {
	return 1
}

func (s *Suffix) Serialize(w *codec.EncodingWriter) error {
	return w.WriteByte(byte(*s))
}

func (s *Suffix) ByteLength() uint64 {
	return 1
}

func (ov *Value) Deserialize(dr *codec.DecodingReader) error {
	selector, err := dr.ReadByte()
	if err != nil {
		return err
	}

	if selector == 0 {
		return nil
	}

	*ov = Value(make([]byte, 32))
	n, err := dr.Read((*ov)[:])
	if err != nil {
		return err
	}
	if n != 32 {
		return fmt.Errorf("invalid read length: 32 != %d", n)
	}
	return nil
}

func (ov *Value) ByteLength() uint64 {
	if ov == nil {
		return 1
	}

	return 33
}

func (ov *Value) FixedLength() uint64 {
	return 0
}

func (ov *Value) Serialize(w *codec.EncodingWriter) error {
	if ov == nil || len(*ov) == 0 {
		return w.Union(0, nil)
	}

	err := w.WriteByte(1)
	if err != nil {
		return err
	}

	return w.Write((*ov)[:])
}

func (ov *Value) HashTreeRoot(h tree.HashFn) tree.Root {
	if ov == nil {
		return h.Union(0, nil)
	}

	return h.Union(1, ov)
}

// tree.import[32]byte{0:31}
type SuffixStateDiff struct {
	Suffix       Suffix
	CurrentValue Value
	// Uncomment in post-Kaustinen testnets
	NewValue Value
}

func (ssd *SuffixStateDiff) ByteLength() (out uint64) {
	out = 1
	if ssd.CurrentValue != nil {
		out += 32
	}
	if len(ssd.NewValue) != 0 {
		out += 32
	}
	return
}

func (ssd *SuffixStateDiff) Serialize(w *codec.EncodingWriter) error {
	return w.Container(&ssd.Suffix, &ssd.CurrentValue, &ssd.NewValue)
}

func (ssd *SuffixStateDiff) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&ssd.Suffix, &ssd.CurrentValue, &ssd.NewValue)
}

func (ssd *SuffixStateDiff) FixedLength() uint64 {
	return 0
}

func (ssd *SuffixStateDiff) HashTreeRoot(h tree.HashFn) tree.Root {
	return h.HashTreeRoot(&ssd.Suffix, &ssd.CurrentValue)
}

type SuffixStateDiffs []SuffixStateDiff

func (ssds SuffixStateDiffs) ByteLength() (out uint64) {
	for _, v := range ssds {
		out += v.ByteLength() + codec.OFFSET_SIZE
	}
	return
}

func (ssds SuffixStateDiffs) Serialize(w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &(ssds[i])
	}, 0, uint64(len(ssds)))
}

func (ssds SuffixStateDiffs) FixedLength() uint64 {
	return 0
}

func (ssds *SuffixStateDiffs) Deserialize(dr *codec.DecodingReader) (err error) {
	return dr.List(func() codec.Deserializable {
		i := len(*ssds)
		*ssds = append(*ssds, SuffixStateDiff{})
		return &(*ssds)[i]
	}, 0, VERKLE_WIDTH)
}

func (ssds *SuffixStateDiffs) HashTreeRoot(h tree.HashFn) tree.Root {
	length := uint64(len(*ssds))
	return h.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &(*ssds)[i]
		}
		return nil
	}, length, VERKLE_WIDTH)
}

type StemStateDiff struct {
	Stem        Stem
	SuffixDiffs SuffixStateDiffs
}

func (ssd *StemStateDiff) ByteLength() (out uint64) {
	out = 31
	for _, v := range ssd.SuffixDiffs {
		out += v.ByteLength()
	}
	return
}

func (ssd *StemStateDiff) Serialize(w *codec.EncodingWriter) error {
	return w.Container(&ssd.Stem, &ssd.SuffixDiffs)
}

func (*StemStateDiff) FixedLength() uint64 {
	return 0
}

func (ssd *StemStateDiff) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&ssd.Stem, &ssd.SuffixDiffs)
}

func (sd *StemStateDiff) HashTreeRoot(h tree.HashFn) tree.Root {
	return h.HashTreeRoot(&sd.Stem, &sd.SuffixDiffs)
}

type StateDiff []StemStateDiff

func (sd StateDiff) ByteLength() (out uint64) {
	for _, v := range sd {
		out += v.ByteLength() + codec.OFFSET_SIZE
	}
	return
}

func (sd StateDiff) Serialize(w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &sd[i]
	}, 0, uint64(len(sd)))
}

func (sd *StateDiff) Deserialize(dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*sd)
		*sd = append(*sd, StemStateDiff{})
		return &(*sd)[i]
	}, 0, uint64(MAX_STEMS))
}

func (*StateDiff) FixedLength() uint64 {
	return 0
}

func (sd *StateDiff) HashTreeRoot(h tree.HashFn) tree.Root {
	length := uint64(len(*sd))
	return h.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &(*sd)[i]
		}
		return nil
	}, length, uint64(MAX_STEMS))
}

type Stems []Stem

func (s *Stems) Deserialize(dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*s)
		*s = append(*s, Stem{})
		return &(*s)[i]
	}, 31, MAX_STEMS)
}

func (Stems) FixedLength() uint64 {
	return 0
}

func (s Stems) HashTreeRoot(h tree.HashFn) tree.Root {
	return h.ComplexListHTR(func(i uint64) tree.HTR {
		return &s[i]
	}, 0, uint64(len(s)))
}

func (s Stems) Serialize(w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &s[i]
	}, 31, uint64(len(s)))
}

func (s Stems) ByteLength() uint64 {
	return uint64(len(s)) * (31 /*+ codec.OFFSET_SIZE*/)
}

type BanderwagonGroupElement [32]byte

func (bge *BanderwagonGroupElement) Deserialize(dr *codec.DecodingReader) error {
	dst := make([]byte, 32)
	err := dr.ByteVector(&dst, 32)
	if err != nil {
		return err
	}
	copy(bge[:], dst)
	return nil
}

func (bge *BanderwagonGroupElement) FixedLength() uint64 {
	return 32
}

func (bge *BanderwagonGroupElement) HashTreeRoot(h tree.HashFn) tree.Root {
	return h.ByteVectorHTR(bge[:])
}

func (bge *BanderwagonGroupElement) Serialize(w *codec.EncodingWriter) error {
	return w.Write(bge[:])
}

func (bge *BanderwagonGroupElement) ByteLength() uint64 {
	return 32
}

type BanderwagonGroupElements []BanderwagonGroupElement

func (bge *BanderwagonGroupElements) Deserialize(dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*bge)
		*bge = append(*bge, BanderwagonGroupElement{})
		return &(*bge)[i]
	}, 32, MAX_STEMS)
}

func (bge *BanderwagonGroupElements) FixedLength() uint64 {
	return 0
}

func (bge *BanderwagonGroupElements) HashTreeRoot(h tree.HashFn) tree.Root {
	return h.ComplexVectorHTR(func(i uint64) tree.HTR {
		return &((*bge)[i])
	}, uint64(len(*bge)))
}

func (bge *BanderwagonGroupElements) Serialize(w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable { return &(*bge)[i] }, 32, uint64(len(*bge)))
}

func (bge *BanderwagonGroupElements) ByteLength() uint64 {
	if len(*bge) == 0 {
		return 0
	}
	return uint64(len(*bge)) * ((*bge)[0].ByteLength() + codec.OFFSET_SIZE)
}

type BanderwagonFieldElement [32]byte

func (bfe *BanderwagonFieldElement) Deserialize(dr *codec.DecodingReader) error {
	b := bfe[:]
	return dr.ByteVector(&b, 32)
}

func (bfe *BanderwagonFieldElement) FixedLength() uint64 {
	return 32
}

func (bfe *BanderwagonFieldElement) HashTreeRoot(h tree.HashFn) tree.Root {
	return h.ByteVectorHTR(bfe[:])
}

func (bge *BanderwagonFieldElement) Serialize(w *codec.EncodingWriter) error {
	return w.Write(bge[:])
}

func (bge *BanderwagonFieldElement) ByteLength() uint64 {
	return 32
}

type IPAProofVectors [IPA_PROOF_DEPTH]BanderwagonGroupElement

func (ipv *IPAProofVectors) Deserialize(dr *codec.DecodingReader) error {
	return dr.Vector(func(i uint64) codec.Deserializable {
		return &ipv[i]
	}, 32, uint64(IPA_PROOF_DEPTH))
}

func (ipv *IPAProofVectors) FixedLength() uint64 {
	return IPA_PROOF_DEPTH * 32
}

func (ipv *IPAProofVectors) HashTreeRoot(h tree.HashFn) tree.Root {
	return h.ComplexVectorHTR(func(i uint64) tree.HTR {
		return &ipv[i]
	}, uint64(IPA_PROOF_DEPTH))
}

func (ipv *IPAProofVectors) Serialize(w *codec.EncodingWriter) error {
	return w.Vector(func(i uint64) codec.Serializable {
		return &ipv[i]
	}, uint64(ipv[0].FixedLength()), uint64(len(*ipv)))
}

func (ipv *IPAProofVectors) ByteLength() uint64 {
	return uint64(len(ipv)) * ipv[0].ByteLength()
}

type IPAProof struct {
	CL, CR          IPAProofVectors
	FinalEvaluation BanderwagonFieldElement
}

func (ip *IPAProof) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&ip.CL, &ip.CR, &ip.FinalEvaluation)
}

func (ip *IPAProof) FixedLength() uint64 {
	return 544
}

func (ip *IPAProof) HashTreeRoot(h tree.HashFn) tree.Root {
	return h.HashTreeRoot(&ip.CR, &ip.CL, &ip.FinalEvaluation)
}

func (ip *IPAProof) Serialize(w *codec.EncodingWriter) error {
	return w.Container(&ip.CL, &ip.CR, &ip.FinalEvaluation)
}

func (ip *IPAProof) ByteLength() uint64 {
	return 544
}

type DepthExtensionPresent []byte

func (dep *DepthExtensionPresent) Deserialize(dr *codec.DecodingReader) error {
	return dr.ByteList((*[]byte)(dep), MAX_STEMS)
}

func (DepthExtensionPresent) FixedLength() uint64 {
	return 0
}

func (dep DepthExtensionPresent) HashTreeRoot(h tree.HashFn) tree.Root {
	return h.ByteListHTR(dep[:], MAX_STEMS)
}

func (dep DepthExtensionPresent) Serialize(w *codec.EncodingWriter) error {
	return w.Write(dep)
}

func (dep DepthExtensionPresent) ByteLength() uint64 {
	return uint64(len(dep))
}

type VerkleProof struct {
	OtherStems            Stems
	DepthExtensionPresent DepthExtensionPresent
	CommitmentsByPath     BanderwagonGroupElements
	D                     BanderwagonGroupElement
	IPAProof              IPAProof
}

func (vp *VerkleProof) FixedLength() uint64 {
	return 0
}

func (vp *VerkleProof) ByteLength() uint64 {
	return vp.OtherStems.ByteLength() + vp.DepthExtensionPresent.ByteLength() + vp.CommitmentsByPath.ByteLength() + vp.D.ByteLength() + vp.IPAProof.ByteLength()
}

func (vp *VerkleProof) Serialize(w *codec.EncodingWriter) error {
	return w.Container(&vp.OtherStems, &vp.DepthExtensionPresent, &vp.CommitmentsByPath, &vp.D, &vp.IPAProof)
}

func (vp *VerkleProof) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&vp.OtherStems, &vp.DepthExtensionPresent, &vp.CommitmentsByPath, &vp.D, &vp.IPAProof)
}

func (vp *VerkleProof) HashTreeRoot(h tree.HashFn) tree.Root {
	return h.HashTreeRoot(&vp.OtherStems, vp.DepthExtensionPresent, &vp.CommitmentsByPath, &vp.D, &vp.IPAProof)
}

type ExecutionWitness struct {
	StateDiff   StateDiff
	VerkleProof VerkleProof
}

func (ew *ExecutionWitness) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&ew.StateDiff, &ew.VerkleProof)
}

func (ew *ExecutionWitness) FixedLength() uint64 {
	return 0
}

func (ew *ExecutionWitness) Serialize(w *codec.EncodingWriter) error {
	return w.Container(&ew.StateDiff, &ew.VerkleProof)
}

func (ew *ExecutionWitness) ByteLength() uint64 {
	return ew.StateDiff.ByteLength() + ew.VerkleProof.ByteLength()
}

func (ew *ExecutionWitness) HashTreeRoot(fn tree.HashFn) common.Root {
	return fn.HashTreeRoot(&ew.StateDiff, &ew.VerkleProof)
}
