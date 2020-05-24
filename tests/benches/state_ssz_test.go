package benches

import (
	"bytes"
	"encoding/gob"
	"github.com/minio/sha256-simd"
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	"github.com/protolambda/ztyp/tree"
	"testing"
)

const stateValidatorFill = 300000

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
		err := stateTree.Serialize(h)
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
	state, err := stateTree.Raw()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		root := ssz.HashTreeRoot(state, BeaconStateSSZ)
		res += root[0]
	}
	//b.Logf("res: %d, N: %d", res, b.N)
}

func BenchmarkRawStateFlatHash(b *testing.B) {
	stateTree, _ := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	state, err := stateTree.Raw()
	if err != nil {
		b.Fatal(err)
	}
	h := sha256.New()
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		_, err := zssz.Encode(h, state, BeaconStateSSZ)
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
	state, err := stateTree.Raw()
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	_, err = zssz.Encode(&buf, state, BeaconStateSSZ)
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
		err := stateTree.Serialize(&buf)
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
	state, err := stateTree.Raw()
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	h := sha256.New()
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		_, err := zssz.Encode(&buf, state, BeaconStateSSZ)
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
	state, err := stateTree.Raw()
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		_, err := zssz.Encode(&buf, state, BeaconStateSSZ)
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
		err := stateTree.Serialize(&buf)
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
	state, err := stateTree.Raw()
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
