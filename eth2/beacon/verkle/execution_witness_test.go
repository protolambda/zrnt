package verkle

import (
	"bytes"
	"testing"

	"github.com/protolambda/ztyp/codec"
)

var testStem = Stem([31]byte{0: 1, 2: 3, 4: 5, 30: 0xff})
var testValueBytes = [32]byte{0: 0xaa, 7: 8, 23: 0xde}
var testValue = Value(testValueBytes[:])
var testValueNewBytes = [32]byte{0: 0xaa, 7: 8, 23: 0xde, 30: 42}
var testValueNew = Value(testValueNewBytes[:])

func TestStemSerialization(t *testing.T) {
	var buf bytes.Buffer
	if err := testStem.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(buf.Bytes()[:31], testStem[:]) {
		t.Fatalf("invalid output %x != %x", buf.Bytes(), testStem[:])
	}
}

func TestStemSerde(t *testing.T) {
	var buf bytes.Buffer
	if err := testStem.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		t.Fatal(err)
	}

	var newStem Stem
	if err := newStem.Deserialize(codec.NewDecodingReader(&buf, uint64(buf.Len()))); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(newStem[:], testStem[:]) {
		t.Fatalf("invalid deserialized stem %x != %x", newStem, testStem)
	}
}

func TestValueSerde(t *testing.T) {
	var buf bytes.Buffer
	if err := testValue.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		t.Fatal(err)
	}
	t.Logf("serialized=%x", buf.Bytes())

	var newValue Value
	if err := newValue.Deserialize(codec.NewDecodingReader(&buf, uint64(buf.Len()))); err != nil {
		t.Fatal(err)
	}

	t.Logf("%v", newValue)
	if !bytes.Equal(newValue[:], testValue[:]) {
		t.Fatalf("invalid deserialized stem %x != %x", newValue, testValue)
	}
}

func TestSuffixStateDiffSerde(t *testing.T) {
	var buf bytes.Buffer
	sd := &SuffixStateDiff{
		Suffix:       0xde,
		CurrentValue: testValue,
	}
	if err := sd.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		t.Fatal(err)
	}
	var newssd SuffixStateDiff
	if err := newssd.Deserialize(codec.NewDecodingReader(&buf, uint64(buf.Len()))); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(newssd.CurrentValue[:], sd.CurrentValue[:]) {
		t.Fatalf("invalid deserialized current value %x != %x", newssd.CurrentValue, sd.CurrentValue)
	}

	if newssd.Suffix != sd.Suffix {
		t.Fatalf("Differing suffixes %x != %x", newssd.Suffix, sd.Suffix)
	}

	// Same thing with a non-nil NewValue
	sd = &SuffixStateDiff{
		Suffix:       0xde,
		CurrentValue: testValue,
		NewValue:     testValueNew,
	}
	if err := sd.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		t.Fatal(err)
	}
	t.Logf("serialized=%x", buf.Bytes())
	if err := newssd.Deserialize(codec.NewDecodingReader(&buf, uint64(buf.Len()))); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(newssd.CurrentValue[:], sd.CurrentValue[:]) {
		t.Fatalf("invalid deserialized current value %x != %x", newssd.CurrentValue, sd.CurrentValue)
	}

	if newssd.Suffix != sd.Suffix {
		t.Fatalf("Differing suffixes %x != %x", newssd.Suffix, sd.Suffix)
	}
}

func TestStemStateDiffSerde(t *testing.T) {
	var buf bytes.Buffer
	sd := &StemStateDiff{
		Stem: testStem,
		SuffixDiffs: []SuffixStateDiff{{
			Suffix:       0xde,
			CurrentValue: testValue,
		}},
	}
	if err := sd.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		t.Fatal(err)
	}
	var newssd StemStateDiff
	if err := newssd.Deserialize(codec.NewDecodingReader(&buf, uint64(buf.Len()))); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(sd.Stem[:], newssd.Stem[:]) {
		t.Fatalf("invalid deserialized stem %x != %x", sd.Stem, newssd.Stem)
	}

	if len(newssd.SuffixDiffs) != 1 {
		t.Fatalf("invalid length for suffix diffs %d != 1", len(newssd.SuffixDiffs))
	}

	if !bytes.Equal(newssd.SuffixDiffs[0].CurrentValue[:], sd.SuffixDiffs[0].CurrentValue[:]) {
		t.Fatalf("invalid deserialized current value %x != %x", newssd.SuffixDiffs[0].CurrentValue, sd.SuffixDiffs[0].CurrentValue)
	}

	if newssd.SuffixDiffs[0].Suffix != sd.SuffixDiffs[0].Suffix {
		t.Fatalf("Differing suffixes %x != %x", newssd.SuffixDiffs[0].Suffix, sd.SuffixDiffs[0].Suffix)
	}
}
func TestIPAProofSerde(t *testing.T) {
	var buf bytes.Buffer
	ipp := &IPAProof{}
	if err := ipp.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		t.Fatal(err)
	}
	var newipp IPAProof
	if err := newipp.Deserialize(codec.NewDecodingReader(&buf, uint64(buf.Len()))); err != nil {
		t.Fatal(err)
	}
}

func TestVerkleProofSerde(t *testing.T) {
	var buf bytes.Buffer
	vp := &VerkleProof{
		OtherStems:            Stems{{0: 0x33, 30: 0x44}, {0: 0x55, 30: 0x66}},
		DepthExtensionPresent: []byte{1, 2, 3},
		CommitmentsByPath: []BanderwagonGroupElement{
			[32]byte{0: 0x77, 31: 0x88},
			[32]byte{0: 0x99, 31: 0xaa}},
		D: BanderwagonGroupElement{0: 0x12, 31: 0x34},
		IPAProof: IPAProof{
			CL:              [IPA_PROOF_DEPTH]BanderwagonGroupElement{0: {0: 0xee, 31: 0xff}, 7: {31: 0xbb}},
			CR:              [IPA_PROOF_DEPTH]BanderwagonGroupElement{0: {0: 0xcc, 31: 0xdd}, 7: {31: 0x11}},
			FinalEvaluation: BanderwagonFieldElement{31: 0x22},
		},
	}
	if err := vp.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		t.Fatal(err)
	}
	var newvp VerkleProof
	if err := newvp.Deserialize(codec.NewDecodingReader(&buf, uint64(buf.Len()))); err != nil {
		t.Fatalf("deserializing proof: %v", err)
	}
	if newvp.DepthExtensionPresent[2] != vp.DepthExtensionPresent[2] {
		t.Fatalf("could not deserialize depth + extension presence indicator: %d != %d", newvp.DepthExtensionPresent[2], vp.DepthExtensionPresent[2])
	}
	if !bytes.Equal(newvp.CommitmentsByPath[1][:], vp.CommitmentsByPath[1][:]) {
		t.Fatalf("differing second commitment %x != %x", newvp.CommitmentsByPath[1][:], vp.CommitmentsByPath[1][:])
	}
	if !bytes.Equal(newvp.IPAProof.CL[0][:], vp.IPAProof.CL[0][:]) {
		t.Fatalf("differing CL proof element %x != %x", newvp.IPAProof.CL[0][:], vp.IPAProof.CL[0][:])
	}
}
