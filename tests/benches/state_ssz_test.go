package benches

import (
	"bytes"
	//"crypto/sha256"
	"encoding/gob"
	"github.com/minio/sha256-simd"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	"testing"
)

const stateValidatorFill = 300000

func BenchmarkStateHash(b *testing.B) {
	full := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		root := ssz.HashTreeRoot(full.BeaconState, phase0.BeaconStateSSZ)
		res ^= root[0]
	}
	b.Logf("res: %d", res)
}

func BenchmarkFlatHash(b *testing.B) {
	full := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	h := sha256.New()
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		_, err := zssz.Encode(h, full.BeaconState, phase0.BeaconStateSSZ)
		if err != nil {
			b.Fatal(err)
		}
		res ^= h.Sum(nil)[0]
		h.Reset()
	}
	b.Logf("res: %d", res)
}

func BenchmarkStateNoEncodingFlatHash(b *testing.B) {
	full := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	var buf bytes.Buffer
	_, err := zssz.Encode(&buf, full.BeaconState, phase0.BeaconStateSSZ)
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
		res ^= h.Sum(nil)[0]
		h.Reset()
	}
	b.Logf("res: %d", res)
}

func BenchmarkStateFullFlatHash(b *testing.B) {
	full := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	var buf bytes.Buffer
	h := sha256.New()
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		_, err := zssz.Encode(&buf, full.BeaconState, phase0.BeaconStateSSZ)
		if err != nil {
			b.Fatal(err)
		}
		h.Write(buf.Bytes())
		res ^= h.Sum(nil)[0]
		h.Reset()
		buf.Reset()
	}
	b.Logf("res: %d", res)
}

func BenchmarkStateSerialize(b *testing.B) {
	full := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	var buf bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		_, err := zssz.Encode(&buf, full.BeaconState, phase0.BeaconStateSSZ)
		if err != nil {
			b.Fatal(err)
		}
		res ^= buf.Bytes()[0]
		buf.Reset()
	}
	b.Logf("res: %d", res)
}

func BenchmarkStateSerializeGob(b *testing.B) {
	full := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	var buf bytes.Buffer
	g := gob.NewEncoder(&buf)
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		err := g.Encode(full.BeaconState)
		if err != nil {
			b.Fatal(err)
		}
		res ^= buf.Bytes()[0]
		buf.Reset()
	}
	b.Logf("res: %d", res)
}
