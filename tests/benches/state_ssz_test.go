package benches

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/minio/sha256-simd"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
)

const stateValidatorFill = 30000

var spec = configs.Mainnet
var MAX_EFFECTIVE_BALANCE = spec.MAX_EFFECTIVE_BALANCE

func BenchmarkTreeStateHash(b *testing.B) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	hFn := tree.GetHashFn()
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		root := stateTree.HashTreeRoot(hFn)
		res += root[0]
	}
	//b.Logf("res: %d, N: %d", res, b.N)
}

func BenchmarkTreeStateFlatHash(b *testing.B) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	h := sha256.New()
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		err := stateTree.Serialize(codec.NewEncodingWriter(h))
		if err != nil {
			b.Fatal(err)
		}
		res += h.Sum(nil)[0]
		h.Reset()
	}
	//b.Logf("res: %d, N: %d", res, b.N)
}

func BenchmarkRawStateHash(b *testing.B) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	state, err := stateTree.Raw(spec)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	h := tree.GetHashFn()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		root := state.HashTreeRoot(spec, h)
		res += root[0]
	}
	//b.Logf("res: %d, N: %d", res, b.N)
}

func BenchmarkRawStateFlatHash(b *testing.B) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	state, err := stateTree.Raw(spec)
	if err != nil {
		b.Fatal(err)
	}
	h := sha256.New()
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		err := state.Serialize(spec, codec.NewEncodingWriter(h))
		if err != nil {
			b.Fatal(err)
		}
		res += h.Sum(nil)[0]
		h.Reset()
	}
	//b.Logf("res: %d, N: %d", res, b.N)
}

func BenchmarkStateNoEncodingFlatHash(b *testing.B) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	state, err := stateTree.Raw(spec)
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	err = state.Serialize(spec, codec.NewEncodingWriter(&buf))
	if err != nil {
		b.Fatal(err)
	}
	data := buf.Bytes()
	h := sha256.New()
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		h.Write(data)
		res += h.Sum(nil)[0]
		h.Reset()
	}
	//b.Logf("res: %d, N: %d", res, b.N)
}

func BenchmarkTreeStateBufferedFlatHash(b *testing.B) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	var buf bytes.Buffer
	h := sha256.New()
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		err := stateTree.Serialize(codec.NewEncodingWriter(&buf))
		if err != nil {
			b.Fatal(err)
		}
		h.Write(buf.Bytes())
		res += h.Sum(nil)[0]
		h.Reset()
		buf.Reset()
	}
	//b.Logf("res: %d, N: %d", res, b.N)
}

func BenchmarkRawStateBufferedFlatHash(b *testing.B) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	state, err := stateTree.Raw(spec)
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	h := sha256.New()
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		err := state.Serialize(spec, codec.NewEncodingWriter(&buf))
		if err != nil {
			b.Fatal(err)
		}
		h.Write(buf.Bytes())
		res += h.Sum(nil)[0]
		h.Reset()
		buf.Reset()
	}
	//b.Logf("res: %d, N: %d", res, b.N)
}

func BenchmarkRawStateSerialize(b *testing.B) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	state, err := stateTree.Raw(spec)
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		err := state.Serialize(spec, codec.NewEncodingWriter(&buf))
		if err != nil {
			b.Fatal(err)
		}
		res += buf.Bytes()[0]
		buf.Reset()
	}
	//b.Logf("res: %d, N: %d", res, b.N)
}

func BenchmarkTreeStateSerialize(b *testing.B) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	var buf bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		err := stateTree.Serialize(codec.NewEncodingWriter(&buf))
		if err != nil {
			b.Fatal(err)
		}
		res += buf.Bytes()[0]
		buf.Reset()
	}
	//b.Logf("res: %d, N: %d", res, b.N)
}

func BenchmarkRawStateSerializeGob(b *testing.B) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	state, err := stateTree.Raw(spec)
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	g := gob.NewEncoder(&buf)
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		err := g.Encode(state)
		if err != nil {
			b.Fatal(err)
		}
		res += buf.Bytes()[0]
		buf.Reset()
	}
	//b.Logf("res: %d, N: %d", res, b.N)
}

func TestRawStateSerialize(t *testing.T) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	state, err := stateTree.Raw(spec)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = state.Serialize(spec, codec.NewEncodingWriter(&buf))
	if err != nil {
		t.Fatal(err)
	}
}

func TestRawHashTreeRoot(t *testing.T) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	state, err := stateTree.Raw(spec)
	if err != nil {
		t.Fatal(err)
	}
	if root := state.HashTreeRoot(spec, tree.GetHashFn()); root == (tree.Root{}) {
		t.Fatal("unexpected zero root")
	}
}
