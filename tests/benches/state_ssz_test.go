package benches

import (
	"bytes"
	"encoding/gob"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	"testing"
)

const stateValidatorFill = 100

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

func BenchmarkStateSerialize(b *testing.B) {
	full := CreateTestState(stateValidatorFill, MAX_EFFECTIVE_BALANCE)
	var buf bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		err := zssz.Encode(&buf, full.BeaconState, phase0.BeaconStateSSZ)
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
